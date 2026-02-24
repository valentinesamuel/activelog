package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/valentinesamuel/activelog/internal/models"
)

// ExportRepository handles database operations for export records.
type ExportRepository struct {
	db DBConn
}

// NewExportRepository creates a new ExportRepository.
func NewExportRepository(db DBConn) *ExportRepository {
	return &ExportRepository{db: db}
}

// Create inserts a new export record and sets its ID from RETURNING.
func (r *ExportRepository) Create(ctx context.Context, record *models.ExportRecord) error {
	query := `
		INSERT INTO exports (user_id, format, status)
		VALUES ($1, $2, $3)
		RETURNING id, created_at`

	err := r.db.QueryRowContext(ctx, query,
		record.UserID,
		record.Format,
		record.Status,
	).Scan(&record.ID, &record.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create export record: %w", err)
	}

	return nil
}

// UpdateStatus updates the status, s3_key, error_message, and completed_at fields.
func (r *ExportRepository) UpdateStatus(ctx context.Context, id string, status models.ExportStatus, s3Key *string, errMsg *string) error {
	var completedAt *time.Time
	if status == models.StatusCompleted || status == models.StatusFailed {
		now := time.Now()
		completedAt = &now
	}

	query := `
		UPDATE exports
		SET status = $1, s3_key = $2, error_message = $3, completed_at = $4
		WHERE id = $5`

	result, err := r.db.ExecContext(ctx, query, status, s3Key, errMsg, completedAt, id)
	if err != nil {
		return fmt.Errorf("failed to update export status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("export record not found: %s", id)
	}

	return nil
}

// GetByID fetches an export record by UUID string.
func (r *ExportRepository) GetByID(ctx context.Context, id string) (*models.ExportRecord, error) {
	query := `
		SELECT id, user_id, format, status, s3_key, error_message, created_at, completed_at
		FROM exports
		WHERE id = $1`

	record := &models.ExportRecord{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&record.ID,
		&record.UserID,
		&record.Format,
		&record.Status,
		&record.S3Key,
		&record.ErrorMessage,
		&record.CreatedAt,
		&record.CompletedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("export record not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get export record: %w", err)
	}

	return record, nil
}

// ListByUser fetches all exports for a user ordered by created_at DESC.
func (r *ExportRepository) ListByUser(ctx context.Context, userID int) ([]*models.ExportRecord, error) {
	query := `
		SELECT id, user_id, format, status, s3_key, error_message, created_at, completed_at
		FROM exports
		WHERE user_id = $1
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list exports: %w", err)
	}
	defer rows.Close()

	var records []*models.ExportRecord
	for rows.Next() {
		record := &models.ExportRecord{}
		err := rows.Scan(
			&record.ID,
			&record.UserID,
			&record.Format,
			&record.Status,
			&record.S3Key,
			&record.ErrorMessage,
			&record.CreatedAt,
			&record.CompletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan export record: %w", err)
		}
		records = append(records, record)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating export rows: %w", err)
	}

	return records, nil
}
