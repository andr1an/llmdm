# D&D Campaign MCP Server Documentation

Welcome to the comprehensive documentation for the D&D Campaign Management MCP Server.

## Overview

This MCP server provides AI assistants with tools to manage D&D 5e campaigns, including:

- **Dice Rolling**: D20 mechanics with advantage/disadvantage
- **Campaign Memory**: Full character sheets, plot tracking, world state
- **Session Management**: Session briefs, checkpoints, AI-compressed summaries

All campaign data is persisted in SQLite databases (one per campaign).

## Tool Reference

### Dice Rolling Tools (4 tools)

Complete dice rolling system with logging and D&D mechanics.

[View Dice Rolling Tools Documentation →](./tools/dice-rolling.md)

- `roll` - Standard dice notation with advantage/disadvantage
- `roll_contested` - Contested checks (attack vs AC, skill vs skill)
- `roll_saving_throw` - Saving throws with DC checks
- `get_roll_history` - Query roll history with filters

### Campaign Memory Tools (11 tools)

Comprehensive campaign state management with full D&D 5e character sheet support.

[View Campaign Memory Tools Documentation →](./tools/campaign-memory.md)

- `create_campaign` - Create new campaign with dedicated database
- `list_campaigns` - List all campaigns
- `save_character` - Create or fully replace character (PC/NPC)
- `update_character` - Patch specific character fields
- `get_character` - Retrieve full character sheet
- `list_characters` - List characters with filters
- `save_plot_event` - Record narrative events and create plot hooks
- `list_open_hooks` - Get unresolved plot threads
- `resolve_hook` - Mark plot hooks as resolved
- `set_world_flag` - Set key/value world state
- `get_world_flags` - Retrieve all world flags

### Session Management Tools (7 tools)

Session workflow with AI-powered compression and context restoration.

[View Session Management Tools Documentation →](./tools/session-management.md)

- `start_session` - Load campaign state and render DM brief
- `end_session` - Compress session with AI summary
- `checkpoint` - Save mid-session turn snapshots
- `get_turn_history` - Retrieve turn-by-turn history
- `get_session_brief` - Get compact markdown brief
- `list_sessions` - List historical sessions
- `get_npc_relationships` - Query NPC relationship graph
- `export_session_recap` - Export markdown recap across sessions

## Guides

### Quick Start

New to the server? Start here for a step-by-step tutorial.

[View Quick Start Guide →](./guides/quick-start.md)

### Character Creation

Deep dive into creating D&D 5e character sheets with all supported fields.

[View Character Creation Guide →](./guides/character-creation.md)

### Session Workflow

Best practices for running sessions with checkpoints and compression.

[View Session Workflow Guide →](./guides/session-workflow.md)

### Deployment & Configuration

Server deployment, transport modes, and MCP client setup.

[View Deployment Guide →](./guides/deployment.md)

### Architecture & Development

Technical architecture, database model, and development practices.

[View Architecture Guide →](./guides/architecture.md)

### Troubleshooting

Common issues and solutions for configuration, sessions, and HTTP transport.

[View Troubleshooting Guide →](./guides/troubleshooting.md)

## Examples

Real-world usage examples for all 22 tools with sample inputs and outputs.

[View Tool Examples →](./examples/tool-examples.json)

## Version Information

**Current Version**: 1.0.0
**MCP SDK**: go-sdk v1.4.1
**D&D Edition**: 5th Edition
**Last Updated**: 2026-03-16

## Contributing

For contributors and maintainers, see [CLAUDE.md](../CLAUDE.md) for documentation maintenance guidelines and critical rules about keeping docs in sync with code changes.
