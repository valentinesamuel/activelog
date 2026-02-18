package memory

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/valentinesamuel/activelog/internal/adapters/queue/types"
)

// Provider is an in-process queue backed by buffered channels.
// Suitable for tests and local development (no Redis required).
type Provider struct {
	mu      sync.Mutex
	jobs    map[types.QueueName]chan types.JobPayload
	bufSize int
}

// New creates a Provider with a per-queue buffer of bufferSize.
func New(bufferSize int) *Provider {
	return &Provider{
		jobs:    make(map[types.QueueName]chan types.JobPayload),
		bufSize: bufferSize,
	}
}

// Enqueue sends the payload to the queue's channel non-blocking.
func (p *Provider) Enqueue(_ context.Context, queue types.QueueName, payload types.JobPayload) (string, error) {
	ch := p.channel(queue)
	select {
	case ch <- payload:
		return fmt.Sprintf("mem-%s-%d", queue, len(ch)), nil
	default:
		return "", fmt.Errorf("memory: queue %q is full (buffer=%d)", queue, p.bufSize)
	}
}

// StartWorking drains the queue in a background goroutine until ctx is cancelled.
func (p *Provider) StartWorking(ctx context.Context, queue types.QueueName, handler func(context.Context, types.JobPayload) error) {
	ch := p.channel(queue)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case job := <-ch:
				if err := handler(ctx, job); err != nil {
					log.Printf("memory: handler error for event %q: %v", job.Event, err)
				}
			}
		}
	}()
}

// channel returns (or creates) the buffered channel for the given queue.
func (p *Provider) channel(queue types.QueueName) chan types.JobPayload {
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, ok := p.jobs[queue]; !ok {
		p.jobs[queue] = make(chan types.JobPayload, p.bufSize)
	}
	return p.jobs[queue]
}
