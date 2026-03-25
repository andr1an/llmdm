# Deployment Guide

This guide covers server deployment, transport configuration, and MCP client setup.

## Transport Modes

The server supports three transport modes controlled by the `MCP_TRANSPORT` environment variable:

### 1. stdio (Default)

Use when launching the server as a local MCP subprocess from an MCP client.

```bash
MCP_TRANSPORT=stdio ./bin/dnd-mcp serve
```

**When to use:**
- Running with Claude Code or other MCP clients that launch local processes
- Development and testing
- Single-user local setups

**Characteristics:**
- Communication over standard input/output
- No network ports required
- Process lifecycle managed by MCP client
- Most common deployment mode

### 2. HTTP

Run the server as a standalone HTTP service.

```bash
MCP_TRANSPORT=http HTTP_ADDR=127.0.0.1:8080 MCP_HTTP_ENDPOINT=/mcp ./bin/dnd-mcp serve
```

**When to use:**
- Remote MCP client connections
- Service-based deployments
- Testing with multiple clients

**Endpoints:**
- MCP endpoint: `http://127.0.0.1:8080/mcp` (configurable via `MCP_HTTP_ENDPOINT`)
- Health check: `http://127.0.0.1:8080/health`

**Security Note:**
HTTP transport does not provide built-in authentication. Keep `HTTP_ADDR` bound to loopback (e.g., `127.0.0.1:8080`) unless you place it behind a trusted authentication/authorization proxy.

**Health Check Example:**
```bash
curl http://127.0.0.1:8080/health
```

### 3. Streamable HTTP

Alias for `http` mode - same behavior as HTTP transport.

```bash
MCP_TRANSPORT=streamable-http ./bin/dnd-mcp serve
```

## MCP Client Configuration

### Claude Code (stdio transport)

For local subprocess execution with Claude Code:

**1. Create `.mcp.json` in your project directory:**

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

**2. Start Claude Code:**

```bash
cd your-project
claude-code
```

The `game/` directory in this repository includes a pre-configured `.mcp.json` for the stdio transport.

### Claude Code (HTTP transport)

For connecting to a running HTTP MCP server:

**1. Start the server in a separate terminal:**

```bash
MCP_TRANSPORT=http HTTP_ADDR=127.0.0.1:8080 ./bin/dnd-mcp serve
```

**2. Create `.mcp.json` in your project directory:**

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

**3. Start Claude Code:**

```bash
cd your-project
claude-code
```

### Other MCP Clients

Consult your MCP client's documentation for configuration details. Generally, you'll need:

- **stdio mode**: Command path, args, and environment variables
- **HTTP mode**: MCP endpoint URL

## Environment Variables

Configure the server using environment variables or a `.env` file:

| Variable | Default | Description |
|---|---|---|
| `ANTHROPIC_API_KEY` | _(empty)_ | Optional key for Anthropic summarization in `end_session`. If not set, server falls back to deterministic compression. |
| `DB_PATH` | `./data/campaigns` | Base directory for campaign database files. Each campaign gets its own SQLite file: `<DB_PATH>/<campaign_id>.db` |
| `LOG_LEVEL` | `info` | Logging verbosity: `debug`, `info`, `warn`, `error` |
| `MCP_TRANSPORT` | `stdio` | Transport mode: `stdio`, `http`, or `streamable-http` |
| `HTTP_ADDR` | `127.0.0.1:8080` | Bind address for HTTP mode. Use loopback (`127.0.0.1`) for security unless behind auth proxy. |
| `MCP_HTTP_ENDPOINT` | `/mcp` | MCP endpoint path in HTTP mode. Health check is always at `/health`. |
| `READ_TIMEOUT` | `15s` | HTTP read timeout (also used for read header timeout) |
| `WRITE_TIMEOUT` | `60s` | HTTP write timeout |
| `IDLE_TIMEOUT` | `60s` | HTTP idle timeout |

### Example `.env` File

Create a `.env` file in the same directory as your binary:

```bash
# AI Compression (optional)
ANTHROPIC_API_KEY=sk-ant-api03-...

# Database Location
DB_PATH=./data/campaigns

# Logging
LOG_LEVEL=info

# Transport Configuration
MCP_TRANSPORT=stdio

# HTTP Configuration (only used if MCP_TRANSPORT=http)
HTTP_ADDR=127.0.0.1:8080
MCP_HTTP_ENDPOINT=/mcp
READ_TIMEOUT=15s
WRITE_TIMEOUT=60s
IDLE_TIMEOUT=60s
```

**Tip:** Copy `.env.example` from the repository:

```bash
cp .env.example .env
# Edit .env with your values
```

## Build and Run

### Build Binary

```bash
make build
```

This creates `bin/dnd-mcp`.

### Run Binary

```bash
./bin/dnd-mcp serve
```

Or with environment variables:

```bash
MCP_TRANSPORT=http HTTP_ADDR=127.0.0.1:8080 ./bin/dnd-mcp serve
```

### Direct Go Run

```bash
go run ./cmd/server serve
```

### Additional Make Targets

```bash
make test          # Run tests with race detector
make lint          # Run golangci-lint (requires golangci-lint installed)
make build-linux   # Cross-compile for Linux (output: bin/dnd-mcp-linux)
make clean         # Remove bin/ directory
```

## Production Deployment

### Recommendations

1. **Use HTTP transport** with a reverse proxy (nginx, caddy) for authentication
2. **Set `DB_PATH`** to a persistent volume or directory with backups
3. **Set `LOG_LEVEL=warn`** or `LOG_LEVEL=error` to reduce log volume
4. **Monitor `/health` endpoint** for availability checks
5. **Set appropriate timeouts** based on expected session lengths

### Systemd Service Example

```ini
[Unit]
Description=D&D Campaign MCP Server
After=network.target

[Service]
Type=simple
User=dnd
WorkingDirectory=/opt/dnd-mcp
EnvironmentFile=/opt/dnd-mcp/.env
ExecStart=/opt/dnd-mcp/bin/dnd-mcp serve
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

### Docker Example

```dockerfile
FROM golang:1.25.6-alpine AS builder
WORKDIR /build
COPY . .
RUN go build -o dnd-mcp ./cmd/server

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /build/dnd-mcp .
ENV MCP_TRANSPORT=http
ENV HTTP_ADDR=0.0.0.0:8080
EXPOSE 8080
CMD ["./dnd-mcp", "serve"]
```

## Requirements

- **Go**: `>= 1.25` (for building from source)
- **SQLite**: Embedded via `modernc.org/sqlite` - no separate installation needed
- **Anthropic API Key**: Optional - only required for AI-powered session compression

## Next Steps

- [Quick Start Guide](./quick-start.md) - Get started with your first campaign
- [Session Workflow Guide](./session-workflow.md) - Best practices for running sessions
- [Troubleshooting Guide](./troubleshooting.md) - Common issues and solutions
