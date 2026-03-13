package mcpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/andr1an/llmdm/internal/db"
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
	s.log().Debug("creating campaign", "campaign_id", campaignID, "name", name)

	database, err := s.openCampaignDB(campaignID)
	if err != nil {
		s.log().Error("failed to open campaign database for create_campaign", "campaign_id", campaignID, "error", err)
		return mcp.NewToolResultError(err.Error()), nil
	}
	defer database.Close()

	store := memory.NewStore(database.DB)
	campaign, err := store.CreateCampaignWithID(campaignID, name, description)
	if err != nil {
		s.log().Error("failed to create campaign", "campaign_id", campaignID, "error", err)
		return mcp.NewToolResultError(err.Error()), nil
	}

	s.log().Info("campaign created successfully", "campaign_id", campaignID, "name", name, "db_path", database.Path())

	result := map[string]interface{}{
		"campaign_id": campaignID,
		"db_path":     database.Path(),
		"campaign":    campaign,
	}
	jsonResult, _ := json.Marshal(result)
	return mcp.NewToolResultText(string(jsonResult)), nil
}

func (s *Server) handleListCampaigns(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.log().Debug("listing campaigns", "db_path", s.DBPath())

	// Read all .db files from the database directory
	entries, err := os.ReadDir(s.DBPath())
	if err != nil {
		if os.IsNotExist(err) {
			// No campaigns directory yet, return empty list
			s.log().Debug("campaigns directory does not exist, returning empty list")
			jsonResult, _ := json.Marshal([]types.Campaign{})
			return mcp.NewToolResultText(string(jsonResult)), nil
		}
		s.log().Error("failed to read campaigns directory", "db_path", s.DBPath(), "error", err)
		return mcp.NewToolResultError(fmt.Sprintf("failed to read campaigns directory: %v", err)), nil
	}

	var campaigns []types.Campaign
	skipped := 0
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".db") {
			continue
		}

		campaignID := strings.TrimSuffix(entry.Name(), ".db")
		if !db.IsValidCampaignID(campaignID) {
			skipped++
			continue
		}

		dbPath := filepath.Join(s.DBPath(), entry.Name())
		database, err := db.Open(dbPath)
		if err != nil {
			s.log().Warn("failed to open campaign database during list", "campaign_id", campaignID, "error", err)
			skipped++
			continue // Skip databases that can't be opened
		}

		store := memory.NewStore(database.DB)
		campaign, err := store.GetCampaign(campaignID)
		database.Close()

		if err != nil || campaign == nil {
			s.log().Warn("failed to load campaign metadata during list", "campaign_id", campaignID, "error", err)
			skipped++
			continue
		}

		campaigns = append(campaigns, *campaign)
	}

	if campaigns == nil {
		campaigns = []types.Campaign{}
	}

	s.log().Info("campaigns listed", "count", len(campaigns), "skipped", skipped)

	jsonResult, _ := json.Marshal(campaigns)
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
		Gold:          int(req.GetFloat("gold", 0)),
		Backstory:     backstory,
		Status:        req.GetString("status", "active"),
		Notes:         notes,
		Inventory:     []string{},
		Conditions:    []string{},
		PlotFlags:     []string{},
		Relationships: map[string]string{},
	}

	s.log().Debug("saving character", "campaign_id", campaignID, "name", name, "type", charType, "class", char.Class)

	database, err := s.openCampaignDB(campaignID)
	if err != nil {
		s.log().Error("failed to open campaign database for save_character", "campaign_id", campaignID, "error", err)
		return mcp.NewToolResultError(err.Error()), nil
	}
	defer database.Close()

	store := memory.NewStore(database.DB)
	if err := store.SaveCharacter(char); err != nil {
		s.log().Error("failed to save character", "campaign_id", campaignID, "name", name, "error", err)
		return mcp.NewToolResultError(err.Error()), nil
	}

	s.log().Info("character saved successfully", "campaign_id", campaignID, "name", name, "type", charType, "character_id", char.ID)

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
	if gold, ok := args["gold"]; ok {
		if goldFloat, ok := gold.(float64); ok {
			goldInt := int(goldFloat)
			update.Gold = &goldInt
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
	// Parse inventory array
	if inv, ok := args["inventory"]; ok {
		if invSlice, ok := inv.([]interface{}); ok {
			inventory := make([]string, 0, len(invSlice))
			for _, item := range invSlice {
				if itemStr, ok := item.(string); ok {
					inventory = append(inventory, itemStr)
				}
			}
			update.Inventory = inventory
		}
	}
	// Parse conditions array
	if cond, ok := args["conditions"]; ok {
		if condSlice, ok := cond.([]interface{}); ok {
			conditions := make([]string, 0, len(condSlice))
			for _, c := range condSlice {
				if condStr, ok := c.(string); ok {
					conditions = append(conditions, condStr)
				}
			}
			update.Conditions = conditions
		}
	}
	// Parse plot_flags array
	if pf, ok := args["plot_flags"]; ok {
		if pfSlice, ok := pf.([]interface{}); ok {
			plotFlags := make([]string, 0, len(pfSlice))
			for _, f := range pfSlice {
				if flagStr, ok := f.(string); ok {
					plotFlags = append(plotFlags, flagStr)
				}
			}
			update.PlotFlags = plotFlags
		}
	}
	// Parse relationships map
	if rel, ok := args["relationships"]; ok {
		if relMap, ok := rel.(map[string]interface{}); ok {
			relationships := make(map[string]string, len(relMap))
			for k, v := range relMap {
				if vStr, ok := v.(string); ok {
					relationships[k] = vStr
				}
			}
			update.Relationships = relationships
		}
	}

	s.log().Debug("updating character", "campaign_id", campaignID, "name", name)

	database, err := s.openCampaignDB(campaignID)
	if err != nil {
		s.log().Error("failed to open campaign database for update_character", "campaign_id", campaignID, "error", err)
		return mcp.NewToolResultError(err.Error()), nil
	}
	defer database.Close()

	store := memory.NewStore(database.DB)
	updatedFields, err := store.UpdateCharacter(campaignID, name, update)
	if err != nil {
		s.log().Error("failed to update character", "campaign_id", campaignID, "name", name, "error", err)
		return mcp.NewToolResultError(err.Error()), nil
	}

	s.log().Info("character updated successfully", "campaign_id", campaignID, "name", name, "updated_fields", updatedFields)

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

	// Parse optional hooks array
	var hooks []string
	if hooksArg, ok := req.GetArguments()["hooks"]; ok {
		if hooksSlice, ok := hooksArg.([]interface{}); ok {
			for _, h := range hooksSlice {
				if hookStr, ok := h.(string); ok {
					hooks = append(hooks, hookStr)
				}
			}
		}
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

	s.log().Debug("saving plot event", "campaign_id", campaignID, "session", int(session), "hooks_count", len(hooks))

	database, err := s.openCampaignDB(campaignID)
	if err != nil {
		s.log().Error("failed to open campaign database for save_plot_event", "campaign_id", campaignID, "error", err)
		return mcp.NewToolResultError(err.Error()), nil
	}
	defer database.Close()

	store := memory.NewStore(database.DB)
	if err := store.SavePlotEvent(event, hooks); err != nil {
		s.log().Error("failed to save plot event", "campaign_id", campaignID, "session", int(session), "error", err)
		return mcp.NewToolResultError(err.Error()), nil
	}

	s.log().Info("plot event saved successfully", "campaign_id", campaignID, "session", int(session), "event_id", event.ID, "hooks_opened", len(hooks))

	result := map[string]interface{}{
		"success":      true,
		"event_id":     event.ID,
		"hooks_opened": len(hooks),
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

	s.log().Debug("resolving hook", "campaign_id", campaignID, "hook_id", hookID)

	database, err := s.openCampaignDB(campaignID)
	if err != nil {
		s.log().Error("failed to open campaign database for resolve_hook", "campaign_id", campaignID, "error", err)
		return mcp.NewToolResultError(err.Error()), nil
	}
	defer database.Close()

	store := memory.NewStore(database.DB)
	if err := store.ResolveHook(campaignID, hookID, resolution); err != nil {
		s.log().Error("failed to resolve hook", "campaign_id", campaignID, "hook_id", hookID, "error", err)
		return mcp.NewToolResultError(err.Error()), nil
	}

	s.log().Info("hook resolved successfully", "campaign_id", campaignID, "hook_id", hookID)

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
