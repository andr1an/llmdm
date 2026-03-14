package mcpserver

import "github.com/modelcontextprotocol/go-sdk/mcp"

// registerSessionTools registers all session management tools with the MCP server.
func (s *Server) registerSessionTools() {
	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "start_session",
		Description: "Load campaign state and render a session brief for DM context.",
	}, s.handleStartSession)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "end_session",
		Description: "Compress and store end-of-session narrative summary.",
	}, s.handleEndSession)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "checkpoint",
		Description: "Save a mid-session checkpoint note.",
	}, s.handleCheckpoint)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "get_turn_history",
		Description: "Retrieve turn history from checkpoints for a session.",
	}, s.handleGetTurnHistory)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "get_session_brief",
		Description: "Get a compact markdown briefing for quick context restoration.",
	}, s.handleGetSessionBrief)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "list_sessions",
		Description: "List historical sessions with summary previews.",
	}, s.handleListSessions)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "get_npc_relationships",
		Description: "Query relationship edges involving NPCs. Optionally filter to one NPC name.",
	}, s.handleGetNPCRelationships)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "export_session_recap",
		Description: "Export a markdown recap for a campaign across all or selected sessions.",
	}, s.handleExportSessionRecap)
}
