package memory

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/andr1an/llmdm/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDBWithNewFields(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)

	// Run schema with new fields
	schema := `
		CREATE TABLE campaigns (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT,
			created_at TEXT DEFAULT CURRENT_TIMESTAMP,
			current_session INTEGER DEFAULT 1
		);
		CREATE TABLE characters (
			id TEXT PRIMARY KEY,
			campaign_id TEXT NOT NULL REFERENCES campaigns(id),
			name TEXT NOT NULL,
			type TEXT CHECK(type IN ('pc','npc')) DEFAULT 'pc',
			class TEXT,
			race TEXT,
			level INTEGER DEFAULT 1,
			hp_current INTEGER NOT NULL,
			hp_max INTEGER NOT NULL,
			stat_str INTEGER,
			stat_dex INTEGER,
			stat_con INTEGER,
			stat_int INTEGER,
			stat_wis INTEGER,
			stat_cha INTEGER,
			alignment TEXT,
			ac INTEGER DEFAULT 10,
			speed TEXT DEFAULT '30 ft',
			experience_points INTEGER DEFAULT 0,
			proficiencies TEXT,
			skills TEXT,
			languages TEXT,
			features TEXT,
			spellcasting TEXT,
			gold INTEGER DEFAULT 0,
			backstory TEXT,
			inventory TEXT,
			conditions TEXT,
			relationships TEXT,
			plot_flags TEXT,
			notes TEXT,
			status TEXT DEFAULT 'active',
			updated_at TEXT DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(campaign_id, name)
		);
	`
	_, err = db.Exec(schema)
	require.NoError(t, err)

	// Create test campaign
	_, err = db.Exec(`INSERT INTO campaigns (id, name) VALUES ('test-campaign', 'Test Campaign')`)
	require.NoError(t, err)

	return db
}

func TestSaveCharacter_WithNewFields(t *testing.T) {
	db := setupTestDBWithNewFields(t)
	defer db.Close()

	store := NewStore(db)

	// Create a full spellcaster (wizard) with all new fields populated
	wizard := &types.Character{
		CampaignID: "test-campaign",
		Name:       "Gandalf",
		Type:       "pc",
		Class:      "Wizard",
		Race:       "Human",
		Level:      5,
		HP:         types.HP{Current: 28, Max: 28},
		Stats:      types.Stats{STR: 10, DEX: 14, CON: 12, INT: 18, WIS: 15, CHA: 11},
		Alignment:  "Chaotic Good",
		AC:         15,
		Speed:      "30 ft",
		ExperiencePoints: 6500,
		Proficiencies: types.Proficiencies{
			Armor:        []string{"Light Armor"},
			Weapons:      []string{"Simple Weapons", "Longswords"},
			Tools:        []string{},
			SavingThrows: []string{"INT", "WIS"},
			Skills:       []string{"Arcana", "History", "Insight", "Investigation"},
		},
		Skills: []types.Skill{
			{Name: "Arcana", Proficient: true, Modifier: 7},
			{Name: "History", Proficient: true, Modifier: 7},
			{Name: "Insight", Proficient: true, Modifier: 5},
			{Name: "Investigation", Proficient: true, Modifier: 7},
		},
		Languages: []string{"Common", "Elvish", "Draconic"},
		Features: []types.Feature{
			{Name: "Spellcasting", Description: "Can cast wizard spells", Source: "class"},
			{Name: "Arcane Recovery", Description: "Recover spell slots on short rest", Source: "class"},
		},
		Spellcasting: &types.Spellcasting{
			Ability: "INT",
			SpellSlots: map[string]int{
				"1": 4,
				"2": 3,
				"3": 2,
			},
			Cantrips:       []string{"Fire Bolt", "Mage Hand", "Prestidigitation"},
			PreparedSpells: []string{"Magic Missile", "Shield", "Detect Magic", "Misty Step", "Fireball"},
		},
		Gold:          150,
		Backstory:     "A wise wizard from the Grey Havens",
		Status:        "active",
		Inventory:     []string{"Spellbook", "Staff", "Component Pouch"},
		Conditions:    []string{},
		PlotFlags:     []string{},
		Relationships: map[string]string{},
	}

	err := store.SaveCharacter(wizard)
	require.NoError(t, err)

	// Retrieve and verify
	retrieved, err := store.GetCharacter("test-campaign", "Gandalf")
	require.NoError(t, err)
	require.NotNil(t, retrieved)

	assert.Equal(t, "Chaotic Good", retrieved.Alignment)
	assert.Equal(t, 15, retrieved.AC)
	assert.Equal(t, "30 ft", retrieved.Speed)
	assert.Equal(t, 6500, retrieved.ExperiencePoints)
	assert.Equal(t, []string{"INT", "WIS"}, retrieved.Proficiencies.SavingThrows)
	assert.Len(t, retrieved.Skills, 4)
	assert.Equal(t, "Arcana", retrieved.Skills[0].Name)
	assert.Equal(t, 7, retrieved.Skills[0].Modifier)
	assert.Equal(t, []string{"Common", "Elvish", "Draconic"}, retrieved.Languages)
	assert.Len(t, retrieved.Features, 2)
	assert.Equal(t, "Spellcasting", retrieved.Features[0].Name)
	assert.NotNil(t, retrieved.Spellcasting)
	assert.Equal(t, "INT", retrieved.Spellcasting.Ability)
	assert.Equal(t, 4, retrieved.Spellcasting.SpellSlots["1"])
	assert.Equal(t, 3, len(retrieved.Spellcasting.Cantrips))
	assert.Equal(t, 5, len(retrieved.Spellcasting.PreparedSpells))
}

