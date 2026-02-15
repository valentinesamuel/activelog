package di

import (
	"log"

	"github.com/valentinesamuel/activelog/internal/config"
	"github.com/valentinesamuel/activelog/internal/container"
	"github.com/valentinesamuel/activelog/internal/queue/asynq"
	"github.com/valentinesamuel/activelog/internal/queue/types"
)

func RegisterQueue(c *container.Container) {
	c.Register(QueueProviderKey, func(c *container.Container) (interface{}, error) {
		return createProvider(), nil
	})
}

func createProvider() types.QueueProviderKey {
	switch config.Queue.Provider {
	case "asynq":
		provider, err := asynq.New()
		if err != nil {
			log.Printf("Warning: Failed to initialize asynq provider: %v. Queue operations will fail.", err)
			return nil
		}
		log.Printf("🤖 Queue provider initialized: asynq")
		return provider

	default:
		log.Printf("Warning: Unknown queue provider '%s'. Queue operations will fail.", config.Queue.Provider)
		return nil
	}
}
