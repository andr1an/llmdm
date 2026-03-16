package mcpserver

import (
	"context"
	"fmt"
	"sort"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/andr1an/llmdm/internal/dm"
	"github.com/andr1an/llmdm/internal/session"
	"github.com/andr1an/llmdm/internal/types"
)

func (s *Server) handleStartSession(ctx context.Context, req *mcp.CallToolRequest, input StartSessionInput) (*mcp.CallToolResult, StartSessionOutput, error) {
	recentSessions := input.RecentSessions
	if recentSessions <= 0 {
		recentSessions = 3
	}

	s.logToolEntry("start_session", input.CampaignID, "session", input.Session)

	var brief types.SessionBrief
	err := s.withDB(ctx, input.CampaignID, func(dbCtx *DBContext) error {
		campaign, err := dbCtx.Store.GetCampaign(input.CampaignID)
		if err != nil {
			return wrapError("get campaign", err)
		}
		if campaign == nil {
			return fmt.Errorf("campaign not found: %s", input.CampaignID)
		}

		activeCharacters, err := dbCtx.Store.ListCharacters(input.CampaignID, "pc", "active")
		if err != nil {
			return wrapError("list active characters", err)
		}
		openHooks, err := dbCtx.Store.ListOpenHooks(input.CampaignID)
		if err != nil {
			return wrapError("list open hooks", err)
		}
		worldFlags, err := dbCtx.Store.GetWorldFlags(input.CampaignID)
		if err != nil {
			return wrapError("get world flags", err)
		}

		lastSession, err := dbCtx.Store.GetLastSessionBefore(input.CampaignID, input.Session)
		if err != nil {
			return wrapError("get last session", err)
		}
		recent, err := dbCtx.Store.ListRecentSessionsBefore(input.CampaignID, input.Session, recentSessions)
		if err != nil {
			return wrapError("list recent sessions", err)
		}

		lastSessionNumber := 0
		lastSessionSummary := ""
		lastCheckpointText := "No checkpoint recorded."
		if lastSession != nil {
			lastSessionNumber = lastSession.Session
			lastSessionSummary = lastSession.Summary
			latestCheckpoint, cpErr := dbCtx.Store.GetLatestCheckpoint(input.CampaignID, lastSession.Session)
			if cpErr != nil {
				return wrapError("get latest checkpoint", cpErr)
			}
			if latestCheckpoint != nil {
				lastCheckpointText = latestCheckpoint.Note
			}
		}

		briefText, err := session.RenderBrief(session.BriefData{
			Session:         input.Session,
			CampaignName:    campaign.Name,
			Characters:      activeCharacters,
			RecentSummaries: session.FormatRecentSummaries(recent),
			OpenHooks:       openHooks,
			WorldFlags:      worldFlags,
			LastCheckpoint:  lastCheckpointText,
		})
		if err != nil {
			return wrapError("render brief", err)
		}

		brief = types.SessionBrief{
			SessionBrief:       briefText,
			ActiveCharacters:   activeCharacters,
			OpenHooks:          openHooks,
			WorldFlags:         worldFlags,
			LastSessionNumber:  lastSessionNumber,
			LastSessionSummary: lastSessionSummary,
			DMSystemPrompt:     dm.BuildSystemPrompt(briefText, openHooks, worldFlags),
		}

		return nil
	})
	if err != nil {
		s.logToolError("start_session", input.CampaignID, err, "session", input.Session)
		return nil, StartSessionOutput{}, err
	}

	s.logToolExit("start_session", input.CampaignID, "session", input.Session, "active_characters", len(brief.ActiveCharacters), "open_hooks", len(brief.OpenHooks))
	return nil, StartSessionOutput{Brief: brief}, nil
}

