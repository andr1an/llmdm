# Character Creation Guide

Complete guide to creating D&D 5e character sheets using the `save_character` tool.

## Overview

The MCP server supports full D&D 5e character sheets with all standard fields including ability scores, proficiencies, skills, features, and spellcasting.

## Character Type

Every character must specify:

```json
{
  "type": "pc"  // or "npc"
}
```

- **PC (Player Character)**: Characters controlled by players
- **NPC (Non-Player Character)**: DM-controlled characters

## Required Fields

Only these fields are required:

```json
{
  "campaign_id": "camp_abc123",
  "name": "Character Name",
  "type": "pc",
  "hp_current": 12.0,
  "hp_max": 12.0
}
```

All other fields have sensible defaults or are optional.

## Core Statistics

### Basic Info

```json
{
  "name": "Thorin Oakenshield",
  "type": "pc",
  "class": "Fighter",
  "race": "Dwarf",
  "level": 5,
  "alignment": "Lawful Good"
}
```

**Defaults:**
- `level`: 1
- `alignment`: Empty (can use any D&D 5e alignment)

**Valid Alignments:**
- `"Lawful Good"`, `"Neutral Good"`, `"Chaotic Good"`
- `"Lawful Neutral"`, `"True Neutral"`, `"Chaotic Neutral"`
- `"Lawful Evil"`, `"Neutral Evil"`, `"Chaotic Evil"`

### Hit Points

```json
{
  "hp_current": 38.0,
  "hp_max": 45.0
}
```

Both required. Use floats to support temporary HP or fractional damage.

### Ability Scores

```json
{
  "str": 16,
  "dex": 12,
  "con": 15,
  "int_stat": 10,  // Note: "int_stat" not "int" (reserved keyword)
  "wis": 11,
  "cha": 13
}
```

All optional. Standard D&D 5e ability scores (3-20 typical range, no hard limit).

### Combat Stats

```json
{
  "ac": 18,
  "speed": "25 ft"
}
```

**Defaults:**
- `ac`: 10 (unarmored)
- `speed`: `"30 ft"` (standard humanoid)

**Speed Examples:**
- `"30 ft"` - Standard humanoid
- `"25 ft"` - Dwarf, heavily armored
- `"40 ft"` - Monk, Barbarian
- `"30 ft, fly 60 ft"` - Flying creatures
- `"30 ft, swim 30 ft"` - Aquatic creatures

### Experience

```json
{
  "experience_points": 6500,
  "gold": 250
}
```

**Defaults:**
- `experience_points`: 0
- `gold`: 0

## Proficiencies

Proficiencies object has four arrays:

```json
{
  "proficiencies": {
    "armor": ["Light Armor", "Medium Armor", "Heavy Armor", "Shields"],
    "weapons": ["Simple Weapons", "Martial Weapons"],
    "tools": ["Smith's Tools", "Thieves' Tools"],
    "saving_throws": ["STR", "CON"],
    "skills": ["Athletics", "Intimidation", "Perception"]
  }
}
```

### Armor Proficiencies

**Standard values:**
- `"Light Armor"` - Rogues, Rangers, light classes
- `"Medium Armor"` - Clerics, Druids, mid-weight classes
- `"Heavy Armor"` - Fighters, Paladins, heavy classes
- `"Shields"` - Most martial classes

**Specific armor:**
- `"Padded"`, `"Leather"`, `"Studded Leather"` (Light)
- `"Hide"`, `"Chain Shirt"`, `"Scale Mail"`, `"Breastplate"`, `"Half Plate"` (Medium)
- `"Ring Mail"`, `"Chain Mail"`, `"Splint"`, `"Plate"` (Heavy)

### Weapon Proficiencies

**Standard values:**
- `"Simple Weapons"` - Most classes
- `"Martial Weapons"` - Fighters, Paladins, Rangers, Barbarians

**Specific weapons:**
- `"Longsword"`, `"Shortsword"`, `"Rapier"`
- `"Longbow"`, `"Shortbow"`, `"Hand Crossbow"`
- Any specific weapon name

### Tool Proficiencies

**Common tools:**
- `"Thieves' Tools"` - Rogues
- `"Smith's Tools"`, `"Carpenter's Tools"` - Artisan tools
- `"Herbalism Kit"`, `"Poisoner's Kit"` - Kits
- `"Lute"`, `"Flute"`, `"Drum"` - Musical instruments

### Saving Throw Proficiencies

Must use uppercase ability codes:

```json
{
  "saving_throws": ["STR", "CON"]
}
```

**Valid values:** `"STR"`, `"DEX"`, `"CON"`, `"INT"`, `"WIS"`, `"CHA"`

**Class saving throws:**
- Fighter: `["STR", "CON"]`
- Wizard: `["INT", "WIS"]`
- Rogue: `["DEX", "INT"]`
- Cleric: `["WIS", "CHA"]`

### Skill Proficiencies

List of skill names:

```json
{
  "skills": ["Athletics", "Intimidation", "Perception"]
}
```

