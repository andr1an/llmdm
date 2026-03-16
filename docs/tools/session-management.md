# Session Management Tools

Session workflow tools with AI-powered compression, context restoration, and turn-by-turn tracking.

## Tools Overview

| Tool | Description |
|------|-------------|
| `start_session` | Load campaign state and render session brief for DM context |
| `end_session` | Compress and store end-of-session narrative summary |
| `checkpoint` | Save mid-session checkpoint note with turn data |
| `get_turn_history` | Retrieve turn history from checkpoints for a session |
| `get_session_brief` | Get compact markdown briefing for quick context restoration |
| `list_sessions` | List historical sessions with summary previews |
| `get_npc_relationships` | Query relationship edges involving NPCs |
| `export_session_recap` | Export markdown recap across all or selected sessions |

---

## Session Workflow

```
1. start_session     → Loads campaign state, renders DM brief
   ↓
2. During play       → Use checkpoint to track turns/events
   ↓                  → Use roll tools for dice mechanics
   ↓                  → Use save_plot_event for major events
   ↓
3. end_session       → Compress session with AI summary
   ↓
4. Between sessions  → export_session_recap for narrative recap
   ↓
5. Next session      → start_session restores context
```

---

## start_session

Load campaign state and render a session brief for DM context. Increments campaign session number.

### Input Parameters

| Parameter | Type | Required | Description | Example |
|-----------|------|----------|-------------|---------|
| `campaign_id` | string | Yes | Campaign ID | `"camp_abc123"` |
| `session` | int | Yes | Session number to start | `4` |
| `recent_sessions` | int | No | How many prior sessions to include (default: 3) | `5` |

### Input Schema

```json
{
  "campaign_id": "camp_abc123",
  "session": 4,
  "recent_sessions": 3
}
```

### Output Schema

```json
{
  "brief": {
    "session_brief": "# Session 4\n\n## Last Session Summary\n[AI-compressed summary of session 3]\n\n## Campaign Status\n...",
    "active_characters": [
      {
        "name": "Thorin",
        "type": "pc",
        "class": "Fighter",
        "race": "Dwarf",
        "level": 5,
        "ac": 18,
        "hp": {"current": 38, "max": 45},
        "status": "active",
        "conditions": []
      }
    ],
    "open_hooks": [
      {
        "id": "hook_001",
        "campaign_id": "camp_abc123",
        "hook": "Investigate Cragmaw Castle",
        "session_opened": 3,
        "event_id": "event_xyz",
        "resolved": false,
        "resolution": ""
      }
    ],
    "world_flags": {
      "goblins_defeated": "true",
      "sildar_rescued": "true"
    },
    "last_session_number": 3,
    "last_session_summary": "[Compressed summary from session 3]",
    "dm_system_prompt": "You are the DM for this D&D 5e campaign. Use the session brief to maintain consistency..."
  }
}
```

### Brief Structure

The `session_brief` markdown string contains:

1. **Session Header** - Current session number
2. **Last Session Summary** - AI-compressed recap of previous session
3. **Active Characters** - Party composition with HP/conditions
4. **Open Plot Hooks** - Unresolved story threads
5. **World Flags** - Current world state
6. **Recent Events** - Highlights from last `recent_sessions` sessions

### Notes

- Increments `current_session` in the campaign metadata
- `recent_sessions` controls how much history is included in the brief
- `dm_system_prompt` provides context for AI assistants
- Brief is optimized for quick DM context restoration
- Should be called at the start of each session

### See Also