func (s *Server) handleEndSession(ctx context.Context, req *mcp.CallToolRequest, input EndSessionInput) (*mcp.CallToolResult, EndSessionOutput, error) {
	if err := validateMaxLength("raw_events", input.RawEvents, maxSessionRawEventsLength); err != nil {
		return nil, EndSessionOutput{}, wrapError("validate raw_events", err)
	}
	if err := validateMaxLength("dm_notes", input.DMNotes, maxSessionDMNotesLength); err != nil {
		return nil, EndSessionOutput{}, wrapError("validate dm_notes", err)
	}

	s.logToolEntry("end_session", input.CampaignID, "session", input.Session)

	compressor := session.NewCompressor(s.cfg.AnthropicAPIKey)
	compressedSummary, err := compressor.Compress(ctx, input.RawEvents)
	if err != nil {
		s.logToolError("end_session", input.CampaignID, err, "session", input.Session)
		return nil, EndSessionOutput{}, wrapError("compress session", err)
	}

	err = s.withDB(ctx, input.CampaignID, func(dbCtx *DBContext) error {
		hooksOpened, err := dbCtx.Store.CountHooksOpenedInSession(input.CampaignID, input.Session)
		if err != nil {
			return wrapError("count hooks opened", err)
		}
		hooksResolved, err := dbCtx.Store.CountHooksResolvedSincePreviousSession(input.CampaignID, input.Session)
		if err != nil {
			return wrapError("count hooks resolved", err)
		}

		if err := dbCtx.Store.UpsertSession(&types.Session{
			CampaignID:    input.CampaignID,
			Session:       input.Session,
			Summary:       compressedSummary,
			DMNotes:       input.DMNotes,
			HooksOpened:   hooksOpened,
			HooksResolved: hooksResolved,
		}); err != nil {
			return wrapError("upsert session", err)
		}

		if err := dbCtx.Store.AdvanceCampaignSession(input.CampaignID, input.Session+1); err != nil {
			return wrapError("advance campaign session", err)
		}

		return nil
	})
	if err != nil {
		s.logToolError("end_session", input.CampaignID, err, "session", input.Session)
		return nil, EndSessionOutput{}, err
	}

	s.logToolExit("end_session", input.CampaignID, "session", input.Session, "summary_length", len(compressedSummary))
	return nil, EndSessionOutput{Summary: compressedSummary}, nil
}

func (s *Server) handleCheckpoint(ctx context.Context, req *mcp.CallToolRequest, input CheckpointInput) (*mcp.CallToolResult, CheckpointOutput, error) {
	if err := validateMaxLength("note", input.Note, maxCheckpointNoteLength); err != nil {
		return nil, CheckpointOutput{}, wrapError("validate note", err)
	}

	s.logToolEntry("checkpoint", input.CampaignID, "session", input.Session, "has_data", len(input.Data) > 0)

	var checkpointID string
	err := s.withDB(ctx, input.CampaignID, func(dbCtx *DBContext) error {
		checkpoint, err := dbCtx.Store.CreateCheckpoint(input.CampaignID, input.Session, input.Note, input.Data)
		if err != nil {
			return wrapError("create checkpoint", err)
		}

		checkpointID = checkpoint.ID
		return nil
	})
	if err != nil {
		s.logToolError("checkpoint", input.CampaignID, err, "session", input.Session)
		return nil, CheckpointOutput{}, err
	}

	s.logToolExit("checkpoint", input.CampaignID, "session", input.Session)
	return nil, CheckpointOutput{CheckpointID: checkpointID}, nil
}

func (s *Server) handleGetTurnHistory(ctx context.Context, req *mcp.CallToolRequest, input GetTurnHistoryInput) (*mcp.CallToolResult, GetTurnHistoryOutput, error) {
	limit := input.Limit
	if limit <= 0 {
		limit = 50
	}

	s.logToolEntry("get_turn_history", input.CampaignID, "session", input.Session, "limit", limit)

	var turns []types.Checkpoint
	err := s.withDB(ctx, input.CampaignID, func(dbCtx *DBContext) error {
		checkpoints, err := dbCtx.Store.ListCheckpoints(input.CampaignID, input.Session, limit)
		if err != nil {
			return wrapError("list checkpoints", err)
		}

		// Filter checkpoints that have turn data
		for _, cp := range checkpoints {
			if len(cp.Data) > 0 {
				turns = append(turns, cp)
			}
		}

		return nil
	})
	if err != nil {
		s.logToolError("get_turn_history", input.CampaignID, err, "session", input.Session)
		return nil, GetTurnHistoryOutput{}, err
	}

	s.logToolExit("get_turn_history", input.CampaignID, "session", input.Session, "turns", len(turns))
	return nil, GetTurnHistoryOutput{Turns: turns}, nil
}

