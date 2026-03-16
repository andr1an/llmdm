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
	s.logToolEntry("create_campaign", "", "name", input.Name)

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
		s.logToolError("create_campaign", campaignID, err)
		return nil, CreateCampaignOutput{}, wrapError("create campaign", err)
	}

	// Get db path by reconstructing it
	dbPath, _ = db.CampaignDBPath(s.dbPath, campaignID)
	s.logToolExit("create_campaign", campaignID, "name", input.Name)

	return nil, CreateCampaignOutput{
		CampaignID: campaignID,
		DBPath:     dbPath,
		Campaign:   *campaign,
	}, nil
}

func (s *Server) handleListCampaigns(ctx context.Context, req *mcp.CallToolRequest, input ListCampaignsInput) (*mcp.CallToolResult, ListCampaignsOutput, error) {
	s.logToolEntry("list_campaigns", "")

	entries, err := os.ReadDir(s.DBPath())
	if err != nil {
		if os.IsNotExist(err) {
			s.logToolExit("list_campaigns", "", "count", 0)
			return nil, ListCampaignsOutput{Campaigns: []types.Campaign{}}, nil
		}
		s.logToolError("list_campaigns", "", err)
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
			skipped++
			continue
		}

		store := memory.NewStore(database.DB)
		campaign, err := store.GetCampaign(campaignID)
		database.Close()

		if err != nil || campaign == nil {
			skipped++
			continue
		}

		campaigns = append(campaigns, *campaign)
	}

	if campaigns == nil {
		campaigns = []types.Campaign{}
	}

	s.logToolExit("list_campaigns", "", "count", len(campaigns))
	return nil, ListCampaignsOutput{Campaigns: campaigns}, nil
}

func (s *Server) handleSaveCharacter(ctx context.Context, req *mcp.CallToolRequest, input SaveCharacterInput) (*mcp.CallToolResult, SaveCharacterOutput, error) {
	if err := validateMaxLength("backstory", input.Backstory, maxCharacterBackstoryLength); err != nil {
		return nil, SaveCharacterOutput{}, err
	}
	if err := validateMaxLength("notes", input.Notes, maxCharacterNotesLength); err != nil {
		return nil, SaveCharacterOutput{}, err
	}
	if err := validateAlignment(input.Alignment); err != nil {
		return nil, SaveCharacterOutput{}, err
	}
	if input.AC != 0 {
		if err := validateAC(input.AC); err != nil {
			return nil, SaveCharacterOutput{}, err
		}
	}
	if input.Spellcasting != nil {
		if err := validateSpellcastingAbility(input.Spellcasting.Ability); err != nil {
			return nil, SaveCharacterOutput{}, err
		}
	}

	level := input.Level
	if level == 0 {
		level = 1
	}
	status := input.Status
	if status == "" {
		status = "active"
	}

	ac := input.AC
	if ac == 0 {
		ac = 10
	}
	speed := input.Speed
	if speed == "" {
		speed = "30 ft"
	}

	proficiencies := types.Proficiencies{
		Armor:        []string{},
		Weapons:      []string{},
		Tools:        []string{},
		SavingThrows: []string{},
		Skills:       []string{},
	}
	if input.Proficiencies != nil {
		proficiencies = *input.Proficiencies
	}

	skills := input.Skills
	if skills == nil {
		skills = []types.Skill{}
	}

	languages := input.Languages
	if languages == nil {
		languages = []string{}
	}

	features := input.Features
	if features == nil {
		features = []types.Feature{}
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
		Alignment:        input.Alignment,
		AC:               ac,
		Speed:            speed,
		ExperiencePoints: input.ExperiencePoints,
		Proficiencies:    proficiencies,
		Skills:           skills,
		Languages:        languages,
		Features:         features,
		Spellcasting:     input.Spellcasting,
		Gold:             input.Gold,
		Backstory:        input.Backstory,
		Status:           status,
		Notes:            input.Notes,
		Inventory:        []string{},
		Conditions:       []string{},
		PlotFlags:        []string{},
		Relationships:    map[string]string{},
	}

	s.logToolEntry("save_character", input.CampaignID, "name", input.Name, "type", input.Type)

	err := s.withDB(ctx, input.CampaignID, func(dbCtx *DBContext) error {
		return dbCtx.Store.SaveCharacter(char)
	})
	if err != nil {
		s.logToolError("save_character", input.CampaignID, err, "name", input.Name)
		return nil, SaveCharacterOutput{}, wrapError("save character", err)
	}

	s.logToolExit("save_character", input.CampaignID, "name", input.Name, "type", input.Type)
	return nil, SaveCharacterOutput{Character: *char}, nil
}

