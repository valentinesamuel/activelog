package redis

import (
	"context"
	"encoding/json"
	"log"

	goredis "github.com/redis/go-redis/v9"
	"github.com/valentinesamuel/activelog/internal/config"
	webhookTypes "github.com/valentinesamuel/activelog/internal/webhook/types"
)

const channelName = "webhook:events"

// Provider is a Redis pub/sub webhook bus provider
type Provider struct {
	client *goredis.Client
}

// New creates a new Redis-backed webhook bus provider
func New() (*Provider, error) {
	client := goredis.NewClient(&goredis.Options{
		Addr:     config.Cache.Redis.Address,
		Password: config.Cache.Redis.Password,
		DB:       config.Cache.Redis.DB,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}

	return &Provider{client: client}, nil
}

// Publish serializes a WebhookEvent and publishes it to the Redis channel
func (p *Provider) Publish(ctx context.Context, event webhookTypes.WebhookEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return p.client.Publish(ctx, channelName, data).Err()
}

// Subscribe starts a goroutine that reads from the Redis channel and calls the handler
func (p *Provider) Subscribe(ctx context.Context, handler func(ctx context.Context, event webhookTypes.WebhookEvent)) error {
	pubsub := p.client.Subscribe(ctx, channelName)

	go func() {
		defer pubsub.Close()
		ch := pubsub.Channel()
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}
				var event webhookTypes.WebhookEvent
				if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
					log.Printf("Error unmarshaling webhook event: %v", err)
					continue
				}
				handler(ctx, event)
			}
		}
	}()

	return nil
}
