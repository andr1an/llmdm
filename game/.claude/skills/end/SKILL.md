---
name: end
description: End a D&D session by saving all state (NPCs, hooks, checkpoints) and creating compressed summary
version: 1.0.0
---

# End Session Skill

This skill properly ends a D&D session by saving all campaign state, resolving plot hooks, and creating an AI-compressed session summary for future continuity.

## When to Use

Use this skill when:
- The session is ending (time to wrap up)
- Player says "let's end here", "that's all for today", or similar
- Natural stopping point is reached

## Workflow

Follow these steps in order:

### 1. Verbal Session Recap

Provide a brief summary of what happened this session:
- Major events and accomplishments
- Significant NPCs encountered
- Loot or items acquired
- Combat encounters
- Plot developments
- Current location and status

This helps the player remember the session and confirms what should be saved.

### 2. Save Any Uncaptured NPCs

Review NPCs that appeared during the session. For each NPC that doesn't already have a character sheet:

Use `mcp__dnd-campaign__save_character` with:
- `type`: "npc"
- `name`: NPC name
- `hp_current` and `hp_max`: Appropriate for their role
- Additional details: race, class, alignment, backstory, notes, relationships

**Important NPCs**: Full character sheets with stats
**Minor NPCs**: Minimal details (name, HP, notes about their role)

### 3. Save Uncaptured Plot Events

Review major story events from the session. For any significant events not yet recorded:

Use `mcp__dnd-campaign__save_plot_event` with:
- `summary`: 2-4 sentences describing what happened
- `consequences`: How the world changed
- `hooks`: Array of new unresolved plot threads created by this event

Examples of events to save:
- Major combat victories
- Important discoveries (items, secrets, locations)
- NPC interactions that revealed information
- Quest milestones reached
- Character decisions with lasting consequences

### 4. Resolve Completed Plot Hooks

Review the list of open hooks. For any that were completed this session:

Use `mcp__dnd-campaign__resolve_hook` with:
- `hook_id`: The ID of the completed hook
- `resolution`: How it was resolved

Examples:
- "Rescued Gundren from Cragmaw Castle"
- "Defeated the Redbrand leader Glasstaff"
- "Delivered supplies to Barthen's Provisions"

### 5. Update World State Flags

Set or update world flags for persistent campaign state:

Use `mcp__dnd-campaign__set_world_flag` for:
- Time progression: `"days_elapsed"`, `"current_date"`
- Quest status: `"quest_cragmaw_complete": "true"`
- Faction reputation: `"harpers_reputation": "friendly"`
- Important counters: `"days_until_ritual": "3"`
- Location tracking: `"party_location": "Phandalin"`
- Any other persistent world state

### 6. Update Character Status

Ensure all character sheets reflect final session state:

Use `mcp__dnd-campaign__update_character` for:
- **HP**: Final HP values after any healing
- **Conditions**: Remove temporary conditions (unless they persist)
- **Inventory**: Add/remove items gained or used
- **Gold**: Update gold amounts
- **Experience**: Add XP if awarded
- **Level**: Level up if XP threshold reached
- **Status**: Change to dead, missing, or retired if applicable

### 7. Create Session Transcript

Compile a comprehensive turn-by-turn transcript of the session in narrative form. This should be a detailed chronological account including:

- Opening scene and player actions
- All significant dice rolls and their outcomes
- Combat turn-by-turn descriptions
- NPC dialogues and interactions
- Exploration discoveries
- Puzzle solutions or challenges
- Character decisions and consequences
- Closing situation

**Format**: Full narrative prose, 1000-2500+ words depending on session length.

This transcript is what will be AI-compressed into a 100-200 word summary.

### 8. End the Session

Use `mcp__dnd-campaign__end_session` with:
- `campaign_id`: Current campaign ID
- `session`: Current session number
- `raw_events`: The full transcript from step 7
- `dm_notes`: Optional DM-only notes about the session

