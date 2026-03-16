# Campaign Memory Tools

Comprehensive campaign state management with full D&D 5e character sheet support, plot tracking, and world state persistence.

## Tools Overview

| Tool | Description |
|------|-------------|
| `create_campaign` | Create new campaign with dedicated SQLite database |
| `list_campaigns` | List all existing campaigns |
| `save_character` | Create or fully replace a character (PC or NPC) |
| `update_character` | Patch specific fields on a character |
| `get_character` | Retrieve full character sheet by name |
| `list_characters` | List characters with type/status filters |
| `save_plot_event` | Record narrative events and create plot hooks |
| `list_open_hooks` | Get all unresolved plot threads |
| `resolve_hook` | Mark a plot hook as resolved |
| `set_world_flag` | Set key/value world state flag |
| `get_world_flags` | Retrieve all world flags for a campaign |

---

## create_campaign

Create a new D&D campaign with a dedicated SQLite database.

### Input Parameters

| Parameter | Type | Required | Description | Example |
|-----------|------|----------|-------------|---------|
| `name` | string | Yes | Campaign name | `"Lost Mines of Phandelver"` |
| `description` | string | No | Brief setting description | `"Classic D&D starter adventure"` |

### Input Schema

```json
{
  "name": "Lost Mines of Phandelver",
  "description": "Classic D&D 5e starter adventure in the Sword Coast"
}
```

### Output Schema

```json
{
  "campaign_id": "camp_abc123",
  "db_path": "/path/to/campaigns/camp_abc123.db",
  "campaign": {
    "id": "camp_abc123",
    "name": "Lost Mines of Phandelver",
    "description": "Classic D&D 5e starter adventure in the Sword Coast",
    "created_at": "2026-03-16T10:00:00Z",
    "current_session": 0
  }
}
```

### Notes

- Each campaign gets its own SQLite database file
- `campaign_id` is automatically generated (format: `camp_` + UUID)
- `current_session` starts at 0 and increments when sessions are started
- Database path is in the configured campaigns directory
- Campaign name does not need to be unique (IDs handle uniqueness)

### See Also

