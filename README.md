# llmdm

A Go MCP server for D&D campaign state management, dice rolling, and session memory.

## What It Does

- Serves a Model Context Protocol (MCP) toolset over `stdio` or streamable HTTP.
- Stores campaign data in per-campaign SQLite databases.
- Tracks characters, plot events/hooks, world flags, roll history, checkpoints, and session summaries.
- Generates session briefs and markdown recaps for fast DM context restoration.
- Compresses end-of-session logs using Anthropic when available, with deterministic local fallback.

## Tech Stack

- Go `1.25.6`
- MCP framework: `github.com/modelcontextprotocol/go-sdk`
- Database: SQLite (`modernc.org/sqlite`)
- Config loading: `.env` via `github.com/joho/godotenv`

## Repository Layout

```text
cmd/server/main.go         # Entry point and server bootstrap
config/config.go           # Environment config parsing and defaults
internal/mcpserver/        # MCP server core, transport, tool registration, and handlers
internal/db/               # SQLite open/migrations and embedded schema
internal/dice/             # Dice parser/roller and log models
internal/memory/           # Data access layer for campaign entities
internal/session/          # Brief rendering, recap rendering, event compression
internal/dm/               # DM system prompt generation
internal/types/            # Shared domain structs
```

## Requirements

- Go `>= 1.25`
- Optional: Anthropic API key for AI session compression

## Configuration

Environment variables (see `.env.example`):

| Variable | Default | Description |
|---|---|---|
| `ANTHROPIC_API_KEY` | _(empty)_ | Optional key for Anthropic summarization in `end_session` |
| `DB_PATH` | `./data/campaigns` | Base directory for campaign database files |
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn`, `error` |
| `MCP_TRANSPORT` | `stdio` | `stdio`, `http`, or `streamable-http` |
| `HTTP_ADDR` | `127.0.0.1:8080` | Bind address for HTTP mode (loopback by default) |
| `MCP_HTTP_ENDPOINT` | `/mcp` | MCP endpoint path in HTTP mode |
| `READ_TIMEOUT` | `15s` | HTTP read timeout (also used for read header timeout) |
| `WRITE_TIMEOUT` | `60s` | HTTP write timeout |
| `IDLE_TIMEOUT` | `60s` | HTTP idle timeout |

### Example `.env`

```bash
cp .env.example .env
```

## Build, Test, Run

```bash
make build
make test
make run
```

Additional targets:

- `make lint` (requires `golangci-lint`)
- `make build-linux` (cross-compile to `bin/dnd-mcp-linux`)
- `make clean`

### Direct Run

```bash
go run ./cmd/server serve
```

Binary usage:

```bash
./bin/dnd-mcp serve
```

## Transport Modes

### 1) `stdio` (default)

Use when launching the server as a local MCP subprocess from an MCP client.

```bash
MCP_TRANSPORT=stdio ./bin/dnd-mcp serve
```

### 2) Streamable HTTP

```bash
MCP_TRANSPORT=http HTTP_ADDR=127.0.0.1:8080 MCP_HTTP_ENDPOINT=/mcp ./bin/dnd-mcp serve
```

Security note: HTTP transport does not provide built-in authentication. Keep `HTTP_ADDR` bound to loopback (for example, `127.0.0.1:8080`) unless you place it behind a trusted authn/authz proxy.

Health endpoint:

```bash
curl http://127.0.0.1:8080/health
```

MCP endpoint:

- `http://127.0.0.1:8080/mcp`

## MCP Client Examples

This repo includes `.mcp.json` for a local HTTP MCP server:

```json
{
  "mcpServers": {
    "dnd-campaign": {
      "type": "http",
      "url": "http://127.0.0.1:8080/mcp"
    }
  }
}
```

### Claude (`stdio` transport)

Use this when configuring Claude to launch the MCP server as a local subprocess:

```json
{
  "mcpServers": {
    "dnd-campaign": {
      "command": "/absolute/path/to/bin/dnd-mcp",
      "args": ["serve"],
      "env": {
        "MCP_TRANSPORT": "stdio"
      }
    }
  }
}
```

## Available MCP Tools

### Dice

- `roll` - Roll dice using standard notation (e.g., "1d20+5", "4d6kh3")
- `roll_contested` - Roll contested checks between attacker and defender
- `roll_saving_throw` - Roll a saving throw against a DC
- `get_roll_history` - Retrieve roll history with optional filters

### Campaign Memory

