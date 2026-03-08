package mcpserver

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/andr1an/llmdm/internal/db"
)

func (s *Server) openCampaignDB(campaignID string) (*db.DB, error) {
	dbPath, err := db.CampaignDBPath(s.dbPath, campaignID)
	if err != nil {
		return nil, fmt.Errorf("invalid campaign_id: %w", err)
	}
	s.log().Debug("opening campaign database", "campaign_id", campaignID, "db_path", dbPath)
	database, err := db.Open(dbPath)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	if err := database.Migrate(); err != nil {
		database.Close()
		return nil, fmt.Errorf("migrate database: %w", err)
	}
	if err := s.reconcileCampaignID(database.DB, campaignID); err != nil {
		database.Close()
		return nil, fmt.Errorf("reconcile campaign id: %w", err)
	}
	s.log().Debug("campaign database ready", "campaign_id", campaignID)
	return database, nil
}

func (s *Server) reconcileCampaignID(rawDB *sql.DB, campaignID string) error {
	if strings.TrimSpace(campaignID) == "" {
		return nil
	}

	var exactMatch int
	if err := rawDB.QueryRow(`SELECT COUNT(1) FROM campaigns WHERE id = ?`, campaignID).Scan(&exactMatch); err != nil {
		return fmt.Errorf("check campaign id existence: %w", err)
	}
	if exactMatch > 0 {
		return nil
	}

	var legacyID string
	var name string
	var description sql.NullString
	var createdAt string
	var currentSession int
	if err := rawDB.QueryRow(`
		SELECT id, name, description, created_at, current_session
		FROM campaigns
		LIMIT 1
	`).Scan(&legacyID, &name, &description, &createdAt, &currentSession); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return fmt.Errorf("load legacy campaign: %w", err)
	}

	var campaignCount int
	if err := rawDB.QueryRow(`SELECT COUNT(1) FROM campaigns`).Scan(&campaignCount); err != nil {
		return fmt.Errorf("count campaigns: %w", err)
	}
	if campaignCount != 1 || legacyID == campaignID {
		return nil
	}

	desc := interface{}(nil)
	if description.Valid {
		desc = description.String
	}

	tx, err := rawDB.Begin()
	if err != nil {
		return fmt.Errorf("begin campaign id reconciliation tx: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`
		INSERT INTO campaigns (id, name, description, created_at, current_session)
		VALUES (?, ?, ?, ?, ?)
	`, campaignID, name, desc, createdAt, currentSession); err != nil {
		return fmt.Errorf("insert reconciled campaign: %w", err)
	}

	tables := []string{"characters", "plot_events", "plot_hooks", "world_flags", "roll_log", "sessions", "checkpoints"}
	for _, table := range tables {
		query := fmt.Sprintf("UPDATE %s SET campaign_id = ? WHERE campaign_id = ?", table)
		if _, err := tx.Exec(query, campaignID, legacyID); err != nil {
			return fmt.Errorf("update campaign references in %s: %w", table, err)
		}
	}

	if _, err := tx.Exec(`DELETE FROM campaigns WHERE id = ?`, legacyID); err != nil {
		return fmt.Errorf("delete legacy campaign row: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit campaign id reconciliation tx: %w", err)
	}

	s.log().Info("reconciled legacy campaign id", "legacy_id", legacyID, "campaign_id", campaignID)
	return nil
}
