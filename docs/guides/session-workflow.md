# Session Workflow Guide

Best practices for running D&D sessions with the MCP server's session management tools.

## Overview

The session workflow follows a structured pattern:

```
Pre-Session → During Play → Post-Session → Between Sessions
```

Each phase uses specific tools to maintain campaign continuity and context.

## Pre-Session: Preparation

### 1. Review Open Hooks

Before starting the session, review unresolved plot threads:

```json
{
  "campaign_id": "camp_abc123"
}
```

Use `list_open_hooks` to see what story threads are dangling from previous sessions.

### 2. Check Character Status

Review active characters:

```json
{
  "campaign_id": "camp_abc123",
  "type": "pc",
  "status": "active"
}
```

Use `list_characters` to see party composition and status.

### 3. Start the Session

Launch the session with context restoration:

```json
{
  "campaign_id": "camp_abc123",
  "session": 5,
  "recent_sessions": 3
}
```

Use `start_session` to:
- Load AI-compressed summary of previous sessions
- Get active character summaries with HP/conditions
- View open plot hooks
- Access world state flags

**The session brief includes:**
- Markdown-formatted DM brief
- Last session summary (AI-compressed)
- Active characters with stats
- Open hooks
- World flags
- DM system prompt for AI assistants

## During Play: Active Session

### Tracking Turns and Events

Use `checkpoint` to create snapshots throughout the session.

#### Simple Checkpoint (Narrative Marker)

For non-combat events:

```json
{
  "campaign_id": "camp_abc123",
  "session": 5,
  "note": "Party arrives at Cragmaw Castle"
}
```

#### Combat Turn Checkpoint

For detailed turn tracking:

```json
{
  "campaign_id": "camp_abc123",
  "session": 5,
  "note": "Thorin's turn - attacks King Grol",
  "data": {
    "turn_id": 3,
    "sequence": 1,
    "player_action": "Attack with greatsword",
    "narrative": "Thorin charges King Grol and swings his greatsword with both hands",
    "tool_results": {
      "attack_roll": {
        "total": 21,
        "notation": "1d20+7",
        "roll_id": "roll_123"
      },
      "damage_roll": {
        "total": 15,
        "notation": "2d6+4",
        "roll_id": "roll_124"
      }
    }
  }
}
```

**Turn Data Structure:**
- `turn_id`: Combat turn number (1, 2, 3...)
- `sequence`: Action order within turn (1st action, 2nd action...)
- `player_action`: What the player declared
- `narrative`: Descriptive text of what happened
- `tool_results`: Roll results, ability check results, etc.

### When to Use Checkpoints

**Use checkpoints for:**
- Combat turns (one per character action)
- Major decisions (party splits, important choices)
- Environmental changes (traps triggered, room entered)
- Dialogue milestones (key NPC interactions)

**Don't overuse:**
- Avoid checkpoints for trivial actions
- Don't checkpoint every single dice roll
- Focus on narrative significance

### Rolling Dice

Use dice tools throughout the session:

```json
{
  "campaign_id": "camp_abc123",
  "notation": "1d20+7",
  "reason": "Attack King Grol",
  "character": "Thorin",
  "session": 5
}
```

All rolls are automatically logged for later review.

### Updating Character State

Update HP, conditions, and inventory as they change:

```json
{
  "campaign_id": "camp_abc123",
  "name": "Thorin",
  "hp_current": 22.0,
  "conditions": ["Poisoned"]
}
```

Use `update_character` to keep state current.

### Recording Plot Events

After major story beats, record the event:

```json
{
  "campaign_id": "camp_abc123",
  "session": 5,
  "summary": "The party defeated King Grol and rescued Gundren Rockseeker from Cragmaw Castle. They discovered that the Black Spider has already reached Wave Echo Cave and is searching for the Forge of Spells.",
  "consequences": "Gundren is reunited with the party and reveals the location of Wave Echo Cave. The Cragmaw tribe is leaderless and scattered.",
  "hooks": [
    "Travel to Wave Echo Cave",
    "Stop the Black Spider from claiming the Forge of Spells",
    "Investigate what happened to Gundren's brothers"
  ]
}
```

