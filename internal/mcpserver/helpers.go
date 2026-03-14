package mcpserver

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/andr1an/llmdm/internal/memory"
)

// wrapError creates a wrapped error with operation context.
func wrapError(operation string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", operation, err)
}

// DBContext provides access to database and store for tool handlers.
type DBContext struct {
	// DB is the raw sql.DB for low-level operations (like dice logger)
	DB *sql.DB
	// Store provides high-level database operations
	Store *memory.Store
}

// withDB opens a campaign database, runs the provided function with it, and ensures cleanup.
// This eliminates repetitive database open/close/error handling code.
func (s *Server) withDB(ctx context.Context, campaignID string, fn func(*DBContext) error) error {
	database, err := s.openCampaignDB(campaignID)
	if err != nil {
		s.log().Error("failed to open campaign database", "campaign_id", campaignID, "error", err)
		return wrapError("open database", err)
	}
	defer database.Close()

	dbCtx := &DBContext{
		DB:    database.DB,
		Store: memory.NewStore(database.DB),
	}

	if err := fn(dbCtx); err != nil {
		return err
	}

	return nil
}