**What This Does**:
- Sends `raw_events` to Anthropic API for AI compression
- Creates a concise 100-200 word summary preserving key plot points
- Stores the compressed summary for future session briefs
- This summary will appear in the next session's `start_session` brief

### 9. Confirm Completion

Tell the player:
- Session ended successfully
- Summary saved for next session
- What session number this was
- Preview what's next (open hooks)

Optionally offer to create a markdown recap:

Use `mcp__dnd-campaign__export_session_recap` to generate a player-friendly recap document.

## Example Interaction

**Player**: "Let's end the session here"

**Claude**: "Great session! Let me wrap things up.

**Session 4 Recap**:
- Traveled from High Road to Phandalin
- Defeated Redbrand thugs at the inn (4 enemies)
- Met Toblen Stonehill (innkeeper) and Sister Garaele (priestess)
- Learned about Redbrand hideout beneath Tresendar Manor
- Acquired 35 gold from Barthen for delivering supplies
- Currently resting at Stonehill Inn

Let me save all the session data..."

[Saves NPCs: Toblen Stonehill, Sister Garaele]

[Saves plot event: "Party arrives in Phandalin"]

[Resolves hook: "Deliver supplies to Barthen's Provisions"]

[Updates world flags: "party_location": "Phandalin", "days_elapsed": "5"]

[Updates characters: Thorin HP 16/18, Elara HP 12/12, added 35gp each]

[Creates and compresses session transcript]

[Ends session]

"Session 4 saved! Next session will load a brief with everything that happened.

**Open Hooks for Next Time**:
- Investigate the Redbrand hideout
- Help Sister Garaele find Bowgentle's spellbook
- Rescue Gundren from Cragmaw Castle

See you next time!"

## Important Notes

### Session Transcript Quality

The transcript quality directly affects future session briefs. Include:
- **Chronological order**: Events in sequence
- **Key decisions**: Player choices and reasoning
- **Roll outcomes**: Important dice rolls and results
- **NPC interactions**: Dialogue and relationship developments
- **Combat details**: Who fought, tactics used, outcomes
- **Discoveries**: Clues, items, locations revealed
- **Consequences**: How actions changed the world

### What Gets Compressed

The AI compression preserves:
- Major plot points and story beats
- Significant character actions
- Important discoveries and revelations
- Quest progress and milestones
- Key NPC interactions
- Combat outcomes (not every roll)

The compression discards:
- Minor flavor text
- Routine skill checks
- Detailed combat turn mechanics
- Trivial conversations
- Redundant descriptions

### NPCs to Save

**Always Save**:
- Named NPCs with speaking roles
- Recurring NPCs
- Quest-givers
- Bosses and lieutenants
- NPCs with relationships to PCs

**Optional**:
- Generic enemies (Goblin #3)
- One-off merchants
- Unnamed townsfolk
- Dead NPCs with no future role

### Character Updates

Ensure final HP accounts for:
- Short rests taken
- Long rests taken
- Healing potions used
- Spell healing received

Ensure conditions reflect:
- Remove temporary effects (Bless, Bane, etc.)
- Keep persistent conditions (Disease, Curse, etc.)
- Update status (dead if HP reached 0 and failed saves)

## Required Tools

This skill uses these MCP tools:
- `save_character` (for uncaptured NPCs)
- `save_plot_event` (for uncaptured events)
- `resolve_hook` (for completed hooks)
- `set_world_flag` (for world state)
- `update_character` (for final character state)
- `end_session` (to compress and save)
- `export_session_recap` (optional)
- `list_open_hooks` (to review hooks)

## Tips for DM

- Keep a mental or written note of NPCs as they appear during the session
- Save plot events as they happen, not just at the end
- Use checkpoints during the session to reduce end-session work
- Write the transcript chronologically while session is fresh
- Be thorough in the transcript - it's the only record that gets compressed
- DM notes can include meta-information, future plans, or private observations