Use `save_plot_event` to:
- Record narrative summary (2-4 sentences)
- Track consequences (world changes)
- Create new plot hooks (unresolved threads)

### Setting World Flags

Update world state as events unfold:

```json
{
  "campaign_id": "camp_abc123",
  "key": "gundren_rescued",
  "value": "true"
}
```

```json
{
  "campaign_id": "camp_abc123",
  "key": "cragmaw_castle_cleared",
  "value": "true"
}
```

Use `set_world_flag` for boolean flags, counters, or state tracking.

### Resolving Hooks

When plot threads are resolved:

```json
{
  "campaign_id": "camp_abc123",
  "hook_id": "hook_002",
  "resolution": "Party rescued Gundren from Cragmaw Castle. He was wounded but alive."
}
```

Use `resolve_hook` to close story threads and track narrative progression.

## Post-Session: Wrap-Up

### 1. End the Session

Compress the session with AI-powered summarization:

```json
{
  "campaign_id": "camp_abc123",
  "session": 5,
  "raw_events": "Turn 1: The party approached Cragmaw Castle under cover of darkness. Thorin scouted ahead and spotted two goblin sentries on the walls.\n\nTurn 2: The party decided to sneak around to the side entrance. Stealth checks: Thorin 18, Gandalf 15, Bilbo 22.\n\nTurn 3: Successfully entered the castle. Found Gundren imprisoned in the throne room.\n\nTurn 4: King Grol confronted the party. Initiative rolled.\n\nTurn 5-12: Combat encounter with King Grol and his guards.\n\n[... full turn-by-turn narrative ...]\n\nTurn 25: Gundren revealed the location of Wave Echo Cave and warned about the Black Spider.\n\nSession ended with party deciding to rest before traveling to Wave Echo Cave.",
  "dm_notes": "Excellent session! Combat was tense. Players loved rescuing Gundren. Foreshadowing the Black Spider worked well."
}
```

Use `end_session` to:
- Compress raw events with AI (Anthropic API if available)
- Store DM notes
- Create summary for next session's brief

**AI Compression:**
- Reduces 2000+ word transcripts to 100-200 word summaries
- Preserves key plot points, character actions, consequences
- Maintains narrative continuity across sessions

### 2. Review the Session

Optional: Export a recap for immediate review:

```json
{
  "campaign_id": "camp_abc123",
  "from_session": 5.0,
  "to_session": 5.0
}
```

Use `export_session_recap` to generate a markdown summary.

## Between Sessions: Downtime

### Export Campaign Recap

Share a full campaign recap with players:

```json
{
  "campaign_id": "camp_abc123",
  "from_session": 1.0,
  "to_session": 5.0
}
```

Use `export_session_recap` to generate markdown covering all sessions.

**Recap includes:**
- Session-by-session summaries
- Plot events and consequences
- Hooks opened and resolved
- Total session count

### Review Session History

Check previous sessions:

```json
{
  "campaign_id": "camp_abc123"
}
```

Use `list_sessions` to see all sessions with previews.

### Review Turn History

For detailed combat review:

```json
{
  "campaign_id": "camp_abc123",
  "session": 5,
  "limit": 50
}
```

Use `get_turn_history` to reconstruct turn-by-turn events.

### Review Roll History

Analyze dice rolls:

```json
{
  "campaign_id": "camp_abc123",
  "session": 5
}
```

Use `get_roll_history` to see all rolls from a session.

### Prepare Next Session

Use `get_session_brief` for prep without starting the session:

```json
{
  "campaign_id": "camp_abc123"
}
```

This provides the same brief as `start_session` but without incrementing the session counter.

## Best Practices

### Checkpoint Frequency

**Combat:**
- One checkpoint per character's turn
- Include `data` with turn_id, sequence, and tool_results

**Exploration:**
- Checkpoint when entering new areas
- Checkpoint at major discoveries

