package mcpserver

import "github.com/andr1an/llmdm/internal/types"

// === Dice Tool Structs ===

// RollInput contains parameters for the roll tool.
type RollInput struct {
	CampaignID   string `json:"campaign_id" jsonschema:"Campaign ID for logging the roll"`
	Notation     string `json:"notation" jsonschema:"Dice notation (e.g. '1d20' '2d6+3' '4d6kh3')"`
	Reason       string `json:"reason,omitempty" jsonschema:"Reason for the roll (e.g. 'Stealth check')"`
	Character    string `json:"character,omitempty" jsonschema:"Character making the roll"`
	Session      int    `json:"session,omitempty" jsonschema:"Current session number"`
	Advantage    bool   `json:"advantage,omitempty" jsonschema:"Roll with advantage (2d20 keep highest)"`
	Disadvantage bool   `json:"disadvantage,omitempty" jsonschema:"Roll with disadvantage (2d20 keep lowest)"`
}

// RollOutput contains the result of a roll.
type RollOutput struct {
	RollResult types.RollResult `json:"roll_result"`
}

// RollContestedInput contains parameters for the roll_contested tool.
type RollContestedInput struct {
	CampaignID       string `json:"campaign_id" jsonschema:"Campaign ID"`
	Attacker         string `json:"attacker" jsonschema:"Attacker's name"`
	Defender         string `json:"defender" jsonschema:"Defender's name"`
	AttackerNotation string `json:"attacker_notation" jsonschema:"Dice notation for attacker (e.g. '1d20+5')"`
	DefenderNotation string `json:"defender_notation" jsonschema:"Dice notation for defender (e.g. '1d20+3')"`
	ContestType      string `json:"contest_type" jsonschema:"Type of contest (e.g. 'Attack vs AC' 'Deception vs Insight')"`
	Session          int    `json:"session,omitempty" jsonschema:"Current session number"`
}

// RollContestedOutput contains the result of a contested roll.
type RollContestedOutput struct {
	Result types.ContestedRollResult `json:"result"`
}

// RollSavingThrowInput contains parameters for the roll_saving_throw tool.
type RollSavingThrowInput struct {
	CampaignID string  `json:"campaign_id" jsonschema:"Campaign ID"`
	Character  string  `json:"character" jsonschema:"Character making the saving throw"`
	Stat       string  `json:"stat" jsonschema:"Ability score to use,enum=STR|DEX|CON|INT|WIS|CHA"`
	Modifier   float64 `json:"modifier" jsonschema:"Total modifier for the saving throw"`
	DC         float64 `json:"dc" jsonschema:"Difficulty class to beat"`
	Reason     string  `json:"reason,omitempty" jsonschema:"Reason for the saving throw"`
	Session    int     `json:"session,omitempty" jsonschema:"Current session number"`
}

// RollSavingThrowOutput contains the result of a saving throw.
type RollSavingThrowOutput struct {
	Result types.SavingThrowResult `json:"result"`
}

// GetRollHistoryInput contains parameters for the get_roll_history tool.
type GetRollHistoryInput struct {
	CampaignID string `json:"campaign_id" jsonschema:"Campaign ID"`
	Character  string `json:"character,omitempty" jsonschema:"Filter by character name"`
	Session    int    `json:"session,omitempty" jsonschema:"Filter by session number"`
	Limit      int    `json:"limit,omitempty" jsonschema:"Maximum number of rolls to return (default: 50)"`
}

// GetRollHistoryOutput contains roll history records.
type GetRollHistoryOutput struct {
	Records []types.RollRecord `json:"records"`
}

// === Memory Tool Structs ===

// CreateCampaignInput contains parameters for the create_campaign tool.
type CreateCampaignInput struct {
	Name        string `json:"name" jsonschema:"Campaign name (e.g. 'Lost Mines of Phandelver')"`
	Description string `json:"description,omitempty" jsonschema:"Brief setting description"`
}

// CreateCampaignOutput contains the created campaign details.
type CreateCampaignOutput struct {
	CampaignID string         `json:"campaign_id"`
	DBPath     string         `json:"db_path"`
	Campaign   types.Campaign `json:"campaign"`
}

