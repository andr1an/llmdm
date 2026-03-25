# Troubleshooting Guide

Common issues and solutions for the D&D Campaign MCP Server.

## Configuration Issues

### Invalid MCP_TRANSPORT

**Error:**
```
invalid MCP_TRANSPORT value: "xyz"
```

**Cause:**
The `MCP_TRANSPORT` environment variable is set to an invalid value.

**Solution:**
Set `MCP_TRANSPORT` to one of the valid values:
- `stdio` (default)
- `http`
- `streamable-http`

**Example:**
```bash
MCP_TRANSPORT=stdio ./bin/dnd-mcp serve
```

### Failed to Load Config

**Error:**
```
Failed to load configuration
Error parsing .env file
```

**Cause:**
The `.env` file has formatting errors or invalid values.

**Solution:**
1. Check `.env` file formatting (use `.env.example` as reference)
2. Ensure no spaces around `=` in assignments
3. Quote values with spaces: `DB_PATH="/path/with spaces/data"`
4. Verify environment variable names match expected values

**Valid `.env` example:**
```bash
ANTHROPIC_API_KEY=sk-ant-api03-...
DB_PATH=./data/campaigns
LOG_LEVEL=info
MCP_TRANSPORT=stdio
```

### Missing Database Directory

**Error:**
```
Failed to open database: no such file or directory
```

**Cause:**
The `DB_PATH` directory does not exist.

**Solution:**
Create the database directory:

```bash
mkdir -p ./data/campaigns
```

Or set `DB_PATH` to an existing directory:

```bash
DB_PATH=/tmp/campaigns ./bin/dnd-mcp serve
```

## Session Management Issues

### Empty or Short Summaries

**Symptom:**
`end_session` returns very short summaries or just "No events to compress".

**Cause:**
The `raw_events` parameter is empty or contains minimal content.

**Solution:**
1. Ensure `raw_events` contains the full session narrative
2. Check that events were actually recorded during the session
3. Verify `checkpoint` tool was called during gameplay to log events

**Example:**
```json
{
  "campaign_id": "my-campaign",
  "raw_events": "The party explored the dungeon, fought goblins, and found treasure..."
}
```

### Missing ANTHROPIC_API_KEY Behavior

**Symptom:**
Session summaries are generic and don't reflect session details.

**Cause:**
No `ANTHROPIC_API_KEY` is set, so the server falls back to deterministic compression.

**Expected Behavior:**
This is **not an error**. The server is designed to work without an API key by using deterministic compression (truncation + open hooks scaffold).