func TestSaveCharacter_NonCaster(t *testing.T) {
	db := setupTestDBWithNewFields(t)
	defer db.Close()

	store := NewStore(db)

	// Create a non-caster (barbarian) with no spellcasting
	barbarian := &types.Character{
		CampaignID: "test-campaign",
		Name:       "Conan",
		Type:       "pc",
		Class:      "Barbarian",
		Race:       "Half-Orc",
		Level:      3,
		HP:         types.HP{Current: 34, Max: 34},
		Stats:      types.Stats{STR: 18, DEX: 14, CON: 16, INT: 8, WIS: 12, CHA: 10},
		Alignment:  "Chaotic Neutral",
		AC:         14,
		Speed:      "40 ft",
		ExperiencePoints: 900,
		Proficiencies: types.Proficiencies{
			Armor:        []string{"Light Armor", "Medium Armor", "Shields"},
			Weapons:      []string{"Simple Weapons", "Martial Weapons"},
			Tools:        []string{},
			SavingThrows: []string{"STR", "CON"},
			Skills:       []string{"Athletics", "Intimidation", "Survival"},
		},
		Skills: []types.Skill{
			{Name: "Athletics", Proficient: true, Modifier: 6},
			{Name: "Intimidation", Proficient: true, Modifier: 2},
			{Name: "Survival", Proficient: true, Modifier: 3},
		},
		Languages: []string{"Common", "Orc"},
		Features: []types.Feature{
			{Name: "Rage", Description: "Enter a rage as a bonus action", Source: "class"},
			{Name: "Unarmored Defense", Description: "AC = 10 + DEX + CON when not wearing armor", Source: "class"},
			{Name: "Darkvision", Description: "See in dim light within 60 feet", Source: "race"},
		},
		Spellcasting:  nil, // Non-caster
		Gold:          75,
		Status:        "active",
		Inventory:     []string{"Greataxe", "Javelin", "Backpack"},
		Conditions:    []string{},
		PlotFlags:     []string{},
		Relationships: map[string]string{},
	}

	err := store.SaveCharacter(barbarian)
	require.NoError(t, err)

	// Retrieve and verify
	retrieved, err := store.GetCharacter("test-campaign", "Conan")
	require.NoError(t, err)
	require.NotNil(t, retrieved)

	assert.Equal(t, "Chaotic Neutral", retrieved.Alignment)
	assert.Equal(t, 14, retrieved.AC)
	assert.Equal(t, "40 ft", retrieved.Speed)
	assert.Equal(t, 900, retrieved.ExperiencePoints)
	assert.Len(t, retrieved.Features, 3)
	assert.Nil(t, retrieved.Spellcasting) // Should be nil for non-casters
}

