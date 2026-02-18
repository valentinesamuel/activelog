package repository

import (
	"context"
	"database/sql"
	"fmt"

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