- [`list_campaigns`](#list_campaigns) - List all campaigns
- [`start_session`](./session-management.md#start_session) - Begin first session

---

## list_campaigns

List all existing campaigns with metadata.

### Input Parameters

No parameters required.

### Input Schema

```json
{}
```

### Output Schema

```json
{
  "campaigns": [
    {
      "id": "camp_abc123",
      "name": "Lost Mines of Phandelver",
      "description": "Classic D&D 5e starter adventure",
      "created_at": "2026-03-15T10:00:00Z",
      "current_session": 5
    },
    {
      "id": "camp_def456",
      "name": "Curse of Strahd",
      "description": "Gothic horror in Barovia",
      "created_at": "2026-03-01T14:30:00Z",
      "current_session": 12
    }
  ]
}
```

### Notes

- Returns all campaigns sorted by creation date (newest first)
- `current_session` shows the last started session number
- Empty array if no campaigns exist

### See Also

- [`create_campaign`](#create_campaign) - Create new campaign

---

## save_character

Create or fully replace a character (PC or NPC). For partial updates, use `update_character` instead.

### Input Parameters

| Parameter | Type | Required | Description | Example |
|-----------|------|----------|-------------|---------|
| `campaign_id` | string | Yes | Campaign ID | `"camp_abc123"` |
| `name` | string | Yes | Character name | `"Thorin Oakenshield"` |
| `type` | string | Yes | Character type (pc or npc) | `"pc"` |
| `class` | string | No | Character class | `"Fighter"` |
| `race` | string | No | Character race | `"Dwarf"` |
| `level` | int | No | Character level (default: 1) | `5` |
| `hp_current` | float | Yes | Current hit points | `38.0` |
| `hp_max` | float | Yes | Maximum hit points | `45.0` |
| `str` | int | No | Strength score | `16` |
| `dex` | int | No | Dexterity score | `12` |
| `con` | int | No | Constitution score | `15` |
| `int_stat` | int | No | Intelligence score | `10` |
| `wis` | int | No | Wisdom score | `11` |
| `cha` | int | No | Charisma score | `13` |
| `alignment` | string | No | D&D 5e alignment | `"Lawful Good"` |
| `ac` | int | No | Armor class (default: 10) | `18` |
| `speed` | string | No | Movement speed (default: "30 ft") | `"25 ft"` |
| `experience_points` | int | No | Experience points (default: 0) | `6500` |
| `proficiencies` | object | No | Proficiencies object | See below |
| `skills` | array | No | Skills array | See below |
| `languages` | array | No | Known languages | `["Common", "Dwarvish"]` |
| `features` | array | No | Features/traits | See below |
| `spellcasting` | object | No | Spellcasting (for casters only) | See below |
| `gold` | int | No | Gold pieces (default: 0) | `250` |
| `backstory` | string | No | Character backstory (max 8000 chars) | `"..."` |
| `status` | string | No | Character status | `"active"` |
| `notes` | string | No | DM private notes (max 4000 chars) | `"Secret villain"` |

### Valid Enum Values

- **type**: `"pc"`, `"npc"`
- **status**: `"active"`, `"dead"`, `"missing"`, `"retired"`
- **alignment**: `"Lawful Good"`, `"Neutral Good"`, `"Chaotic Good"`, `"Lawful Neutral"`, `"True Neutral"`, `"Chaotic Neutral"`, `"Lawful Evil"`, `"Neutral Evil"`, `"Chaotic Evil"`

### Proficiencies Object Structure

```json
{
  "armor": ["Light Armor", "Medium Armor", "Heavy Armor", "Shields"],
  "weapons": ["Simple Weapons", "Martial Weapons"],
  "tools": ["Smith's Tools"],
  "saving_throws": ["STR", "CON"],
  "skills": ["Athletics", "Intimidation"]
}
```

### Skills Array Structure

```json
[
  {
    "name": "Perception",
    "proficient": true,
    "modifier": 5
  },
  {
    "name": "Athletics",
    "proficient": true,
    "modifier": 7
  }
]
```

### Features Array Structure

```json
[
  {
    "name": "Darkvision",
    "description": "See in dim light within 60 feet as if bright light",
    "source": "race"
  },
  {
    "name": "Action Surge",
    "description": "Take an additional action on your turn",
    "source": "class"
  }
]
```

### Spellcasting Object Structure (Optional - Casters Only)

```json
{
  "ability": "INT",
  "spell_slots": {
    "1": 4,
    "2": 3,
    "3": 2
  },
  "cantrips": ["Fire Bolt", "Mage Hand", "Prestidigitation"],
  "prepared_spells": ["Shield", "Magic Missile", "Detect Magic", "Fireball"]
}
```

### Input Schema (Full Example - Fighter PC)

```json
{
  "campaign_id": "camp_abc123",
  "name": "Thorin Oakenshield",
  "type": "pc",
  "class": "Fighter",
  "race": "Dwarf",
  "level": 5,
  "hp_current": 38.0,
  "hp_max": 45.0,
  "str": 16,
  "dex": 12,
  "con": 15,
  "int_stat": 10,
  "wis": 11,
  "cha": 13,
  "alignment": "Lawful Good",
  "ac": 18,
  "speed": "25 ft",
  "experience_points": 6500,
  "proficiencies": {
    "armor": ["Light Armor", "Medium Armor", "Heavy Armor", "Shields"],
    "weapons": ["Simple Weapons", "Martial Weapons"],
    "tools": ["Smith's Tools"],
    "saving_throws": ["STR", "CON"],
    "skills": ["Athletics", "Intimidation"]
  },
  "skills": [
    {"name": "Athletics", "proficient": true, "modifier": 7},
    {"name": "Intimidation", "proficient": true, "modifier": 5},
    {"name": "Perception", "proficient": false, "modifier": 0}
  ],
  "languages": ["Common", "Dwarvish"],
  "features": [
    {
      "name": "Darkvision",
      "description": "See in dim light within 60 feet",
      "source": "race"
    },
    {
      "name": "Action Surge",
      "description": "Take an additional action",
      "source": "class"
    }
  ],
  "gold": 250,
  "backstory": "Exiled prince seeking to reclaim his homeland...",
  "status": "active"
}
```

### Input Schema (Wizard with Spellcasting)

```json
{
  "campaign_id": "camp_abc123",
  "name": "Gandalf the Grey",
  "type": "npc",
  "class": "Wizard",
  "race": "Human",
  "level": 18,
  "hp_current": 120.0,
  "hp_max": 120.0,
  "str": 10,
  "dex": 14,
  "con": 14,
  "int_stat": 20,
  "wis": 18,
  "cha": 16,
  "alignment": "Neutral Good",
  "ac": 15,
  "speed": "30 ft",
  "proficiencies": {
    "armor": [],
    "weapons": ["Daggers", "Darts", "Slings", "Quarterstaffs", "Light Crossbows"],
    "tools": [],
    "saving_throws": ["INT", "WIS"],
    "skills": ["Arcana", "History", "Investigation", "Insight"]
  },
  "spellcasting": {
    "ability": "INT",
    "spell_slots": {
      "1": 4,
      "2": 3,
      "3": 3,
      "4": 3,
      "5": 3,
      "6": 2,
      "7": 2,
      "8": 1,
      "9": 1
    },
    "cantrips": ["Fire Bolt", "Light", "Mage Hand", "Prestidigitation"],
    "prepared_spells": ["Shield", "Counterspell", "Fireball", "Teleport"]
  },
  "gold": 5000,
  "status": "active",
  "notes": "Secret Maiar, plot-critical NPC"
}
```

### Output Schema

```json
{
  "character": {
    "id": "char_xyz789",
    "campaign_id": "camp_abc123",
    "name": "Thorin Oakenshield",
    "type": "pc",
    "class": "Fighter",
    "race": "Dwarf",
    "level": 5,
    "hp": {
      "current": 38,
      "max": 45
    },
    "stats": {
      "STR": 16,
      "DEX": 12,
      "CON": 15,
      "INT": 10,
      "WIS": 11,
      "CHA": 13
    },
    "alignment": "Lawful Good",
    "ac": 18,
    "speed": "25 ft",
    "experience_points": 6500,
    "proficiencies": {
      "armor": ["Light Armor", "Medium Armor", "Heavy Armor", "Shields"],
      "weapons": ["Simple Weapons", "Martial Weapons"],
      "tools": ["Smith's Tools"],
      "saving_throws": ["STR", "CON"],
      "skills": ["Athletics", "Intimidation"]
    },
    "skills": [
      {"name": "Athletics", "proficient": true, "modifier": 7}
    ],
    "languages": ["Common", "Dwarvish"],
    "features": [
      {"name": "Darkvision", "description": "...", "source": "race"}
    ],
    "spellcasting": null,
    "gold": 250,
    "backstory": "Exiled prince...",
    "inventory": [],
    "conditions": [],
    "relationships": {},
    "plot_flags": [],
    "notes": "",
    "status": "active",
    "updated_at": "2026-03-16T11:00:00Z"
  }
}
```

### Notes

- This is a **full replacement** operation - if character exists, all fields are overwritten
- For partial updates (e.g., just HP or gold), use `update_character` instead
- `backstory` has a 8000 character limit
- `notes` (DM-only) has a 4000 character limit
- Default values: `level=1`, `ac=10`, `speed="30 ft"`, `gold=0`, `experience_points=0`
- `spellcasting` is optional and should only be provided for spellcasting classes
- Character names must be unique within a campaign
- Empty arrays/objects are valid for `proficiencies`, `skills`, `languages`, `features`

### See Also

- [`update_character`](#update_character) - For partial field updates
- [`get_character`](#get_character) - Retrieve character sheet
- [Character Creation Guide](../guides/character-creation.md) - Detailed character creation walkthrough

---

## update_character

Patch specific fields on a character. Only provided fields are updated.

### Input Parameters

| Parameter | Type | Required | Description | Example |
|-----------|------|----------|-------------|---------|
| `campaign_id` | string | Yes | Campaign ID | `"camp_abc123"` |
| `name` | string | Yes | Character name | `"Thorin Oakenshield"` |
| `hp_current` | float | No | New current HP | `25.0` |
| `level` | int | No | New level | `6` |
| `gold` | int | No | New gold amount | `500` |
| `status` | string | No | New status (active/dead/missing/retired) | `"dead"` |
| `notes` | string | No | New DM notes | `"Killed by dragon"` |
| `inventory` | array | No | Replace inventory | `["Sword", "Shield"]` |
| `conditions` | array | No | Replace conditions | `["Poisoned"]` |
| `plot_flags` | array | No | Replace plot flags | `["knows_secret"]` |
| `relationships` | object | No | Replace relationships map | `{"Bilbo": "friend"}` |
| `alignment` | string | No | New alignment | `"Chaotic Good"` |
| `ac` | int | No | New armor class | `19` |
| `speed` | string | No | New speed | `"30 ft"` |
| `experience_points` | int | No | New XP | `14000` |
| `proficiencies` | object | No | Replace proficiencies | See `save_character` |
| `skills` | array | No | Replace skills array | See `save_character` |
| `languages` | array | No | Replace languages | `["Common", "Elvish"]` |
| `features` | array | No | Replace features | See `save_character` |
| `spellcasting` | object | No | Replace spellcasting | See `save_character` |

### Input Schema (Common Use Case - Update HP and Gold)

```json
{
  "campaign_id": "camp_abc123",
  "name": "Thorin Oakenshield",
  "hp_current": 25.0,
  "gold": 500
}
```

### Input Schema (Mark Character as Dead)

```json
{
  "campaign_id": "camp_abc123",
  "name": "Thorin Oakenshield",
  "status": "dead",
  "notes": "Killed in battle with Smaug",
  "hp_current": 0.0
}
```

### Input Schema (Add Condition)

```json
{
  "campaign_id": "camp_abc123",
  "name": "Gandalf",
  "conditions": ["Poisoned", "Frightened"]
}
```

### Output Schema

```json
{
  "character": {
    "id": "char_xyz789",
    "campaign_id": "camp_abc123",
    "name": "Thorin Oakenshield",
    "type": "pc",
    "level": 5,
    "hp": {
      "current": 25,
      "max": 45
    },
    "gold": 500,
    "status": "active",
    "updated_at": "2026-03-16T11:30:00Z"
  }
}
```

### Notes

- Only provided fields are modified - all other fields remain unchanged
- Arrays and objects are **replaced**, not merged (e.g., providing `inventory` replaces entire inventory)
- Character must exist (will error if not found)
- `name` cannot be changed (it's the identifier)
- Use null or omit fields you don't want to change

### See Also

- [`save_character`](#save_character) - For full character creation/replacement
- [`get_character`](#get_character) - Retrieve current state before updating

---

## get_character

Get full character sheet by name.

### Input Parameters

| Parameter | Type | Required | Description | Example |
|-----------|------|----------|-------------|---------|
| `campaign_id` | string | Yes | Campaign ID | `"camp_abc123"` |
| `name` | string | Yes | Character name | `"Thorin Oakenshield"` |

### Input Schema

```json
{
  "campaign_id": "camp_abc123",
  "name": "Thorin Oakenshield"
}
```

### Output Schema

```json
{
  "character": {
    "id": "char_xyz789",
    "campaign_id": "camp_abc123",
    "name": "Thorin Oakenshield",
    "type": "pc",
    "class": "Fighter",
    "race": "Dwarf",
    "level": 5,
    "hp": {
      "current": 38,
      "max": 45
    },
    "stats": {
      "STR": 16,
      "DEX": 12,
      "CON": 15,
      "INT": 10,
      "WIS": 11,
      "CHA": 13
    },
    "alignment": "Lawful Good",
    "ac": 18,
    "speed": "25 ft",
    "experience_points": 6500,
    "proficiencies": {
      "armor": ["Light Armor", "Medium Armor", "Heavy Armor"],
      "weapons": ["Simple Weapons", "Martial Weapons"],
      "tools": ["Smith's Tools"],
      "saving_throws": ["STR", "CON"],
      "skills": ["Athletics", "Intimidation"]
    },
    "skills": [
      {"name": "Athletics", "proficient": true, "modifier": 7}
    ],
    "languages": ["Common", "Dwarvish"],
    "features": [
      {"name": "Darkvision", "description": "...", "source": "race"}
    ],
    "spellcasting": null,
    "gold": 250,
    "backstory": "Exiled prince...",
    "inventory": ["Longsword", "Shield", "Plate Armor"],
    "conditions": [],
    "relationships": {
      "Bilbo": "friend",
      "Smaug": "enemy"
    },
    "plot_flags": ["quest_accepted"],
    "notes": "",
    "status": "active",
    "updated_at": "2026-03-16T11:00:00Z"
  }
}
```

### Notes

- Returns complete character sheet with all fields
- Errors if character not found
- `notes` field is DM-only (not intended for player view)
- `spellcasting` is null for non-spellcasters

### See Also

- [`list_characters`](#list_characters) - Get summary of multiple characters
- [`save_character`](#save_character) - Create/update characters

---

## list_characters

List characters in a campaign, optionally filtered by type and status.

### Input Parameters

| Parameter | Type | Required | Description | Example |
|-----------|------|----------|-------------|---------|
| `campaign_id` | string | Yes | Campaign ID | `"camp_abc123"` |
| `type` | string | No | Filter by type (pc or npc) | `"pc"` |
| `status` | string | No | Filter by status | `"active"` |

### Valid Filter Values

- **type**: `"pc"`, `"npc"`
- **status**: `"active"`, `"dead"`, `"missing"`, `"retired"`

### Input Schema (All Characters)

```json
{
  "campaign_id": "camp_abc123"
}
```

### Input Schema (Active PCs Only)

```json
{
  "campaign_id": "camp_abc123",
  "type": "pc",
  "status": "active"
}
```

### Output Schema

```json
{
  "characters": [
    {
      "name": "Thorin Oakenshield",
      "type": "pc",
      "class": "Fighter",
      "race": "Dwarf",
      "level": 5,
      "ac": 18,
      "hp": {
        "current": 38,
        "max": 45
      },
      "status": "active",
      "conditions": []
    },
    {
      "name": "Gandalf",
      "type": "npc",
      "class": "Wizard",
      "race": "Human",
      "level": 18,
      "ac": 15,
      "hp": {
        "current": 120,
        "max": 120
      },
      "status": "active",
      "conditions": []
    }
  ]
}
```

### Notes

- Returns condensed summaries (not full character sheets)
- Filters can be combined (e.g., active NPCs only)
- Results are sorted by name alphabetically
- Empty array if no characters match filters

### See Also

- [`get_character`](#get_character) - Get full sheet for specific character
- [`save_character`](#save_character) - Create characters

---

## save_plot_event

Record a narrative event in the campaign. Creates plot hooks if provided.

### Input Parameters

| Parameter | Type | Required | Description | Example |
|-----------|------|----------|-------------|---------|
| `campaign_id` | string | Yes | Campaign ID | `"camp_abc123"` |
| `session` | int | Yes | Session number | `3` |
| `summary` | string | Yes | 2-4 sentence narrative description | `"The party defeated..."` |
| `consequences` | string | No | What changed in the world | `"Goblins flee the area"` |
| `hooks` | array | No | Array of plot hook strings | `["Find goblin lair"]` |

### Input Schema

```json
{
  "campaign_id": "camp_abc123",
  "session": 3,
  "summary": "The party ambushed a goblin patrol on the road. After a fierce battle, they captured one goblin alive and interrogated him about a nearby lair.",
  "consequences": "Goblin patrols in the area are now on high alert. The captured goblin revealed the location of Cragmaw Hideout.",
  "hooks": [
    "Investigate Cragmaw Hideout",
    "Rescue Sildar Hallwinter from goblins"
  ]
}
```

### Output Schema

```json
{
  "event": {
    "id": "event_abc123",
    "campaign_id": "camp_abc123",
    "session": 3,
    "summary": "The party ambushed a goblin patrol...",
    "npcs_involved": [],
    "pcs_involved": [],
    "consequences": "Goblin patrols in the area...",
    "tags": [],
    "created_at": "2026-03-16T12:00:00Z"
  },
  "hooks": [
    {
      "id": "hook_001",
      "campaign_id": "camp_abc123",
      "hook": "Investigate Cragmaw Hideout",
      "session_opened": 3,
      "event_id": "event_abc123",
      "resolved": false,
      "resolution": ""
    },
    {
      "id": "hook_002",
      "campaign_id": "camp_abc123",
      "hook": "Rescue Sildar Hallwinter from goblins",
      "session_opened": 3,
      "event_id": "event_abc123",
      "resolved": false,
      "resolution": ""
    }
  ]
}
```

### Notes

- `summary` should be narrative description of what happened
- `hooks` are unresolved plot threads to track across sessions
- Each hook is created as a separate database record
- Hooks remain open until explicitly resolved with `resolve_hook`
- `npcs_involved` and `pcs_involved` are auto-extracted from summary text (not in input)
- Events are ordered chronologically for session recaps

### See Also

- [`list_open_hooks`](#list_open_hooks) - View unresolved hooks
- [`resolve_hook`](#resolve_hook) - Close a hook
- [`export_session_recap`](./session-management.md#export_session_recap) - Generate narrative recap

---

## list_open_hooks

List all unresolved plot threads for a campaign.

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
  "hooks": [
    {
      "id": "hook_001",
      "campaign_id": "camp_abc123",
      "hook": "Investigate Cragmaw Hideout",
      "session_opened": 3,
      "event_id": "event_abc123",
      "resolved": false,
      "resolution": ""
    },
    {
      "id": "hook_002",
      "campaign_id": "camp_abc123",
      "hook": "Find the missing caravan",
      "session_opened": 1,
      "event_id": "event_xyz789",
      "resolved": false,
      "resolution": ""
    }
  ]
}
```

### Notes

- Only returns hooks where `resolved = false`
- Ordered by `session_opened` (oldest first)
- Useful for DM prep to see dangling plot threads
- Empty array if all hooks are resolved

### See Also

- [`save_plot_event`](#save_plot_event) - Create new hooks
- [`resolve_hook`](#resolve_hook) - Mark hook as resolved

---

## resolve_hook

Mark a plot hook as resolved with a resolution description.

### Input Parameters

| Parameter | Type | Required | Description | Example |
|-----------|------|----------|-------------|---------|
| `campaign_id` | string | Yes | Campaign ID | `"camp_abc123"` |
| `hook_id` | string | Yes | Hook ID to resolve | `"hook_001"` |
| `resolution` | string | Yes | How the hook was resolved | `"Party cleared the hideout"` |

### Input Schema

```json
{
  "campaign_id": "camp_abc123",
  "hook_id": "hook_001",
  "resolution": "Party infiltrated Cragmaw Hideout, defeated Klarg, and rescued Sildar Hallwinter."
}
```

### Output Schema

```json
{
  "hook": {
    "id": "hook_001",
    "campaign_id": "camp_abc123",
    "hook": "Investigate Cragmaw Hideout",
    "session_opened": 3,
    "event_id": "event_abc123",
    "resolved": true,
    "resolution": "Party infiltrated Cragmaw Hideout, defeated Klarg, and rescued Sildar Hallwinter."
  }
}
```

### Notes

- Sets `resolved = true` on the hook
- `resolution` is stored for narrative history
- Resolved hooks no longer appear in `list_open_hooks`
- Can be used in session recaps to show story progression

### See Also

- [`list_open_hooks`](#list_open_hooks) - View unresolved hooks
- [`save_plot_event`](#save_plot_event) - Create new hooks

---

## set_world_flag

Set a key/value world state flag. Use for persistent world state like quest completion, faction reputation, or binary flags.

### Input Parameters

| Parameter | Type | Required | Description | Example |
|-----------|------|----------|-------------|---------|
| `campaign_id` | string | Yes | Campaign ID | `"camp_abc123"` |
| `key` | string | Yes | Flag key | `"dragons_awakened"` |
| `value` | string | Yes | Flag value | `"true"` |

### Input Schema (Boolean Flag)

```json
{
  "campaign_id": "camp_abc123",
  "key": "dragons_awakened",
  "value": "true"
}
```

### Input Schema (String Value)

```json
{
  "campaign_id": "camp_abc123",
  "key": "harpers_reputation",
  "value": "friendly"
}
```

### Input Schema (Numeric Value)

```json
{
  "campaign_id": "camp_abc123",
  "key": "days_until_ritual",
  "value": "7"
}
```

### Output Schema

```json
{
  "success": true
}
```

### Notes

- All values are stored as strings
- Use `"true"`/`"false"` for boolean flags
- Key must be unique within campaign (will overwrite existing)
- Common use cases: quest flags, faction reputation, timers, world state
- Flags persist across sessions until explicitly changed

### See Also

- [`get_world_flags`](#get_world_flags) - Retrieve all flags

---

## get_world_flags

Get all world state flags for a campaign.

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
  "flags": {
    "dragons_awakened": "true",
    "harpers_reputation": "friendly",
    "days_until_ritual": "7",
    "quest_goblin_lair_complete": "true",
    "sildar_rescued": "true"
  }
}
```

### Notes

- Returns all flags as a key-value map
- Empty object if no flags are set
- Values are always strings (parse booleans/numbers as needed)
- Useful for session briefs and DM context

### See Also

- [`set_world_flag`](#set_world_flag) - Set individual flags
- [`start_session`](./session-management.md#start_session) - Session briefs include world flags