func TestUpdateCharacter_NewFields(t *testing.T) {
	db := setupTestDBWithNewFields(t)
	defer db.Close()

	store := NewStore(db)

	// Create initial character
	char := &types.Character{
		CampaignID: "test-campaign",
		Name:       "TestChar",
		Type:       "pc",
		Class:      "Fighter",
		Race:       "Human",
		Level:      1,
		HP:         types.HP{Current: 12, Max: 12},
		Stats:      types.Stats{STR: 16, DEX: 14, CON: 14, INT: 10, WIS: 12, CHA: 10},
		Alignment:  "Lawful Good",
		AC:         16,
		Speed:      "30 ft",
		ExperiencePoints: 0,
		Proficiencies: types.Proficiencies{
			SavingThrows: []string{"STR", "CON"},
		},
		Skills:        []types.Skill{},
		Languages:     []string{"Common"},
		Features:      []types.Feature{},
		Spellcasting:  nil,
		Gold:          100,
		Status:        "active",
		Inventory:     []string{},
		Conditions:    []string{},
		PlotFlags:     []string{},
		Relationships: map[string]string{},
	}

	err := store.SaveCharacter(char)
	require.NoError(t, err)

	// Update alignment, AC, and XP
	newAlignment := "Neutral Good"
	newAC := 18
	newXP := 300
	update := CharacterUpdate{
		Alignment:        &newAlignment,
		AC:               &newAC,
		ExperiencePoints: &newXP,
	}

	fields, err := store.UpdateCharacter("test-campaign", "TestChar", update)
	require.NoError(t, err)
	assert.Contains(t, fields, "alignment")
	assert.Contains(t, fields, "ac")
	assert.Contains(t, fields, "experience_points")

	// Verify updates
	updated, err := store.GetCharacter("test-campaign", "TestChar")
	require.NoError(t, err)
	assert.Equal(t, "Neutral Good", updated.Alignment)
	assert.Equal(t, 18, updated.AC)
	assert.Equal(t, 300, updated.ExperiencePoints)
}

func TestUpdateCharacter_Skills(t *testing.T) {
	db := setupTestDBWithNewFields(t)
	defer db.Close()

	store := NewStore(db)

	// Create character
	char := &types.Character{
		CampaignID: "test-campaign",
		Name:       "Rogue",
		Type:       "pc",
		Class:      "Rogue",
		Level:      1,
		HP:         types.HP{Current: 10, Max: 10},
		Stats:      types.Stats{STR: 10, DEX: 16, CON: 12, INT: 14, WIS: 12, CHA: 10},
		AC:         14,
		Skills:     []types.Skill{},
		Languages:  []string{"Common"},
		Features:   []types.Feature{},
		Status:     "active",
		Inventory:  []string{},
		Conditions: []string{},
		PlotFlags:  []string{},
		Relationships: map[string]string{},
	}

	err := store.SaveCharacter(char)
	require.NoError(t, err)

	// Update skills
	newSkills := []types.Skill{
		{Name: "Stealth", Proficient: true, Modifier: 5},
		{Name: "Sleight of Hand", Proficient: true, Modifier: 5},
		{Name: "Perception", Proficient: true, Modifier: 3},
	}

	update := CharacterUpdate{
		Skills: newSkills,
	}

	fields, err := store.UpdateCharacter("test-campaign", "Rogue", update)
	require.NoError(t, err)
	assert.Contains(t, fields, "skills")

	// Verify
	updated, err := store.GetCharacter("test-campaign", "Rogue")
	require.NoError(t, err)
	assert.Len(t, updated.Skills, 3)
	assert.Equal(t, "Stealth", updated.Skills[0].Name)
	assert.True(t, updated.Skills[0].Proficient)
}

func TestUpdateCharacter_Spellcasting(t *testing.T) {
	db := setupTestDBWithNewFields(t)
	defer db.Close()

	store := NewStore(db)

	// Create wizard
	wizard := &types.Character{
		CampaignID: "test-campaign",
		Name:       "Merlin",
		Type:       "pc",
		Class:      "Wizard",
		Level:      3,
		HP:         types.HP{Current: 18, Max: 18},
		Stats:      types.Stats{STR: 8, DEX: 14, CON: 12, INT: 18, WIS: 13, CHA: 10},
		AC:         12,
		Spellcasting: &types.Spellcasting{
			Ability:    "INT",
			SpellSlots: map[string]int{"1": 4, "2": 2},
			Cantrips:   []string{"Fire Bolt", "Mage Hand"},
			PreparedSpells: []string{"Magic Missile", "Shield"},
		},
		Skills:     []types.Skill{},
		Languages:  []string{"Common"},
		Features:   []types.Feature{},
		Status:     "active",
		Inventory:  []string{},
		Conditions: []string{},
		PlotFlags:  []string{},
		Relationships: map[string]string{},
	}

	err := store.SaveCharacter(wizard)
	require.NoError(t, err)

	// Level up - update spell slots and prepared spells
	newSpellcasting := &types.Spellcasting{
		Ability:    "INT",
		SpellSlots: map[string]int{"1": 4, "2": 3},
		Cantrips:   []string{"Fire Bolt", "Mage Hand", "Prestidigitation"},
		PreparedSpells: []string{"Magic Missile", "Shield", "Misty Step", "Detect Magic"},
	}

	update := CharacterUpdate{
		Spellcasting: newSpellcasting,
	}

	fields, err := store.UpdateCharacter("test-campaign", "Merlin", update)
	require.NoError(t, err)
	assert.Contains(t, fields, "spellcasting")

	// Verify
	updated, err := store.GetCharacter("test-campaign", "Merlin")
	require.NoError(t, err)
	require.NotNil(t, updated.Spellcasting)
	assert.Equal(t, 3, updated.Spellcasting.SpellSlots["2"])
	assert.Len(t, updated.Spellcasting.Cantrips, 3)
	assert.Len(t, updated.Spellcasting.PreparedSpells, 4)
}

