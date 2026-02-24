package models

import "time"

// ExportFormat represents the output format of an export.
type ExportFormat string

const (
	FormatCSV ExportFormat = "csv"
	FormatPDF ExportFormat = "pdf"
)

// ExportStatus represents the current state of an export job.
type ExportStatus string

const (
	StatusPending    ExportStatus = "pending"
	StatusProcessing ExportStatus = "processing"
	StatusCompleted  ExportStatus = "completed"
	StatusFailed     ExportStatus = "failed"
)

// ExportRecord represents a row in the exports table.
type ExportRecord struct {
	ID           string       `json:"id"`
	UserID       int          `json:"user_id"`
	Format       ExportFormat `json:"format"`
	Status       ExportStatus `json:"status"`
	S3Key        *string      `json:"s3_key,omitempty"`
	ErrorMessage *string      `json:"error_message,omitempty"`
	CreatedAt    time.Time    `json:"created_at"`
	CompletedAt  *time.Time   `json:"completed_at,omitempty"`
}
