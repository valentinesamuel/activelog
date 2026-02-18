package webhook

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/valentinesamuel/activelog/internal/repository"
	webhookTypes "github.com/valentinesamuel/activelog/internal/webhook/types"
)

const retryPollInterval = 30 * time.Second

// RetryWorker polls the DB for failed deliveries due for retry
type RetryWorker struct {
	webhookRepo *repository.WebhookRepository
	delivery    *Delivery
}

// NewRetryWorker creates a new RetryWorker
func NewRetryWorker(repo *repository.WebhookRepository, delivery *Delivery) *RetryWorker {
	return &RetryWorker{webhookRepo: repo, delivery: delivery}
}

// Start launches the retry polling loop in a goroutine
func (w *RetryWorker) Start(ctx context.Context) {
	go w.run(ctx)
}

func (w *RetryWorker) run(ctx context.Context) {
	ticker := time.NewTicker(retryPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.poll(ctx)
		}
	}
}

func (w *RetryWorker) poll(ctx context.Context) {
	deliveries, err := w.webhookRepo.ListPendingRetries(ctx, 100)
	if err != nil {
		log.Printf("RetryWorker: error listing pending retries: %v", err)
		return
	}

	for _, d := range deliveries {
		wh, err := w.webhookRepo.GetByID(ctx, d.WebhookID)
		if err != nil {
			log.Printf("RetryWorker: error fetching webhook %s: %v", d.WebhookID, err)
			continue
		}

		var event webhookTypes.WebhookEvent
		if err := json.Unmarshal(d.Payload, &event); err != nil {
			log.Printf("RetryWorker: error unmarshaling payload for delivery %s: %v", d.ID, err)
			continue
		}

		go w.retryDelivery(ctx, wh, d, event)
	}
}

func (w *RetryWorker) retryDelivery(ctx context.Context, wh *webhookTypes.Webhook, d *webhookTypes.WebhookDelivery, event webhookTypes.WebhookEvent) {
	body, err := json.Marshal(event)
	if err != nil {
		log.Printf("RetryWorker: marshal error for delivery %s: %v", d.ID, err)
		return
	}
	sig := computeSignature(wh.Secret, body)
	w.delivery.executeDelivery(ctx, wh.URL, event.EventType, sig, body, d)
}
