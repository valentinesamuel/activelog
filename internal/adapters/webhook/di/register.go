package di

import (
	"log"

	"github.com/valentinesamuel/activelog/internal/platform/config"
	"github.com/valentinesamuel/activelog/internal/platform/container"
	"github.com/valentinesamuel/activelog/internal/repository"
	repoDI "github.com/valentinesamuel/activelog/internal/repository/di"
	"github.com/valentinesamuel/activelog/internal/adapters/webhook"
	webhookMemory "github.com/valentinesamuel/activelog/internal/adapters/webhook/memory"
	webhookNATS "github.com/valentinesamuel/activelog/internal/adapters/webhook/nats"
	webhookRedis "github.com/valentinesamuel/activelog/internal/adapters/webhook/redis"
	webhookTypes "github.com/valentinesamuel/activelog/internal/adapters/webhook/types"
)

// RegisterWebhookBus registers the webhook bus provider in the DI container
func RegisterWebhookBus(c *container.Container) {
	c.Register(WebhookBusKey, func(c *container.Container) (interface{}, error) {
		return createProvider(), nil
	})
}

// RegisterWebhookDelivery registers the webhook delivery handler in the DI container
func RegisterWebhookDelivery(c *container.Container) {
	c.Register(WebhookDeliveryKey, func(c *container.Container) (interface{}, error) {
		repo := c.MustResolve(repoDI.WebhookRepoKey).(*repository.WebhookRepository)
		return webhook.NewDelivery(repo), nil
	})
}

// RegisterRetryWorker registers the webhook retry worker in the DI container
func RegisterRetryWorker(c *container.Container) {
	c.Register(RetryWorkerKey, func(c *container.Container) (interface{}, error) {
		repo := c.MustResolve(repoDI.WebhookRepoKey).(*repository.WebhookRepository)
		delivery := c.MustResolve(WebhookDeliveryKey).(*webhook.Delivery)
		return webhook.NewRetryWorker(repo, delivery), nil
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
		log.Printf("Webhook bus provider initialized: redis (streams)")
		return provider

	case "nats":
		provider, err := webhookNATS.New(config.Webhook.NATSUrl)
		if err != nil {
			log.Printf("Warning: Failed to initialize NATS webhook bus: %v. Falling back to memory.", err)
			return webhookMemory.New(100)
		}
		log.Printf("Webhook bus provider initialized: nats (jetstream)")
		return provider

	default:
		log.Printf("Webhook bus provider initialized: memory (buffer=100)")
		return webhookMemory.New(100)
	}
}