**To Enable AI Compression:**
1. Get an API key from [Anthropic](https://console.anthropic.com/)
2. Add to `.env` file:
   ```bash
   ANTHROPIC_API_KEY=sk-ant-api03-...
   ```
3. Restart the server

**Fallback Compression:**
When `ANTHROPIC_API_KEY` is not set, `end_session` returns:
- Truncated event log
- List of open plot hooks
- Basic session metadata

This keeps workflows resilient in offline or degraded network conditions.

## HTTP Transport Issues

### Connection Refused

**Error:**
```
Failed to connect to MCP server at http://127.0.0.1:8080/mcp
Connection refused
```

**Cause:**
The server is not running in HTTP mode, or the address/port doesn't match.

**Solution:**
1. Verify server is running:
   ```bash
   MCP_TRANSPORT=http HTTP_ADDR=127.0.0.1:8080 ./bin/dnd-mcp serve
   ```

2. Check server logs for startup confirmation:
   ```
   {"level":"info","msg":"Starting MCP server","transport":"http","addr":"127.0.0.1:8080"}
   ```

3. Verify MCP client URL matches exactly:
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

### Endpoint Mismatch

**Error:**
```
404 Not Found
```

**Cause:**
MCP client URL doesn't match the server's `MCP_HTTP_ENDPOINT`.

**Solution:**
1. Check server's `MCP_HTTP_ENDPOINT` (default: `/mcp`)
2. Ensure client URL includes the correct endpoint:
   - Server: `MCP_HTTP_ENDPOINT=/mcp`
   - Client: `http://127.0.0.1:8080/mcp` ✓
   - Wrong: `http://127.0.0.1:8080` ✗

### Health Check Not Working

**Symptom:**
Health endpoint returns 404.

**Cause:**
Health endpoint is always at `/health`, regardless of `MCP_HTTP_ENDPOINT`.

**Solution:**
Use the correct health endpoint URL:

```bash
curl http://127.0.0.1:8080/health
```

**Not:**
```bash
curl http://127.0.0.1:8080/mcp/health  # Wrong
```

## Character Management Issues

### Character Not Found

**Error:**
```
character not found: <character_id>
```

**Cause:**
Character ID doesn't exist in the campaign database, or you're querying the wrong campaign.

**Solution:**
1. Verify campaign ID is correct
2. List all characters to find the correct ID:
   ```json
   {
     "campaign_id": "my-campaign"
   }
   ```
   Tool: `list_characters`

3. Check for typos in character ID (IDs are case-sensitive)

### Spellcasting Fields Not Saving

**Symptom:**
Spellcasting data is lost after saving a character.

**Cause:**
The `spellcasting` field must be properly structured JSON (or omitted for non-spellcasters).

**Solution:**
For spellcasters, include complete `spellcasting` object:

```json
{
  "spellcasting": {
    "ability": "INT",
    "spell_slots": {
      "1": 4,
      "2": 3
    },
    "cantrips": ["Fire Bolt"],
    "prepared_spells": ["Shield", "Magic Missile"]
  }
}
```

For non-spellcasters, **omit the field entirely** (don't set to `null`).

## Database Issues

### Database Locked

**Error:**
```
database is locked
```

**Cause:**
Another process is accessing the same database file with an exclusive lock.

**Solution:**
1. SQLite uses WAL mode and busy timeout (5 seconds) - this should be rare
2. Check for other processes accessing the database:
   ```bash
   lsof ./data/campaigns/my-campaign.db
   ```
3. Ensure only one MCP server instance accesses each database
4. If stuck, restart the MCP server

### Schema Migration Failed

**Error:**
```
Failed to run migrations
```

**Cause:**
Database file is corrupted or schema version is incompatible.

**Solution:**
1. **Backup the database file first**
2. Check SQLite integrity:
   ```bash
   sqlite3 ./data/campaigns/my-campaign.db "PRAGMA integrity_check;"
   ```
3. If corrupted, restore from backup or recreate the campaign

### Disk Space Issues

**Symptom:**
Database operations fail silently or with "disk full" errors.

**Solution:**
1. Check available disk space:
   ```bash
   df -h
   ```
2. SQLite WAL files can grow large - check for `-wal` files:
   ```bash
   ls -lh ./data/campaigns/*.db-wal
   ```
3. Clean up old campaigns if needed
4. Set `DB_PATH` to a directory with more space

## Performance Issues

### Slow Tool Responses

**Symptom:**
MCP tool calls take several seconds to respond.

**Cause:**
Large database files, slow disk I/O, or complex queries.

**Solution:**
1. Check database file sizes:
   ```bash
   du -sh ./data/campaigns/*.db
   ```
2. Optimize SQLite:
   ```bash
   sqlite3 ./data/campaigns/my-campaign.db "VACUUM;"
   ```
3. Review query patterns in logs (`LOG_LEVEL=debug`)
4. Consider archiving old sessions

### High Memory Usage

**Symptom:**
Server process uses excessive memory.

**Cause:**
Large `raw_events` strings in `end_session` or memory leaks.

**Solution:**
1. Monitor memory with `top` or `htop`
2. Limit `raw_events` size (recommend < 100KB)
3. Restart server periodically if running long-term
4. Check for goroutine leaks with pprof (development)

## Logging and Debugging

### Enable Debug Logging

To see detailed request/response logs:

```bash
LOG_LEVEL=debug ./bin/dnd-mcp serve
```

### View Logs in JSON Format

Logs are structured JSON by default:

```bash
./bin/dnd-mcp serve 2>&1 | jq
```

### Check Server Version

View version information:

```bash
./bin/dnd-mcp --version
```

## Getting Help

### Check Documentation

- [Quick Start Guide](./quick-start.md)
- [Deployment Guide](./deployment.md)
- [Tool Reference](../tools/)
- [Architecture Guide](./architecture.md)

### Verify Setup

1. **Binary built successfully**: `./bin/dnd-mcp --help`
2. **Configuration valid**: Check `.env` file
3. **Database directory exists**: `ls -la ./data/campaigns/`
4. **Transport mode correct**: `echo $MCP_TRANSPORT`

### Common Checklist

- [ ] Server is running (check process with `ps aux | grep dnd-mcp`)
- [ ] MCP client configuration matches transport mode (stdio vs HTTP)
- [ ] Database directory exists and is writable
- [ ] `.env` file has correct formatting (no spaces around `=`)
- [ ] URLs/endpoints match exactly (HTTP mode)
- [ ] Campaign ID exists (use `list_campaigns` to verify)

### File an Issue

If you've tried the solutions above and still have problems:

1. Enable debug logging: `LOG_LEVEL=debug`
2. Reproduce the issue
3. Collect relevant logs
4. File an issue at the repository with:
   - MCP server version
   - Transport mode
   - Configuration (sanitized - remove API keys)
   - Error messages
   - Steps to reproduce

## Next Steps

- [Quick Start Guide](./quick-start.md) - Get started with your first campaign
- [Deployment Guide](./deployment.md) - Server configuration options
- [Session Workflow Guide](./session-workflow.md) - Best practices for DMs
