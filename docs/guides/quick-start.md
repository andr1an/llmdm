# Quick Start Guide

Get started with the D&D Campaign MCP Server in minutes. This guide walks through creating your first campaign, adding characters, and running a session.

## Prerequisites

- MCP server installed and running
- AI assistant with MCP support (Claude, etc.)

## Step 1: Create Your First Campaign

Use the `create_campaign` tool to create a new campaign:

```json
{
  "name": "Lost Mines of Phandelver",
  "description": "Classic D&D 5e starter adventure in the Sword Coast"
}
```

**Response:**
```json
{
  "campaign_id": "camp_abc123",
  "db_path": "/path/to/campaigns/camp_abc123.db",
  "campaign": {
    "id": "camp_abc123",
    "name": "Lost Mines of Phandelver",
    "current_session": 0
  }
}
```

Save the `campaign_id` - you'll need it for all future tool calls.

## Step 2: Add Characters

### Create a Player Character

Use `save_character` to add a PC:

```json
{
  "campaign_id": "camp_abc123",
  "name": "Thorin Oakenshield",
  "type": "pc",
  "class": "Fighter",
  "race": "Dwarf",
  "level": 1,
  "hp_current": 12.0,
  "hp_max": 12.0,
  "str": 16,
  "dex": 12,
  "con": 15,
  "int_stat": 10,
  "wis": 11,
  "cha": 13,
  "ac": 16,
  "speed": "25 ft",
  "proficiencies": {
    "armor": ["Light Armor", "Medium Armor", "Heavy Armor", "Shields"],
    "weapons": ["Simple Weapons", "Martial Weapons"],
    "tools": [],
    "saving_throws": ["STR", "CON"],
    "skills": ["Athletics", "Intimidation"]
  },
  "skills": [
    {"name": "Athletics", "proficient": true, "modifier": 5}
  ],
  "languages": ["Common", "Dwarvish"],
  "gold": 50,
  "status": "active"
}
```

### Create an NPC

```json
{
  "campaign_id": "camp_abc123",
  "name": "Sildar Hallwinter",
  "type": "npc",
  "class": "Fighter",
  "race": "Human",
  "level": 4,
  "hp_current": 35.0,
  "hp_max": 35.0,
  "str": 14,
  "dex": 12,
  "con": 14,
  "int_stat": 12,
  "wis": 13,
  "cha": 11,
  "ac": 17,
  "status": "active",
  "notes": "Allied NPC, member of the Lords' Alliance"
}
```

## Step 3: Start Your First Session

Use `start_session` to begin session 1:

```json
{
  "campaign_id": "camp_abc123",
  "session": 1
}
```

**Response includes:**
- Session brief with campaign context
- Active character summaries
- Open plot hooks (empty for first session)
- World flags
- DM system prompt

## Step 4: Make Some Rolls

### Attack Roll

Use the `roll` tool for a standard attack:

```json
{
  "campaign_id": "camp_abc123",
  "notation": "1d20+5",
  "reason": "Attack goblin with longsword",
  "character": "Thorin Oakenshield",
  "session": 1
}
```

**Response:**
```json
{
  "roll_result": {
    "total": 18,
    "rolls": [13],
    "modifier": 5,
    "notation": "1d20+5",
    "roll_id": "roll_001"
  }
}
```

### Roll with Advantage

```json
{
  "campaign_id": "camp_abc123",
  "notation": "1d20+3",
  "reason": "Stealth check",
  "character": "Thorin Oakenshield",
  "session": 1,
  "advantage": true
}
```

**Response:**
```json
{
  "roll_result": {
    "total": 21,
    "rolls": [18, 7],
    "kept": [18],
    "modifier": 3
  }
}
```

### Saving Throw

```json
{
  "campaign_id": "camp_abc123",
  "character": "Thorin Oakenshield",
  "stat": "DEX",
  "modifier": 1.0,
  "dc": 15.0,
  "reason": "Dodge falling rocks",
  "session": 1
}
```

## Step 5: Track Turn-by-Turn Events

Use `checkpoint` to save turn snapshots during combat:

```json
{
  "campaign_id": "camp_abc123",
  "session": 1,
  "note": "Thorin attacks goblin #1",
  "data": {
    "turn_id": 1,
    "sequence": 1,
    "player_action": "Attack with longsword",
    "narrative": "Thorin charges the goblin and swings his sword",
    "tool_results": {
      "attack_roll": {"total": 18},
      "damage_roll": {"total": 9}
    }
  }
}
```

