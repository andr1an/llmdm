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

- `roll`
- `roll_contested`
- `roll_saving_throw`
- `get_roll_history`

### Campaign Memory

- `create_campaign`
- `list_campaigns`
- `save_character`
- `update_character`
- `get_character`
- `list_characters`
- `save_plot_event`
- `list_open_hooks`
- `resolve_hook`
- `set_world_flag`
- `get_world_flags`

### Session Management

- `start_session`
- `end_session`
- `checkpoint`
- `get_turn_history`
- `get_session_brief`
- `list_sessions`
- `get_npc_relationships`
- `export_session_recap`

## Persistence Model

Each campaign is persisted in its own SQLite file:

- `<DB_PATH>/<campaign_id>.db`

Core tables (auto-migrated on access):

- `campaigns`
- `characters`
- `plot_events`
- `plot_hooks`
- `world_flags`
- `roll_log`
- `sessions`
- `checkpoints`

Schema source: `internal/db/schema.sql`.

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
