package types

import (
	"context"
	"encoding/json"
	"time"
)

// WebhookEvent is the internal event published when something happens
type WebhookEvent struct {
	EventType string          `json:"event_type"`
	UserID    int             `json:"user_id"`
	Payload   json.RawMessage `json:"payload"`
	Timestamp time.Time       `json:"timestamp"`
}

// WebhookBusProvider is the event-bus that fans events to subscribers
type WebhookBusProvider interface {
	Publish(ctx context.Context, event WebhookEvent) error
	Subscribe(ctx context.Context, handler func(ctx context.Context, event WebhookEvent)) error
}

// Webhook event type constants
const (
	EventActivityCreated = "activity.created"
	EventActivityDeleted = "activity.deleted"
	EventActivityUpdated = "activity.updated"
)

// Webhook represents a registered webhook endpoint
type Webhook struct {
	ID        string    `json:"id"`
	UserID    int       `json:"user_id"`
	URL       string    `json:"url"`
	Events    []string  `json:"events"`
	Secret    string    `json:"-"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
}

// DeliveryStatus represents the status of a webhook delivery attempt
type DeliveryStatus string

const (
	DeliveryStatusPending   DeliveryStatus = "pending"
	DeliveryStatusSucceeded DeliveryStatus = "succeeded"
	DeliveryStatusFailed    DeliveryStatus = "failed"
	DeliveryStatusExhausted DeliveryStatus = "exhausted"
)

// WebhookDelivery represents a single delivery attempt persisted in the DB
type WebhookDelivery struct {
	ID             string          `db:"id"`
	WebhookID      string          `db:"webhook_id"`
	EventType      string          `db:"event_type"`
	Payload        json.RawMessage `db:"payload"`
	Status         DeliveryStatus  `db:"status"`
	AttemptCount   int             `db:"attempt_count"`
	MaxAttempts    int             `db:"max_attempts"`
	LastHTTPStatus *int            `db:"last_http_status"`
	LastError      *string         `db:"last_error"`
	NextRetryAt    time.Time       `db:"next_retry_at"`
	CreatedAt      time.Time       `db:"created_at"`
	UpdatedAt      time.Time       `db:"updated_at"`
}