## Step 6: Record Major Plot Events

Use `save_plot_event` to track narrative events:

```json
{
  "campaign_id": "camp_abc123",
  "session": 1,
  "summary": "The party was ambushed by goblins on the road to Phandalin. After defeating the goblins, they discovered a map showing the location of Cragmaw Hideout. They also found evidence that Gundren Rockseeker has been captured.",
  "consequences": "Goblins in the area are now on high alert. The party knows where to find the hideout.",
  "hooks": [
    "Rescue Gundren from Cragmaw Hideout",
    "Investigate the Black Spider mentioned in the goblin's notes"
  ]
}
```

**Response includes:**
- Created event record
- Two new plot hooks (unresolved)

## Step 7: Update Character HP

Use `update_character` to modify specific fields:

```json
{
  "campaign_id": "camp_abc123",
  "name": "Thorin Oakenshield",
  "hp_current": 8.0
}
```

## Step 8: End the Session

Use `end_session` to compress the session with AI:

```json
{
  "campaign_id": "camp_abc123",
  "session": 1,
  "raw_events": "Turn 1: The party traveled along the road to Phandalin when Thorin spotted goblin tracks.\n\nTurn 2: Goblins ambushed from the bushes. Initiative was rolled: Thorin 15, Sildar 12, Goblins 8.\n\nTurn 3: Thorin attacked goblin #1, rolling 18 to hit, dealing 9 damage and killing it.\n\n[... full session transcript ...]\n\nThe party found a map and notes about Gundren's capture.",
  "dm_notes": "Great first session! Players engaged well with combat and roleplay."
}
```

**Response:**
```json
{
  "summary": "The party was ambushed by goblins on the road to Phandalin. Thorin led the combat effectively, while Sildar provided support. After defeating the goblins, they discovered crucial intelligence: Gundren Rockseeker has been captured by the Cragmaw goblins and taken to their hideout. A mysterious figure called 'The Black Spider' was mentioned in the goblin's notes."
}
```

## Step 9: Between Sessions

### View Open Plot Hooks

```json
{
  "campaign_id": "camp_abc123"
}
```

Use `list_open_hooks` to see unresolved threads.

### Export a Recap

```json
{
  "campaign_id": "camp_abc123",
  "from_session": 1.0,
  "to_session": 1.0
}
```

Use `export_session_recap` to generate markdown for players.

## Step 10: Start Next Session

When ready for session 2:

```json
{
  "campaign_id": "camp_abc123",
  "session": 2,
  "recent_sessions": 3
}
```

Use `start_session` - it will load:
- Last session's AI-compressed summary
- Active characters with current HP
- Open plot hooks
- World state flags

## Common Workflows

### Update Character After Level Up

```json
{
  "campaign_id": "camp_abc123",
  "name": "Thorin Oakenshield",
  "level": 2,
  "hp_current": 20.0,
  "hp_max": 20.0,
  "experience_points": 300
}
```

### Set World State Flags

```json
{
  "campaign_id": "camp_abc123",
  "key": "gundren_rescued",
  "value": "true"
}
```

### Resolve a Plot Hook

First get hooks with `list_open_hooks`, then:

```json
{
  "campaign_id": "camp_abc123",
  "hook_id": "hook_001",
  "resolution": "Party successfully rescued Gundren from Cragmaw Hideout"
}
```

### View Roll History

```json
{
  "campaign_id": "camp_abc123",
  "session": 1,
  "limit": 20
}
```

Use `get_roll_history` to review all rolls from a session.

## Tips for DMs

1. **Start each session** with `start_session` to load context
2. **Use checkpoints** liberally during combat for turn tracking
3. **Record plot events** after major story beats
4. **Update character HP/status** immediately when changed
5. **End sessions** with `end_session` for AI compression
6. **Review open hooks** between sessions for prep
7. **Export recaps** to share with players between sessions

## Next Steps

- [Character Creation Guide](./character-creation.md) - Deep dive into D&D 5e character sheets
- [Session Workflow Guide](./session-workflow.md) - Best practices for session management
- [Tool Reference](../tools/) - Complete documentation for all 22 tools
