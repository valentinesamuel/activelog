package nats

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/nats-io/nats.go"
	webhookTypes "github.com/valentinesamuel/activelog/internal/webhook/types"
)

const (
	streamSubject  = "webhook.events"
	streamName     = "WEBHOOK_EVENTS"
	consumerName   = "webhook-delivery"
	fetchBatchSize = 10
	fetchMaxWait   = 5 * time.Second
)

// Provider is a NATS JetStream webhook bus provider
type Provider struct {
	nc *nats.Conn
	js nats.JetStreamContext
}

// New creates a new NATS JetStream webhook bus provider
func New(url string) (*Provider, error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, err
	}

	js, err := nc.JetStream()
	if err != nil {
		nc.Close()
		return nil, err
	}

	// Create or bind the stream
	_, err = js.AddStream(&nats.StreamConfig{
		Name:     streamName,
		Subjects: []string{streamSubject},
	})
	if err != nil {
		// Stream may already exist, try to get it
		_, err = js.StreamInfo(streamName)
		if err != nil {
			nc.Close()
			return nil, err
		}
	}

	return &Provider{nc: nc, js: js}, nil
}

// Publish serializes a WebhookEvent and publishes it to NATS JetStream
func (p *Provider) Publish(ctx context.Context, event webhookTypes.WebhookEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	_, err = p.js.Publish(streamSubject, data)
	return err
}

// Subscribe creates a durable pull consumer and starts reading messages
func (p *Provider) Subscribe(ctx context.Context, handler func(ctx context.Context, event webhookTypes.WebhookEvent)) error {
	sub, err := p.js.PullSubscribe(streamSubject, consumerName)
	if err != nil {
		return err
	}
	go p.readLoop(ctx, sub, handler)
	return nil
}

func (p *Provider) readLoop(ctx context.Context, sub *nats.Subscription, handler func(ctx context.Context, event webhookTypes.WebhookEvent)) {
	for {
		select {
		case <-ctx.Done():
			sub.Unsubscribe()
			return
		default:
		}

		msgs, err := sub.Fetch(fetchBatchSize, nats.MaxWait(fetchMaxWait))
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			// Timeout is expected when no messages available
			continue
		}

		for _, msg := range msgs {
			var event webhookTypes.WebhookEvent
			if err := json.Unmarshal(msg.Data, &event); err != nil {
				log.Printf("Failed to unmarshal NATS webhook event, dropping: %v", err)
				msg.Ack()
				continue
			}
			handler(ctx, event)
			msg.Ack()
		}
	}
}