func (s *Server) handleUpdateCharacter(ctx context.Context, req *mcp.CallToolRequest, input UpdateCharacterInput) (*mcp.CallToolResult, UpdateCharacterOutput, error) {
	if input.Notes != "" {
		if err := validateMaxLength("notes", input.Notes, maxCharacterNotesLength); err != nil {
			return nil, UpdateCharacterOutput{}, err
		}
	}
	if input.Alignment != "" {
		if err := validateAlignment(input.Alignment); err != nil {
			return nil, UpdateCharacterOutput{}, err
		}
	}
	if input.AC != nil {
		if err := validateAC(*input.AC); err != nil {
			return nil, UpdateCharacterOutput{}, err
		}
	}
	if input.Spellcasting != nil {
		if err := validateSpellcastingAbility(input.Spellcasting.Ability); err != nil {
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
	if input.Alignment != "" {
		update.Alignment = &input.Alignment
	}
	if input.AC != nil {
		update.AC = input.AC
	}
	if input.Speed != "" {
		update.Speed = &input.Speed
	}
	if input.ExperiencePoints != nil {
		update.ExperiencePoints = input.ExperiencePoints
	}
	if input.Proficiencies != nil {
		update.Proficiencies = input.Proficiencies
	}
	if input.Skills != nil {
		update.Skills = input.Skills
	}
	if input.Languages != nil {
		update.Languages = input.Languages
	}
	if input.Features != nil {
		update.Features = input.Features
	}
	if input.Spellcasting != nil {
		update.Spellcasting = input.Spellcasting
	}

	s.logToolEntry("update_character", input.CampaignID, "name", input.Name)

	var character *types.Character
	err := s.withDB(ctx, input.CampaignID, func(dbCtx *DBContext) error {
		_, err := dbCtx.Store.UpdateCharacter(input.CampaignID, input.Name, update)
		if err != nil {
			return err
		}

		// Fetch the updated character
		character, err = dbCtx.Store.GetCharacter(input.CampaignID, input.Name)
		return err
	})
	if err != nil {
		s.logToolError("update_character", input.CampaignID, err, "name", input.Name)
		return nil, UpdateCharacterOutput{}, wrapError("update character", err)
	}

	if character == nil {
		return nil, UpdateCharacterOutput{}, fmt.Errorf("character not found after update: %s", input.Name)
	}

	s.logToolExit("update_character", input.CampaignID, "name", input.Name)
	return nil, UpdateCharacterOutput{Character: *character}, nil
}

func (s *Server) handleGetCharacter(ctx context.Context, req *mcp.CallToolRequest, input GetCharacterInput) (*mcp.CallToolResult, GetCharacterOutput, error) {
	s.logToolEntry("get_character", input.CampaignID, "name", input.Name)

	var char *types.Character
	err := s.withDB(ctx, input.CampaignID, func(dbCtx *DBContext) error {
		var err error
		char, err = dbCtx.Store.GetCharacter(input.CampaignID, input.Name)
		return err
	})
	if err != nil {
		s.logToolError("get_character", input.CampaignID, err, "name", input.Name)
		return nil, GetCharacterOutput{}, wrapError("get character", err)
	}
	if char == nil {
		return nil, GetCharacterOutput{}, fmt.Errorf("character not found: %s", input.Name)
	}

	s.logToolExit("get_character", input.CampaignID, "name", input.Name)
	return nil, GetCharacterOutput{Character: *char}, nil
}

func (s *Server) handleListCharacters(ctx context.Context, req *mcp.CallToolRequest, input ListCharactersInput) (*mcp.CallToolResult, ListCharactersOutput, error) {
	s.logToolEntry("list_characters", input.CampaignID, "type", input.Type, "status", input.Status)

	var characters []types.CharacterSummary
	err := s.withDB(ctx, input.CampaignID, func(dbCtx *DBContext) error {
		var err error
		characters, err = dbCtx.Store.ListCharacters(input.CampaignID, input.Type, input.Status)
		return err
	})
	if err != nil {
		s.logToolError("list_characters", input.CampaignID, err)
		return nil, ListCharactersOutput{}, wrapError("list characters", err)
	}

	s.logToolExit("list_characters", input.CampaignID, "count", len(characters))
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

	s.logToolEntry("save_plot_event", input.CampaignID, "session", input.Session, "hooks", len(hooks))

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
		s.logToolError("save_plot_event", input.CampaignID, err, "session", input.Session)
		return nil, SavePlotEventOutput{}, wrapError("save plot event", err)
	}

	s.logToolExit("save_plot_event", input.CampaignID, "session", input.Session, "hooks_opened", len(hooks))
	return nil, SavePlotEventOutput{Event: *event, Hooks: savedHooks}, nil
}

func (s *Server) handleListOpenHooks(ctx context.Context, req *mcp.CallToolRequest, input ListOpenHooksInput) (*mcp.CallToolResult, ListOpenHooksOutput, error) {
	s.logToolEntry("list_open_hooks", input.CampaignID)

	var hooks []types.Hook
	err := s.withDB(ctx, input.CampaignID, func(dbCtx *DBContext) error {
		var err error
		hooks, err = dbCtx.Store.ListOpenHooks(input.CampaignID)
		return err
	})
	if err != nil {
		s.logToolError("list_open_hooks", input.CampaignID, err)
		return nil, ListOpenHooksOutput{}, wrapError("list open hooks", err)
	}

	s.logToolExit("list_open_hooks", input.CampaignID, "count", len(hooks))
	return nil, ListOpenHooksOutput{Hooks: hooks}, nil
}

func (s *Server) handleResolveHook(ctx context.Context, req *mcp.CallToolRequest, input ResolveHookInput) (*mcp.CallToolResult, ResolveHookOutput, error) {
	s.logToolEntry("resolve_hook", input.CampaignID, "hook_id", input.HookID)

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
		s.logToolError("resolve_hook", input.CampaignID, err, "hook_id", input.HookID)
		return nil, ResolveHookOutput{}, wrapError("resolve hook", err)
	}

	s.logToolExit("resolve_hook", input.CampaignID, "hook_id", input.HookID)
	return nil, ResolveHookOutput{Hook: *hook}, nil
}

