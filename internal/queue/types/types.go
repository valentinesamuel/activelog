package types

import (
	"context"
	"encoding/json"
)

// QueueName identifies which queue a job should go into
type QueueName string

const (
	InboxQueue  QueueName = "inbox"
	OutboxQueue QueueName = "outbox"
)

// EventType identifies which handler should process a job
type EventType string

// Inbox events
const (
	EventWelcomeEmail             EventType = "welcome_email"
	EventWeeklySummary            EventType = "weekly_summary"
	EventGenerateExport           EventType = "generate_export"
	EventSendVerificationEmail    EventType = "send_verification_email"
	EventRefreshRateLimitConfig   EventType = "refresh_rate_limit_config"
)

// Outbox events
const (
	EventActivityCreated EventType = "activity_created"
	EventActivityDeleted EventType = "activity_deleted"
)

// JobPayload is the envelope for every queued job
type JobPayload struct {
	Event EventType       `json:"event"`
	Data  json.RawMessage `json:"data"`
}

// QueueProvider is the interface all queue backends must implement
type QueueProvider interface {
	Enqueue(ctx context.Context, queue QueueName, payload JobPayload) (taskID string, err error)
}
