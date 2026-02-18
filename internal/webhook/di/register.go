package di

import (
	"log"

	"github.com/valentinesamuel/activelog/internal/config"
	"github.com/valentinesamuel/activelog/internal/container"
	webhookMemory "github.com/valentinesamuel/activelog/internal/webhook/memory"
	webhookRedis "github.com/valentinesamuel/activelog/internal/webhook/redis"
	webhookTypes "github.com/valentinesamuel/activelog/internal/webhook/types"
)

// RegisterWebhookBus registers the webhook bus provider in the DI container
func RegisterWebhookBus(c *container.Container) {
	c.Register(WebhookBusKey, func(c *container.Container) (interface{}, error) {
		return createProvider(), nil
	})
}

func createProvider() webhookTypes.WebhookBusProvider {
	switch config.Webhook.Provider {
	case "redis":
		provider, err := webhookRedis.New()
		if err != nil {
			log.Printf("Warning: Failed to initialize Redis webhook bus: %v. Falling back to memory.", err)
			return webhookMemory.New(100)
		}
		log.Printf("Webhook bus provider initialized: redis")
		return provider

	default:
		log.Printf("Webhook bus provider initialized: memory (buffer=100)")
		return webhookMemory.New(100)
	}
}
