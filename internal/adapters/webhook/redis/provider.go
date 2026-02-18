package redis

import (
	"context"
	"encoding/json"
	"log"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"github.com/valentinesamuel/activelog/internal/platform/config"
	webhookTypes "github.com/valentinesamuel/activelog/internal/adapters/webhook/types"
)

const (
	streamName    = "webhook:events"
	groupName     = "webhook-delivery"
	consumerName  = "activelog-worker-1"
	blockDuration = 5 * time.Second
)

// Provider is a Redis Streams webhook bus provider
type Provider struct {
	client *goredis.Client
}

// New creates a new Redis Streams webhook bus provider
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

// Publish serializes a WebhookEvent and adds it to the Redis stream
func (p *Provider) Publish(ctx context.Context, event webhookTypes.WebhookEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return p.client.XAdd(ctx, &goredis.XAddArgs{
		Stream: streamName,
		MaxLen: config.Webhook.StreamMaxLen,
		Approx: true,
		Values: map[string]interface{}{"event": string(data)},
	}).Err()
}

// Subscribe creates the consumer group and starts reading from the stream
func (p *Provider) Subscribe(ctx context.Context, handler func(ctx context.Context, event webhookTypes.WebhookEvent)) error {
	if err := p.ensureConsumerGroup(ctx); err != nil {
		return err
	}
	go p.readLoop(ctx, handler)
	return nil
}

func (p *Provider) ensureConsumerGroup(ctx context.Context) error {
	err := p.client.XGroupCreateMkStream(ctx, streamName, groupName, "$").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		return err
	}
	return nil
}

func (p *Provider) readLoop(ctx context.Context, handler func(ctx context.Context, event webhookTypes.WebhookEvent)) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		streams, err := p.client.XReadGroup(ctx, &goredis.XReadGroupArgs{
			Group:    groupName,
			Consumer: consumerName,
			Streams:  []string{streamName, ">"},
			Count:    10,
			Block:    blockDuration,
		}).Result()

		if err != nil {
			if ctx.Err() != nil {
				return
			}
			// Timeout is expected (no new messages), just continue
			if err.Error() == "redis: nil" {
				continue
			}
			log.Printf("Redis stream read error: %v", err)
			time.Sleep(1 * time.Second)
			continue
		}

		for _, stream := range streams {
			for _, msg := range stream.Messages {
				p.processMessage(ctx, msg, handler)
			}
		}
	}
}

func (p *Provider) processMessage(ctx context.Context, msg goredis.XMessage, handler func(ctx context.Context, event webhookTypes.WebhookEvent)) {
	data, ok := msg.Values["event"]
	if !ok {
		log.Printf("Malformed stream message (no 'event' field), dropping: %s", msg.ID)
		p.client.XAck(ctx, streamName, groupName, msg.ID)
		return
	}

	var event webhookTypes.WebhookEvent
	if err := json.Unmarshal([]byte(data.(string)), &event); err != nil {
		log.Printf("Failed to unmarshal webhook event, dropping: %v", err)
		p.client.XAck(ctx, streamName, groupName, msg.ID)
		return
	}

	handler(ctx, event)
	p.client.XAck(ctx, streamName, groupName, msg.ID)
}