// ListCampaignsInput contains parameters for the list_campaigns tool.
type ListCampaignsInput struct {
	// No parameters needed
}

// ListCampaignsOutput contains the list of campaigns.
type ListCampaignsOutput struct {
	Campaigns []types.Campaign `json:"campaigns"`
}

// SaveCharacterInput contains parameters for the save_character tool.
type SaveCharacterInput struct {
	CampaignID       string               `json:"campaign_id" jsonschema:"Campaign ID"`
	Name             string               `json:"name" jsonschema:"Character name"`
	Type             string               `json:"type" jsonschema:"Character type,enum=pc|npc"`
	Class            string               `json:"class,omitempty" jsonschema:"Character class (e.g. 'Paladin')"`
	Race             string               `json:"race,omitempty" jsonschema:"Character race (e.g. 'Human')"`
	Level            int                  `json:"level,omitempty" jsonschema:"Character level (default: 1)"`
	HPCurrent        float64              `json:"hp_current" jsonschema:"Current hit points"`
	HPMax            float64              `json:"hp_max" jsonschema:"Maximum hit points"`
	STR              int                  `json:"str,omitempty" jsonschema:"Strength score"`
	DEX              int                  `json:"dex,omitempty" jsonschema:"Dexterity score"`
	CON              int                  `json:"con,omitempty" jsonschema:"Constitution score"`
	INTStat          int                  `json:"int_stat,omitempty" jsonschema:"Intelligence score"`
	WIS              int                  `json:"wis,omitempty" jsonschema:"Wisdom score"`
	CHA              int                  `json:"cha,omitempty" jsonschema:"Charisma score"`
	Alignment        string               `json:"alignment,omitempty" jsonschema:"D&D 5e alignment (e.g. 'Lawful Good' 'Chaotic Neutral')"`
	AC               int                  `json:"ac,omitempty" jsonschema:"Armor class (default: 10)"`
	Speed            string               `json:"speed,omitempty" jsonschema:"Movement speed (default: '30 ft')"`
	ExperiencePoints int                  `json:"experience_points,omitempty" jsonschema:"Experience points (default: 0)"`
	Proficiencies    *types.Proficiencies `json:"proficiencies,omitempty"`
	Skills           []types.Skill        `json:"skills,omitempty"`
	Languages        []string             `json:"languages,omitempty"`
	Features         []types.Feature      `json:"features,omitempty"`
	Spellcasting     *types.Spellcasting  `json:"spellcasting,omitempty"`
	Gold             int                  `json:"gold,omitempty" jsonschema:"Gold pieces (default: 0)"`
	Backstory        string               `json:"backstory,omitempty" jsonschema:"Character backstory"`
	Status           string               `json:"status,omitempty" jsonschema:"Character status,enum=active|dead|missing|retired"`
	Notes            string               `json:"notes,omitempty" jsonschema:"DM private notes (for NPCs)"`
}

// SaveCharacterOutput contains the saved character.
type SaveCharacterOutput struct {
	Character types.Character `json:"character"`
}

// UpdateCharacterInput contains parameters for the update_character tool.
type UpdateCharacterInput struct {
	CampaignID       string                 `json:"campaign_id" jsonschema:"Campaign ID"`
	Name             string                 `json:"name" jsonschema:"Character name"`
	HPCurrent        *float64               `json:"hp_current,omitempty" jsonschema:"New current hit points"`
	Level            *int                   `json:"level,omitempty" jsonschema:"New level"`
	Gold             *int                   `json:"gold,omitempty" jsonschema:"New gold amount"`
	Status           string                 `json:"status,omitempty" jsonschema:"New status,enum=active|dead|missing|retired"`
	Notes            string                 `json:"notes,omitempty" jsonschema:"New DM notes"`
	Inventory        []string               `json:"inventory,omitempty" jsonschema:"Replace inventory with array of item strings"`
	Conditions       []string               `json:"conditions,omitempty" jsonschema:"Replace conditions with array of condition strings"`
	PlotFlags        []string               `json:"plot_flags,omitempty" jsonschema:"Replace plot flags with array of flag strings"`
	Relationships    map[string]interface{} `json:"relationships,omitempty" jsonschema:"Replace relationships map (name -> relation)"`
	Alignment        string                 `json:"alignment,omitempty" jsonschema:"D&D 5e alignment (e.g. 'Lawful Good' 'Chaotic Neutral')"`
	AC               *int                   `json:"ac,omitempty"`
	Speed            string                 `json:"speed,omitempty"`
	ExperiencePoints *int                   `json:"experience_points,omitempty"`
	Proficiencies    *types.Proficiencies   `json:"proficiencies,omitempty"`
	Skills           []types.Skill          `json:"skills,omitempty"`
	Languages        []string               `json:"languages,omitempty"`
	Features         []types.Feature        `json:"features,omitempty"`
	Spellcasting     *types.Spellcasting    `json:"spellcasting,omitempty"`
}

