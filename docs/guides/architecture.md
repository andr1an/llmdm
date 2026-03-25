# Architecture & Persistence

This guide covers the technical architecture, code organization, and database persistence model.

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

### Key Directories

**`cmd/server/`**
- Application entry point
- Server initialization and configuration loading
- CLI command handling (`serve` command)

**`config/`**
- Environment variable parsing
- Configuration defaults
- `.env` file loading via `godotenv`

**`internal/mcpserver/`**
- MCP protocol implementation
- Transport layer (stdio, HTTP)
- Tool registration and routing
- Handler functions for all 22 tools
- Request/response serialization

**`internal/db/`**
- Database connection management
- SQLite pragma configuration (WAL, foreign keys, busy timeout)
- Embedded schema migrations (`schema.sql`)
- Automatic migration on database open

**`internal/dice/`**
- Dice notation parser (supports d20, advantage/disadvantage, modifiers)
- Dice roller with random number generation
- Roll history logging models

**`internal/memory/`**
- Data access layer for campaigns, characters, plot events, hooks, world flags
- CRUD operations with SQLite
- Transaction management
- JSON column handling for complex fields

**`internal/session/`**
- Session brief rendering (markdown format)
- Session recap export
- Event compression (AI-powered via Anthropic API, deterministic fallback)
- Checkpoint management

**`internal/dm/`**
- DM system prompt generation
- Context assembly for session briefs
- Campaign state serialization for AI context

**`internal/types/`**
- Shared domain structs (Campaign, Character, PlotEvent, etc.)
- D&D 5e data models (ability scores, skills, proficiencies, spellcasting)
- Common types used across packages

## Database Persistence Model

### Database Files

Each campaign is stored in its own SQLite database file:

```
<DB_PATH>/<campaign_id>.db
```

**Default DB_PATH**: `./data/campaigns/`

**Example:**
- Campaign ID: `lost-mines-2026`
- Database file: `./data/campaigns/lost-mines-2026.db`

### Schema Tables

The database schema is defined in `internal/db/schema.sql` and includes:

#### `campaigns`
Campaign metadata including name, description, setting, and creation timestamp.

#### `characters`
Full D&D 5e character sheets with:
- Core stats: name, type (PC/NPC), class, race, level, HP, ability scores, AC, speed
- Proficiencies: armor, weapons, tools, saving throws
- Skills: individual skills with proficiency status
- Features: racial traits, class features, feats (with descriptions and sources)
- Spellcasting: ability, spell slots, cantrips, prepared spells (optional, stored as JSON)
- Narrative fields: backstory, inventory, conditions, plot flags, relationships, notes
- Status: active/dead/missing/retired

Complex fields (proficiencies, skills, features, spellcasting) are stored as JSON columns for flexibility.

#### `plot_events`
Narrative events by session with session number, timestamp, and event description.

#### `plot_hooks`
Unresolved story threads with:
- Hook description
- Session created
- Resolved status
- Resolution timestamp (if resolved)

#### `world_flags`
Campaign-wide key/value state flags for tracking world state (e.g., `"dragon_slain": "true"`, `"king_attitude": "friendly"`).

#### `roll_log`
Complete dice roll history with:
- Roll notation (e.g., `"1d20+5"`)
- Result value
- Character name (optional)
- Context/description
- Session number
- Timestamp

#### `sessions`
Session metadata and AI-compressed summaries:
- Session number
- Start/end timestamps
- Raw event log (turn-by-turn narrative)
- Compressed summary (AI-generated or deterministic fallback)

#### `checkpoints`
Mid-session turn snapshots for turn-by-turn reconstruction:
- Session number
- Turn number
- Checkpoint data (JSON with state snapshot)
- Timestamp

### JSON Columns

Complex character fields are stored as JSON for flexibility:

**Proficiencies:**
```json
{
  "armor": ["Light", "Medium"],
  "weapons": ["Simple", "Martial"],
  "tools": ["Thieves' Tools"],
  "saving_throws": ["DEX", "INT"],
  "skills": ["Stealth", "Investigation"]
}
```

