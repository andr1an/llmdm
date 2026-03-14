// Package memory provides campaign memory storage and retrieval.
package memory

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/andr1an/llmdm/internal/types"
)

// Store handles all campaign memory operations.
type Store struct {
	db *sql.DB
}

// NewStore creates a new memory store.
func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

// CreateCampaign creates a new campaign record.
func (s *Store) CreateCampaign(name, description string) (*types.Campaign, error) {
	return s.CreateCampaignWithID(uuid.New().String(), name, description)
}

// CreateCampaignWithID creates a campaign record with a caller-provided ID.
func (s *Store) CreateCampaignWithID(id, name, description string) (*types.Campaign, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("campaign id cannot be empty")
	}

	now := time.Now().UTC().Format(time.RFC3339)

	_, err := s.db.Exec(`
		INSERT INTO campaigns (id, name, description, created_at, current_session)
		VALUES (?, ?, ?, ?, 1)
	`, id, name, description, now)
	if err != nil {
		return nil, fmt.Errorf("insert campaign: %w", err)
	}

	return &types.Campaign{
		ID:             id,
		Name:           name,
		Description:    description,
		CreatedAt:      now,
		CurrentSession: 1,
	}, nil
}

// GetCampaign retrieves a campaign by ID.
func (s *Store) GetCampaign(id string) (*types.Campaign, error) {
	var c types.Campaign
	var descNull sql.NullString
	err := s.db.QueryRow(`
		SELECT id, name, description, created_at, current_session
		FROM campaigns WHERE id = ?
	`, id).Scan(&c.ID, &c.Name, &descNull, &c.CreatedAt, &c.CurrentSession)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query campaign: %w", err)
	}
	if descNull.Valid {
		c.Description = descNull.String
	}
	return &c, nil
}

// SaveCharacter creates or fully replaces a character.
func (s *Store) SaveCharacter(char *types.Character) error {
	if char.ID == "" {
		char.ID = uuid.New().String()
	}
	char.UpdatedAt = time.Now().UTC().Format(time.RFC3339)

	inventoryJSON, _ := json.Marshal(char.Inventory)
	conditionsJSON, _ := json.Marshal(char.Conditions)
	relationshipsJSON, _ := json.Marshal(char.Relationships)
	plotFlagsJSON, _ := json.Marshal(char.PlotFlags)

	_, err := s.db.Exec(`
		INSERT INTO characters (id, campaign_id, name, type, class, race, level, hp_current, hp_max,
			stat_str, stat_dex, stat_con, stat_int, stat_wis, stat_cha, gold, backstory, inventory,
			conditions, relationships, plot_flags, notes, status, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(campaign_id, name) DO UPDATE SET
			id = excluded.id,
			type = excluded.type,
			class = excluded.class,
			race = excluded.race,
			level = excluded.level,
			hp_current = excluded.hp_current,
			hp_max = excluded.hp_max,
			stat_str = excluded.stat_str,
			stat_dex = excluded.stat_dex,
			stat_con = excluded.stat_con,
			stat_int = excluded.stat_int,
			stat_wis = excluded.stat_wis,
			stat_cha = excluded.stat_cha,
			gold = excluded.gold,
			backstory = excluded.backstory,
			inventory = excluded.inventory,
			conditions = excluded.conditions,
			relationships = excluded.relationships,
			plot_flags = excluded.plot_flags,
			notes = excluded.notes,
			status = excluded.status,
			updated_at = excluded.updated_at
	`, char.ID, char.CampaignID, char.Name, char.Type, char.Class, char.Race, char.Level,
		char.HP.Current, char.HP.Max, char.Stats.STR, char.Stats.DEX, char.Stats.CON,
		char.Stats.INT, char.Stats.WIS, char.Stats.CHA, char.Gold, char.Backstory, string(inventoryJSON),
		string(conditionsJSON), string(relationshipsJSON), string(plotFlagsJSON),
		char.Notes, char.Status, char.UpdatedAt)

	if err != nil {
		return fmt.Errorf("upsert character: %w", err)
	}
	return nil
}

