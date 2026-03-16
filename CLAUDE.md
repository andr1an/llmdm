# Instructions for AI Assistants

## CRITICAL RULE: Tool Changes Require Documentation Updates

**IMPORTANT**: Every time you modify, add, or remove a tool, you MUST update the corresponding documentation.

### When Changing Tools, Update:

1. `/docs/tools/[category].md` - Tool reference documentation
2. `/docs/examples/tool-examples.json` - Usage examples
3. `/docs/README.md` - If adding/removing tools, update counts
4. This file (CLAUDE.md) - If documentation structure changes

### Tool Categories and Their Files

- **Dice Rolling** (4 tools) → `/docs/tools/dice-rolling.md`
  - Code: `internal/mcpserver/tools_dice.go`, `handlers_dice.go`
  - Tools: `roll`, `roll_contested`, `roll_saving_throw`, `get_roll_history`

- **Campaign Memory** (11 tools) → `/docs/tools/campaign-memory.md`
  - Code: `internal/mcpserver/tools_memory.go`, `handlers_memory.go`
  - Tools: `create_campaign`, `list_campaigns`, `save_character`, `update_character`, `get_character`, `list_characters`, `save_plot_event`, `list_open_hooks`, `resolve_hook`, `set_world_flag`, `get_world_flags`

- **Session Management** (7 tools) → `/docs/tools/session-management.md`
  - Code: `internal/mcpserver/tools_session.go`, `handlers_session.go`
  - Tools: `start_session`, `end_session`, `checkpoint`, `get_turn_history`, `get_session_brief`, `list_sessions`, `get_npc_relationships`, `export_session_recap`

**Total**: 22 tools

## Documentation Format Standards

Each tool must be documented with:

1. **Description**: From tool registration in `tools_*.go`
2. **Input Parameters**: Table with name, type, required, description, example
3. **Input Schema**: JSON example with all parameters
4. **Output Schema**: JSON example showing structure
5. **Notes**: Special behaviors, validation, side effects
6. **See Also**: Related tools

### Parameter Documentation Format

```markdown
| Parameter | Type | Required | Description | Example |
|-----------|------|----------|-------------|---------|
| `param_name` | string | Yes | What it does | `"example_value"` |
```

### Input Schema Format

```json
{
  "param1": "value1",
  "param2": 123,
  "param3": true
}
```

### Output Schema Format

```json
{
  "result": {
    "field1": "value1",
    "field2": 123
  }
}
```

## Validation Rules to Document

When documenting tools, always include these validation details:

### String Length Constraints

- `backstory`: 8000 characters max
- `notes`: 4000 characters max
- `summary`: 2-4 sentences (guideline, not enforced)

### Enum Values

Document all valid enum values:

- **type**: `"pc"`, `"npc"`
- **status**: `"active"`, `"dead"`, `"missing"`, `"retired"`
- **stat**: `"STR"`, `"DEX"`, `"CON"`, `"INT"`, `"WIS"`, `"CHA"`
- **alignment**: D&D 5e alignments (9 total)

### Required vs Optional Fields

Check `omitempty` in `internal/mcpserver/structs.go` to determine if fields are optional.

**Required fields** (no `omitempty`):
- `campaign_id`, `name`, `type`, `hp_current`, `hp_max` (in `save_character`)

**Optional fields** (with `omitempty`):
- Most other character fields (class, race, level, stats, etc.)

### Default Values

Document defaults clearly:

- `level`: 1
- `ac`: 10
- `speed`: `"30 ft"`
- `gold`: 0
- `experience_points`: 0

## When Adding a New Tool

Follow this checklist:

1. **Code Changes**:
   - Add tool to `internal/mcpserver/tools_*.go`
   - Add handler to `internal/mcpserver/handlers_*.go`
   - Add input/output structs to `internal/mcpserver/structs.go` with `jsonschema` tags
   - Update `internal/types/types.go` if new domain types needed

2. **Documentation Updates**:
   - Add tool to `/docs/tools/[category].md` with full documentation
   - Add 2-3 examples to `/docs/examples/tool-examples.json`
   - Update tool count in `/docs/README.md`
   - Update tool list in this file (CLAUDE.md)

3. **Testing**:
   - Test tool with MCP client
   - Verify input validation
   - Verify output format matches documentation

## When Modifying a Tool

1. **Update Code**:
   - Modify tool registration, handler, or structs
   - Update `jsonschema` tags if parameter descriptions change