func TestCharacterDefaults(t *testing.T) {
	db := setupTestDBWithNewFields(t)
	defer db.Close()

	store := NewStore(db)

	// Create character with minimal fields, applying defaults like the handler does
	char := &types.Character{
		CampaignID: "test-campaign",
		Name:       "MinimalChar",
		Type:       "pc",
		Level:      1,
		HP:         types.HP{Current: 10, Max: 10},
		Stats:      types.Stats{STR: 10, DEX: 10, CON: 10, INT: 10, WIS: 10, CHA: 10},
		AC:            10,       // Default AC (applied by handler)
		Speed:         "30 ft",  // Default Speed (applied by handler)
		Skills:        []types.Skill{},
		Languages:     []string{},
		Features:      []types.Feature{},
		Status:        "active",
		Inventory:     []string{},
		Conditions:    []string{},
		PlotFlags:     []string{},
		Relationships: map[string]string{},
	}

	err := store.SaveCharacter(char)
	require.NoError(t, err)

	// Retrieve and verify defaults were preserved
	retrieved, err := store.GetCharacter("test-campaign", "MinimalChar")
	require.NoError(t, err)
	require.NotNil(t, retrieved)

	assert.Equal(t, 10, retrieved.AC)       // Default AC
	assert.Equal(t, "30 ft", retrieved.Speed) // Default Speed
	assert.Equal(t, 0, retrieved.ExperiencePoints) // Default XP
	assert.NotNil(t, retrieved.Skills)
	assert.NotNil(t, retrieved.Languages)
	assert.NotNil(t, retrieved.Features)
}

func TestListCharacters_WithNewFields(t *testing.T) {
	db := setupTestDBWithNewFields(t)
	defer db.Close()

	store := NewStore(db)

	// Create multiple characters
	chars := []*types.Character{
		{
			CampaignID: "test-campaign",
			Name:       "Fighter",
			Type:       "pc",
			Class:      "Fighter",
			Race:       "Human",
			Level:      5,
			HP:         types.HP{Current: 45, Max: 45},
			Stats:      types.Stats{STR: 16, DEX: 14, CON: 14, INT: 10, WIS: 12, CHA: 10},
			AC:         18,
			Status:     "active",
			Skills:        []types.Skill{},
			Languages:     []string{},
			Features:      []types.Feature{},
			Inventory:     []string{},
			Conditions:    []string{},
			PlotFlags:     []string{},
			Relationships: map[string]string{},
		},
		{
			CampaignID: "test-campaign",
			Name:       "Wizard",
			Type:       "pc",
			Class:      "Wizard",
			Race:       "Elf",
			Level:      5,
			HP:         types.HP{Current: 28, Max: 28},
			Stats:      types.Stats{STR: 8, DEX: 14, CON: 12, INT: 18, WIS: 13, CHA: 10},
			AC:         12,
			Status:     "active",
			Skills:        []types.Skill{},
			Languages:     []string{},
			Features:      []types.Feature{},
			Inventory:     []string{},
			Conditions:    []string{},
			PlotFlags:     []string{},
			Relationships: map[string]string{},
		},
	}

	for _, char := range chars {
		err := store.SaveCharacter(char)
		require.NoError(t, err)
	}

	// List characters
	summaries, err := store.ListCharacters("test-campaign", "", "")
	require.NoError(t, err)
	require.Len(t, summaries, 2)

	// Verify Race and AC are included in summaries
	for _, summary := range summaries {
		assert.NotEmpty(t, summary.Race)
		assert.Greater(t, summary.AC, 0)
		if summary.Name == "Fighter" {
			assert.Equal(t, "Human", summary.Race)
			assert.Equal(t, 18, summary.AC)
		} else if summary.Name == "Wizard" {
			assert.Equal(t, "Elf", summary.Race)
			assert.Equal(t, 12, summary.AC)
		}
	}
}