// CharacterUpdate represents fields to patch on a character.
type CharacterUpdate struct {
	HPCurrent     *int
	Level         *int
	Gold          *int
	Conditions    []string
	Inventory     []string
	PlotFlags     []string
	Relationships map[string]string
	Status        *string
	Notes         *string
}

// UpdateCharacter patches specific fields on a character.
func (s *Store) UpdateCharacter(campaignID, name string, update CharacterUpdate) ([]string, error) {
	var updatedFields []string
	now := time.Now().UTC().Format(time.RFC3339)

	// Build dynamic update
	updates := []string{"updated_at = ?"}
	args := []interface{}{now}

	if update.HPCurrent != nil {
		updates = append(updates, "hp_current = ?")
		args = append(args, *update.HPCurrent)
		updatedFields = append(updatedFields, "hp_current")
	}
	if update.Level != nil {
		updates = append(updates, "level = ?")
		args = append(args, *update.Level)
		updatedFields = append(updatedFields, "level")
	}
	if update.Gold != nil {
		updates = append(updates, "gold = ?")
		args = append(args, *update.Gold)
		updatedFields = append(updatedFields, "gold")
	}
	if update.Conditions != nil {
		conditionsJSON, _ := json.Marshal(update.Conditions)
		updates = append(updates, "conditions = ?")
		args = append(args, string(conditionsJSON))
		updatedFields = append(updatedFields, "conditions")
	}
	if update.Inventory != nil {
		inventoryJSON, _ := json.Marshal(update.Inventory)
		updates = append(updates, "inventory = ?")
		args = append(args, string(inventoryJSON))
		updatedFields = append(updatedFields, "inventory")
	}
	if update.PlotFlags != nil {
		plotFlagsJSON, _ := json.Marshal(update.PlotFlags)
		updates = append(updates, "plot_flags = ?")
		args = append(args, string(plotFlagsJSON))
		updatedFields = append(updatedFields, "plot_flags")
	}
	if update.Relationships != nil {
		relationshipsJSON, _ := json.Marshal(update.Relationships)
		updates = append(updates, "relationships = ?")
		args = append(args, string(relationshipsJSON))
		updatedFields = append(updatedFields, "relationships")
	}
	if update.Status != nil {
		updates = append(updates, "status = ?")
		args = append(args, *update.Status)
		updatedFields = append(updatedFields, "status")
	}
	if update.Notes != nil {
		updates = append(updates, "notes = ?")
		args = append(args, *update.Notes)
		updatedFields = append(updatedFields, "notes")
	}

	if len(updatedFields) == 0 {
		return nil, nil
	}

	args = append(args, campaignID, name)

	query := "UPDATE characters SET " + joinStrings(updates, ", ") + " WHERE campaign_id = ? AND name = ?"
	result, err := s.db.Exec(query, args...)
	if err != nil {
		return nil, fmt.Errorf("update character: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return nil, fmt.Errorf("character not found: %s", name)
	}

	return updatedFields, nil
}

// GetCharacter retrieves a full character by name.
func (s *Store) GetCharacter(campaignID, name string) (*types.Character, error) {
	var c types.Character
	var classNull, raceNull, backstoryNull, notesNull sql.NullString
	var inventoryJSON, conditionsJSON, relationshipsJSON, plotFlagsJSON sql.NullString

	err := s.db.QueryRow(`
		SELECT id, campaign_id, name, type, class, race, level, hp_current, hp_max,
			stat_str, stat_dex, stat_con, stat_int, stat_wis, stat_cha, gold, backstory,
			inventory, conditions, relationships, plot_flags, notes, status, updated_at
		FROM characters WHERE campaign_id = ? AND name = ?
	`, campaignID, name).Scan(
		&c.ID, &c.CampaignID, &c.Name, &c.Type, &classNull, &raceNull, &c.Level,
		&c.HP.Current, &c.HP.Max, &c.Stats.STR, &c.Stats.DEX, &c.Stats.CON,
		&c.Stats.INT, &c.Stats.WIS, &c.Stats.CHA, &c.Gold, &backstoryNull,
		&inventoryJSON, &conditionsJSON, &relationshipsJSON, &plotFlagsJSON,
		&notesNull, &c.Status, &c.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query character: %w", err)
	}

	if classNull.Valid {
		c.Class = classNull.String
	}
	if raceNull.Valid {
		c.Race = raceNull.String
	}
	if backstoryNull.Valid {
		c.Backstory = backstoryNull.String
	}
	if notesNull.Valid {
		c.Notes = notesNull.String
	}
	if inventoryJSON.Valid {
		json.Unmarshal([]byte(inventoryJSON.String), &c.Inventory)
	}
	if conditionsJSON.Valid {
		json.Unmarshal([]byte(conditionsJSON.String), &c.Conditions)
	}
	if relationshipsJSON.Valid {
		json.Unmarshal([]byte(relationshipsJSON.String), &c.Relationships)
	}
	if plotFlagsJSON.Valid {
		json.Unmarshal([]byte(plotFlagsJSON.String), &c.PlotFlags)
	}

	// Ensure non-nil slices and maps
	if c.Inventory == nil {
		c.Inventory = []string{}
	}
	if c.Conditions == nil {
		c.Conditions = []string{}
	}
	if c.PlotFlags == nil {
		c.PlotFlags = []string{}
	}
	if c.Relationships == nil {
		c.Relationships = map[string]string{}
	}

	return &c, nil
}

// ListCharacters returns characters filtered by type and status.
func (s *Store) ListCharacters(campaignID, charType, status string) ([]types.CharacterSummary, error) {
	query := `SELECT name, type, class, level, hp_current, hp_max, status, conditions
		FROM characters WHERE campaign_id = ?`
	args := []interface{}{campaignID}

	if charType != "" {
		query += " AND type = ?"
		args = append(args, charType)
	}
	if status != "" {
		query += " AND status = ?"
		args = append(args, status)
	}
	query += " ORDER BY name"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query characters: %w", err)
	}
	defer rows.Close()

	var result []types.CharacterSummary
	for rows.Next() {
		var cs types.CharacterSummary
		var classNull, conditionsJSON sql.NullString
		err := rows.Scan(&cs.Name, &cs.Type, &classNull, &cs.Level, &cs.HP.Current, &cs.HP.Max, &cs.Status, &conditionsJSON)
		if err != nil {
			return nil, fmt.Errorf("scan character: %w", err)
		}
		if classNull.Valid {
			cs.Class = classNull.String
		}
		if conditionsJSON.Valid {
			json.Unmarshal([]byte(conditionsJSON.String), &cs.Conditions)
		}
		if cs.Conditions == nil {
			cs.Conditions = []string{}
		}
		result = append(result, cs)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate characters: %w", err)
	}

	if result == nil {
		result = []types.CharacterSummary{}
	}
	return result, nil
}

