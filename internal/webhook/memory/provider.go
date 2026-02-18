package memory

import (
	"context"
	"log"
	"sync"

	webhookTypes "github.com/valentinesamuel/activelog/internal/webhook/types"
)

// Provider is an in-process webhook bus using a buffered channel
type Provider struct {
	ch       chan webhookTypes.WebhookEvent
	mu       sync.RWMutex
	handlers []func(ctx context.Context, event webhookTypes.WebhookEvent)
}

// New creates a new in-memory webhook bus provider
func New(bufferSize int) *Provider {
	return &Provider{
		ch: make(chan webhookTypes.WebhookEvent, bufferSize),
	}
}

// Publish sends a webhook event to the internal channel (non-blocking)
func (p *Provider) Publish(ctx context.Context, event webhookTypes.WebhookEvent) error {
	select {
	case p.ch <- event:
		return nil
	default:
		log.Printf("Warning: webhook bus channel full, dropping event: %s", event.EventType)
		return nil
	}
}

// Subscribe registers a handler and starts a goroutine to process events
func (p *Provider) Subscribe(ctx context.Context, handler func(ctx context.Context, event webhookTypes.WebhookEvent)) error {
	p.mu.Lock()
	p.handlers = append(p.handlers, handler)
	p.mu.Unlock()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case event, ok := <-p.ch:
				if !ok {
					return
				}
				p.mu.RLock()
				handlers := p.handlers
				p.mu.RUnlock()
				for _, h := range handlers {
					h(ctx, event)
				}
			}
		}
	}()

	return nil
}