// UpdateCharacterOutput contains the updated character.
type UpdateCharacterOutput struct {
	Character types.Character `json:"character"`
}

// GetCharacterInput contains parameters for the get_character tool.
type GetCharacterInput struct {
	CampaignID string `json:"campaign_id" jsonschema:"Campaign ID"`
	Name       string `json:"name" jsonschema:"Character name"`
}

// GetCharacterOutput contains the character sheet.
type GetCharacterOutput struct {
	Character types.Character `json:"character"`
}

// ListCharactersInput contains parameters for the list_characters tool.
type ListCharactersInput struct {
	CampaignID string `json:"campaign_id" jsonschema:"Campaign ID"`
	Type       string `json:"type,omitempty" jsonschema:"Filter by type,enum=pc|npc"`
	Status     string `json:"status,omitempty" jsonschema:"Filter by status,enum=active|dead|missing|retired"`
}

// ListCharactersOutput contains the list of characters.
type ListCharactersOutput struct {
	Characters []types.CharacterSummary `json:"characters"`
}

// SavePlotEventInput contains parameters for the save_plot_event tool.
type SavePlotEventInput struct {
	CampaignID   string   `json:"campaign_id" jsonschema:"Campaign ID"`
	Session      int      `json:"session" jsonschema:"Session number"`
	Summary      string   `json:"summary" jsonschema:"2-4 sentence narrative description"`
	Consequences string   `json:"consequences,omitempty" jsonschema:"What changed in the world"`
	Hooks        []string `json:"hooks,omitempty" jsonschema:"Array of plot hook strings - unresolved story threads to track"`
}

// SavePlotEventOutput contains the saved plot event.
type SavePlotEventOutput struct {
	Event types.PlotEvent `json:"event"`
	Hooks []types.Hook    `json:"hooks"`
}

// ListOpenHooksInput contains parameters for the list_open_hooks tool.
type ListOpenHooksInput struct {
	CampaignID string `json:"campaign_id" jsonschema:"Campaign ID"`
}

// ListOpenHooksOutput contains the list of open hooks.
type ListOpenHooksOutput struct {
	Hooks []types.Hook `json:"hooks"`
}

// ResolveHookInput contains parameters for the resolve_hook tool.
type ResolveHookInput struct {
	CampaignID string `json:"campaign_id" jsonschema:"Campaign ID"`
	HookID     string `json:"hook_id" jsonschema:"Hook ID to resolve"`
	Resolution string `json:"resolution" jsonschema:"How the hook was resolved"`
}

// ResolveHookOutput contains the resolved hook.
type ResolveHookOutput struct {
	Hook types.Hook `json:"hook"`
}

// SetWorldFlagInput contains parameters for the set_world_flag tool.
type SetWorldFlagInput struct {
	CampaignID string `json:"campaign_id" jsonschema:"Campaign ID"`
	Key        string `json:"key" jsonschema:"Flag key"`
	Value      string `json:"value" jsonschema:"Flag value"`
}

// SetWorldFlagOutput contains confirmation.
type SetWorldFlagOutput struct {
	Success bool `json:"success"`
}

// GetWorldFlagsInput contains parameters for the get_world_flags tool.
type GetWorldFlagsInput struct {
	CampaignID string `json:"campaign_id" jsonschema:"Campaign ID"`
}