// SavePlotEvent creates a plot event and any open hooks.
func (s *Store) SavePlotEvent(event *types.PlotEvent, openHooks []string) error {
	event.ID = uuid.New().String()
	event.CreatedAt = time.Now().UTC().Format(time.RFC3339)

	npcsJSON, _ := json.Marshal(event.NPCs)
	pcsJSON, _ := json.Marshal(event.PCs)
	tagsJSON, _ := json.Marshal(event.Tags)

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		INSERT INTO plot_events (id, campaign_id, session, summary, npcs_involved, pcs_involved, consequences, tags, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, event.ID, event.CampaignID, event.Session, event.Summary, string(npcsJSON), string(pcsJSON), event.Consequences, string(tagsJSON), event.CreatedAt)
	if err != nil {
		return fmt.Errorf("insert plot event: %w", err)
	}

	// Create hooks
	for _, hook := range openHooks {
		hookID := uuid.New().String()
		_, err = tx.Exec(`
			INSERT INTO plot_hooks (id, campaign_id, hook, session_opened, event_id, resolved)
			VALUES (?, ?, ?, ?, ?, 0)
		`, hookID, event.CampaignID, hook, event.Session, event.ID)
		if err != nil {
			return fmt.Errorf("insert hook: %w", err)
		}
	}

	return tx.Commit()
}

