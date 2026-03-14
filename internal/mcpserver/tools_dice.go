package mcpserver

import "github.com/modelcontextprotocol/go-sdk/mcp"

// registerDiceTools registers all dice-rolling tools with the MCP server.
func (s *Server) registerDiceTools() {
	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "roll",
		Description: "Roll dice using standard D&D notation. Supports NdM, NdM+K, NdM-K, NdMkhX (keep highest), NdMklX (keep lowest). All rolls are logged.",
	}, s.handleRoll)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "roll_contested",
		Description: "Roll a contested check between attacker and defender. Both rolls are logged.",
	}, s.handleRollContested)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "roll_saving_throw",
		Description: "Roll a saving throw and check against a DC.",
	}, s.handleRollSavingThrow)

	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "get_roll_history",
		Description: "Retrieve dice roll history for a campaign.",
	}, s.handleGetRollHistory)
}
