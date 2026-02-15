package asynq

import (
	"context"

	"github.com/hibiken/asynq"
	"github.com/valentinesamuel/activelog/internal/config"
	"github.com/valentinesamuel/activelog/internal/queue/types"
)

type Provider struct {
	client *asynq.Client
}

func New() (*Provider, error) {

	address := config.GetEnv("REDIS_ADDRESS", "localhost")

	client := asynq.NewClient(asynq.RedisClientOpt{Addr: address})
	defer client.Close()

	return &Provider{
		client: client,
	}, nil
}

func (p *Provider) Enqueue(ctx context.Context, input types.EmailTask) error {
	return nil
}
