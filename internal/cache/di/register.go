package di

import (
	"fmt"
	"log"

	redisadapter "github.com/valentinesamuel/activelog/internal/cache/adapter/redis"
	"github.com/valentinesamuel/activelog/internal/config"
	"github.com/valentinesamuel/activelog/internal/container"
)

// RegisterCacheAdapter registers the multi-DB cache adapter.
func RegisterCacheAdapter(c *container.Container) {
	c.Register(CacheAdapterKey, func(c *container.Container) (interface{}, error) {
		switch config.Cache.Provider {
		case "redis":
			adapter := redisadapter.New()
			log.Printf("Cache adapter initialized: Redis multi-DB")
			return adapter, nil
		default:
			return nil, fmt.Errorf("unsupported cache provider for adapter: %s", config.Cache.Provider)
		}
	})
}