**All D&D 5e skills:**
- **STR**: Athletics
- **DEX**: Acrobatics, Sleight of Hand, Stealth
- **INT**: Arcana, History, Investigation, Nature, Religion
- **WIS**: Animal Handling, Insight, Medicine, Perception, Survival
- **CHA**: Deception, Intimidation, Performance, Persuasion

## Skills Array

Detailed skill modifiers with proficiency:

```json
{
  "skills": [
    {
      "name": "Perception",
      "proficient": true,
      "modifier": 5
    },
    {
      "name": "Athletics",
      "proficient": true,
      "modifier": 7
    },
    {
      "name": "Stealth",
      "proficient": false,
      "modifier": 1
    }
  ]
}
```

**Fields:**
- `name`: Skill name (see list above)
- `proficient`: Is character proficient? (adds proficiency bonus)
- `modifier`: Total modifier (ability mod + proficiency if applicable)

**Calculating modifiers:**
```
Modifier = Ability Modifier + (Proficiency Bonus if proficient)

Example (Level 5, WIS 14, Perception proficient):
- Ability Mod: +2 (WIS 14)
- Proficiency Bonus: +3 (Level 5)
- Total: +5
```

## Languages

Simple string array:

```json
{
  "languages": ["Common", "Dwarvish", "Elvish"]
}
```

**Common D&D languages:**
- **Standard**: Common, Dwarvish, Elvish, Giant, Gnomish, Goblin, Halfling, Orc
- **Exotic**: Abyssal, Celestial, Draconic, Deep Speech, Infernal, Primordial, Sylvan, Undercommon

## Features and Traits

Array of feature objects:

```json
{
  "features": [
    {
      "name": "Darkvision",
      "description": "You can see in dim light within 60 feet of you as if it were bright light",
      "source": "race"
    },
    {
      "name": "Action Surge",
      "description": "On your turn, you can take one additional action",
      "source": "class"
    },
    {
      "name": "Lucky",
      "description": "When you roll a 1 on an attack roll, ability check, or saving throw, you can reroll the die",
      "source": "feat"
    }
  ]
}
```

**Source values:**
- `"race"` - Racial traits (Darkvision, Fey Ancestry, etc.)
- `"class"` - Class features (Action Surge, Rage, Sneak Attack, etc.)
- `"feat"` - Feats (Lucky, Alert, Great Weapon Master, etc.)
- `"background"` - Background features
- `"other"` - Magic items, boons, etc.

## Spellcasting (Optional)

Only include for spellcasting classes. Null/omit for non-casters.

```json
{
  "spellcasting": {
    "ability": "INT",
    "spell_slots": {
      "1": 4,
      "2": 3,
      "3": 2
    },
    "cantrips": ["Fire Bolt", "Mage Hand", "Prestidigitation"],
    "prepared_spells": ["Shield", "Magic Missile", "Detect Magic", "Fireball", "Haste"]
  }
}
```

### Spellcasting Ability

```json
{
  "ability": "INT"  // or "WIS" or "CHA"
}
```

**By class:**
- **INT**: Wizard, Artificer
- **WIS**: Cleric, Druid, Ranger
- **CHA**: Bard, Sorcerer, Warlock, Paladin

### Spell Slots

Map of spell level to number of slots:

```json
{
  "spell_slots": {
    "1": 4,
    "2": 3,
    "3": 3,
    "4": 3,
    "5": 2
  }
}
```

**Keys:** Spell level as string (`"1"` through `"9"`)
**Values:** Number of slots (integer)

**Example progression (Wizard level 9):**
```json
{
  "spell_slots": {
    "1": 4,
    "2": 3,
    "3": 3,
    "4": 3,
    "5": 1
  }
}
```

### Cantrips

String array of known cantrips:

```json
{
  "cantrips": ["Fire Bolt", "Light", "Mage Hand", "Prestidigitation"]
}
```

Cantrips have unlimited uses.

### Prepared Spells

String array of prepared/known spells:

```json
{
  "prepared_spells": [
    "Shield",
    "Magic Missile",
    "Detect Magic",
    "Fireball",
    "Counterspell"
  ]
}
```

**Note:** Some classes (Wizards) prepare spells daily. Others (Sorcerers, Warlocks) have fixed known spells.

## Character Status

```json
{
  "status": "active"
}
```

**Valid values:**
- `"active"` - Default, character is active in campaign
- `"dead"` - Character has died
- `"missing"` - Character's whereabouts unknown
- `"retired"` - Character retired from adventuring

## Dynamic State

These fields track in-game state changes:

```json
{
  "inventory": ["Longsword", "Shield", "Plate Armor", "Potion of Healing (x2)"],
  "conditions": ["Poisoned", "Frightened"],
  "plot_flags": ["knows_about_black_spider", "has_map_to_wave_echo_cave"],
  "relationships": {
    "Sildar Hallwinter": "ally",
    "Gundren Rockseeker": "employer",
    "The Black Spider": "enemy"
  }
}
```

All optional, default to empty arrays/objects.

### Inventory

String array of items:

