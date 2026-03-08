package memory

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/andr1an/llmdm/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)

	// Run schema
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
		CREATE TABLE plot_events (
			id TEXT PRIMARY KEY,
			campaign_id TEXT NOT NULL REFERENCES campaigns(id),
			session INTEGER NOT NULL,
			summary TEXT NOT NULL,
			npcs_involved TEXT,
			pcs_involved TEXT,
			consequences TEXT,
			tags TEXT,
			created_at TEXT DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE plot_hooks (
			id TEXT PRIMARY KEY,
			campaign_id TEXT NOT NULL REFERENCES campaigns(id),
			hook TEXT NOT NULL,
			session_opened INTEGER NOT NULL,
			event_id TEXT REFERENCES plot_events(id),
			resolved INTEGER DEFAULT 0,
			resolution TEXT,
			resolved_at TEXT
		);
		CREATE TABLE world_flags (
			campaign_id TEXT NOT NULL REFERENCES campaigns(id),
			key TEXT NOT NULL,
			value TEXT,
			updated_at TEXT DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (campaign_id, key)
		);
		CREATE TABLE sessions (
			campaign_id TEXT NOT NULL REFERENCES campaigns(id),
			session INTEGER NOT NULL,
			summary TEXT,
			dm_notes TEXT,
			hooks_opened INTEGER DEFAULT 0,
			hooks_resolved INTEGER DEFAULT 0,
			created_at TEXT DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (campaign_id, session)
		);
		CREATE TABLE checkpoints (
			id TEXT PRIMARY KEY,
			campaign_id TEXT NOT NULL REFERENCES campaigns(id),
			session INTEGER NOT NULL,
			note TEXT NOT NULL,
			created_at TEXT DEFAULT CURRENT_TIMESTAMP
		);
	`
	_, err = db.Exec(schema)
	require.NoError(t, err)

	return db
}

func TestCreateCampaign(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	store := NewStore(db)

	campaign, err := store.CreateCampaign("Test Campaign", "A test campaign")
	require.NoError(t, err)
	assert.NotEmpty(t, campaign.ID)
	assert.Equal(t, "Test Campaign", campaign.Name)
	assert.Equal(t, "A test campaign", campaign.Description)
	assert.Equal(t, 1, campaign.CurrentSession)
	assert.NotEmpty(t, campaign.CreatedAt)
}

func TestCreateCampaignWithID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	store := NewStore(db)

	campaign, err := store.CreateCampaignWithID("the-obsidian-throne", "The Obsidian Throne", "A dark court conspiracy.")
	require.NoError(t, err)
	assert.Equal(t, "the-obsidian-throne", campaign.ID)
	assert.Equal(t, "The Obsidian Throne", campaign.Name)
}

func TestGetCampaign(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	store := NewStore(db)

	// Create a campaign
	created, err := store.CreateCampaign("Test Campaign", "A test campaign")
	require.NoError(t, err)

	// Retrieve it
	campaign, err := store.GetCampaign(created.ID)
	require.NoError(t, err)
	require.NotNil(t, campaign)
	assert.Equal(t, created.ID, campaign.ID)
	assert.Equal(t, "Test Campaign", campaign.Name)

	// Non-existent campaign
	notFound, err := store.GetCampaign("nonexistent")
	require.NoError(t, err)
	assert.Nil(t, notFound)
}

func TestSaveCharacter(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	store := NewStore(db)

	// Create campaign first
	campaign, err := store.CreateCampaign("Test Campaign", "")
	require.NoError(t, err)

	char := &types.Character{
		CampaignID: campaign.ID,
		Name:       "Aldric",
		Type:       "pc",
		Class:      "Paladin",
		Race:       "Human",
		Level:      5,
		HP:         types.HP{Current: 45, Max: 45},
		Stats:      types.Stats{STR: 16, DEX: 12, CON: 14, INT: 10, WIS: 14, CHA: 16},
		Backstory:  "A noble knight",
		Status:     "active",
		Inventory:  []string{"Longsword", "Shield"},
		Conditions: []string{},
		PlotFlags:  []string{"met_the_king"},
		Relationships: map[string]string{
			"Zara": "ally",
		},
	}

	err = store.SaveCharacter(char)
	require.NoError(t, err)
	assert.NotEmpty(t, char.ID)

	// Verify it was saved
	retrieved, err := store.GetCharacter(campaign.ID, "Aldric")
	require.NoError(t, err)
	require.NotNil(t, retrieved)
	assert.Equal(t, "Aldric", retrieved.Name)
	assert.Equal(t, "pc", retrieved.Type)
	assert.Equal(t, "Paladin", retrieved.Class)
	assert.Equal(t, 5, retrieved.Level)
	assert.Equal(t, 45, retrieved.HP.Current)
	assert.Equal(t, 16, retrieved.Stats.STR)
	assert.Contains(t, retrieved.Inventory, "Longsword")
	assert.Equal(t, "ally", retrieved.Relationships["Zara"])
}

func TestSaveCharacter_Upsert(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	store := NewStore(db)

	campaign, err := store.CreateCampaign("Test Campaign", "")
	require.NoError(t, err)

	// Create initial character
	char := &types.Character{
		CampaignID: campaign.ID,
		Name:       "Aldric",
		Type:       "pc",
		HP:         types.HP{Current: 45, Max: 45},
		Status:     "active",
	}
	err = store.SaveCharacter(char)
	require.NoError(t, err)
	originalID := char.ID

	// Upsert with same name - should update
	char2 := &types.Character{
		CampaignID: campaign.ID,
		Name:       "Aldric",
		Type:       "pc",
		Class:      "Fighter", // Changed
		HP:         types.HP{Current: 30, Max: 50},
		Status:     "active",
	}
	err = store.SaveCharacter(char2)
	require.NoError(t, err)

	// Verify update
	retrieved, err := store.GetCharacter(campaign.ID, "Aldric")
	require.NoError(t, err)
	assert.Equal(t, "Fighter", retrieved.Class)
	assert.Equal(t, 30, retrieved.HP.Current)
	assert.Equal(t, 50, retrieved.HP.Max)
	// ID should have changed since we're replacing
	assert.NotEqual(t, originalID, retrieved.ID)
}

func TestUpdateCharacter(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	store := NewStore(db)

	campaign, err := store.CreateCampaign("Test Campaign", "")
	require.NoError(t, err)

	// Create character
	char := &types.Character{
		CampaignID: campaign.ID,
		Name:       "Aldric",
		Type:       "pc",
		Level:      5,
		HP:         types.HP{Current: 45, Max: 45},
		Status:     "active",
	}
	err = store.SaveCharacter(char)
	require.NoError(t, err)

	// Update HP
	newHP := 30
	updated, err := store.UpdateCharacter(campaign.ID, "Aldric", CharacterUpdate{
		HPCurrent: &newHP,
	})
	require.NoError(t, err)
	assert.Contains(t, updated, "hp_current")

	// Verify update
	retrieved, err := store.GetCharacter(campaign.ID, "Aldric")
	require.NoError(t, err)
	assert.Equal(t, 30, retrieved.HP.Current)
	assert.Equal(t, 5, retrieved.Level) // Unchanged
}

func TestUpdateCharacter_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	store := NewStore(db)

	campaign, err := store.CreateCampaign("Test Campaign", "")
	require.NoError(t, err)

	newHP := 30
	_, err = store.UpdateCharacter(campaign.ID, "NonExistent", CharacterUpdate{
		HPCurrent: &newHP,
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "character not found")
}

func TestListCharacters(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	store := NewStore(db)

	campaign, err := store.CreateCampaign("Test Campaign", "")
	require.NoError(t, err)

	// Create multiple characters
	chars := []*types.Character{
		{CampaignID: campaign.ID, Name: "Aldric", Type: "pc", HP: types.HP{Current: 45, Max: 45}, Status: "active"},
		{CampaignID: campaign.ID, Name: "Zara", Type: "pc", HP: types.HP{Current: 30, Max: 30}, Status: "active"},
		{CampaignID: campaign.ID, Name: "Goblin King", Type: "npc", HP: types.HP{Current: 50, Max: 50}, Status: "active"},
		{CampaignID: campaign.ID, Name: "Dead NPC", Type: "npc", HP: types.HP{Current: 0, Max: 20}, Status: "dead"},
	}
	for _, c := range chars {
		err = store.SaveCharacter(c)
		require.NoError(t, err)
	}

	// List all
	all, err := store.ListCharacters(campaign.ID, "", "")
	require.NoError(t, err)
	assert.Len(t, all, 4)

	// Filter by type
	pcs, err := store.ListCharacters(campaign.ID, "pc", "")
	require.NoError(t, err)
	assert.Len(t, pcs, 2)

	// Filter by status
	active, err := store.ListCharacters(campaign.ID, "", "active")
	require.NoError(t, err)
	assert.Len(t, active, 3)

	// Filter by both
	activeNPCs, err := store.ListCharacters(campaign.ID, "npc", "active")
	require.NoError(t, err)
	assert.Len(t, activeNPCs, 1)
	assert.Equal(t, "Goblin King", activeNPCs[0].Name)
}

func TestSavePlotEvent(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	store := NewStore(db)

	campaign, err := store.CreateCampaign("Test Campaign", "")
	require.NoError(t, err)

	event := &types.PlotEvent{
		CampaignID:   campaign.ID,
		Session:      1,
		Summary:      "The party discovered a secret passage behind the tavern.",
		Consequences: "The passage leads to an underground network.",
		NPCs:         []string{"Tavern Keeper"},
		PCs:          []string{"Aldric", "Zara"},
		Tags:         []string{"discovery", "exploration"},
	}

	hooks := []string{
		"Where does the passage lead?",
		"Why did the Tavern Keeper hide this?",
	}

	err = store.SavePlotEvent(event, hooks)
	require.NoError(t, err)
	assert.NotEmpty(t, event.ID)

	// Verify hooks were created
	openHooks, err := store.ListOpenHooks(campaign.ID)
	require.NoError(t, err)
	assert.Len(t, openHooks, 2)
	assert.Equal(t, "Where does the passage lead?", openHooks[0].Hook)
	assert.Equal(t, 1, openHooks[0].SessionOpened)
	assert.Equal(t, event.ID, openHooks[0].EventID)
}

func TestResolveHook(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	store := NewStore(db)

	campaign, err := store.CreateCampaign("Test Campaign", "")
	require.NoError(t, err)

	// Create event with hook
	event := &types.PlotEvent{
		CampaignID: campaign.ID,
		Session:    1,
		Summary:    "Test event",
	}
	err = store.SavePlotEvent(event, []string{"Mystery hook"})
	require.NoError(t, err)

	// Get the hook
	hooks, err := store.ListOpenHooks(campaign.ID)
	require.NoError(t, err)
	require.Len(t, hooks, 1)
	hookID := hooks[0].ID

	// Resolve it
	err = store.ResolveHook(campaign.ID, hookID, "The mystery was solved!")
	require.NoError(t, err)

	// Verify it's no longer in open hooks
	openHooks, err := store.ListOpenHooks(campaign.ID)
	require.NoError(t, err)
	assert.Len(t, openHooks, 0)
}

func TestResolveHook_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	store := NewStore(db)

	campaign, err := store.CreateCampaign("Test Campaign", "")
	require.NoError(t, err)

	err = store.ResolveHook(campaign.ID, "nonexistent", "Resolution")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "hook not found")
}

func TestWorldFlags(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	store := NewStore(db)

	campaign, err := store.CreateCampaign("Test Campaign", "")
	require.NoError(t, err)

	// Set flags
	err = store.SetWorldFlag(campaign.ID, "dragon_defeated", "true")
	require.NoError(t, err)
	err = store.SetWorldFlag(campaign.ID, "current_season", "winter")
	require.NoError(t, err)

	// Get flags
	flags, err := store.GetWorldFlags(campaign.ID)
	require.NoError(t, err)
	assert.Len(t, flags, 2)
	assert.Equal(t, "true", flags["dragon_defeated"])
	assert.Equal(t, "winter", flags["current_season"])

	// Update a flag
	err = store.SetWorldFlag(campaign.ID, "current_season", "spring")
	require.NoError(t, err)

	flags, err = store.GetWorldFlags(campaign.ID)
	require.NoError(t, err)
	assert.Equal(t, "spring", flags["current_season"])
}

func TestGetWorldFlags_Empty(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	store := NewStore(db)

	campaign, err := store.CreateCampaign("Test Campaign", "")
	require.NoError(t, err)

	flags, err := store.GetWorldFlags(campaign.ID)
	require.NoError(t, err)
	assert.NotNil(t, flags)
	assert.Len(t, flags, 0)
}

func TestCharacterConditionsAndPlotFlags(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	store := NewStore(db)

	campaign, err := store.CreateCampaign("Test Campaign", "")
	require.NoError(t, err)

	// Create character with conditions
	char := &types.Character{
		CampaignID: campaign.ID,
		Name:       "Aldric",
		Type:       "pc",
		HP:         types.HP{Current: 45, Max: 45},
		Status:     "active",
		Conditions: []string{"Blessed", "Inspired"},
		PlotFlags:  []string{"knows_secret", "met_the_king"},
	}
	err = store.SaveCharacter(char)
	require.NoError(t, err)

	// Verify
	retrieved, err := store.GetCharacter(campaign.ID, "Aldric")
	require.NoError(t, err)
	assert.Contains(t, retrieved.Conditions, "Blessed")
	assert.Contains(t, retrieved.Conditions, "Inspired")
	assert.Contains(t, retrieved.PlotFlags, "knows_secret")
	assert.Contains(t, retrieved.PlotFlags, "met_the_king")

	// Update conditions
	newConditions := []string{"Poisoned"}
	_, err = store.UpdateCharacter(campaign.ID, "Aldric", CharacterUpdate{
		Conditions: newConditions,
	})
	require.NoError(t, err)

	retrieved, err = store.GetCharacter(campaign.ID, "Aldric")
	require.NoError(t, err)
	assert.Len(t, retrieved.Conditions, 1)
	assert.Contains(t, retrieved.Conditions, "Poisoned")
	// Plot flags should be unchanged
	assert.Len(t, retrieved.PlotFlags, 2)
}

func TestCheckpointLifecycle(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	store := NewStore(db)

	campaign, err := store.CreateCampaign("Test Campaign", "")
	require.NoError(t, err)

	cp1, err := store.CreateCheckpoint(campaign.ID, 3, "Party reaches the ruined watchtower.")
	require.NoError(t, err)
	require.NotEmpty(t, cp1.ID)

	cp2, err := store.CreateCheckpoint(campaign.ID, 3, "They uncover a hidden cellar map.")
	require.NoError(t, err)
	require.NotEmpty(t, cp2.ID)

	latest, err := store.GetLatestCheckpoint(campaign.ID, 3)
	require.NoError(t, err)
	require.NotNil(t, latest)
	assert.Equal(t, cp2.ID, latest.ID)
	assert.Equal(t, "They uncover a hidden cellar map.", latest.Note)
}

func TestSessionsLifecycle(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	store := NewStore(db)

	campaign, err := store.CreateCampaign("Test Campaign", "")
	require.NoError(t, err)

	err = store.UpsertSession(&types.Session{
		CampaignID:    campaign.ID,
		Session:       1,
		Summary:       "Session one summary",
		HooksOpened:   2,
		HooksResolved: 0,
	})
	require.NoError(t, err)

	err = store.UpsertSession(&types.Session{
		CampaignID:    campaign.ID,
		Session:       2,
		Summary:       "Session two summary",
		HooksOpened:   1,
		HooksResolved: 1,
	})
	require.NoError(t, err)

	latest, err := store.GetLatestSession(campaign.ID)
	require.NoError(t, err)
	require.NotNil(t, latest)
	assert.Equal(t, 2, latest.Session)

	prev, err := store.GetLastSessionBefore(campaign.ID, 3)
	require.NoError(t, err)
	require.NotNil(t, prev)
	assert.Equal(t, 2, prev.Session)

	all, err := store.ListSessions(campaign.ID)
	require.NoError(t, err)
	require.Len(t, all, 2)
	assert.Equal(t, 2, all[0].Session)
	assert.Equal(t, 1, all[1].Session)

	recent, err := store.ListRecentSessionsBefore(campaign.ID, 3, 1)
	require.NoError(t, err)
	require.Len(t, recent, 1)
	assert.Equal(t, 2, recent[0].Session)

	err = store.AdvanceCampaignSession(campaign.ID, 5)
	require.NoError(t, err)

	reloaded, err := store.GetCampaign(campaign.ID)
	require.NoError(t, err)
	require.NotNil(t, reloaded)
	assert.Equal(t, 5, reloaded.CurrentSession)
}

func TestQueryNPCRelationships(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	store := NewStore(db)

	campaign, err := store.CreateCampaign("Test Campaign", "")
	require.NoError(t, err)

	chars := []*types.Character{
		{
			CampaignID: campaign.ID,
			Name:       "Aldric",
			Type:       "pc",
			HP:         types.HP{Current: 40, Max: 40},
			Status:     "active",
			Relationships: map[string]string{
				"Zara":        "ally",
				"Goblin King": "enemy",
			},
		},
		{
			CampaignID: campaign.ID,
			Name:       "Zara",
			Type:       "npc",
			HP:         types.HP{Current: 25, Max: 25},
			Status:     "active",
			Relationships: map[string]string{
				"Aldric": "trusted",
			},
		},
		{
			CampaignID: campaign.ID,
			Name:       "Goblin King",
			Type:       "npc",
			HP:         types.HP{Current: 50, Max: 50},
			Status:     "active",
		},
	}

	for _, c := range chars {
		require.NoError(t, store.SaveCharacter(c))
	}

	allEdges, err := store.QueryNPCRelationships(campaign.ID, "")
	require.NoError(t, err)
	require.Len(t, allEdges, 3)

	filtered, err := store.QueryNPCRelationships(campaign.ID, "zara")
	require.NoError(t, err)
	require.Len(t, filtered, 2)
	assert.Equal(t, "Zara", filtered[1].Source)
	assert.Equal(t, "Aldric", filtered[1].Target)
}