**Social:**
- Checkpoint at key dialogue milestones
- Checkpoint when important information is revealed

### Plot Event Timing

Record plot events:
- **During session**: After major story beats conclude
- **End of session**: Summarize any uncaptured events

### World Flag Usage

Use world flags for:
- Quest completion: `"quest_cragmaw_complete": "true"`
- Faction reputation: `"harpers_reputation": "friendly"`
- Time tracking: `"days_since_start": "15"`
- Binary state: `"dragons_awakened": "false"`

### Hook Management

**Creating hooks:**
- Keep hooks specific and actionable
- Create hooks when new story threads emerge
- Average 2-4 hooks per session

**Resolving hooks:**
- Mark hooks resolved when story thread closes
- Include resolution description for narrative tracking
- Don't leave hooks open indefinitely

## Common Workflows

### Starting a New Campaign

```
1. create_campaign
2. save_character (for each PC)
3. save_character (for key NPCs)
4. start_session (session 1)
```

### Running a Combat Encounter

```
1. checkpoint - "Combat begins"
2. For each turn:
   a. roll (attack, damage, saves)
   b. checkpoint (with turn data)
   c. update_character (HP, conditions)
3. checkpoint - "Combat ends"
4. save_plot_event (if combat had plot significance)
```

### Ending a Session

```
1. save_plot_event (for any uncaptured events)
2. resolve_hook (for any closed threads)
3. set_world_flag (for world state changes)
4. end_session (with full transcript)
```

### Between-Session Prep

```
1. list_open_hooks (review dangling threads)
2. list_characters (check party status)
3. get_session_brief (prep context)
4. Plan next session encounters
```

## Advanced Patterns

### Multi-Party Tracking

For campaigns with party splits:

```json
{
  "campaign_id": "camp_abc123",
  "session": 5,
  "note": "Party A explores the dungeon",
  "data": {
    "party": "A",
    "location": "Cragmaw Castle - East Wing"
  }
}
```

```json
{
  "campaign_id": "camp_abc123",
  "session": 5,
  "note": "Party B negotiates in town",
  "data": {
    "party": "B",
    "location": "Phandalin - Town Square"
  }
}
```

### Time Tracking

Use world flags for in-game time:

```json
{
  "campaign_id": "camp_abc123",
  "key": "current_date",
  "value": "15th of Hammer, 1492 DR"
}
```

```json
{
  "campaign_id": "camp_abc123",
  "key": "days_until_ritual",
  "value": "3"
}
```

Update after each session or long rest.

### Long-Form Plot Tracking

For complex story arcs:

```json
{
  "campaign_id": "camp_abc123",
  "session": 5,
  "summary": "Party discovered ancient prophecy tablets in Wave Echo Cave...",
  "consequences": "Prophecy reveals that one of the PCs is destined to face the Demon Prince...",
  "hooks": [
    "Decipher the remaining prophecy fragments",
    "Investigate the PC's mysterious bloodline",
    "Prevent the Demon Prince's awakening (5 days remaining)"
  ]
}
```

Set flags for arc progression:

```json
{
  "campaign_id": "camp_abc123",
  "key": "prophecy_arc_stage",
  "value": "2"
}
```

## Troubleshooting

### Sessions Feel Fragmented

**Solution:**
- Use more checkpoints to capture context
- Include narrative in checkpoint notes
- Use `recent_sessions: 5` when starting sessions for more history

### Too Much Data Overhead

**Solution:**
- Use simple checkpoints without `data` for non-combat
- Only include tool_results for significant rolls
- Focus plot events on major story beats only

### Lost Context Between Sessions

**Solution:**
- Always use `end_session` to compress (enables AI summary)
- Record plot events during the session, not after
- Set world flags for critical state changes
- Export recaps for players between sessions

## See Also

- [Quick Start Guide](./quick-start.md) - Basic workflow tutorial
- [Session Management Tools](../tools/session-management.md) - Complete tool reference
- [Campaign Memory Tools](../tools/campaign-memory.md) - State management tools
