package di

import (
	"log"

	"github.com/valentinesamuel/activelog/internal/cache/redis"
	"github.com/valentinesamuel/activelog/internal/cache/types"
	"github.com/valentinesamuel/activelog/internal/config"
	"github.com/valentinesamuel/activelog/internal/container"
)

func RegisterCache(c *container.Container) {
	c.Register(CacheProviderKey, func(c *container.Container) (interface{}, error) {
		return createProvider(), nil
	})
}

// createProvider creates the appropriate storage provider based on configuration
func createProvider() types.CacheProvider {
	switch config.Cache.Provider {
	case "redis":
		provider, err := redis.Connect()
		if err != nil {
			log.Printf("Warning: Failed to initialize redis provider: %v. Cache operations will fail.", err)
			return nil
		}
		log.Printf("🗑️ Cache provider initialized: Redis (DB: %v)", config.Cache.Redis.DB)
		return provider

	case "memcached":
		log.Printf("Warning: Memcached cache provider not yet implemented")
		return nil

	default:
		log.Printf("Warning: Unknown cache provider '%s'. Cache operations will fail.", config.Cache.Provider)
		return nil
	}
}