```json
{
  "inventory": [
    "Longsword +1",
    "Shield",
    "Plate Armor",
    "Potion of Healing (x3)",
    "Rope (50 ft)",
    "Torch (x5)",
    "Backpack"
  ]
}
```

### Conditions

String array of active conditions:

```json
{
  "conditions": ["Poisoned", "Frightened", "Exhaustion (1)"]
}
```

**Common D&D 5e conditions:**
- Blinded, Charmed, Deafened, Exhaustion, Frightened, Grappled
- Incapacitated, Invisible, Paralyzed, Petrified, Poisoned, Prone
- Restrained, Stunned, Unconscious

### Plot Flags

String array of narrative flags:

```json
{
  "plot_flags": [
    "knows_about_black_spider",
    "has_map_to_wave_echo_cave",
    "saved_sildar",
    "suspicious_of_halia"
  ]
}
```

Use for character-specific plot tracking.

### Relationships

Map of character names to relationship strings:

```json
{
  "relationships": {
    "Sildar Hallwinter": "trusted ally",
    "Gundren Rockseeker": "employer and friend",
    "The Black Spider": "sworn enemy",
    "Sister Garaele": "friendly contact",
    "Halia Thornton": "business contact, suspicious"
  }
}
```

## Backstory and Notes

```json
{
  "backstory": "Thorin was once a prince of the Iron Hills, but was exiled after...",
  "notes": "Secret: Thorin is actually searching for the Arkenstone. Don't reveal until session 10."
}
```

**Limits:**
- `backstory`: 8000 characters (for player-facing lore)
- `notes`: 4000 characters (DM-only secrets)

## Complete Examples

### Example 1: Level 1 Fighter (Minimal)

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
  "con": 14,
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
  "languages": ["Common", "Dwarvish"],
  "gold": 50,
  "status": "active"
}
```

### Example 2: Level 5 Wizard (With Spellcasting)

```json
{
  "campaign_id": "camp_abc123",
  "name": "Gandalf the Grey",
  "type": "npc",
  "class": "Wizard",
  "race": "Human",
  "level": 5,
  "hp_current": 28.0,
  "hp_max": 28.0,
  "str": 8,
  "dex": 14,
  "con": 12,
  "int_stat": 18,
  "wis": 14,
  "cha": 12,
  "ac": 12,
  "speed": "30 ft",
  "proficiencies": {
    "armor": [],
    "weapons": ["Daggers", "Darts", "Slings", "Quarterstaffs", "Light Crossbows"],
    "tools": [],
    "saving_throws": ["INT", "WIS"],
    "skills": ["Arcana", "History", "Investigation"]
  },
  "skills": [
    {"name": "Arcana", "proficient": true, "modifier": 7},
    {"name": "History", "proficient": true, "modifier": 7},
    {"name": "Investigation", "proficient": true, "modifier": 7},
    {"name": "Perception", "proficient": false, "modifier": 2}
  ],
  "languages": ["Common", "Draconic", "Elvish", "Dwarvish"],
  "features": [
    {
      "name": "Arcane Recovery",
      "description": "Once per day, recover spell slots on short rest",
      "source": "class"
    }
  ],
  "spellcasting": {
    "ability": "INT",
    "spell_slots": {
      "1": 4,
      "2": 3,
      "3": 2
    },
    "cantrips": ["Fire Bolt", "Light", "Mage Hand", "Prestidigitation"],
    "prepared_spells": [
      "Shield",
      "Magic Missile",
      "Detect Magic",
      "Fireball",
      "Counterspell",
      "Haste"
    ]
  },
  "gold": 200,
  "inventory": ["Quarterstaff", "Spellbook", "Component Pouch", "Robe"],
  "status": "active",
  "notes": "Plot-critical NPC, cannot die"
}
```

### Example 3: NPC Villain (Minimal)

```json
{
  "campaign_id": "camp_abc123",
  "name": "The Black Spider",
  "type": "npc",
  "class": "Rogue",
  "race": "Drow",
  "level": 8,
  "hp_current": 55.0,
  "hp_max": 55.0,
  "str": 10,
  "dex": 18,
  "con": 12,
  "int_stat": 14,
  "wis": 13,
  "cha": 16,
  "ac": 17,
  "status": "active",
  "notes": "Main campaign villain. Wants to control Wave Echo Cave. Drow supremacist with spider familiar."
}
```

## Tips

1. **Start minimal** - Only provide required fields initially
2. **Use update_character** - Incrementally add details as needed
3. **Spellcasting is optional** - Omit for non-casters to keep it clean
4. **Skill modifiers** - Pre-calculate and store for quick reference
5. **Notes field** - Use for DM secrets, not player-facing info
6. **Relationships** - Update as the story progresses
7. **Status tracking** - Mark characters as dead/missing when appropriate

## See Also

- [`save_character`](../tools/campaign-memory.md#save_character) - Tool reference
- [`update_character`](../tools/campaign-memory.md#update_character) - Partial updates
- [Quick Start Guide](./quick-start.md) - Get started quickly
