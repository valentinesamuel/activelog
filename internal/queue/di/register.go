package di

import (
	"log"

	"github.com/valentinesamuel/activelog/internal/config"
	"github.com/valentinesamuel/activelog/internal/container"
	"github.com/valentinesamuel/activelog/internal/queue/asynq"
	"github.com/valentinesamuel/activelog/internal/queue/memory"
	"github.com/valentinesamuel/activelog/internal/queue/types"
)

// RegisterQueue registers the queue provider in the DI container.
func RegisterQueue(c *container.Container) {
	c.Register(QueueProviderKey, func(c *container.Container) (interface{}, error) {
		return createProvider(), nil
	})
}

// createProvider selects a queue backend based on QUEUE_PROVIDER env var.
func createProvider() types.QueueProvider {
	switch config.Queue.Provider {
	case "asynq":
		provider, err := asynq.New()
		if err != nil {
			log.Printf("Warning: Failed to initialize asynq provider: %v. Queue operations will fail.", err)
			return nil
		}
		log.Printf("Queue provider initialized: asynq")
		return provider

	default:
		log.Printf("Queue provider initialized: memory (buffer=100)")
		return memory.New(100)
	}
}