// GetWorldFlagsOutput contains world flags.
type GetWorldFlagsOutput struct {
	Flags map[string]string `json:"flags"`
}

// === Session Tool Structs ===

// StartSessionInput contains parameters for the start_session tool.
type StartSessionInput struct {
	CampaignID     string `json:"campaign_id" jsonschema:"Campaign ID"`
	Session        int    `json:"session" jsonschema:"Session number to start"`
	RecentSessions int    `json:"recent_sessions,omitempty" jsonschema:"How many prior sessions to include (default: 3)"`
}

// StartSessionOutput contains the session brief.
type StartSessionOutput struct {
	Brief types.SessionBrief `json:"brief"`
}

// EndSessionInput contains parameters for the end_session tool.
type EndSessionInput struct {
	CampaignID string `json:"campaign_id" jsonschema:"Campaign ID"`
	Session    int    `json:"session" jsonschema:"Session number to end"`
	RawEvents  string `json:"raw_events" jsonschema:"Full narrative event log for the session"`
	DMNotes    string `json:"dm_notes,omitempty" jsonschema:"Optional DM notes"`
}

// EndSessionOutput contains the session summary.
type EndSessionOutput struct {
	Summary string `json:"summary"`
}

// CheckpointInput contains parameters for the checkpoint tool.
type CheckpointInput struct {
	CampaignID string                 `json:"campaign_id" jsonschema:"Campaign ID"`
	Session    int                    `json:"session" jsonschema:"Current session number"`
	Note       string                 `json:"note" jsonschema:"Checkpoint note"`
	Data       map[string]interface{} `json:"data,omitempty" jsonschema:"Optional turn data (turn_id sequence player_action narrative tool_results)"`
}

// CheckpointOutput contains the checkpoint ID.
type CheckpointOutput struct {
	CheckpointID string `json:"checkpoint_id"`
}

// GetTurnHistoryInput contains parameters for the get_turn_history tool.
type GetTurnHistoryInput struct {
	CampaignID string `json:"campaign_id" jsonschema:"Campaign ID"`
	Session    int    `json:"session" jsonschema:"Session number"`
	Limit      int    `json:"limit,omitempty" jsonschema:"Maximum number of turns to return (default: 50)"`
}

// GetTurnHistoryOutput contains turn history.
type GetTurnHistoryOutput struct {
	Turns []types.Checkpoint `json:"turns"`
}

// GetSessionBriefInput contains parameters for the get_session_brief tool.
type GetSessionBriefInput struct {
	CampaignID string `json:"campaign_id" jsonschema:"Campaign ID"`
}

// GetSessionBriefOutput contains the session brief.
type GetSessionBriefOutput struct {
	Brief string `json:"brief"`
}

// ListSessionsInput contains parameters for the list_sessions tool.
type ListSessionsInput struct {
	CampaignID string `json:"campaign_id" jsonschema:"Campaign ID"`
}

// ListSessionsOutput contains session metadata.
type ListSessionsOutput struct {
	Sessions []types.SessionMeta `json:"sessions"`
}

// GetNPCRelationshipsInput contains parameters for the get_npc_relationships tool.
type GetNPCRelationshipsInput struct {
	CampaignID string `json:"campaign_id" jsonschema:"Campaign ID"`
	NPCName    string `json:"npc_name,omitempty" jsonschema:"Optional NPC name filter"`
}

// GetNPCRelationshipsOutput contains relationship edges.
type GetNPCRelationshipsOutput struct {
	Relationships []types.RelationshipEdge `json:"relationships"`
}

// ExportSessionRecapInput contains parameters for the export_session_recap tool.
type ExportSessionRecapInput struct {
	CampaignID  string   `json:"campaign_id" jsonschema:"Campaign ID"`
	FromSession *float64 `json:"from_session,omitempty" jsonschema:"Optional lower inclusive session bound"`
	ToSession   *float64 `json:"to_session,omitempty" jsonschema:"Optional upper inclusive session bound"`
}

// ExportSessionRecapOutput contains the markdown recap.
type ExportSessionRecapOutput struct {
	Markdown string `json:"markdown"`
}
