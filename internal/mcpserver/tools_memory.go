package mcpserver

import "github.com/modelcontextprotocol/go-sdk/mcp"

// registerMemoryTools registers all campaign memory tools with the MCP server.
func (s *Server) registerMemoryTools() {
	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "create_campaign",
		Description: "Create a new D&D campaign. Returns the campaign ID and database path.",
	}, s.handleCreateCampaign)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "list_campaigns",
		Description: "List all existing campaigns. Returns campaign metadata including ID, name, description, and current session.",
	}, s.handleListCampaigns)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "save_character",
		Description: "Create or fully replace a character (PC or NPC). Use update_character for partial updates.",
	}, s.handleSaveCharacter)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "update_character",
		Description: "Patch specific fields on a character. Only provided fields are updated.",
	}, s.handleUpdateCharacter)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "get_character",
		Description: "Get full character sheet by name.",
	}, s.handleGetCharacter)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "list_characters",
		Description: "List characters in a campaign, optionally filtered by type and status.",
	}, s.handleListCharacters)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "save_plot_event",
		Description: "Record a narrative event in the campaign. Creates plot hooks if provided.",
	}, s.handleSavePlotEvent)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "list_open_hooks",
		Description: "List all unresolved plot threads for a campaign.",
	}, s.handleListOpenHooks)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "resolve_hook",
		Description: "Mark a plot hook as resolved.",
	}, s.handleResolveHook)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "set_world_flag",
		Description: "Set a key/value world state flag. Use 'true'/'false' for booleans.",
	}, s.handleSetWorldFlag)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "get_world_flags",
		Description: "Get all world state flags for a campaign.",
	}, s.handleGetWorldFlags)
}
