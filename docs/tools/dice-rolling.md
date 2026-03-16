# Dice Rolling Tools

Complete dice rolling system supporting D&D 5e mechanics with automatic logging.

## Tools Overview

| Tool | Description |
|------|-------------|
| `roll` | Standard dice rolling with advantage/disadvantage support |
| `roll_contested` | Contested checks between attacker and defender |
| `roll_saving_throw` | Saving throws with DC validation |
| `get_roll_history` | Query roll history with filters |

---

## roll

Roll dice using standard D&D notation. Supports advantage, disadvantage, keep highest/lowest mechanics. All rolls are automatically logged.

### Input Parameters

| Parameter | Type | Required | Description | Example |
|-----------|------|----------|-------------|---------|
| `campaign_id` | string | Yes | Campaign ID for logging the roll | `"camp_abc123"` |
| `notation` | string | Yes | Dice notation (e.g. '1d20' '2d6+3' '4d6kh3') | `"1d20+5"` |
| `reason` | string | No | Reason for the roll | `"Stealth check"` |
| `character` | string | No | Character making the roll | `"Gandalf"` |
| `session` | int | No | Current session number | `5` |
| `advantage` | bool | No | Roll with advantage (2d20 keep highest) | `true` |
| `disadvantage` | bool | No | Roll with disadvantage (2d20 keep lowest) | `false` |

### Supported Notation

- **Basic**: `NdM` - Roll N dice with M sides (e.g., `1d20`, `2d6`)
- **With modifier**: `NdM+K` or `NdM-K` - Add/subtract modifier (e.g., `1d20+5`, `2d8-2`)
- **Keep highest**: `NdMkhX` - Roll N dice, keep X highest (e.g., `4d6kh3`)
- **Keep lowest**: `NdMklX` - Roll N dice, keep X lowest (e.g., `4d6kl1`)
- **Advantage**: Set `advantage: true` (equivalent to `2d20kh1`)
- **Disadvantage**: Set `disadvantage: true` (equivalent to `2d20kl1`)

### Input Schema

```json
{
  "campaign_id": "camp_abc123",
  "notation": "1d20+5",
  "reason": "Attack roll against goblin",
  "character": "Thorin",
  "session": 3,
  "advantage": false,
  "disadvantage": false
}
```

### Output Schema

```json
{
  "roll_result": {
    "total": 18,
    "rolls": [13],
    "kept": [],
    "modifier": 5,
    "notation": "1d20+5",
    "roll_id": "roll_xyz789",
    "timestamp": "2026-03-16T10:30:00Z"
  }
}
```

**With advantage example:**

```json
{
  "roll_result": {
    "total": 22,
    "rolls": [17, 9],
    "kept": [17],
    "modifier": 5,
    "notation": "1d20+5",
    "roll_id": "roll_xyz790",
    "timestamp": "2026-03-16T10:31:00Z"
  }
}
```

### Notes

- All rolls are logged to the campaign database
- `advantage` and `disadvantage` cannot both be true (mutual exclusion)
- When advantage/disadvantage is used, `rolls` contains both d20 results and `kept` shows which was used
- The `roll_id` is unique and can be used for audit trails
- Session number is optional but recommended for turn-based tracking

### See Also

