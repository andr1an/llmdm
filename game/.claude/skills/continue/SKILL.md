---
name: continue
description: Continue an existing D&D campaign by starting a new session
version: 1.0.0
---

# Continue Campaign Skill

This skill resumes an existing D&D campaign by loading campaign state and starting a new session.

## When to Use

Use this skill when:
- Continuing an existing campaign
- Starting a new session (session 2, 3, 4, etc.)
- Resuming after a break

## Workflow

Follow these steps in order:

### 1. List Available Campaigns

Use `mcp__dnd-campaign__list_campaigns` to show all campaigns.

Display the campaigns to the player in a clear format:
```
Available Campaigns:
1. The Dragon's Lair (Session 3)
2. Curse of Strahd (Session 7)
3. Custom Adventure (Session 1)
```

Ask the player which campaign to continue.

### 2. Get Campaign Details (Optional but Recommended)

Before starting the session, you can review:
- `mcp__dnd-campaign__list_characters` - See active party members
- `mcp__dnd-campaign__list_open_hooks` - Check unresolved plot threads
- `mcp__dnd-campaign__get_world_flags` - Review world state
- `mcp__dnd-campaign__list_sessions` - See session history

This helps you prepare better session content.

### 3. Determine Next Session Number

From the campaign list or session history, identify the next session number.
If current session is 5, next session is 6.

### 4. Start the New Session

Use `mcp__dnd-campaign__start_session` with:
- `campaign_id`: The selected campaign ID
- `session`: Next session number
- `recent_sessions`: 3 (default, loads last 3 session summaries)

This returns a comprehensive **DM Brief** containing:
- **Campaign Overview**: Name, current session
- **Previous Sessions**: AI-compressed summaries of recent sessions
- **Active Characters**: PCs and important NPCs with current HP and status
- **Open Plot Hooks**: Unresolved story threads
- **World State**: Current flags and conditions

### 5. Present the Brief to Player

Share the session brief with the player:
- Recap previous session highlights
- Remind them of open quests/hooks
- Confirm their character's current status (HP, conditions, inventory)
- Ask if they want to review anything before starting

### 6. Begin the Session

Describe the current situation:
- Where they are (based on last session)
- What they see/hear/smell
- Time of day, weather, atmosphere
- Any immediate threats or opportunities

Then ask: "What do you do?"

### 7. During Session

Follow standard D&D gameplay:
- **All dice rolls** via `mcp__dnd-campaign__roll`, `mcp__dnd-campaign__roll_contested`, or `mcp__dnd-campaign__roll_saving_throw`
- **Checkpoints** via `mcp__dnd-campaign__checkpoint` for major events and every combat turn
- **Character updates** via `mcp__dnd-campaign__update_character` for HP, conditions, inventory, etc.
- **Plot events** via `mcp__dnd-campaign__save_plot_event` after major story beats
- **Hook resolution** via `mcp__dnd-campaign__resolve_hook` when plot threads close
- **World state** via `mcp__dnd-campaign__set_world_flag` for persistent changes
- **New NPCs** via `mcp__dnd-campaign__save_character` when they appear

## Example Interaction

**Player**: "Let's continue my campaign"

**Claude**: [Lists campaigns]
```
Available Campaigns:
1. The Dragon's Lair (Session 3) - campaign_abc123
2. Curse of Strahd (Session 1) - campaign_def456
```
"Which campaign would you like to continue?"

**Player**: "The Dragon's Lair"

**Claude**: [Starts session 4, gets brief]

"Welcome back to The Dragon's Lair - Session 4!

**Previous Session Recap**: You defeated the goblin ambush and discovered a map leading to Cragmaw Castle. You rescued the dwarf Sildar Hallwinter and learned that Gundren Rockseeker was kidnapped by the goblin king.

**Active Characters**:
- Thorin Ironfist (Fighter 2): HP 18/18, AC 16
- Elara Moonwhisper (Wizard 2): HP 10/12, AC 12

**Open Hooks**:
- Rescue Gundren from Cragmaw Castle
- Deliver supplies to Barthen's Provisions in Phandalin
- Investigate the Redbrands gang

**Current Location**: You're camping on the High Road, one day from Phandalin.

You wake to the sound of wolves howling in the distance. The campfire has burned low, and dawn is breaking through the trees. What do you do?"

## Important Notes

- The `start_session` brief is your **system prompt** - it contains everything you need to know
- ALWAYS use MCP tools for dice rolls - NEVER simulate
- Update characters immediately when they take damage or gain conditions
- Checkpoint frequently (combat turns, major decisions, room transitions)
- Save plot events as they happen, not at end of session
- Resolve hooks when plot threads close
- Track time with world flags if relevant

## Continuity Best Practices

- Honor previous session events - they're in the compressed summaries
- Check character HP/conditions from the brief before narrating
- Reference open hooks to drive the story forward
- Maintain consistency with world flags
- Use NPC relationships if they exist

## Required Tools

This skill uses these MCP tools:
- `list_campaigns`
- `start_session`
- `list_characters` (optional prep)
- `list_open_hooks` (optional prep)
- `get_world_flags` (optional prep)
- `list_sessions` (optional prep)
- `roll` / `roll_contested` / `roll_saving_throw`
- `checkpoint`
- `update_character`
- `save_plot_event`
- `resolve_hook`
- `set_world_flag`
- `save_character` (for new NPCs)
