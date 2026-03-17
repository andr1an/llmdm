---
name: create-campaign
description: Create a new D&D campaign with a player character and start the first session
version: 1.0.0
---

# Create Campaign Skill

This skill creates a brand new D&D 5e campaign, sets up the player character, and starts the first session.

## When to Use

Use this skill when:
- Starting a completely new D&D campaign
- The player wants to create their first character
- Beginning a fresh adventure

## Workflow

Follow these steps in order:

### 1. Create the Campaign

Use `mcp__dnd-campaign__create_campaign` with:
- `name`: Ask the player for a campaign name (e.g., "Lost Mines of Phandelver", "Curse of Strahd", "Custom Adventure")
- `description`: Optional setting description

Store the returned `campaign_id` for subsequent operations.

### 2. Create the Player Character

Gather character information from the player:
- **Name**: Character's name
- **Race**: Elf, Dwarf, Human, Halfling, etc.
- **Class**: Fighter, Wizard, Rogue, Cleric, etc.
- **Background**: Brief backstory
- **Ability Scores**: STR, DEX, CON, INT, WIS, CHA (offer to roll 4d6kh3 six times if needed)
- **HP**: Based on class hit die + CON modifier
- **AC**: Based on armor and DEX modifier
- **Equipment**: Starting equipment for class

Use `mcp__dnd-campaign__save_character` to create the PC with full details including:
- Basic info (name, race, class, level 1)
- Ability scores
- HP (current and max)
- AC and speed
- Proficiencies (armor, weapons, tools, saving throws, skills)
- Languages
- Features (racial and class features)
- Spellcasting (if applicable)
- Starting inventory
- Backstory

### 3. Set Initial World Flags

Use `mcp__dnd-campaign__set_world_flag` to establish initial campaign state:
- `"campaign_start_date"`: Current in-game date
- `"days_elapsed"`: "0"
- Any campaign-specific flags (e.g., `"main_quest_started": "false"`)

### 4. Start Session 1

Use `mcp__dnd-campaign__start_session` with:
- `campaign_id`: The campaign ID from step 1
- `session`: 1
- `recent_sessions`: 3 (default, though this is session 1)

This will return a session brief (though empty for first session).

### 5. Begin the Adventure

Present the opening scene:
- Describe the setting vividly
- Introduce the character's situation
- Set up the initial hook or conflict
- Ask: "What do you do?"

### 6. During First Session

Follow standard D&D gameplay:
- Describe situations
- Handle player actions with `mcp__dnd-campaign__roll` for all dice rolls
- Use `mcp__dnd-campaign__checkpoint` for major events
- Update character with `mcp__dnd-campaign__update_character` for HP/condition changes
- Create NPCs with `mcp__dnd-campaign__save_character` as they appear
- Record major events with `mcp__dnd-campaign__save_plot_event`

## Example Interaction

**Player**: "I want to start a new D&D campaign"

**Claude**: "Great! Let's create a new campaign. What would you like to call it?"

**Player**: "The Dragon's Lair"

**Claude**: [Creates campaign] "Perfect! Now let's create your character. What's your character's name, race, and class?"

[Continue character creation dialog...]

[After character is created]

**Claude**: [Starts session 1] "You awaken in a tavern called The Prancing Pony. The smell of ale and roasted meat fills the air. A cloaked figure in the corner catches your eye. What do you do?"

## Important Notes

- Always use MCP tools for ALL dice rolls - NEVER simulate
- Create NPCs immediately when they appear in the story
- Checkpoint liberally during first session
- Save plot events as they happen
- Be descriptive and engaging
- Follow D&D 5e rules accurately

## Required Tools

This skill uses these MCP tools:
- `create_campaign`
- `save_character`
- `set_world_flag`
- `start_session`
- `roll`
- `checkpoint`
- `update_character`
- `save_plot_event`
