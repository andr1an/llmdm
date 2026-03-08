// Package db provides SQLite database connection and migrations.
package db

import (
	"database/sql"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schema string

var campaignIDPattern = regexp.MustCompile(`^[a-z]+(?:-[a-z]+)*$`)

// DB wraps the SQLite database connection.
type DB struct {
	*sql.DB
	path string
}

// Open opens or creates a SQLite database at the given path.
func Open(dbPath string) (*DB, error) {
	// Ensure parent directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("create db directory: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// Enable foreign keys and WAL mode
	pragmas := []string{
		"PRAGMA foreign_keys = ON",
		"PRAGMA journal_mode = WAL",
		"PRAGMA busy_timeout = 5000",
	}
	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			db.Close()
			return nil, fmt.Errorf("set pragma: %w", err)
		}
	}

	return &DB{DB: db, path: dbPath}, nil
}

// Migrate runs the schema migrations.
func (db *DB) Migrate() error {
	_, err := db.Exec(schema)
	if err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}
	return nil
}

// Path returns the database file path.
func (db *DB) Path() string {
	return db.path
}

// IsValidCampaignID validates campaign IDs used for database filenames.
func IsValidCampaignID(campaignID string) bool {
	return campaignIDPattern.MatchString(campaignID)
}

// CampaignDBPath returns the path for a campaign's database file.
func CampaignDBPath(basePath, campaignID string) (string, error) {
	if !IsValidCampaignID(campaignID) {
		return "", fmt.Errorf("campaign_id must match regex %q", campaignIDPattern.String())
	}
	return filepath.Join(basePath, campaignID+".db"), nil
}
