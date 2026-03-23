# Instructions for AI Assistants

## Documentation Structure

All user-facing documentation lives in `/docs/`:

- **Tool Reference**: `/docs/tools/` - Complete documentation for all 22 tools
- **Guides**: `/docs/guides/` - Quick start, character creation, session workflow
- **Examples**: `/docs/examples/tool-examples.json` - Real-world usage examples
- **Overview**: `/docs/README.md` - Documentation hub with tool counts and links

**For documentation format, style, and examples**: Follow the patterns in existing `/docs/` files.

## CRITICAL RULE: Keep Code and Docs in Sync

**Every tool change requires documentation updates.** Before committing:

- [ ] Code updated: `tools_*.go`, `handlers_*.go`, `structs.go` (with `jsonschema` tags)
- [ ] Docs updated: `/docs/tools/[category].md` (follow existing format)
- [ ] Examples updated: `/docs/examples/tool-examples.json` (add 2-3 examples)
- [ ] Counts updated: `/docs/README.md` (if tools added/removed)

## Tool Categories and Code Locations

| Category | Docs | Code |
|----------|------|------|
| **Dice Rolling** (4 tools) | `/docs/tools/dice-rolling.md` | `tools_dice.go`, `handlers_dice.go` |
| **Campaign Memory** (11 tools) | `/docs/tools/campaign-memory.md` | `tools_memory.go`, `handlers_memory.go` |
| **Session Management** (7 tools) | `/docs/tools/session-management.md` | `tools_session.go`, `handlers_session.go` |

**Total**: 22 tools

## Key Code Files

- `internal/mcpserver/tools_*.go` - Tool registrations with descriptions
- `internal/mcpserver/handlers_*.go` - Tool implementation logic
- `internal/mcpserver/structs.go` - Tool input/output structs with `jsonschema` tags
- `internal/types/types.go` - Domain structs (Character, Campaign, etc.)
- `internal/db/schema.sql` - SQLite table definitions

## Project Architecture

```
MCP Client (Claude, etc.)
    ↓
MCP Server (this project)
    ↓
Tool Router (by category)
    ↓
Handler Functions
    ↓
Database Layer (SQLite)
```

**Design Principles**:
- Explicit over implicit (always require `campaign_id`)
- Structured data with `jsonschema` validation
- Clear error messages with context
- All rolls and events are logged

---

**Remember**: Documentation is code. Keep it in sync with every tool change.
