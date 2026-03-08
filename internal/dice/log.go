// Package dice provides dice notation parsing and rolling.
package dice

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/andr1an/llmdm/internal/types"
)

// Logger handles dice roll logging to the database.
type Logger struct {
	db *sql.DB
}

// NewLogger creates a new dice roll logger.
func NewLogger(db *sql.DB) *Logger {
	return &Logger{db: db}
}

// Log saves a roll result to the database.
func (l *Logger) Log(campaignID string, session int, character, reason string, result types.RollResult, advantage, disadvantage bool) error {
	rollsJSON, err := json.Marshal(result.Rolls)
	if err != nil {
		return fmt.Errorf("marshal rolls: %w", err)
	}

	var keptJSON []byte
	if len(result.Kept) > 0 {
		keptJSON, err = json.Marshal(result.Kept)
		if err != nil {
			return fmt.Errorf("marshal kept: %w", err)
		}
	}

	advInt := 0
	if advantage {
		advInt = 1
	}
	disadvInt := 0
	if disadvantage {
		disadvInt = 1
	}

	_, err = l.db.Exec(`
		INSERT INTO roll_log (id, campaign_id, session, character, notation, total, rolls, kept, modifier, reason, advantage, disadvantage, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, result.RollID, campaignID, session, character, result.Notation, result.Total, string(rollsJSON), string(keptJSON), result.Modifier, reason, advInt, disadvInt, result.Timestamp)

	if err != nil {
		return fmt.Errorf("insert roll log: %w", err)
	}
	return nil
}

// GetHistory retrieves roll history from the database.
func (l *Logger) GetHistory(campaignID string, character string, session int, limit int) ([]types.RollRecord, error) {
	if limit <= 0 {
		limit = 50
	}

	query := `SELECT id, campaign_id, session, character, notation, total, rolls, kept, modifier, reason, advantage, disadvantage, created_at
		FROM roll_log WHERE campaign_id = ?`
	args := []interface{}{campaignID}

	if character != "" {
		query += " AND character = ?"
		args = append(args, character)
	}
	if session > 0 {
		query += " AND session = ?"
		args = append(args, session)
	}

	query += " ORDER BY created_at DESC LIMIT ?"
	args = append(args, limit)

	rows, err := l.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query roll log: %w", err)
	}
	defer rows.Close()

	var records []types.RollRecord
	for rows.Next() {
		var r types.RollRecord
		var rollsJSON, keptJSON sql.NullString
		var advInt, disadvInt int
		var sessionNull sql.NullInt64
		var charNull, reasonNull sql.NullString

		err := rows.Scan(&r.ID, &r.CampaignID, &sessionNull, &charNull, &r.Notation, &r.Total, &rollsJSON, &keptJSON, &r.Modifier, &reasonNull, &advInt, &disadvInt, &r.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan roll log: %w", err)
		}

		if sessionNull.Valid {
			r.Session = int(sessionNull.Int64)
		}
		if charNull.Valid {
			r.Character = charNull.String
		}
		if reasonNull.Valid {
			r.Reason = reasonNull.String
		}
		r.Advantage = advInt == 1
		r.Disadvantage = disadvInt == 1

		if rollsJSON.Valid {
			json.Unmarshal([]byte(rollsJSON.String), &r.Rolls)
		}
		if keptJSON.Valid && keptJSON.String != "" {
			json.Unmarshal([]byte(keptJSON.String), &r.Kept)
		}

		records = append(records, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate roll log: %w", err)
	}

	return records, nil
}