- `create_campaign` - Create a new campaign database
- `list_campaigns` - List all available campaigns
- `save_character` - Save or replace a character (PC or NPC)
- `update_character` - Partially update character fields
- `get_character` - Retrieve full character sheet
- `list_characters` - List character summaries with optional filters
- `save_plot_event` - Record a narrative event with plot hooks
- `list_open_hooks` - Get all unresolved plot hooks
- `resolve_hook` - Mark a plot hook as resolved
- `set_world_flag` - Set a campaign world flag
- `get_world_flags` - Get all world flags

### Session Management

- `start_session` - Generate a session brief with context
- `end_session` - End session and compress event log
- `checkpoint` - Save mid-session checkpoint with turn data
- `get_turn_history` - Retrieve checkpoints from a session
- `get_session_brief` - Get brief for current session
- `list_sessions` - List all session metadata
- `get_npc_relationships` - Query NPC relationship graph
- `export_session_recap` - Export markdown recap for session range

## Documentation

Comprehensive tool documentation is available in the [`docs/`](./docs/) directory:

- **Tool Reference**: Detailed documentation for all 22 tools
  - [Dice Rolling Tools](./docs/tools/dice-rolling.md) (4 tools)
  - [Campaign Memory Tools](./docs/tools/campaign-memory.md) (11 tools)
  - [Session Management Tools](./docs/tools/session-management.md) (7 tools)

- **Guides**:
  - [Quick Start Guide](./docs/guides/quick-start.md) - Get started in minutes
  - [Character Creation Guide](./docs/guides/character-creation.md) - D&D 5e character sheets
  - [Session Workflow Guide](./docs/guides/session-workflow.md) - Best practices for DMs

- **Examples**: [Real-world tool usage examples](./docs/examples/tool-examples.json)

**For Contributors**: See [CLAUDE.md](./CLAUDE.md) for documentation maintenance guidelines.

## Character Data Model

Characters support full D&D 5e character sheets with the following fields:

### Core Stats
- Name, Type (PC/NPC), Class, Race, Level
- HP (Current/Max)
- Ability Scores (STR, DEX, CON, INT, WIS, CHA)
- Alignment (e.g., "Lawful Good", "Chaotic Neutral")
- Armor Class (AC), Speed
- Experience Points
- Gold

### D&D 5e Features
- **Proficiencies**: Armor, Weapons, Tools, Saving Throws, Skills
- **Skills**: Individual skills with proficiency status and modifiers
- **Languages**: Array of known languages
- **Features**: Racial traits, class features, feats (with descriptions and sources)
- **Spellcasting** (optional, for spellcasters):
  - Spellcasting ability (INT/WIS/CHA)
  - Spell slots by level
  - Known cantrips
  - Prepared spells

### Narrative Fields
- Backstory (up to 8,000 characters)
- Inventory (array of items)
- Conditions (active status effects)
- Plot Flags (narrative state markers)
- Relationships (NPC connections)
- Notes (DM-only, up to 4,000 characters)
- Status (active/dead/missing/retired)

### Defaults
- AC defaults to 10 if not specified
- Speed defaults to "30 ft" if not specified
- Experience Points default to 0
- All array fields initialize to empty arrays (never null)

## Persistence Model

Each campaign is persisted in its own SQLite file:

- `<DB_PATH>/<campaign_id>.db`

Core tables (auto-migrated on access):

- `campaigns` - Campaign metadata
- `characters` - Full D&D 5e character sheets with ability scores, proficiencies, skills, spellcasting, etc.
- `plot_events` - Narrative events by session
- `plot_hooks` - Unresolved story threads
- `world_flags` - Campaign-wide state flags
- `roll_log` - Complete dice roll history
- `sessions` - Session summaries and metadata
- `checkpoints` - Mid-session turn snapshots

Schema source: `internal/db/schema.sql`.

Character data is stored with JSON columns for complex fields (proficiencies, skills, features, spellcasting) and supports both spellcasters and non-spellcasters.

## End-Session Compression Behavior

`end_session` always returns a summary:

- If `ANTHROPIC_API_KEY` is set and the API call succeeds, summary is model-generated.
- If key is missing or API fails, server falls back to deterministic truncation + `OPEN HOOKS` scaffold.

This keeps workflows resilient in offline or degraded network/API conditions.

## Development Notes

- SQLite pragmas are enabled on open: foreign keys, WAL, and busy timeout.
- Logging is structured JSON via `log/slog`.
- Migrations are embedded and run automatically when opening campaign DBs.
- Test suite uses race detector by default via `make test`.

## Troubleshooting

- `invalid MCP_TRANSPORT ...`: set to one of `stdio`, `http`, `streamable-http`.
- `Failed to load config`: verify env values and `.env` formatting.
- Empty/short summary in `end_session`: confirm `raw_events` is non-empty.
- HTTP connection issues: verify `HTTP_ADDR`, `MCP_HTTP_ENDPOINT`, and that client URL matches exactly.
