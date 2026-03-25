# llmdm

A Go MCP server for D&D campaign state management, dice rolling, and session memory.

## What It Does

- Serves Model Context Protocol (MCP) toolset over stdio or HTTP
- Stores campaign data in per-campaign SQLite databases
- Tracks characters, plot hooks, world flags, roll history, and session summaries
- Generates session briefs and markdown recaps for fast DM context restoration
- Compresses session logs using Anthropic API with local fallback

## How to Play

The `game/` directory contains a ready-to-play workspace for running D&D campaigns.

### Setup

1. Build and copy the server:
   ```bash
   make build
   cp bin/dnd-mcp game/dnd-mcp
   ```

2. (Optional) Set up AI session compression:
   ```bash
   cd game
   echo "ANTHROPIC_API_KEY=your-key-here" > .env
   ```

### Running Campaigns

1. Navigate to game directory and start Claude Code:
   ```bash
   cd game
   claude
   ```

2. Start a new campaign or continue existing:
   ```
   /create-campaign
   /continue
   ```

3. Play! Claude acts as DM with:
   - Real dice rolls via MCP server
   - Persistent character sheets in SQLite
   - AI-compressed session history
   - Automatic plot hook tracking

4. End your session:
   ```
   /end
   ```

### Model Recommendations

- **Sonnet 4.5**: Best for regular play (creativity + speed + cost)
- **Opus 4.5**: Use for major story moments and boss fights
- **Haiku**: Quick queries only (not recommended for gameplay)

### Context Window Management

Campaign state persists in SQLite, so start fresh Claude Code sessions without losing progress.

**When to start fresh:**

- After 2-3 D&D sessions (50k-150k tokens each)
- When responses become slow or repetitive
- At natural story break points
- When context exceeds ~150k tokens

**Resume after refresh:**

```bash
exit
claude
/continue
```

Campaign data is stored in `./data/campaigns/` - never needs cleaning.

## Documentation

Complete documentation in [`docs/`](./docs/):

- [Quick Start Guide](./docs/guides/quick-start.md) - Step-by-step tutorial
- [Tool Reference](./docs/tools/) - All 22 MCP tools documented
- [Deployment Guide](./docs/guides/deployment.md) - Server setup & configuration
- [Architecture Guide](./docs/guides/architecture.md) - Technical details
- [Troubleshooting](./docs/guides/troubleshooting.md) - Common issues & solutions

## Tech Stack

- Go 1.25.6 | MCP SDK: `github.com/modelcontextprotocol/go-sdk`
- SQLite: `modernc.org/sqlite` | Config: `.env` via `github.com/joho/godotenv`

## Build & Run

```bash
make build
make test
make run
```

See [Deployment Guide](./docs/guides/deployment.md) for transport modes and configuration.

## License

MIT License