// ListOpenHooks returns all unresolved plot hooks.
func (s *Store) ListOpenHooks(campaignID string) ([]types.Hook, error) {
	rows, err := s.db.Query(`
		SELECT id, campaign_id, hook, session_opened, event_id, resolved, resolution
		FROM plot_hooks WHERE campaign_id = ? AND resolved = 0
		ORDER BY session_opened
	`, campaignID)
	if err != nil {
		return nil, fmt.Errorf("query hooks: %w", err)
	}
	defer rows.Close()

	var hooks []types.Hook
	for rows.Next() {
		var h types.Hook
		var eventIDNull, resolutionNull sql.NullString
		var resolvedInt int
		err := rows.Scan(&h.ID, &h.CampaignID, &h.Hook, &h.SessionOpened, &eventIDNull, &resolvedInt, &resolutionNull)
		if err != nil {
			return nil, fmt.Errorf("scan hook: %w", err)
		}
		h.Resolved = resolvedInt == 1
		if eventIDNull.Valid {
			h.EventID = eventIDNull.String
		}
		if resolutionNull.Valid {
			h.Resolution = resolutionNull.String
		}
		hooks = append(hooks, h)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate hooks: %w", err)
	}

	if hooks == nil {
		hooks = []types.Hook{}
	}
	return hooks, nil
}

