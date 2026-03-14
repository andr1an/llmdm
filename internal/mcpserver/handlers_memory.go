package mcpserver

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/andr1an/llmdm/internal/db"
	"github.com/andr1an/llmdm/internal/memory"
	"github.com/andr1an/llmdm/internal/types"
)

func (s *Server) handleCreateCampaign(ctx context.Context, req *mcp.CallToolRequest, input CreateCampaignInput) (*mcp.CallToolResult, CreateCampaignOutput, error) {
	if err := validateCampaignName(input.Name); err != nil {
		return nil, CreateCampaignOutput{}, err
	}

	campaignID := generateCampaignID(input.Name)
	s.log().Debug("creating campaign", "campaign_id", campaignID, "name", input.Name)

	var campaign *types.Campaign
	var dbPath string
	err := s.withDB(ctx, campaignID, func(dbCtx *DBContext) error {
		var err error
		campaign, err = dbCtx.Store.CreateCampaignWithID(campaignID, input.Name, input.Description)
		if err != nil {
			return err
		}
		// Get db path (we need to access the underlying database struct)
		return nil
	})
	if err != nil {
		s.log().Error("failed to create campaign", "campaign_id", campaignID, "error", err)
		return nil, CreateCampaignOutput{}, wrapError("create campaign", err)
	}

	// Get db path by reconstructing it
	dbPath, _ = db.CampaignDBPath(s.dbPath, campaignID)
	s.log().Info("campaign created successfully", "campaign_id", campaignID, "name", input.Name, "db_path", dbPath)

	return nil, CreateCampaignOutput{
		CampaignID: campaignID,
		DBPath:     dbPath,
		Campaign:   *campaign,
	}, nil
}

