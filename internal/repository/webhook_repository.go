package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"
	webhookTypes "github.com/valentinesamuel/activelog/internal/webhook/types"
)

// WebhookRepository handles database operations for webhooks
type WebhookRepository struct {
	db DBConn
}

// NewWebhookRepository creates a new WebhookRepository
func NewWebhookRepository(db DBConn) *WebhookRepository {
	return &WebhookRepository{db: db}
}

// Create inserts a new webhook and sets its ID from RETURNING
func (r *WebhookRepository) Create(ctx context.Context, wh *webhookTypes.Webhook) error {
	query := `
		INSERT INTO webhooks (user_id, url, events, secret, active)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at`

	return r.db.QueryRowContext(ctx, query,
		wh.UserID,
		wh.URL,
		pq.Array(wh.Events),
		wh.Secret,
		wh.Active,
	).Scan(&wh.ID, &wh.CreatedAt)
}

// Delete removes a webhook by ID for a specific user
func (r *WebhookRepository) Delete(ctx context.Context, id string, userID int) error {
	query := `DELETE FROM webhooks WHERE id = $1 AND user_id = $2`
	result, err := r.db.ExecContext(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete webhook: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("webhook not found: %s", id)
	}
	return nil
}

// ListByUserID returns all webhooks for a user
func (r *WebhookRepository) ListByUserID(ctx context.Context, userID int) ([]*webhookTypes.Webhook, error) {
	query := `
		SELECT id, user_id, url, events, secret, active, created_at
		FROM webhooks WHERE user_id = $1 ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list webhooks: %w", err)
	}
	defer rows.Close()

	var webhooks []*webhookTypes.Webhook
	for rows.Next() {
		wh := &webhookTypes.Webhook{}
		if err := rows.Scan(
			&wh.ID, &wh.UserID, &wh.URL, pq.Array(&wh.Events), &wh.Secret, &wh.Active, &wh.CreatedAt,
		); err != nil {
			return nil, err
		}
		webhooks = append(webhooks, wh)
	}
	return webhooks, rows.Err()
}

// ListByEvent returns all active webhooks subscribed to a given event type
func (r *WebhookRepository) ListByEvent(ctx context.Context, eventType string) ([]*webhookTypes.Webhook, error) {
	query := `
		SELECT id, user_id, url, events, secret, active, created_at
		FROM webhooks WHERE active = true AND $1 = ANY(events)`

	rows, err := r.db.QueryContext(ctx, query, eventType)
	if err != nil {
		return nil, fmt.Errorf("failed to list webhooks by event: %w", err)
	}
	defer rows.Close()

	var webhooks []*webhookTypes.Webhook
	for rows.Next() {
		wh := &webhookTypes.Webhook{}
		if err := rows.Scan(
			&wh.ID, &wh.UserID, &wh.URL, pq.Array(&wh.Events), &wh.Secret, &wh.Active, &wh.CreatedAt,
		); err != nil {
			return nil, err
		}
		webhooks = append(webhooks, wh)
	}
	return webhooks, rows.Err()
}

// GetByID fetches a webhook by its ID
func (r *WebhookRepository) GetByID(ctx context.Context, id string) (*webhookTypes.Webhook, error) {
	query := `
		SELECT id, user_id, url, events, secret, active, created_at
		FROM webhooks WHERE id = $1`

	wh := &webhookTypes.Webhook{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&wh.ID, &wh.UserID, &wh.URL, pq.Array(&wh.Events), &wh.Secret, &wh.Active, &wh.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("webhook not found: %s", id)
	}
	return wh, err
}

// CreateDelivery inserts a new webhook delivery record
func (r *WebhookRepository) CreateDelivery(ctx context.Context, d *webhookTypes.WebhookDelivery) error {
	query := `
		INSERT INTO webhook_deliveries (webhook_id, event_type, payload, status, attempt_count, max_attempts, next_retry_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at`

	return r.db.QueryRowContext(ctx, query,
		d.WebhookID,
		d.EventType,
		d.Payload,
		d.Status,
		d.AttemptCount,
		d.MaxAttempts,
		d.NextRetryAt,
	).Scan(&d.ID, &d.CreatedAt, &d.UpdatedAt)
}

// MarkDeliverySucceeded updates a delivery to succeeded status
func (r *WebhookRepository) MarkDeliverySucceeded(ctx context.Context, id string, httpStatus int) error {
	query := `
		UPDATE webhook_deliveries
		SET status = 'succeeded', last_http_status = $2, updated_at = NOW()
		WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id, httpStatus)
	return err
}

// MarkDeliveryFailed updates a delivery to failed or exhausted status
func (r *WebhookRepository) MarkDeliveryFailed(ctx context.Context, id string, httpStatus *int, errMsg string, nextRetryAt *time.Time) error {
	var status webhookTypes.DeliveryStatus
	if nextRetryAt == nil {
		status = webhookTypes.DeliveryStatusExhausted
	} else {
		status = webhookTypes.DeliveryStatusFailed
	}

	query := `
		UPDATE webhook_deliveries
		SET status = $2, last_http_status = $3, last_error = $4, next_retry_at = COALESCE($5, next_retry_at),
		    attempt_count = attempt_count + 1, updated_at = NOW()
		WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id, status, httpStatus, errMsg, nextRetryAt)
	return err
}

// ListPendingRetries returns deliveries that are due for retry
func (r *WebhookRepository) ListPendingRetries(ctx context.Context, limit int) ([]*webhookTypes.WebhookDelivery, error) {
	query := `
		SELECT wd.id, wd.webhook_id, wd.event_type, wd.payload, wd.status,
		       wd.attempt_count, wd.max_attempts, wd.last_http_status, wd.last_error,
		       wd.next_retry_at, wd.created_at, wd.updated_at
		FROM webhook_deliveries wd
		JOIN webhooks w ON w.id = wd.webhook_id
		WHERE w.active = true
		  AND wd.status IN ('pending', 'failed')
		  AND wd.next_retry_at <= NOW()
		ORDER BY wd.next_retry_at
		LIMIT $1`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list pending retries: %w", err)
	}
	defer rows.Close()

	var deliveries []*webhookTypes.WebhookDelivery
	for rows.Next() {
		d := &webhookTypes.WebhookDelivery{}
		if err := rows.Scan(
			&d.ID, &d.WebhookID, &d.EventType, &d.Payload, &d.Status,
			&d.AttemptCount, &d.MaxAttempts, &d.LastHTTPStatus, &d.LastError,
			&d.NextRetryAt, &d.CreatedAt, &d.UpdatedAt,
		); err != nil {
			return nil, err
		}
		deliveries = append(deliveries, d)
	}
	return deliveries, rows.Err()
}