**Features:**
```json
[
  {
    "name": "Darkvision",
    "description": "You can see in dim light within 60 feet...",
    "source": "Elf racial trait"
  }
]
```

**Spellcasting (optional):**
```json
{
  "ability": "INT",
  "spell_slots": {
    "1": 4,
    "2": 3,
    "3": 2
  },
  "cantrips": ["Fire Bolt", "Mage Hand"],
  "prepared_spells": ["Shield", "Magic Missile", "Fireball"]
}
```

### SQLite Configuration

The server configures SQLite with optimized pragmas on database open:

**Foreign Keys:**
```sql
PRAGMA foreign_keys = ON;
```
Ensures referential integrity across tables.

**WAL (Write-Ahead Logging):**
```sql
PRAGMA journal_mode = WAL;
```
Improves concurrency and write performance.

**Busy Timeout:**
```sql
PRAGMA busy_timeout = 5000;
```
Prevents lock contention errors during concurrent access (5 second timeout).

### Schema Migrations

Migrations are:
- **Embedded**: `schema.sql` is compiled into the binary via `go:embed`
- **Automatic**: Run on first database open or when schema version changes
- **Idempotent**: Safe to run multiple times (uses `IF NOT EXISTS`)
- **Versioned**: Schema version tracked in database metadata

Migration process:
1. Open database connection
2. Check current schema version
3. Apply missing migrations if needed
4. Update schema version

## Development Practices

### Logging

Structured JSON logging via `log/slog`:

```go
slog.Info("Campaign created",
    "campaign_id", campaignID,
    "name", name,
)
```

**Log levels**: `debug`, `info`, `warn`, `error`

Set via `LOG_LEVEL` environment variable.

### Testing

**Test suite uses:**
- Race detector (`go test -race`)
- Table-driven tests for dice notation parsing
- Temporary SQLite databases for integration tests
- Isolated test fixtures

**Run tests:**
```bash
make test
```

**Coverage:**
```bash
go test -cover ./...
```

### Error Handling

**Principles:**
- Return errors with context (use `fmt.Errorf` with `%w` for wrapping)
- Include campaign_id and character_id in error messages for debugging
- Log errors at appropriate levels before returning
- Provide actionable error messages to MCP clients

**Example:**
```go
if err := db.SaveCharacter(char); err != nil {
    slog.Error("Failed to save character",
        "campaign_id", campaignID,
        "character_id", char.ID,
        "error", err,
    )
    return fmt.Errorf("save character %s: %w", char.ID, err)
}
```

### Code Organization

**Separation of concerns:**
- **Transport layer** (`mcpserver/`): MCP protocol, JSON-RPC handling
- **Business logic** (`memory/`, `session/`, `dice/`): Core functionality
- **Data layer** (`db/`): Database operations
- **Types** (`types/`): Shared data structures

**Dependency flow:**
```
Transport → Handlers → Business Logic → Data Layer → Database
```

### Adding New Tools

When adding a new MCP tool, update:

1. **Code**: `internal/mcpserver/tools_*.go` - Register tool with description
2. **Code**: `internal/mcpserver/handlers_*.go` - Implement handler function
3. **Code**: `internal/mcpserver/structs.go` - Add input/output structs with `jsonschema` tags
4. **Docs**: `/docs/tools/[category].md` - Document tool usage
5. **Examples**: `/docs/examples/tool-examples.json` - Add 2-3 usage examples
6. **Counts**: `/docs/README.md` - Update tool count if category changes

See [CLAUDE.md](../../CLAUDE.md) for the critical rule: **Every tool change requires documentation updates.**

## Tech Stack

- **Go**: `1.25.6`
- **MCP SDK**: `github.com/modelcontextprotocol/go-sdk`
- **SQLite**: `modernc.org/sqlite` (pure Go, no CGO)
- **Config**: `github.com/joho/godotenv` for `.env` loading
- **Logging**: `log/slog` (standard library)
- **AI Compression**: Anthropic API (optional)

## Next Steps

- [Deployment Guide](./deployment.md) - Server configuration and client setup
- [Session Workflow Guide](./session-workflow.md) - Best practices for DMs
- [Tool Reference](../tools/) - Complete tool documentation
