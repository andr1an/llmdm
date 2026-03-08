package mcpserver

import "github.com/mark3labs/mcp-go/mcp"

func (s *Server) registerTools() {
	// roll tool
	rollTool := mcp.NewTool("roll",
		mcp.WithDescription("Roll dice using standard D&D notation. Supports NdM, NdM+K, NdM-K, NdMkhX (keep highest), NdMklX (keep lowest). All rolls are logged."),
		mcp.WithString("campaign_id",
			mcp.Required(),
			mcp.Description("Campaign ID for logging the roll"),
		),
		mcp.WithString("notation",
			mcp.Required(),
			mcp.Description("Dice notation (e.g., '1d20', '2d6+3', '4d6kh3')"),
		),
		mcp.WithString("reason",
			mcp.Description("Reason for the roll (e.g., 'Stealth check')"),
		),
		mcp.WithString("character",
			mcp.Description("Character making the roll"),
		),
		mcp.WithNumber("session",
			mcp.Description("Current session number"),
		),
		mcp.WithBoolean("advantage",
			mcp.Description("Roll with advantage (2d20 keep highest)"),
		),
		mcp.WithBoolean("disadvantage",
			mcp.Description("Roll with disadvantage (2d20 keep lowest)"),
		),
	)
	s.mcp.AddTool(rollTool, s.handleRoll)

	// roll_contested tool
	contestedTool := mcp.NewTool("roll_contested",
		mcp.WithDescription("Roll a contested check between attacker and defender. Both rolls are logged."),
		mcp.WithString("campaign_id",
			mcp.Required(),
			mcp.Description("Campaign ID"),
		),
		mcp.WithString("attacker",
			mcp.Required(),
			mcp.Description("Attacker's name"),
		),
		mcp.WithString("defender",
			mcp.Required(),
			mcp.Description("Defender's name"),
		),
		mcp.WithString("attacker_notation",
			mcp.Required(),
			mcp.Description("Dice notation for attacker (e.g., '1d20+5')"),
		),
		mcp.WithString("defender_notation",
			mcp.Required(),
			mcp.Description("Dice notation for defender (e.g., '1d20+3')"),
		),
		mcp.WithString("contest_type",
			mcp.Required(),
			mcp.Description("Type of contest (e.g., 'Attack vs AC', 'Deception vs Insight')"),
		),
		mcp.WithNumber("session",
			mcp.Description("Current session number"),
		),
	)
	s.mcp.AddTool(contestedTool, s.handleRollContested)

	// roll_saving_throw tool
	savingThrowTool := mcp.NewTool("roll_saving_throw",
		mcp.WithDescription("Roll a saving throw and check against a DC."),
		mcp.WithString("campaign_id",
			mcp.Required(),
			mcp.Description("Campaign ID"),
		),
		mcp.WithString("character",
			mcp.Required(),
			mcp.Description("Character making the saving throw"),
		),
		mcp.WithString("stat",
			mcp.Required(),
			mcp.Description("Ability score to use"),
			mcp.Enum("STR", "DEX", "CON", "INT", "WIS", "CHA"),
		),
		mcp.WithNumber("modifier",
			mcp.Required(),
			mcp.Description("Total modifier for the saving throw"),
		),
		mcp.WithNumber("dc",
			mcp.Required(),
			mcp.Description("Difficulty class to beat"),
		),
		mcp.WithString("reason",
			mcp.Description("Reason for the saving throw"),
		),
		mcp.WithNumber("session",
			mcp.Description("Current session number"),
		),
	)
	s.mcp.AddTool(savingThrowTool, s.handleRollSavingThrow)

	// get_roll_history tool
	historyTool := mcp.NewTool("get_roll_history",
		mcp.WithDescription("Retrieve dice roll history for a campaign."),
		mcp.WithString("campaign_id",
			mcp.Required(),
			mcp.Description("Campaign ID"),
		),
		mcp.WithString("character",
			mcp.Description("Filter by character name"),
		),
		mcp.WithNumber("session",
			mcp.Description("Filter by session number"),
		),
		mcp.WithNumber("limit",
			mcp.Description("Maximum number of rolls to return (default: 50)"),
		),
	)
	s.mcp.AddTool(historyTool, s.handleGetRollHistory)

	// === Campaign Memory Tools ===

	// create_campaign tool
	createCampaignTool := mcp.NewTool("create_campaign",
		mcp.WithDescription("Create a new D&D campaign. Returns the campaign ID and database path."),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("Campaign name (e.g., 'Lost Mines of Phandelver')"),
		),
		mcp.WithString("description",
			mcp.Description("Brief setting description"),
		),
	)
	s.mcp.AddTool(createCampaignTool, s.handleCreateCampaign)

	// save_character tool
	saveCharacterTool := mcp.NewTool("save_character",
		mcp.WithDescription("Create or fully replace a character (PC or NPC). Use update_character for partial updates."),
		mcp.WithString("campaign_id", mcp.Required(), mcp.Description("Campaign ID")),
		mcp.WithString("name", mcp.Required(), mcp.Description("Character name")),
		mcp.WithString("type", mcp.Required(), mcp.Description("Character type"), mcp.Enum("pc", "npc")),
		mcp.WithString("class", mcp.Description("Character class (e.g., 'Paladin')")),
		mcp.WithString("race", mcp.Description("Character race (e.g., 'Human')")),
		mcp.WithNumber("level", mcp.Description("Character level (default: 1)")),
		mcp.WithNumber("hp_current", mcp.Required(), mcp.Description("Current hit points")),
		mcp.WithNumber("hp_max", mcp.Required(), mcp.Description("Maximum hit points")),
		mcp.WithNumber("str", mcp.Description("Strength score")),
		mcp.WithNumber("dex", mcp.Description("Dexterity score")),
		mcp.WithNumber("con", mcp.Description("Constitution score")),
		mcp.WithNumber("int_stat", mcp.Description("Intelligence score")),
		mcp.WithNumber("wis", mcp.Description("Wisdom score")),
		mcp.WithNumber("cha", mcp.Description("Charisma score")),
		mcp.WithString("backstory", mcp.Description("Character backstory"), mcp.MaxLength(maxCharacterBackstoryLength)),
		mcp.WithString("status", mcp.Description("Character status"), mcp.Enum("active", "dead", "missing", "retired")),
		mcp.WithString("notes", mcp.Description("DM private notes (for NPCs)"), mcp.MaxLength(maxCharacterNotesLength)),
	)
	s.mcp.AddTool(saveCharacterTool, s.handleSaveCharacter)

	// update_character tool
	updateCharacterTool := mcp.NewTool("update_character",
		mcp.WithDescription("Patch specific fields on a character. Only provided fields are updated."),
		mcp.WithString("campaign_id", mcp.Required(), mcp.Description("Campaign ID")),
		mcp.WithString("name", mcp.Required(), mcp.Description("Character name")),
		mcp.WithNumber("hp_current", mcp.Description("New current hit points")),
		mcp.WithNumber("level", mcp.Description("New level")),
		mcp.WithString("status", mcp.Description("New status"), mcp.Enum("active", "dead", "missing", "retired")),
		mcp.WithString("notes", mcp.Description("New DM notes"), mcp.MaxLength(maxCharacterNotesLength)),
	)
	s.mcp.AddTool(updateCharacterTool, s.handleUpdateCharacter)

	// get_character tool
	getCharacterTool := mcp.NewTool("get_character",
		mcp.WithDescription("Get full character sheet by name."),
		mcp.WithString("campaign_id", mcp.Required(), mcp.Description("Campaign ID")),
		mcp.WithString("name", mcp.Required(), mcp.Description("Character name")),
	)
	s.mcp.AddTool(getCharacterTool, s.handleGetCharacter)

	// list_characters tool
	listCharactersTool := mcp.NewTool("list_characters",
		mcp.WithDescription("List characters in a campaign, optionally filtered by type and status."),
		mcp.WithString("campaign_id", mcp.Required(), mcp.Description("Campaign ID")),
		mcp.WithString("type", mcp.Description("Filter by type"), mcp.Enum("pc", "npc")),
		mcp.WithString("status", mcp.Description("Filter by status"), mcp.Enum("active", "dead", "missing", "retired")),
	)
	s.mcp.AddTool(listCharactersTool, s.handleListCharacters)

	// save_plot_event tool
	savePlotEventTool := mcp.NewTool("save_plot_event",
		mcp.WithDescription("Record a narrative event in the campaign. Creates plot hooks if provided."),
		mcp.WithString("campaign_id", mcp.Required(), mcp.Description("Campaign ID")),
		mcp.WithNumber("session", mcp.Required(), mcp.Description("Session number")),
		mcp.WithString("summary", mcp.Required(), mcp.Description("2-4 sentence narrative description")),
		mcp.WithString("consequences", mcp.Description("What changed in the world")),
	)
	s.mcp.AddTool(savePlotEventTool, s.handleSavePlotEvent)

	// list_open_hooks tool
	listOpenHooksTool := mcp.NewTool("list_open_hooks",
		mcp.WithDescription("List all unresolved plot threads for a campaign."),
		mcp.WithString("campaign_id", mcp.Required(), mcp.Description("Campaign ID")),
	)
	s.mcp.AddTool(listOpenHooksTool, s.handleListOpenHooks)

	// resolve_hook tool
	resolveHookTool := mcp.NewTool("resolve_hook",
		mcp.WithDescription("Mark a plot hook as resolved."),
		mcp.WithString("campaign_id", mcp.Required(), mcp.Description("Campaign ID")),
		mcp.WithString("hook_id", mcp.Required(), mcp.Description("Hook ID to resolve")),
		mcp.WithString("resolution", mcp.Required(), mcp.Description("How the hook was resolved")),
	)
	s.mcp.AddTool(resolveHookTool, s.handleResolveHook)

	// set_world_flag tool
	setWorldFlagTool := mcp.NewTool("set_world_flag",
		mcp.WithDescription("Set a key/value world state flag. Use 'true'/'false' for booleans."),
		mcp.WithString("campaign_id", mcp.Required(), mcp.Description("Campaign ID")),
		mcp.WithString("key", mcp.Required(), mcp.Description("Flag key")),
		mcp.WithString("value", mcp.Required(), mcp.Description("Flag value")),
	)
	s.mcp.AddTool(setWorldFlagTool, s.handleSetWorldFlag)

	// get_world_flags tool
	getWorldFlagsTool := mcp.NewTool("get_world_flags",
		mcp.WithDescription("Get all world state flags for a campaign."),
		mcp.WithString("campaign_id", mcp.Required(), mcp.Description("Campaign ID")),
	)
	s.mcp.AddTool(getWorldFlagsTool, s.handleGetWorldFlags)

	// === Session Management Tools ===

	// start_session tool
	startSessionTool := mcp.NewTool("start_session",
		mcp.WithDescription("Load campaign state and render a session brief for DM context."),
		mcp.WithString("campaign_id", mcp.Required(), mcp.Description("Campaign ID")),
		mcp.WithNumber("session", mcp.Required(), mcp.Description("Session number to start")),
		mcp.WithNumber("recent_sessions", mcp.Description("How many prior sessions to include (default: 3)")),
	)
	s.mcp.AddTool(startSessionTool, s.handleStartSession)

	// end_session tool
	endSessionTool := mcp.NewTool("end_session",
		mcp.WithDescription("Compress and store end-of-session narrative summary."),
		mcp.WithString("campaign_id", mcp.Required(), mcp.Description("Campaign ID")),
		mcp.WithNumber("session", mcp.Required(), mcp.Description("Session number to end")),
		mcp.WithString("raw_events", mcp.Required(), mcp.Description("Full narrative event log for the session"), mcp.MaxLength(maxSessionRawEventsLength)),
		mcp.WithString("dm_notes", mcp.Description("Optional DM notes"), mcp.MaxLength(maxSessionDMNotesLength)),
	)
	s.mcp.AddTool(endSessionTool, s.handleEndSession)

	// checkpoint tool
	checkpointTool := mcp.NewTool("checkpoint",
		mcp.WithDescription("Save a mid-session checkpoint note."),
		mcp.WithString("campaign_id", mcp.Required(), mcp.Description("Campaign ID")),
		mcp.WithNumber("session", mcp.Required(), mcp.Description("Current session number")),
		mcp.WithString("note", mcp.Required(), mcp.Description("Checkpoint note"), mcp.MaxLength(maxCheckpointNoteLength)),
	)
	s.mcp.AddTool(checkpointTool, s.handleCheckpoint)

	// get_session_brief tool
	getSessionBriefTool := mcp.NewTool("get_session_brief",
		mcp.WithDescription("Get a compact markdown briefing for quick context restoration."),
		mcp.WithString("campaign_id", mcp.Required(), mcp.Description("Campaign ID")),
	)
	s.mcp.AddTool(getSessionBriefTool, s.handleGetSessionBrief)

	// list_sessions tool
	listSessionsTool := mcp.NewTool("list_sessions",
		mcp.WithDescription("List historical sessions with summary previews."),
		mcp.WithString("campaign_id", mcp.Required(), mcp.Description("Campaign ID")),
	)
	s.mcp.AddTool(listSessionsTool, s.handleListSessions)

	// get_npc_relationships tool
	relationshipTool := mcp.NewTool("get_npc_relationships",
		mcp.WithDescription("Query relationship edges involving NPCs. Optionally filter to one NPC name."),
		mcp.WithString("campaign_id", mcp.Required(), mcp.Description("Campaign ID")),
		mcp.WithString("npc_name", mcp.Description("Optional NPC name filter")),
	)
	s.mcp.AddTool(relationshipTool, s.handleGetNPCRelationships)

	// export_session_recap tool
	exportRecapTool := mcp.NewTool("export_session_recap",
		mcp.WithDescription("Export a markdown recap for a campaign across all or selected sessions."),
		mcp.WithString("campaign_id", mcp.Required(), mcp.Description("Campaign ID")),
		mcp.WithNumber("from_session", mcp.Description("Optional lower inclusive session bound")),
		mcp.WithNumber("to_session", mcp.Description("Optional upper inclusive session bound")),
	)
	s.mcp.AddTool(exportRecapTool, s.handleExportSessionRecap)
}
