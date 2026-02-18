package service

import (
	"context"
	"database/sql"
	"log"
)

// CleanupService hard-deletes soft-deleted records that are older than 30 days.
type CleanupService struct {
	db *sql.DB
}

// NewCleanupService creates a CleanupService backed by a raw *sql.DB.
func NewCleanupService(db *sql.DB) *CleanupService {
	return &CleanupService{db: db}
}

// DeleteOldData permanently removes records soft-deleted more than 30 days ago.
func (c *CleanupService) DeleteOldData(ctx context.Context) error {
	query := `
		DELETE FROM activities
		WHERE deleted_at IS NOT NULL
		  AND deleted_at < NOW() - INTERVAL '30 days'
	`

	result, err := c.db.ExecContext(ctx, query)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	log.Printf("[scheduler] cleanup: hard-deleted %d stale activities", rows)
	return nil
}