- [`roll_contested`](#roll_contested) - For opposed checks
- [`roll_saving_throw`](#roll_saving_throw) - For DC-based saves
- [`get_roll_history`](#get_roll_history) - To query logged rolls

---

## roll_contested

Roll a contested check between attacker and defender. Both rolls are logged separately.

### Input Parameters

| Parameter | Type | Required | Description | Example |
|-----------|------|----------|-------------|---------|
| `campaign_id` | string | Yes | Campaign ID | `"camp_abc123"` |
| `attacker` | string | Yes | Attacker's name | `"Thorin"` |
| `defender` | string | Yes | Defender's name | `"Goblin Chief"` |
| `attacker_notation` | string | Yes | Dice notation for attacker | `"1d20+5"` |
| `defender_notation` | string | Yes | Dice notation for defender | `"1d20+3"` |
| `contest_type` | string | Yes | Type of contest | `"Attack vs AC"` |
| `session` | int | No | Current session number | `5` |

### Common Contest Types

- `"Attack vs AC"` - Attack roll against armor class
- `"Deception vs Insight"` - Social skill contests
- `"Stealth vs Perception"` - Hide vs spot checks
- `"Grapple vs Athletics"` - Physical contests
- `"Sleight of Hand vs Perception"` - Theft attempts

### Input Schema

```json
{
  "campaign_id": "camp_abc123",
  "attacker": "Thorin",
  "defender": "Goblin Chief",
  "attacker_notation": "1d20+7",
  "defender_notation": "1d20+2",
  "contest_type": "Attack vs AC",
  "session": 3
}
```

### Output Schema

```json
{
  "result": {
    "winner": "Thorin",
    "attacker_result": {
      "total": 19,
      "rolls": [12],
      "kept": [],
      "modifier": 7,
      "notation": "1d20+7",
      "roll_id": "roll_att001",
      "timestamp": "2026-03-16T10:32:00Z"
    },
    "defender_result": {
      "total": 14,
      "rolls": [12],
      "kept": [],
      "modifier": 2,
      "notation": "1d20+2",
      "roll_id": "roll_def001",
      "timestamp": "2026-03-16T10:32:00Z"
    },
    "margin": 5
  }
}
```

### Notes

- Both rolls are logged with their respective characters and reasons
- `winner` field contains the name of the character with the higher total
- `margin` shows the difference between the winning and losing totals
- In case of a tie, the defender wins (standard D&D rule)
- Each roll gets a unique `roll_id` for tracking

### See Also

- [`roll`](#roll) - For standard non-contested rolls
- [`get_roll_history`](#get_roll_history) - To review contest results

---

## roll_saving_throw

Roll a saving throw and check against a DC (Difficulty Class).

### Input Parameters

| Parameter | Type | Required | Description | Example |
|-----------|------|----------|-------------|---------|
| `campaign_id` | string | Yes | Campaign ID | `"camp_abc123"` |
| `character` | string | Yes | Character making the saving throw | `"Gandalf"` |
| `stat` | string | Yes | Ability score to use (STR/DEX/CON/INT/WIS/CHA) | `"DEX"` |
| `modifier` | float | Yes | Total modifier for the saving throw | `5.0` |
| `dc` | float | Yes | Difficulty class to beat | `15.0` |
| `reason` | string | No | Reason for the saving throw | `"Fireball trap"` |
| `session` | int | No | Current session number | `5` |

### Valid Stat Values

- `"STR"` - Strength saving throw
- `"DEX"` - Dexterity saving throw
- `"CON"` - Constitution saving throw
- `"INT"` - Intelligence saving throw
- `"WIS"` - Wisdom saving throw
- `"CHA"` - Charisma saving throw

### Input Schema

```json
{
  "campaign_id": "camp_abc123",
  "character": "Gandalf",
  "stat": "DEX",
  "modifier": 3.0,
  "dc": 15.0,
  "reason": "Dodge fireball",
  "session": 3
}
```

### Output Schema

```json
{
  "result": {
    "total": 18,
    "rolled": 15,
    "modifier": 3,
    "dc": 15,
    "success": true,
    "roll_id": "roll_save001"
  }
}
```

**Failed save example:**

```json
{
  "result": {
    "total": 12,
    "rolled": 9,
    "modifier": 3,
    "dc": 15,
    "success": false,
    "roll_id": "roll_save002"
  }
}
```

### Notes

- `success` is `true` if `total >= dc`, otherwise `false`
- The roll is automatically logged with the saving throw details
- `modifier` should include ability modifier + proficiency bonus (if proficient)
- Natural 1 is not an automatic failure for saving throws in D&D 5e
- Natural 20 is not an automatic success for saving throws in D&D 5e

### See Also

- [`roll`](#roll) - For general ability checks
- [`get_roll_history`](#get_roll_history) - To review saving throw results

---

## get_roll_history

Retrieve dice roll history for a campaign with optional filters.

### Input Parameters

| Parameter | Type | Required | Description | Example |
|-----------|------|----------|-------------|---------|
| `campaign_id` | string | Yes | Campaign ID | `"camp_abc123"` |
| `character` | string | No | Filter by character name | `"Thorin"` |
| `session` | int | No | Filter by session number | `3` |
| `limit` | int | No | Maximum number of rolls to return (default: 50) | `20` |

### Input Schema

```json
{
  "campaign_id": "camp_abc123",
  "character": "Thorin",
  "session": 3,
  "limit": 10
}
```

### Output Schema

```json
{
  "records": [
    {
      "id": "roll_001",
      "campaign_id": "camp_abc123",
      "session": 3,
      "character": "Thorin",
      "notation": "1d20+5",
      "total": 18,
      "rolls": [13],
      "kept": [],
      "modifier": 5,
      "reason": "Attack roll",
      "advantage": false,
      "disadvantage": false,
      "created_at": "2026-03-16T10:30:00Z"
    },
    {
      "id": "roll_002",
      "campaign_id": "camp_abc123",
      "session": 3,
      "character": "Thorin",
      "notation": "1d20+3",
      "total": 21,
      "rolls": [18, 7],
      "kept": [18],
      "modifier": 3,
      "reason": "Stealth check",
      "advantage": true,
      "disadvantage": false,
      "created_at": "2026-03-16T10:35:00Z"
    }
  ]
}
```

### Notes

- Results are ordered by `created_at` descending (most recent first)
- If no filters are provided, returns all rolls for the campaign (up to `limit`)
- Default limit is 50 to prevent overwhelming output
- `kept` array shows which dice were used when advantage/disadvantage/keep mechanics apply
- All three filters can be combined (e.g., specific character in specific session)

### See Also

- [`roll`](#roll) - To create new roll records
- [`roll_contested`](#roll_contested) - Contested rolls also appear in history
- [`roll_saving_throw`](#roll_saving_throw) - Saving throws also appear in history
