package webhook

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/valentinesamuel/activelog/internal/repository"
	webhookTypes "github.com/valentinesamuel/activelog/internal/webhook/types"
)

// Delivery handles delivering webhook events to registered endpoints
type Delivery struct {
	webhookRepo *repository.WebhookRepository
	httpClient  *http.Client
}

// NewDelivery creates a new Delivery handler
func NewDelivery(webhookRepo *repository.WebhookRepository) *Delivery {
	return &Delivery{
		webhookRepo: webhookRepo,
		httpClient:  &http.Client{Timeout: 10 * time.Second},
	}
}

// Handle is the subscriber handler - it delivers events to all matching webhook endpoints
func (d *Delivery) Handle(ctx context.Context, event webhookTypes.WebhookEvent) {
	webhooks, err := d.webhookRepo.ListByEvent(ctx, event.EventType)
	if err != nil {
		log.Printf("Error fetching webhooks for event %s: %v", event.EventType, err)
		return
	}

	for _, wh := range webhooks {
		if err := d.deliver(ctx, wh, event); err != nil {
			log.Printf("Failed to deliver webhook event %s to %s: %v", event.EventType, wh.URL, err)
		}
	}
}

// deliver sends the event to a single webhook endpoint with retries
func (d *Delivery) deliver(ctx context.Context, wh *webhookTypes.Webhook, event webhookTypes.WebhookEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	sig := computeSignature(wh.Secret, body)

	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			wait := time.Duration(1<<attempt) * time.Second
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(wait):
			}
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, wh.URL, bytes.NewReader(body))
		if err != nil {
			return fmt.Errorf("create request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Webhook-Signature", sig)
		req.Header.Set("X-Webhook-Event", event.EventType)

		resp, err := d.httpClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return nil
		}
		lastErr = fmt.Errorf("non-2xx status: %d", resp.StatusCode)
	}

	return lastErr
}

// computeSignature generates the HMAC-SHA256 signature for a webhook payload
func computeSignature(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}
