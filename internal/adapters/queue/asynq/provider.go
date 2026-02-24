package asynq

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/valentinesamuel/activelog/internal/platform/config"
	"github.com/valentinesamuel/activelog/internal/adapters/queue/types"
)

// Provider wraps an asynq.Client to implement types.QueueProvider
type Provider struct {
	client *asynq.Client
}

// New creates an asynq Provider. The client is NOT closed here.
func New() (*Provider, error) {
	address := config.GetEnv("REDIS_ADDRESS", "localhost:6379")
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: address})
	return &Provider{client: client}, nil
}

// Enqueue marshals the payload and submits a task to the given queue.
func (p *Provider) Enqueue(ctx context.Context, queue types.QueueName, payload types.JobPayload) (string, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("asynq: marshal payload: %w", err)
	}

	task := asynq.NewTask(string(payload.Event), data)
	info, err := p.client.EnqueueContext(ctx, task,
		asynq.Queue(string(queue)),
		asynq.MaxRetry(3),
	)
	if err != nil {
		return "", fmt.Errorf("asynq: enqueue task: %w", err)
	}

	return info.ID, nil
}

// NewWorkerServer creates an asynq server for processing jobs.
func NewWorkerServer(redisAddr string, concurrency int) *asynq.Server {
	return asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
			Concurrency: concurrency,
		},
	)
}