func (s *Server) handleSetWorldFlag(ctx context.Context, req *mcp.CallToolRequest, input SetWorldFlagInput) (*mcp.CallToolResult, SetWorldFlagOutput, error) {
	s.logToolEntry("set_world_flag", input.CampaignID, "key", input.Key)

	err := s.withDB(ctx, input.CampaignID, func(dbCtx *DBContext) error {
		return dbCtx.Store.SetWorldFlag(input.CampaignID, input.Key, input.Value)
	})
	if err != nil {
		s.logToolError("set_world_flag", input.CampaignID, err, "key", input.Key)
		return nil, SetWorldFlagOutput{}, wrapError("set world flag", err)
	}

	s.logToolExit("set_world_flag", input.CampaignID, "key", input.Key)
	return nil, SetWorldFlagOutput{Success: true}, nil
}

func (s *Server) handleGetWorldFlags(ctx context.Context, req *mcp.CallToolRequest, input GetWorldFlagsInput) (*mcp.CallToolResult, GetWorldFlagsOutput, error) {
	s.logToolEntry("get_world_flags", input.CampaignID)

	var flags map[string]string
	err := s.withDB(ctx, input.CampaignID, func(dbCtx *DBContext) error {
		var err error
		flags, err = dbCtx.Store.GetWorldFlags(input.CampaignID)
		return err
	})
	if err != nil {
		s.logToolError("get_world_flags", input.CampaignID, err)
		return nil, GetWorldFlagsOutput{}, wrapError("get world flags", err)
	}

	s.logToolExit("get_world_flags", input.CampaignID, "count", len(flags))
	return nil, GetWorldFlagsOutput{Flags: flags}, nil
}