- [`end_session`](#end_session) - Close session with summary
- [`get_session_brief`](#get_session_brief) - Get brief without starting new session
- [`checkpoint`](#checkpoint) - Track turns during session

---

## end_session

Compress and store end-of-session narrative summary using AI compression.

### Input Parameters

| Parameter | Type | Required | Description | Example |
|-----------|------|----------|-------------|---------|
| `campaign_id` | string | Yes | Campaign ID | `"camp_abc123"` |
| `session` | int | Yes | Session number to end | `3` |
| `raw_events` | string | Yes | Full narrative event log for the session | `"Turn 1: Party entered..."` |
| `dm_notes` | string | No | Optional DM notes | `"Great session, players loved it"` |

### Input Schema

```json
{
  "campaign_id": "camp_abc123",
  "session": 3,
  "raw_events": "Turn 1: The party entered the goblin cave cautiously. Thorin led with his shield raised.\n\nTurn 2: Goblins ambushed from the shadows. Initiative: Thorin 15, Gandalf 12, Goblins 8.\n\nTurn 3: Thorin attacked the goblin leader with his longsword, rolling 18+5=23 to hit, dealing 12 damage...\n\n[Full session narrative]",
  "dm_notes": "Players really enjoyed the combat encounter. Need to introduce Sildar next session."
}
```

### Output Schema

```json
{
  "summary": "The party infiltrated a goblin cave and fought through several encounters. Thorin took the lead in combat, while Gandalf provided magical support. After defeating the goblin leader Klarg, they discovered that a human named Sildar Hallwinter is being held prisoner deeper in the cave. The party decided to rest before continuing their rescue mission."
}
```

### AI Compression Behavior

- **With Anthropic API**: Uses Claude to intelligently compress `raw_events` into a concise summary
- **Without API**: Falls back to truncation-based compression
- Summary preserves key plot points, character actions, and consequences
- Typical compression: 2000+ word transcript → 100-200 word summary

### Notes

- `raw_events` should contain the full session transcript (turn-by-turn narrative)
- Compression is destructive - keep `raw_events` if you need full detail
- `dm_notes` are stored separately and not compressed
- Summary is used in future `start_session` briefs
- Checkpoint data is preserved separately for turn history

### See Also

- [`start_session`](#start_session) - Next session loads this summary
- [`checkpoint`](#checkpoint) - Track detailed turn data during session
- [`export_session_recap`](#export_session_recap) - Generate markdown recap

---

## checkpoint

Save a mid-session checkpoint note with optional turn data for turn-by-turn tracking.

### Input Parameters

| Parameter | Type | Required | Description | Example |
|-----------|------|----------|-------------|---------|
| `campaign_id` | string | Yes | Campaign ID | `"camp_abc123"` |
| `session` | int | Yes | Current session number | `3` |
| `note` | string | Yes | Checkpoint note | `"Thorin attacks goblin"` |
| `data` | object | No | Optional turn data (turn_id, sequence, player_action, narrative, tool_results) | See below |

### Turn Data Structure

```json
{
  "turn_id": 5,
  "sequence": 2,
  "player_action": "Attack with longsword",
  "narrative": "Thorin swings his longsword at the goblin leader",
  "tool_results": {
    "roll": {
      "total": 23,
      "notation": "1d20+5"
    }
  }
}
```

### Input Schema (Simple Checkpoint)

```json
{
  "campaign_id": "camp_abc123",
  "session": 3,
  "note": "Party enters goblin cave"
}
```

### Input Schema (Turn Checkpoint with Data)

```json
{
  "campaign_id": "camp_abc123",
  "session": 3,
  "note": "Thorin attacks goblin leader",
  "data": {
    "turn_id": 5,
    "sequence": 2,
    "player_action": "Attack with longsword",
    "narrative": "Thorin charges forward and swings his longsword at Klarg",
    "tool_results": {
      "attack_roll": {"total": 23, "notation": "1d20+5"},
      "damage_roll": {"total": 12, "notation": "1d8+3"}
    }
  }
}
```

### Output Schema

```json
{
  "checkpoint_id": "ckpt_abc123"
}
```

### Notes

- Checkpoints are ordered chronologically within a session
- `data` field is flexible - store any turn-relevant information
- Common pattern: One checkpoint per combat turn or major decision
- Useful for reconstructing detailed session history
- Checkpoint data survives `end_session` compression

### See Also

- [`get_turn_history`](#get_turn_history) - Retrieve checkpoints for a session
- [`end_session`](#end_session) - Compress session (preserves checkpoints)

---

## get_turn_history

Retrieve turn history from checkpoints for a session.

### Input Parameters

| Parameter | Type | Required | Description | Example |
|-----------|------|----------|-------------|---------|
| `campaign_id` | string | Yes | Campaign ID | `"camp_abc123"` |
| `session` | int | Yes | Session number | `3` |
| `limit` | int | No | Maximum number of turns to return (default: 50) | `20` |

### Input Schema

```json
{
  "campaign_id": "camp_abc123",
  "session": 3,
  "limit": 10
}
```

### Output Schema

```json
{
  "turns": [
    {
      "id": "ckpt_001",
      "campaign_id": "camp_abc123",
      "session": 3,
      "note": "Party enters goblin cave",
      "data": {},
      "created_at": "2026-03-16T14:00:00Z"
    },
    {
      "id": "ckpt_002",
      "campaign_id": "camp_abc123",
      "session": 3,
      "note": "Thorin attacks goblin leader",
      "data": {
        "turn_id": 5,
        "sequence": 2,
        "player_action": "Attack with longsword",
        "narrative": "Thorin swings his longsword",
        "tool_results": {
          "attack_roll": {"total": 23},
          "damage_roll": {"total": 12}
        }
      },
      "created_at": "2026-03-16T14:15:00Z"
    }
  ]
}
```

### Notes

- Results are ordered by `created_at` ascending (chronological order)
- Returns up to `limit` checkpoints (default: 50)
- Useful for reconstructing turn-by-turn session narrative
- Empty array if no checkpoints exist for the session

### See Also

- [`checkpoint`](#checkpoint) - Create checkpoint records
- [`export_session_recap`](#export_session_recap) - Generate narrative recap

---

## get_session_brief

Get a compact markdown briefing for quick context restoration without starting a new session.

### Input Parameters

| Parameter | Type | Required | Description | Example |
|-----------|------|----------|-------------|---------|
| `campaign_id` | string | Yes | Campaign ID | `"camp_abc123"` |

### Input Schema

```json
{
  "campaign_id": "camp_abc123"
}
```

### Output Schema

```json
{
  "brief": "# Campaign: Lost Mines of Phandelver\n\n## Current Session: 3\n\n## Active Characters\n- Thorin (Fighter 5) - HP: 38/45, AC: 18\n- Gandalf (Wizard 18) - HP: 120/120, AC: 15\n\n## Open Plot Hooks\n- Investigate Cragmaw Castle\n- Find the missing caravan\n\n## World State\n- goblins_defeated: true\n- sildar_rescued: true\n\n## Last Session Summary\n[Compressed summary from most recent session]\n"
}
```

### Notes

- Does NOT increment session number (unlike `start_session`)
- Useful for mid-session context restoration or DM prep
- Returns markdown-formatted brief string
- Includes same context as `start_session` but without side effects

### See Also

- [`start_session`](#start_session) - Start new session with brief
- [`export_session_recap`](#export_session_recap) - Full narrative recap

---

## list_sessions

List historical sessions with summary previews.

### Input Parameters

| Parameter | Type | Required | Description | Example |
|-----------|------|----------|-------------|---------|
| `campaign_id` | string | Yes | Campaign ID | `"camp_abc123"` |

### Input Schema

```json
{
  "campaign_id": "camp_abc123"
}
```

### Output Schema

```json
{
  "sessions": [
    {
      "session": 3,
      "date": "2026-03-15T14:00:00Z",
      "summary_preview": "The party infiltrated a goblin cave and fought through several encounters...",
      "hooks_opened": 2,
      "hooks_resolved": 1
    },
    {
      "session": 2,
      "date": "2026-03-08T14:00:00Z",
      "summary_preview": "Party traveled to Phandalin and met with Sildar...",
      "hooks_opened": 3,
      "hooks_resolved": 0
    },
    {
      "session": 1,
      "date": "2026-03-01T14:00:00Z",
      "summary_preview": "Campaign begins with the party accepting a quest...",
      "hooks_opened": 1,
      "hooks_resolved": 0
    }
  ]
}
```

### Notes

- Results are ordered by session number descending (most recent first)
- `summary_preview` is truncated to ~200 characters
- `hooks_opened`/`hooks_resolved` show plot progression per session
- Useful for DM to review campaign history

### See Also

- [`start_session`](#start_session) - Start new session
- [`end_session`](#end_session) - Create session summaries
- [`export_session_recap`](#export_session_recap) - Full narrative export

---

## get_npc_relationships

Query relationship edges involving NPCs. Optionally filter to one NPC name.

### Input Parameters

| Parameter | Type | Required | Description | Example |
|-----------|------|----------|-------------|---------|
| `campaign_id` | string | Yes | Campaign ID | `"camp_abc123"` |
| `npc_name` | string | No | Optional NPC name filter | `"Sildar"` |

### Input Schema (All NPC Relationships)

```json
{
  "campaign_id": "camp_abc123"
}
```

### Input Schema (Specific NPC)

```json
{
  "campaign_id": "camp_abc123",
  "npc_name": "Sildar"
}
```

### Output Schema

```json
{
  "relationships": [
    {
      "source": "Thorin",
      "source_type": "pc",
      "target": "Sildar",
      "target_type": "npc",
      "relation": "ally"
    },
    {
      "source": "Sildar",
      "source_type": "npc",
      "target": "Gundren",
      "target_type": "npc",
      "relation": "friend"
    },
    {
      "source": "Gandalf",
      "source_type": "pc",
      "target": "Sildar",
      "target_type": "npc",
      "relation": "trusted contact"
    }
  ]
}
```

### Notes

- Returns all edges where source OR target is an NPC
- If `npc_name` provided, filters to edges involving that NPC
- Relationships are directional (source → target)
- Extracted from character `relationships` maps
- Useful for tracking NPC social networks and party connections

### See Also

- [`save_character`](./campaign-memory.md#save_character) - Set character relationships
- [`update_character`](./campaign-memory.md#update_character) - Update relationships

---

## export_session_recap

Export a markdown recap for a campaign across all or selected sessions.

### Input Parameters

| Parameter | Type | Required | Description | Example |
|-----------|------|----------|-------------|---------|
| `campaign_id` | string | Yes | Campaign ID | `"camp_abc123"` |
| `from_session` | float | No | Optional lower inclusive session bound | `1.0` |
| `to_session` | float | No | Optional upper inclusive session bound | `5.0` |

### Input Schema (All Sessions)

```json
{
  "campaign_id": "camp_abc123"
}
```

### Input Schema (Sessions 1-5)

```json
{
  "campaign_id": "camp_abc123",
  "from_session": 1.0,
  "to_session": 5.0
}
```

### Output Schema

```json
{
  "markdown": "# Lost Mines of Phandelver - Campaign Recap\n\n## Session 1 (2026-03-01)\n\nThe campaign begins with the party accepting a quest from Gundren Rockseeker...\n\n## Session 2 (2026-03-08)\n\nParty traveled to Phandalin and met with Sildar Hallwinter...\n\n## Session 3 (2026-03-15)\n\nThe party infiltrated a goblin cave and fought through several encounters...\n\n---\n\n**Total Sessions**: 3\n**Open Plot Hooks**: 2\n"
}
```

### Markdown Format

```markdown
# [Campaign Name] - Campaign Recap

## Session N (Date)

[AI-compressed summary]

### Key Events
- Event 1
- Event 2

### Hooks Opened
- Hook 1
- Hook 2

### Hooks Resolved
- Hook 1 (resolved in session N+1)

---

**Total Sessions**: N
**Open Plot Hooks**: N
```

### Notes

- Includes AI-compressed summaries from `end_session`
- Shows plot progression across sessions
- If no session bounds provided, exports entire campaign
- Useful for sharing campaign recaps with players
- Can be saved to file or shared as markdown

### See Also

- [`end_session`](#end_session) - Creates session summaries
- [`list_sessions`](#list_sessions) - Session metadata overview
- [`save_plot_event`](./campaign-memory.md#save_plot_event) - Track plot events