2. **Update Documentation**:
   - Update `/docs/tools/[category].md` to match new behavior
   - Update examples in `/docs/examples/tool-examples.json` if signature changed
   - Note breaking changes prominently in docs

3. **Version Consideration**:
   - If breaking change, consider versioning or migration path
   - Document migration in changelog

## When Removing a Tool

1. **Deprecation First** (if possible):
   - Mark as deprecated in tool description
   - Note in `/docs/tools/[category].md` with deprecation warning
   - Keep tool functional for at least one release

2. **Removal**:
   - Remove from code (`tools_*.go`, `handlers_*.go`, `structs.go`)
   - Remove from or mark as removed in `/docs/tools/[category].md`
   - Remove examples from `/docs/examples/tool-examples.json`
   - Update tool counts in `/docs/README.md` and this file
   - Add migration guide if tool had users

## Documentation Maintenance Checklist

Before committing tool changes:

- [ ] Tool code updated in `tools_*.go` and `handlers_*.go`
- [ ] Input/output structs updated in `structs.go` with `jsonschema` tags
- [ ] Documentation updated in `/docs/tools/[category].md`
- [ ] Examples updated in `/docs/examples/tool-examples.json`
- [ ] Tool counts verified in `/docs/README.md`
- [ ] Tool list updated in this file if tool added/removed
- [ ] Breaking changes noted prominently
- [ ] All tests pass

## Project Context

This is an MCP (Model Context Protocol) server for D&D 5e campaign management with:

- **Persistence**: SQLite databases (one per campaign)
- **Character Sheets**: Full D&D 5e support with stats, proficiencies, skills, features, spellcasting
- **Session Management**: AI-powered compression using Anthropic API (optional)
- **Dice Mechanics**: D20 system with advantage/disadvantage, contested rolls, saving throws
- **Plot Tracking**: Events, hooks, world flags

### Key Files

- **Domain Types**: `internal/types/types.go` - All domain structs (Character, Campaign, RollResult, etc.)
- **Database Schema**: `internal/db/schema.sql` - SQLite table definitions
- **Tool Structs**: `internal/mcpserver/structs.go` - All tool input/output structs with `jsonschema` tags
- **Tool Registrations**: `internal/mcpserver/tools_*.go` - Tool definitions with descriptions
- **Tool Handlers**: `internal/mcpserver/handlers_*.go` - Tool implementation logic

### Architecture

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

### Tool Design Principles

1. **Explicit over implicit**: Require campaign_id, don't assume context
2. **Structured data**: Use structs with jsonschema tags for validation
3. **Idempotency**: Where possible, tools should be safe to re-run
4. **Clear errors**: Return meaningful error messages with context
5. **Logging**: All rolls logged, all events timestamped

## Documentation Style Guide

### Tone

- Professional and clear
- Tutorial-style for guides
- Reference-style for tool docs
- Example-driven

### Formatting

- Use tables for parameters
- Use code blocks for JSON
- Use markdown headers consistently (## for tools, ### for subsections)
- Use **bold** for emphasis on critical points
- Use `code` for field names, tool names, values

### Examples

- Always include realistic examples
- Show both success and failure cases where relevant
- Use consistent campaign_id across examples: `"camp_abc123"`
- Use consistent character names: Thorin, Gandalf, Bilbo, etc.

### Cross-References

- Link related tools with `[tool_name](./category.md#tool_name)`
- Link guides from tool docs: `[Guide Name](../guides/guide-name.md)`
- Always use relative paths for portability

## Common Mistakes to Avoid

1. **Don't forget `jsonschema` tags**: All struct fields in `structs.go` need jsonschema descriptions
2. **Don't skip examples**: Every new tool needs at least 2 examples
3. **Don't ignore counts**: Update tool counts when adding/removing tools
4. **Don't break links**: Verify all internal documentation links work
5. **Don't assume knowledge**: Document all enums, defaults, and constraints

## Getting Help

- **MCP SDK**: https://github.com/modelcontextprotocol/go-sdk
- **D&D 5e SRD**: https://www.dndbeyond.com/sources/basic-rules
- **SQLite Docs**: https://www.sqlite.org/docs.html

## Version Information

**Current Version**: 1.0.0
**MCP SDK**: go-sdk v1.4.1
**Go Version**: 1.21+
**D&D Edition**: 5th Edition

## Changelog

When making changes, update CHANGELOG.md (if it exists) or note breaking changes in commit messages.

---

**Remember**: Documentation is code. Keep it accurate, keep it updated, keep it useful.