// ResolveHook marks a hook as resolved.
func (s *Store) ResolveHook(campaignID, hookID, resolution string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	result, err := s.db.Exec(`
		UPDATE plot_hooks SET resolved = 1, resolution = ?, resolved_at = ?
		WHERE id = ? AND campaign_id = ?
	`, resolution, now, hookID, campaignID)
	if err != nil {
		return fmt.Errorf("update hook: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("hook not found: %s", hookID)
	}
	return nil
}

// SetWorldFlag sets a key/value world flag.
func (s *Store) SetWorldFlag(campaignID, key, value string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := s.db.Exec(`
		INSERT INTO world_flags (campaign_id, key, value, updated_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(campaign_id, key) DO UPDATE SET value = excluded.value, updated_at = excluded.updated_at
	`, campaignID, key, value, now)
	if err != nil {
		return fmt.Errorf("upsert world flag: %w", err)
	}
	return nil
}

// GetWorldFlags returns all world flags for a campaign.
func (s *Store) GetWorldFlags(campaignID string) (map[string]string, error) {
	rows, err := s.db.Query(`
		SELECT key, value FROM world_flags WHERE campaign_id = ?
	`, campaignID)
	if err != nil {
		return nil, fmt.Errorf("query world flags: %w", err)
	}
	defer rows.Close()

	flags := make(map[string]string)
	for rows.Next() {
		var key string
		var value sql.NullString
		if err := rows.Scan(&key, &value); err != nil {
			return nil, fmt.Errorf("scan world flag: %w", err)
		}
		if value.Valid {
			flags[key] = value.String
		} else {
			flags[key] = ""
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate world flags: %w", err)
	}
	return flags, nil
}

// CreateCheckpoint stores a mid-session checkpoint note.
func (s *Store) CreateCheckpoint(campaignID string, session int, note string, data map[string]interface{}) (*types.Checkpoint, error) {
	checkpoint := &types.Checkpoint{
		ID:         uuid.New().String(),
		CampaignID: campaignID,
		Session:    session,
		Note:       note,
		Data:       data,
		CreatedAt:  time.Now().UTC().Format(time.RFC3339),
	}

	var dataJSON []byte
	if len(data) > 0 {
		dataJSON, _ = json.Marshal(data)
	}

	_, err := s.db.Exec(`
		INSERT INTO checkpoints (id, campaign_id, session, note, data, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, checkpoint.ID, checkpoint.CampaignID, checkpoint.Session, checkpoint.Note, string(dataJSON), checkpoint.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("insert checkpoint: %w", err)
	}

	return checkpoint, nil
}

// GetLatestCheckpoint returns the newest checkpoint for the given campaign/session.
func (s *Store) GetLatestCheckpoint(campaignID string, session int) (*types.Checkpoint, error) {
	var cp types.Checkpoint
	var dataJSON sql.NullString
	err := s.db.QueryRow(`
		SELECT id, campaign_id, session, note, data, created_at
		FROM checkpoints
		WHERE campaign_id = ? AND session = ?
		ORDER BY created_at DESC, rowid DESC
		LIMIT 1
	`, campaignID, session).Scan(&cp.ID, &cp.CampaignID, &cp.Session, &cp.Note, &dataJSON, &cp.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query latest checkpoint: %w", err)
	}
	if dataJSON.Valid && dataJSON.String != "" {
		if err := json.Unmarshal([]byte(dataJSON.String), &cp.Data); err != nil {
			return nil, fmt.Errorf("decode checkpoint data: %w", err)
		}
	}
	return &cp, nil
}

// ListCheckpoints returns checkpoints for a campaign/session, ordered by creation time (oldest first).
// If limit > 0, returns only the most recent limit checkpoints.
func (s *Store) ListCheckpoints(campaignID string, session int, limit int) ([]types.Checkpoint, error) {
	query := `
		SELECT id, campaign_id, session, note, data, created_at
		FROM checkpoints
		WHERE campaign_id = ? AND session = ?
		ORDER BY created_at ASC, rowid ASC
	`
	args := []interface{}{campaignID, session}

	if limit > 0 {
		// Get the last N checkpoints by using a subquery to get IDs in reverse order
		query = `
			SELECT id, campaign_id, session, note, data, created_at
			FROM checkpoints
			WHERE campaign_id = ? AND session = ?
			AND id IN (
				SELECT id FROM checkpoints
				WHERE campaign_id = ? AND session = ?
				ORDER BY created_at DESC, rowid DESC
				LIMIT ?
			)
			ORDER BY created_at ASC, rowid ASC
		`
		args = []interface{}{campaignID, session, campaignID, session, limit}
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query checkpoints: %w", err)
	}
	defer rows.Close()

	checkpoints := make([]types.Checkpoint, 0)
	for rows.Next() {
		var cp types.Checkpoint
		var dataJSON sql.NullString
		err := rows.Scan(&cp.ID, &cp.CampaignID, &cp.Session, &cp.Note, &dataJSON, &cp.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan checkpoint: %w", err)
		}
		if dataJSON.Valid && dataJSON.String != "" {
			if err := json.Unmarshal([]byte(dataJSON.String), &cp.Data); err != nil {
				return nil, fmt.Errorf("decode checkpoint data: %w", err)
			}
		}
		checkpoints = append(checkpoints, cp)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate checkpoints: %w", err)
	}

	return checkpoints, nil
}

// UpsertSession stores (or replaces) a completed session summary.
func (s *Store) UpsertSession(session *types.Session) error {
	if session.CreatedAt == "" {
		session.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	}

	_, err := s.db.Exec(`
		INSERT INTO sessions (campaign_id, session, summary, dm_notes, hooks_opened, hooks_resolved, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(campaign_id, session) DO UPDATE SET
			summary = excluded.summary,
			dm_notes = excluded.dm_notes,
			hooks_opened = excluded.hooks_opened,
			hooks_resolved = excluded.hooks_resolved,
			created_at = excluded.created_at
	`, session.CampaignID, session.Session, session.Summary, session.DMNotes, session.HooksOpened, session.HooksResolved, session.CreatedAt)
	if err != nil {
		return fmt.Errorf("upsert session: %w", err)
	}
	return nil
}

// GetLatestSession returns the most recent session in a campaign.
func (s *Store) GetLatestSession(campaignID string) (*types.Session, error) {
	var sess types.Session
	var summaryNull, notesNull sql.NullString
	err := s.db.QueryRow(`
		SELECT campaign_id, session, summary, dm_notes, hooks_opened, hooks_resolved, created_at
		FROM sessions
		WHERE campaign_id = ?
		ORDER BY session DESC
		LIMIT 1
	`, campaignID).Scan(
		&sess.CampaignID,
		&sess.Session,
		&summaryNull,
		&notesNull,
		&sess.HooksOpened,
		&sess.HooksResolved,
		&sess.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query latest session: %w", err)
	}
	if summaryNull.Valid {
		sess.Summary = summaryNull.String
	}
	if notesNull.Valid {
		sess.DMNotes = notesNull.String
	}
	return &sess, nil
}

// GetLastSessionBefore returns the latest session with session number lower than currentSession.
func (s *Store) GetLastSessionBefore(campaignID string, currentSession int) (*types.Session, error) {
	var sess types.Session
	var summaryNull, notesNull sql.NullString
	err := s.db.QueryRow(`
		SELECT campaign_id, session, summary, dm_notes, hooks_opened, hooks_resolved, created_at
		FROM sessions
		WHERE campaign_id = ? AND session < ?
		ORDER BY session DESC
		LIMIT 1
	`, campaignID, currentSession).Scan(
		&sess.CampaignID,
		&sess.Session,
		&summaryNull,
		&notesNull,
		&sess.HooksOpened,
		&sess.HooksResolved,
		&sess.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query previous session: %w", err)
	}
	if summaryNull.Valid {
		sess.Summary = summaryNull.String
	}
	if notesNull.Valid {
		sess.DMNotes = notesNull.String
	}
	return &sess, nil
}

// ListRecentSessionsBefore returns up to limit sessions before currentSession, newest first.
func (s *Store) ListRecentSessionsBefore(campaignID string, currentSession, limit int) ([]types.SessionMeta, error) {
	rows, err := s.db.Query(`
		SELECT session, created_at, summary, hooks_opened, hooks_resolved
		FROM sessions
		WHERE campaign_id = ? AND session < ?
		ORDER BY session DESC
		LIMIT ?
	`, campaignID, currentSession, limit)
	if err != nil {
		return nil, fmt.Errorf("query recent sessions: %w", err)
	}
	defer rows.Close()

	sessions := make([]types.SessionMeta, 0)
	for rows.Next() {
		var meta types.SessionMeta
		var summaryNull sql.NullString
		if err := rows.Scan(&meta.Session, &meta.Date, &summaryNull, &meta.HooksOpened, &meta.HooksResolved); err != nil {
			return nil, fmt.Errorf("scan recent session: %w", err)
		}
		if summaryNull.Valid {
			meta.SummaryPreview = truncatePreview(summaryNull.String, 180)
		}
		sessions = append(sessions, meta)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate recent sessions: %w", err)
	}
	return sessions, nil
}

// ListSessions returns all sessions for a campaign in descending order.
func (s *Store) ListSessions(campaignID string) ([]types.SessionMeta, error) {
	rows, err := s.db.Query(`
		SELECT session, created_at, summary, hooks_opened, hooks_resolved
		FROM sessions
		WHERE campaign_id = ?
		ORDER BY session DESC
	`, campaignID)
	if err != nil {
		return nil, fmt.Errorf("query sessions: %w", err)
	}
	defer rows.Close()

	sessions := make([]types.SessionMeta, 0)
	for rows.Next() {
		var meta types.SessionMeta
		var summaryNull sql.NullString
		if err := rows.Scan(&meta.Session, &meta.Date, &summaryNull, &meta.HooksOpened, &meta.HooksResolved); err != nil {
			return nil, fmt.Errorf("scan session: %w", err)
		}
		if summaryNull.Valid {
			meta.SummaryPreview = truncatePreview(summaryNull.String, 140)
		}
		sessions = append(sessions, meta)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate sessions: %w", err)
	}
	return sessions, nil
}

// CountHooksOpenedInSession returns number of hooks opened from plot events in the provided session.
func (s *Store) CountHooksOpenedInSession(campaignID string, session int) (int, error) {
	var count int
	err := s.db.QueryRow(`
		SELECT COUNT(*)
		FROM plot_hooks h
		INNER JOIN plot_events e ON e.id = h.event_id
		WHERE h.campaign_id = ? AND e.session = ?
	`, campaignID, session).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count hooks opened: %w", err)
	}
	return count, nil
}

// CountHooksResolvedSincePreviousSession approximates hooks resolved during a session by timestamp window.
func (s *Store) CountHooksResolvedSincePreviousSession(campaignID string, currentSession int) (int, error) {
	var prevCreatedAt sql.NullString
	err := s.db.QueryRow(`
		SELECT created_at
		FROM sessions
		WHERE campaign_id = ? AND session < ?
		ORDER BY session DESC
		LIMIT 1
	`, campaignID, currentSession).Scan(&prevCreatedAt)
	if err != nil && err != sql.ErrNoRows {
		return 0, fmt.Errorf("query previous session timestamp: %w", err)
	}

	var count int
	if prevCreatedAt.Valid {
		err = s.db.QueryRow(`
			SELECT COUNT(*)
			FROM plot_hooks
			WHERE campaign_id = ? AND resolved = 1 AND resolved_at > ?
		`, campaignID, prevCreatedAt.String).Scan(&count)
	} else {
		err = s.db.QueryRow(`
			SELECT COUNT(*)
			FROM plot_hooks
			WHERE campaign_id = ? AND resolved = 1
		`, campaignID).Scan(&count)
	}
	if err != nil {
		return 0, fmt.Errorf("count hooks resolved: %w", err)
	}
	return count, nil
}

// AdvanceCampaignSession sets current_session if nextSession is greater than current value.
func (s *Store) AdvanceCampaignSession(campaignID string, nextSession int) error {
	_, err := s.db.Exec(`
		UPDATE campaigns
		SET current_session = CASE WHEN current_session < ? THEN ? ELSE current_session END
		WHERE id = ?
	`, nextSession, nextSession, campaignID)
	if err != nil {
		return fmt.Errorf("advance campaign session: %w", err)
	}
	return nil
}

// QueryNPCRelationships returns relationship edges where source or target is an NPC.
// If npcName is provided, only relationships involving that NPC are returned.
func (s *Store) QueryNPCRelationships(campaignID, npcName string) ([]types.RelationshipEdge, error) {
	rows, err := s.db.Query(`
		SELECT name, type, relationships
		FROM characters
		WHERE campaign_id = ?
	`, campaignID)
	if err != nil {
		return nil, fmt.Errorf("query relationships: %w", err)
	}
	defer rows.Close()

	typesByName := make(map[string]string)
	relationshipsByName := make(map[string]map[string]string)

	for rows.Next() {
		var name, charType string
		var relationshipsJSON sql.NullString
		if err := rows.Scan(&name, &charType, &relationshipsJSON); err != nil {
			return nil, fmt.Errorf("scan relationships: %w", err)
		}
		typesByName[name] = charType

		rels := map[string]string{}
		if relationshipsJSON.Valid && strings.TrimSpace(relationshipsJSON.String) != "" {
			if err := json.Unmarshal([]byte(relationshipsJSON.String), &rels); err != nil {
				return nil, fmt.Errorf("decode relationships for %s: %w", name, err)
			}
		}
		relationshipsByName[name] = rels
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate relationships: %w", err)
	}

	edges := make([]types.RelationshipEdge, 0)
	for source, rels := range relationshipsByName {
		sourceType := typesByName[source]
		if sourceType == "" {
			sourceType = "unknown"
		}

		for target, relation := range rels {
			targetType := typesByName[target]
			if targetType == "" {
				targetType = "unknown"
			}

			if sourceType != "npc" && targetType != "npc" {
				continue
			}
			if npcName != "" && !strings.EqualFold(source, npcName) && !strings.EqualFold(target, npcName) {
				continue
			}

			edges = append(edges, types.RelationshipEdge{
				Source:     source,
				SourceType: sourceType,
				Target:     target,
				TargetType: targetType,
				Relation:   relation,
			})
		}
	}

	sort.Slice(edges, func(i, j int) bool {
		if edges[i].Source == edges[j].Source {
			if edges[i].Target == edges[j].Target {
				return edges[i].Relation < edges[j].Relation
			}
			return edges[i].Target < edges[j].Target
		}
		return edges[i].Source < edges[j].Source
	})

	return edges, nil
}

func truncatePreview(s string, maxLen int) string {
	trimmed := strings.TrimSpace(s)
	if len(trimmed) <= maxLen {
		return trimmed
	}
	if maxLen <= 3 {
		return trimmed[:maxLen]
	}
	return trimmed[:maxLen-3] + "..."
}

func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for _, s := range strs[1:] {
		result += sep + s
	}
	return result
}
