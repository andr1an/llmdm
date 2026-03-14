// Package db provides SQLite database connection and migrations.
package db

import (
	"database/sql"
	_ "embed"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"time"

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
	start := time.Now()
	slog.Debug("opening database", "db_path", dbPath)

	// Ensure parent directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		slog.Error("failed to create database directory", "dir", dir, "error", err)
		return nil, fmt.Errorf("create db directory: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		slog.Error("failed to open database", "db_path", dbPath, "error", err)
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
			slog.Error("failed to set pragma", "pragma", pragma, "error", err)
			db.Close()
			return nil, fmt.Errorf("set pragma: %w", err)
		}
	}

	slog.Debug("database opened successfully", "db_path", dbPath, "duration_ms", time.Since(start).Milliseconds())
	return &DB{DB: db, path: dbPath}, nil
}

// Migrate runs the schema migrations.
func (db *DB) Migrate() error {
	start := time.Now()
	slog.Debug("running database migrations", "db_path", db.path)

	_, err := db.Exec(schema)
	if err != nil {
		slog.Error("migration failed", "db_path", db.path, "error", err)
		return fmt.Errorf("run migrations: %w", err)
	}

	// Add gold column to existing databases (ignore error if column exists)
	_, _ = db.Exec(`ALTER TABLE characters ADD COLUMN gold INTEGER DEFAULT 0`)

	// Add data column to checkpoints table (ignore error if column exists)
	_, _ = db.Exec(`ALTER TABLE checkpoints ADD COLUMN data TEXT`)

	slog.Debug("migrations completed successfully", "db_path", db.path, "duration_ms", time.Since(start).Milliseconds())
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