func (s *Server) handleListCampaigns(ctx context.Context, req *mcp.CallToolRequest, input ListCampaignsInput) (*mcp.CallToolResult, ListCampaignsOutput, error) {
	s.log().Debug("listing campaigns", "db_path", s.DBPath())

	entries, err := os.ReadDir(s.DBPath())
	if err != nil {
		if os.IsNotExist(err) {
			s.log().Debug("campaigns directory does not exist, returning empty list")
			return nil, ListCampaignsOutput{Campaigns: []types.Campaign{}}, nil
		}
		s.log().Error("failed to read campaigns directory", "db_path", s.DBPath(), "error", err)
		return nil, ListCampaignsOutput{}, wrapError("read campaigns directory", err)
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
			continue
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
	return nil, ListCampaignsOutput{Campaigns: campaigns}, nil
}

func (s *Server) handleSaveCharacter(ctx context.Context, req *mcp.CallToolRequest, input SaveCharacterInput) (*mcp.CallToolResult, SaveCharacterOutput, error) {
	if err := validateMaxLength("backstory", input.Backstory, maxCharacterBackstoryLength); err != nil {
		return nil, SaveCharacterOutput{}, err
	}
	if err := validateMaxLength("notes", input.Notes, maxCharacterNotesLength); err != nil {
		return nil, SaveCharacterOutput{}, err
	}

	level := input.Level
	if level == 0 {
		level = 1
	}
	status := input.Status
	if status == "" {
		status = "active"
	}

	char := &types.Character{
		CampaignID: input.CampaignID,
		Name:       input.Name,
		Type:       input.Type,
		Class:      input.Class,
		Race:       input.Race,
		Level:      level,
		HP: types.HP{
			Current: int(input.HPCurrent),
			Max:     int(input.HPMax),
		},
		Stats: types.Stats{
			STR: input.STR,
			DEX: input.DEX,
			CON: input.CON,
			INT: input.INTStat,
			WIS: input.WIS,
			CHA: input.CHA,
		},
		Gold:          input.Gold,
		Backstory:     input.Backstory,
		Status:        status,
		Notes:         input.Notes,
		Inventory:     []string{},
		Conditions:    []string{},
		PlotFlags:     []string{},
		Relationships: map[string]string{},
	}

	s.log().Debug("saving character", "campaign_id", input.CampaignID, "name", input.Name, "type", input.Type, "class", char.Class)

	err := s.withDB(ctx, input.CampaignID, func(dbCtx *DBContext) error {
		return dbCtx.Store.SaveCharacter(char)
	})
	if err != nil {
		s.log().Error("failed to save character", "campaign_id", input.CampaignID, "name", input.Name, "error", err)
		return nil, SaveCharacterOutput{}, wrapError("save character", err)
	}

	s.log().Info("character saved successfully", "campaign_id", input.CampaignID, "name", input.Name, "type", input.Type, "character_id", char.ID)
	return nil, SaveCharacterOutput{Character: *char}, nil
}

func (s *Server) handleUpdateCharacter(ctx context.Context, req *mcp.CallToolRequest, input UpdateCharacterInput) (*mcp.CallToolResult, UpdateCharacterOutput, error) {
	if input.Notes != "" {
		if err := validateMaxLength("notes", input.Notes, maxCharacterNotesLength); err != nil {
			return nil, UpdateCharacterOutput{}, err
		}
	}

	update := memory.CharacterUpdate{}

	// Convert pointer fields from input to update struct
	if input.HPCurrent != nil {
		hpInt := int(*input.HPCurrent)
		update.HPCurrent = &hpInt
	}
	if input.Level != nil {
		update.Level = input.Level
	}
	if input.Gold != nil {
		update.Gold = input.Gold
	}
	if input.Status != "" {
		update.Status = &input.Status
	}
	if input.Notes != "" {
		update.Notes = &input.Notes
	}
	if input.Inventory != nil {
		update.Inventory = input.Inventory
	}
	if input.Conditions != nil {
		update.Conditions = input.Conditions
	}
	if input.PlotFlags != nil {
		update.PlotFlags = input.PlotFlags
	}
	if input.Relationships != nil {
		// Convert map[string]interface{} to map[string]string
		relationships := make(map[string]string, len(input.Relationships))
		for k, v := range input.Relationships {
			if vStr, ok := v.(string); ok {
				relationships[k] = vStr
			}
		}
		update.Relationships = relationships
	}

	s.log().Debug("updating character", "campaign_id", input.CampaignID, "name", input.Name)

	var character *types.Character
	err := s.withDB(ctx, input.CampaignID, func(dbCtx *DBContext) error {
		updatedFields, err := dbCtx.Store.UpdateCharacter(input.CampaignID, input.Name, update)
		if err != nil {
			return err
		}
		s.log().Info("character updated successfully", "campaign_id", input.CampaignID, "name", input.Name, "updated_fields", updatedFields)

		// Fetch the updated character
		character, err = dbCtx.Store.GetCharacter(input.CampaignID, input.Name)
		return err
	})
	if err != nil {
		s.log().Error("failed to update character", "campaign_id", input.CampaignID, "name", input.Name, "error", err)
		return nil, UpdateCharacterOutput{}, wrapError("update character", err)
	}

	if character == nil {
		return nil, UpdateCharacterOutput{}, fmt.Errorf("character not found after update: %s", input.Name)
	}

	return nil, UpdateCharacterOutput{Character: *character}, nil
}

func (s *Server) handleGetCharacter(ctx context.Context, req *mcp.CallToolRequest, input GetCharacterInput) (*mcp.CallToolResult, GetCharacterOutput, error) {
	var char *types.Character
	err := s.withDB(ctx, input.CampaignID, func(dbCtx *DBContext) error {
		var err error
		char, err = dbCtx.Store.GetCharacter(input.CampaignID, input.Name)
		return err
	})
	if err != nil {
		return nil, GetCharacterOutput{}, wrapError("get character", err)
	}
	if char == nil {
		return nil, GetCharacterOutput{}, fmt.Errorf("character not found: %s", input.Name)
	}

	return nil, GetCharacterOutput{Character: *char}, nil
}

func (s *Server) handleListCharacters(ctx context.Context, req *mcp.CallToolRequest, input ListCharactersInput) (*mcp.CallToolResult, ListCharactersOutput, error) {
	var characters []types.CharacterSummary
	err := s.withDB(ctx, input.CampaignID, func(dbCtx *DBContext) error {
		var err error
		characters, err = dbCtx.Store.ListCharacters(input.CampaignID, input.Type, input.Status)
		return err
	})
	if err != nil {
		return nil, ListCharactersOutput{}, wrapError("list characters", err)
	}

	return nil, ListCharactersOutput{Characters: characters}, nil
}

func (s *Server) handleSavePlotEvent(ctx context.Context, req *mcp.CallToolRequest, input SavePlotEventInput) (*mcp.CallToolResult, SavePlotEventOutput, error) {
	event := &types.PlotEvent{
		CampaignID:   input.CampaignID,
		Session:      input.Session,
		Summary:      input.Summary,
		Consequences: input.Consequences,
		NPCs:         []string{},
		PCs:          []string{},
		Tags:         []string{},
	}

	hooks := input.Hooks
	if hooks == nil {
		hooks = []string{}
	}

	s.log().Debug("saving plot event", "campaign_id", input.CampaignID, "session", input.Session, "hooks_count", len(hooks))

	var savedHooks []types.Hook
	err := s.withDB(ctx, input.CampaignID, func(dbCtx *DBContext) error {
		if err := dbCtx.Store.SavePlotEvent(event, hooks); err != nil {
			return err
		}
		// Fetch the created hooks
		var err error
		savedHooks, err = dbCtx.Store.ListOpenHooks(input.CampaignID)
		return err
	})
	if err != nil {
		s.log().Error("failed to save plot event", "campaign_id", input.CampaignID, "session", input.Session, "error", err)
		return nil, SavePlotEventOutput{}, wrapError("save plot event", err)
	}

	s.log().Info("plot event saved successfully", "campaign_id", input.CampaignID, "session", input.Session, "event_id", event.ID, "hooks_opened", len(hooks))
	return nil, SavePlotEventOutput{Event: *event, Hooks: savedHooks}, nil
}

func (s *Server) handleListOpenHooks(ctx context.Context, req *mcp.CallToolRequest, input ListOpenHooksInput) (*mcp.CallToolResult, ListOpenHooksOutput, error) {
	var hooks []types.Hook
	err := s.withDB(ctx, input.CampaignID, func(dbCtx *DBContext) error {
		var err error
		hooks, err = dbCtx.Store.ListOpenHooks(input.CampaignID)
		return err
	})
	if err != nil {
		return nil, ListOpenHooksOutput{}, wrapError("list open hooks", err)
	}

	return nil, ListOpenHooksOutput{Hooks: hooks}, nil
}

func (s *Server) handleResolveHook(ctx context.Context, req *mcp.CallToolRequest, input ResolveHookInput) (*mcp.CallToolResult, ResolveHookOutput, error) {
	s.log().Debug("resolving hook", "campaign_id", input.CampaignID, "hook_id", input.HookID)

	var hook *types.Hook
	err := s.withDB(ctx, input.CampaignID, func(dbCtx *DBContext) error {
		if err := dbCtx.Store.ResolveHook(input.CampaignID, input.HookID, input.Resolution); err != nil {
			return err
		}
		// Fetch the resolved hook to return
		hooks, err := dbCtx.Store.ListOpenHooks(input.CampaignID)
		if err != nil {
			return err
		}
		// Find the resolved hook (it's now in closed hooks, but we can construct it)
		// For simplicity, just return a basic hook object
		hook = &types.Hook{
			ID:         input.HookID,
			CampaignID: input.CampaignID,
			Resolved:   true,
			Resolution: input.Resolution,
		}
		_ = hooks // silence unused variable
		return nil
	})
	if err != nil {
		s.log().Error("failed to resolve hook", "campaign_id", input.CampaignID, "hook_id", input.HookID, "error", err)
		return nil, ResolveHookOutput{}, wrapError("resolve hook", err)
	}

	s.log().Info("hook resolved successfully", "campaign_id", input.CampaignID, "hook_id", input.HookID)
	return nil, ResolveHookOutput{Hook: *hook}, nil
}

func (s *Server) handleSetWorldFlag(ctx context.Context, req *mcp.CallToolRequest, input SetWorldFlagInput) (*mcp.CallToolResult, SetWorldFlagOutput, error) {
	err := s.withDB(ctx, input.CampaignID, func(dbCtx *DBContext) error {
		return dbCtx.Store.SetWorldFlag(input.CampaignID, input.Key, input.Value)
	})
	if err != nil {
		return nil, SetWorldFlagOutput{}, wrapError("set world flag", err)
	}

	return nil, SetWorldFlagOutput{Success: true}, nil
}

func (s *Server) handleGetWorldFlags(ctx context.Context, req *mcp.CallToolRequest, input GetWorldFlagsInput) (*mcp.CallToolResult, GetWorldFlagsOutput, error) {
	var flags map[string]string
	err := s.withDB(ctx, input.CampaignID, func(dbCtx *DBContext) error {
		var err error
		flags, err = dbCtx.Store.GetWorldFlags(input.CampaignID)
		return err
	})
	if err != nil {
		return nil, GetWorldFlagsOutput{}, wrapError("get world flags", err)
	}

	return nil, GetWorldFlagsOutput{Flags: flags}, nil
}