func (s *Server) handleGetSessionBrief(ctx context.Context, req *mcp.CallToolRequest, input GetSessionBriefInput) (*mcp.CallToolResult, GetSessionBriefOutput, error) {
	s.logToolEntry("get_session_brief", input.CampaignID)

	var brief string
	err := s.withDB(ctx, input.CampaignID, func(dbCtx *DBContext) error {
		campaign, err := dbCtx.Store.GetCampaign(input.CampaignID)
		if err != nil {
			return wrapError("get campaign", err)
		}
		if campaign == nil {
			return fmt.Errorf("campaign not found: %s", input.CampaignID)
		}

		latestSession, err := dbCtx.Store.GetLatestSession(input.CampaignID)
		if err != nil {
			return wrapError("get latest session", err)
		}
		activeCharacters, err := dbCtx.Store.ListCharacters(input.CampaignID, "pc", "active")
		if err != nil {
			return wrapError("list active characters", err)
		}
		openHooks, err := dbCtx.Store.ListOpenHooks(input.CampaignID)
		if err != nil {
			return wrapError("list open hooks", err)
		}
		worldFlags, err := dbCtx.Store.GetWorldFlags(input.CampaignID)
		if err != nil {
			return wrapError("get world flags", err)
		}

		brief = session.RenderQuickBrief(campaign.Name, latestSession, activeCharacters, openHooks, worldFlags)
		return nil
	})
	if err != nil {
		s.logToolError("get_session_brief", input.CampaignID, err)
		return nil, GetSessionBriefOutput{}, err
	}

	s.logToolExit("get_session_brief", input.CampaignID)
	return nil, GetSessionBriefOutput{Brief: brief}, nil
}

func (s *Server) handleListSessions(ctx context.Context, req *mcp.CallToolRequest, input ListSessionsInput) (*mcp.CallToolResult, ListSessionsOutput, error) {
	s.logToolEntry("list_sessions", input.CampaignID)

	var sessions []types.SessionMeta
	err := s.withDB(ctx, input.CampaignID, func(dbCtx *DBContext) error {
		var err error
		sessions, err = dbCtx.Store.ListSessions(input.CampaignID)
		if err != nil {
			return wrapError("list sessions", err)
		}
		return nil
	})
	if err != nil {
		s.logToolError("list_sessions", input.CampaignID, err)
		return nil, ListSessionsOutput{}, err
	}

	s.logToolExit("list_sessions", input.CampaignID, "count", len(sessions))
	return nil, ListSessionsOutput{Sessions: sessions}, nil
}

func (s *Server) handleGetNPCRelationships(ctx context.Context, req *mcp.CallToolRequest, input GetNPCRelationshipsInput) (*mcp.CallToolResult, GetNPCRelationshipsOutput, error) {
	s.logToolEntry("get_npc_relationships", input.CampaignID, "npc_name", input.NPCName)

	var edges []types.RelationshipEdge
	err := s.withDB(ctx, input.CampaignID, func(dbCtx *DBContext) error {
		var err error
		edges, err = dbCtx.Store.QueryNPCRelationships(input.CampaignID, input.NPCName)
		if err != nil {
			return wrapError("query npc relationships", err)
		}
		return nil
	})
	if err != nil {
		s.logToolError("get_npc_relationships", input.CampaignID, err, "npc_name", input.NPCName)
		return nil, GetNPCRelationshipsOutput{}, err
	}

	s.logToolExit("get_npc_relationships", input.CampaignID, "npc_name", input.NPCName, "count", len(edges))
	return nil, GetNPCRelationshipsOutput{Relationships: edges}, nil
}

func (s *Server) handleExportSessionRecap(ctx context.Context, req *mcp.CallToolRequest, input ExportSessionRecapInput) (*mcp.CallToolResult, ExportSessionRecapOutput, error) {
	fromSession := 0
	if input.FromSession != nil {
		fromSession = int(*input.FromSession)
	}
	toSession := 0
	if input.ToSession != nil {
		toSession = int(*input.ToSession)
	}

	s.logToolEntry("export_session_recap", input.CampaignID, "from_session", fromSession, "to_session", toSession)

	var markdown string
	err := s.withDB(ctx, input.CampaignID, func(dbCtx *DBContext) error {
		campaign, err := dbCtx.Store.GetCampaign(input.CampaignID)
		if err != nil {
			return wrapError("get campaign", err)
		}
		if campaign == nil {
			return fmt.Errorf("campaign not found: %s", input.CampaignID)
		}

		sessions, err := dbCtx.Store.ListSessions(input.CampaignID)
		if err != nil {
			return wrapError("list sessions", err)
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

		hooks, err := dbCtx.Store.ListOpenHooks(input.CampaignID)
		if err != nil {
			return wrapError("list open hooks", err)
		}
		flags, err := dbCtx.Store.GetWorldFlags(input.CampaignID)
		if err != nil {
			return wrapError("get world flags", err)
		}

		markdown = session.RenderRecap(campaign.Name, filtered, hooks, flags)
		return nil
	})
	if err != nil {
		s.logToolError("export_session_recap", input.CampaignID, err)
		return nil, ExportSessionRecapOutput{}, err
	}

	s.logToolExit("export_session_recap", input.CampaignID, "markdown_length", len(markdown))
	return nil, ExportSessionRecapOutput{Markdown: markdown}, nil
}
