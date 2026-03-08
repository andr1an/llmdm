package mcpserver

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/andr1an/llmdm/internal/memory"
	"github.com/andr1an/llmdm/internal/types"
)

func (s *Server) handleCreateCampaign(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name, err := req.RequireString("name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if err := validateCampaignName(name); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	description := req.GetString("description", "")

	// Create a new campaign DB
	campaignID := generateCampaignID(name)
	database, err := s.openCampaignDB(campaignID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	defer database.Close()

	store := memory.NewStore(database.DB)
	campaign, err := store.CreateCampaignWithID(campaignID, name, description)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	result := map[string]interface{}{
		"campaign_id": campaignID,
		"db_path":     database.Path(),
		"campaign":    campaign,
	}
	jsonResult, _ := json.Marshal(result)
	return mcp.NewToolResultText(string(jsonResult)), nil
}

func (s *Server) handleSaveCharacter(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	campaignID, err := req.RequireString("campaign_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	name, err := req.RequireString("name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	charType, err := req.RequireString("type")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	hpCurrent, err := req.RequireFloat("hp_current")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	hpMax, err := req.RequireFloat("hp_max")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	backstory := req.GetString("backstory", "")
	if err := validateMaxLength("backstory", backstory, maxCharacterBackstoryLength); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	notes := req.GetString("notes", "")
	if err := validateMaxLength("notes", notes, maxCharacterNotesLength); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	char := &types.Character{
		CampaignID: campaignID,
		Name:       name,
		Type:       charType,
		Class:      req.GetString("class", ""),
		Race:       req.GetString("race", ""),
		Level:      int(req.GetFloat("level", 1)),
		HP: types.HP{
			Current: int(hpCurrent),
			Max:     int(hpMax),
		},
		Stats: types.Stats{
			STR: int(req.GetFloat("str", 10)),
			DEX: int(req.GetFloat("dex", 10)),
			CON: int(req.GetFloat("con", 10)),
			INT: int(req.GetFloat("int_stat", 10)),
			WIS: int(req.GetFloat("wis", 10)),
			CHA: int(req.GetFloat("cha", 10)),
		},
		Backstory:     backstory,
		Status:        req.GetString("status", "active"),
		Notes:         notes,
		Inventory:     []string{},
		Conditions:    []string{},
		PlotFlags:     []string{},
		Relationships: map[string]string{},
	}

	database, err := s.openCampaignDB(campaignID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	defer database.Close()

	store := memory.NewStore(database.DB)
	if err := store.SaveCharacter(char); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	result := map[string]interface{}{
		"success":      true,
		"character_id": char.ID,
	}
	jsonResult, _ := json.Marshal(result)
	return mcp.NewToolResultText(string(jsonResult)), nil
}

func (s *Server) handleUpdateCharacter(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	campaignID, err := req.RequireString("campaign_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	name, err := req.RequireString("name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	update := memory.CharacterUpdate{}
	args := req.GetArguments()

	// Check for optional fields
	if hp, ok := args["hp_current"]; ok {
		if hpFloat, ok := hp.(float64); ok {
			hpInt := int(hpFloat)
			update.HPCurrent = &hpInt
		}
	}
	if level, ok := args["level"]; ok {
		if levelFloat, ok := level.(float64); ok {
			levelInt := int(levelFloat)
			update.Level = &levelInt
		}
	}
	if status, ok := args["status"]; ok {
		if statusStr, ok := status.(string); ok {
			update.Status = &statusStr
		}
	}
	if notes, ok := args["notes"]; ok {
		if notesStr, ok := notes.(string); ok {
			if err := validateMaxLength("notes", notesStr, maxCharacterNotesLength); err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			update.Notes = &notesStr
		}
	}

	database, err := s.openCampaignDB(campaignID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	defer database.Close()

	store := memory.NewStore(database.DB)
	updatedFields, err := store.UpdateCharacter(campaignID, name, update)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	result := map[string]interface{}{
		"success":        true,
		"updated_fields": updatedFields,
	}
	jsonResult, _ := json.Marshal(result)
	return mcp.NewToolResultText(string(jsonResult)), nil
}

func (s *Server) handleGetCharacter(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	campaignID, err := req.RequireString("campaign_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	name, err := req.RequireString("name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	database, err := s.openCampaignDB(campaignID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	defer database.Close()

	store := memory.NewStore(database.DB)
	char, err := store.GetCharacter(campaignID, name)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if char == nil {
		return mcp.NewToolResultError(fmt.Sprintf("character not found: %s", name)), nil
	}

	jsonResult, _ := json.Marshal(char)
	return mcp.NewToolResultText(string(jsonResult)), nil
}

func (s *Server) handleListCharacters(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	campaignID, err := req.RequireString("campaign_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	charType := req.GetString("type", "")
	status := req.GetString("status", "")

	database, err := s.openCampaignDB(campaignID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	defer database.Close()

	store := memory.NewStore(database.DB)
	characters, err := store.ListCharacters(campaignID, charType, status)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	jsonResult, _ := json.Marshal(characters)
	return mcp.NewToolResultText(string(jsonResult)), nil
}

func (s *Server) handleSavePlotEvent(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	campaignID, err := req.RequireString("campaign_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	session, err := req.RequireFloat("session")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	summary, err := req.RequireString("summary")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	event := &types.PlotEvent{
		CampaignID:   campaignID,
		Session:      int(session),
		Summary:      summary,
		Consequences: req.GetString("consequences", ""),
		NPCs:         []string{},
		PCs:          []string{},
		Tags:         []string{},
	}

	database, err := s.openCampaignDB(campaignID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	defer database.Close()

	store := memory.NewStore(database.DB)
	if err := store.SavePlotEvent(event, nil); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	result := map[string]interface{}{
		"success":  true,
		"event_id": event.ID,
	}
	jsonResult, _ := json.Marshal(result)
	return mcp.NewToolResultText(string(jsonResult)), nil
}

func (s *Server) handleListOpenHooks(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
	hooks, err := store.ListOpenHooks(campaignID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	jsonResult, _ := json.Marshal(hooks)
	return mcp.NewToolResultText(string(jsonResult)), nil
}

func (s *Server) handleResolveHook(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	campaignID, err := req.RequireString("campaign_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	hookID, err := req.RequireString("hook_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	resolution, err := req.RequireString("resolution")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	database, err := s.openCampaignDB(campaignID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	defer database.Close()

	store := memory.NewStore(database.DB)
	if err := store.ResolveHook(campaignID, hookID, resolution); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	result := map[string]interface{}{
		"success": true,
	}
	jsonResult, _ := json.Marshal(result)
	return mcp.NewToolResultText(string(jsonResult)), nil
}

func (s *Server) handleSetWorldFlag(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	campaignID, err := req.RequireString("campaign_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	key, err := req.RequireString("key")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	value, err := req.RequireString("value")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	database, err := s.openCampaignDB(campaignID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	defer database.Close()

	store := memory.NewStore(database.DB)
	if err := store.SetWorldFlag(campaignID, key, value); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	result := map[string]interface{}{
		"success": true,
	}
	jsonResult, _ := json.Marshal(result)
	return mcp.NewToolResultText(string(jsonResult)), nil
}

func (s *Server) handleGetWorldFlags(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
	flags, err := store.GetWorldFlags(campaignID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	jsonResult, _ := json.Marshal(flags)
	return mcp.NewToolResultText(string(jsonResult)), nil
}
