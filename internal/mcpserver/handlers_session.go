package mcpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/andr1an/llmdm/internal/dm"
	"github.com/andr1an/llmdm/internal/memory"
	"github.com/andr1an/llmdm/internal/session"
	"github.com/andr1an/llmdm/internal/types"
)

func (s *Server) handleStartSession(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	campaignID, err := req.RequireString("campaign_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	sessionNumberRaw, err := req.RequireFloat("session")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	sessionNumber := int(sessionNumberRaw)
	recentSessions := int(req.GetFloat("recent_sessions", 3))
	if recentSessions <= 0 {
		recentSessions = 3
	}

	s.log().Debug("starting session", "campaign_id", campaignID, "session", sessionNumber, "recent_sessions", recentSessions)

	database, err := s.openCampaignDB(campaignID)
	if err != nil {
		s.log().Error("failed to open campaign database for start_session", "campaign_id", campaignID, "error", err)
		return mcp.NewToolResultError(err.Error()), nil
	}
	defer database.Close()

	store := memory.NewStore(database.DB)
	campaign, err := store.GetCampaign(campaignID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if campaign == nil {
		return mcp.NewToolResultError(fmt.Sprintf("campaign not found: %s", campaignID)), nil
	}

	activeCharacters, err := store.ListCharacters(campaignID, "pc", "active")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	openHooks, err := store.ListOpenHooks(campaignID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	worldFlags, err := store.GetWorldFlags(campaignID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	lastSession, err := store.GetLastSessionBefore(campaignID, sessionNumber)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	recent, err := store.ListRecentSessionsBefore(campaignID, sessionNumber, recentSessions)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	lastSessionNumber := 0
	lastSessionSummary := ""
	lastCheckpointText := "No checkpoint recorded."
	if lastSession != nil {
		lastSessionNumber = lastSession.Session
		lastSessionSummary = lastSession.Summary
		latestCheckpoint, cpErr := store.GetLatestCheckpoint(campaignID, lastSession.Session)
		if cpErr != nil {
			return mcp.NewToolResultError(cpErr.Error()), nil
		}
		if latestCheckpoint != nil {
			lastCheckpointText = latestCheckpoint.Note
		}
	}

	briefText, err := session.RenderBrief(session.BriefData{
		Session:         sessionNumber,
		CampaignName:    campaign.Name,
		Characters:      activeCharacters,
		RecentSummaries: session.FormatRecentSummaries(recent),
		OpenHooks:       openHooks,
		WorldFlags:      worldFlags,
		LastCheckpoint:  lastCheckpointText,
	})
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	result := types.SessionBrief{
		SessionBrief:       briefText,
		ActiveCharacters:   activeCharacters,
		OpenHooks:          openHooks,
		WorldFlags:         worldFlags,
		LastSessionNumber:  lastSessionNumber,
		LastSessionSummary: lastSessionSummary,
		DMSystemPrompt:     dm.BuildSystemPrompt(briefText, openHooks, worldFlags),
	}

	s.log().Info("session started successfully",
		"campaign_id", campaignID,
		"session", sessionNumber,
		"active_characters", len(activeCharacters),
		"open_hooks", len(openHooks),
		"world_flags", len(worldFlags))

	jsonResult, _ := json.Marshal(result)
	return mcp.NewToolResultText(string(jsonResult)), nil
}

func (s *Server) handleEndSession(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	campaignID, err := req.RequireString("campaign_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	sessionNumberRaw, err := req.RequireFloat("session")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	rawEvents, err := req.RequireString("raw_events")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if err := validateMaxLength("raw_events", rawEvents, maxSessionRawEventsLength); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	dmNotes := req.GetString("dm_notes", "")
	if err := validateMaxLength("dm_notes", dmNotes, maxSessionDMNotesLength); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	sessionNumber := int(sessionNumberRaw)

	s.log().Debug("ending session", "campaign_id", campaignID, "session", sessionNumber, "raw_events_length", len(rawEvents))

	database, err := s.openCampaignDB(campaignID)
	if err != nil {
		s.log().Error("failed to open campaign database for end_session", "campaign_id", campaignID, "error", err)
		return mcp.NewToolResultError(err.Error()), nil
	}
	defer database.Close()

	store := memory.NewStore(database.DB)
	compressor := session.NewCompressor(s.cfg.AnthropicAPIKey)
	compressedSummary, err := compressor.Compress(ctx, rawEvents)
	if err != nil {
		s.log().Error("session compression failed", "campaign_id", campaignID, "session", sessionNumber, "error", err)
		return mcp.NewToolResultError(err.Error()), nil
	}

	hooksOpened, err := store.CountHooksOpenedInSession(campaignID, sessionNumber)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	hooksResolved, err := store.CountHooksResolvedSincePreviousSession(campaignID, sessionNumber)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if err := store.UpsertSession(&types.Session{
		CampaignID:    campaignID,
		Session:       sessionNumber,
		Summary:       compressedSummary,
		DMNotes:       dmNotes,
		HooksOpened:   hooksOpened,
		HooksResolved: hooksResolved,
	}); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if err := store.AdvanceCampaignSession(campaignID, sessionNumber+1); err != nil {
		s.log().Error("failed to advance campaign session", "campaign_id", campaignID, "session", sessionNumber, "error", err)
		return mcp.NewToolResultError(err.Error()), nil
	}

	s.log().Info("session ended successfully",
		"campaign_id", campaignID,
		"session", sessionNumber,
		"summary_length", len(compressedSummary),
		"hooks_opened", hooksOpened,
		"hooks_resolved", hooksResolved)

	result := map[string]interface{}{
		"success":            true,
		"compressed_summary": compressedSummary,
		"hooks_opened":       hooksOpened,
		"hooks_resolved":     hooksResolved,
		"characters_updated": []string{},
	}
	jsonResult, _ := json.Marshal(result)
	return mcp.NewToolResultText(string(jsonResult)), nil
}

func (s *Server) handleCheckpoint(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	campaignID, err := req.RequireString("campaign_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	sessionNumberRaw, err := req.RequireFloat("session")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	note, err := req.RequireString("note")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if err := validateMaxLength("note", note, maxCheckpointNoteLength); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Get optional data parameter
	args := req.GetArguments()
	var data map[string]interface{}
	if dataRaw, ok := args["data"]; ok {
		if dataMap, ok := dataRaw.(map[string]interface{}); ok {
			data = dataMap
		}
	}

	s.log().Debug("creating checkpoint", "campaign_id", campaignID, "session", int(sessionNumberRaw), "note_length", len(note), "has_data", len(data) > 0)

	database, err := s.openCampaignDB(campaignID)
	if err != nil {
		s.log().Error("failed to open campaign database for checkpoint", "campaign_id", campaignID, "error", err)
		return mcp.NewToolResultError(err.Error()), nil
	}
	defer database.Close()

	store := memory.NewStore(database.DB)
	checkpoint, err := store.CreateCheckpoint(campaignID, int(sessionNumberRaw), note, data)
	if err != nil {
		s.log().Error("failed to create checkpoint", "campaign_id", campaignID, "session", int(sessionNumberRaw), "error", err)
		return mcp.NewToolResultError(err.Error()), nil
	}

	s.log().Info("checkpoint created", "campaign_id", campaignID, "session", int(sessionNumberRaw), "checkpoint_id", checkpoint.ID)

	result := map[string]interface{}{
		"success":       true,
		"checkpoint_id": checkpoint.ID,
	}
	jsonResult, _ := json.Marshal(result)
	return mcp.NewToolResultText(string(jsonResult)), nil
}

func (s *Server) handleGetTurnHistory(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	campaignID, err := req.RequireString("campaign_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	sessionNumberRaw, err := req.RequireFloat("session")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	limit := int(req.GetFloat("limit", 50))
	if limit <= 0 {
		limit = 50
	}

	s.log().Debug("getting turn history", "campaign_id", campaignID, "session", int(sessionNumberRaw), "limit", limit)

	database, err := s.openCampaignDB(campaignID)
	if err != nil {
		s.log().Error("failed to open campaign database for get_turn_history", "campaign_id", campaignID, "error", err)
		return mcp.NewToolResultError(err.Error()), nil
	}
	defer database.Close()

	store := memory.NewStore(database.DB)
	checkpoints, err := store.ListCheckpoints(campaignID, int(sessionNumberRaw), limit)
	if err != nil {
		s.log().Error("failed to list checkpoints", "campaign_id", campaignID, "session", int(sessionNumberRaw), "error", err)
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Extract turn data from checkpoints
	turns := make([]map[string]interface{}, 0, len(checkpoints))
	for _, cp := range checkpoints {
		if len(cp.Data) > 0 {
			turns = append(turns, cp.Data)
		}
	}

	s.log().Info("turn history retrieved", "campaign_id", campaignID, "session", int(sessionNumberRaw), "turns", len(turns))

	result := map[string]interface{}{
		"campaign_id": campaignID,
		"session":     int(sessionNumberRaw),
		"turns":       turns,
		"count":       len(turns),
	}
	jsonResult, _ := json.Marshal(result)
	return mcp.NewToolResultText(string(jsonResult)), nil
}

func (s *Server) handleGetSessionBrief(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	campaignID, err := req.RequireString("campaign_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	database, err := s.openCampaignDB(campaignID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	defer database.Close()

	store := memory.NewStore(database.DB)
	campaign, err := store.GetCampaign(campaignID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if campaign == nil {
		return mcp.NewToolResultError(fmt.Sprintf("campaign not found: %s", campaignID)), nil
	}

	latestSession, err := store.GetLatestSession(campaignID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	activeCharacters, err := store.ListCharacters(campaignID, "pc", "active")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	openHooks, err := store.ListOpenHooks(campaignID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	worldFlags, err := store.GetWorldFlags(campaignID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	result := map[string]string{
		"brief": session.RenderQuickBrief(campaign.Name, latestSession, activeCharacters, openHooks, worldFlags),
	}
	jsonResult, _ := json.Marshal(result)
	return mcp.NewToolResultText(string(jsonResult)), nil
}

func (s *Server) handleListSessions(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	campaignID, err := req.RequireString("campaign_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	database, err := s.openCampaignDB(campaignID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	defer database.Close()

	store := memory.NewStore(database.DB)
	sessions, err := store.ListSessions(campaignID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	jsonResult, _ := json.Marshal(sessions)
	return mcp.NewToolResultText(string(jsonResult)), nil
}

func (s *Server) handleGetNPCRelationships(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	campaignID, err := req.RequireString("campaign_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	npcName := req.GetString("npc_name", "")

	database, err := s.openCampaignDB(campaignID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	defer database.Close()

	store := memory.NewStore(database.DB)
	edges, err := store.QueryNPCRelationships(campaignID, npcName)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	jsonResult, _ := json.Marshal(map[string]interface{}{
		"campaign_id": campaignID,
		"npc_name":    npcName,
		"count":       len(edges),
		"edges":       edges,
	})
	return mcp.NewToolResultText(string(jsonResult)), nil
}

func (s *Server) handleExportSessionRecap(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	campaignID, err := req.RequireString("campaign_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	fromSession := int(req.GetFloat("from_session", 0))
	toSession := int(req.GetFloat("to_session", 0))

	database, err := s.openCampaignDB(campaignID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	defer database.Close()

	store := memory.NewStore(database.DB)
	campaign, err := store.GetCampaign(campaignID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if campaign == nil {
		return mcp.NewToolResultError(fmt.Sprintf("campaign not found: %s", campaignID)), nil
	}

	sessions, err := store.ListSessions(campaignID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	filtered := make([]types.SessionMeta, 0, len(sessions))
	for _, s := range sessions {
		if fromSession > 0 && s.Session < fromSession {
			continue
		}
		if toSession > 0 && s.Session > toSession {
			continue
		}
		filtered = append(filtered, s)
	}
	sort.Slice(filtered, func(i, j int) bool { return filtered[i].Session < filtered[j].Session })

	hooks, err := store.ListOpenHooks(campaignID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	flags, err := store.GetWorldFlags(campaignID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	markdown := session.RenderRecap(campaign.Name, filtered, hooks, flags)
	result := map[string]interface{}{
		"campaign_id":   campaignID,
		"from_session":  fromSession,
		"to_session":    toSession,
		"session_count": len(filtered),
		"markdown":      markdown,
	}
	jsonResult, _ := json.Marshal(result)
	return mcp.NewToolResultText(string(jsonResult)), nil
}
