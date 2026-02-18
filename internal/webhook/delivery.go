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

var retryDelays = []time.Duration{
	1 * time.Minute,
	5 * time.Minute,
	30 * time.Minute,
	2 * time.Hour,
	24 * time.Hour,
}

const maxAttempts = 5

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

// Handle is the subscriber handler - creates DB records and dispatches async goroutines
func (d *Delivery) Handle(ctx context.Context, event webhookTypes.WebhookEvent) {
	webhooks, err := d.webhookRepo.ListByEvent(ctx, event.EventType)
	if err != nil {
		log.Printf("Error fetching webhooks for event %s: %v", event.EventType, err)
		return
	}

	payload, err := json.Marshal(event)
	if err != nil {
		log.Printf("Error marshaling event payload: %v", err)
		return
	}

	for _, wh := range webhooks {
		delivery := &webhookTypes.WebhookDelivery{
			WebhookID:    wh.ID,
			EventType:    event.EventType,
			Payload:      payload,
			Status:       webhookTypes.DeliveryStatusPending,
			AttemptCount: 0,
			MaxAttempts:  maxAttempts,
			NextRetryAt:  time.Now(),
		}
		if err := d.webhookRepo.CreateDelivery(ctx, delivery); err != nil {
			log.Printf("Error creating delivery record for webhook %s: %v", wh.ID, err)
			continue
		}
		go d.dispatchAsync(wh, delivery, event)
	}
}

func (d *Delivery) dispatchAsync(wh *webhookTypes.Webhook, delivery *webhookTypes.WebhookDelivery, event webhookTypes.WebhookEvent) {
	ctx := context.Background()
	body, err := json.Marshal(event)
	if err != nil {
		log.Printf("dispatchAsync: marshal error for delivery %s: %v", delivery.ID, err)
		return
	}
	sig := computeSignature(wh.Secret, body)
	d.executeDelivery(ctx, wh.URL, event.EventType, sig, body, delivery)
}

func (d *Delivery) executeDelivery(ctx context.Context, url, eventType, sig string, body []byte, delivery *webhookTypes.WebhookDelivery) {
	statusCode, err := d.attemptHTTP(ctx, url, eventType, sig, body)

	if err == nil && statusCode >= 200 && statusCode < 300 {
		if dbErr := d.webhookRepo.MarkDeliverySucceeded(ctx, delivery.ID, statusCode); dbErr != nil {
			log.Printf("Error marking delivery succeeded: %v", dbErr)
		}
		return
	}

	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	} else {
		errMsg = fmt.Sprintf("non-2xx status: %d", statusCode)
	}

	nextAttempt := delivery.AttemptCount + 1
	var nextRetryAt *time.Time
	if nextAttempt < maxAttempts && nextAttempt < len(retryDelays) {
		t := time.Now().Add(retryDelays[nextAttempt])
		nextRetryAt = &t
	}

	var httpStatusPtr *int
	if statusCode > 0 {
		httpStatusPtr = &statusCode
	}

	if dbErr := d.webhookRepo.MarkDeliveryFailed(ctx, delivery.ID, httpStatusPtr, errMsg, nextRetryAt); dbErr != nil {
		log.Printf("Error marking delivery failed: %v", dbErr)
	}
}

func (d *Delivery) attemptHTTP(ctx context.Context, url, eventType, sig string, body []byte) (int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return 0, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Signature", sig)
	req.Header.Set("X-Webhook-Event", eventType)

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	resp.Body.Close()
	return resp.StatusCode, nil
}

// computeSignature generates the HMAC-SHA256 signature for a webhook payload
func computeSignature(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}
