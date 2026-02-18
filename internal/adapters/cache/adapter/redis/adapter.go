package redis

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/valentinesamuel/activelog/internal/adapters/cache/types"
	"github.com/valentinesamuel/activelog/internal/platform/config"
)

// Adapter implements both CacheAdapter and RateLimitCacheProvider.
// It lazily creates one *redis.Client per DB number, protected by a mutex.
type Adapter struct {
	addr     string
	password string
	dbMap    map[types.CacheDBName]int // name → DB number
	clients  map[int]*redis.Client     // DB number → client (lazy)
	mu       sync.Mutex
}

// New creates an Adapter using the current cache configuration.
func New() *Adapter {
	return &Adapter{
		addr:     config.Cache.Redis.Address,
		password: config.Cache.Redis.Password,
		dbMap: map[types.CacheDBName]int{
			types.CacheDBActivityData: config.Cache.DBs.ActivityData,
			types.CacheDBStats:        config.Cache.DBs.Stats,
			types.CacheDBRateLimits:   config.Cache.DBs.RateLimits,
		},
		clients: make(map[int]*redis.Client),
	}
}

// client lazily initializes and returns the redis.Client for the given DB name.
func (a *Adapter) client(db types.CacheDBName) (*redis.Client, error) {
	dbNum, ok := a.dbMap[db]
	if !ok {
		return nil, fmt.Errorf("unknown cache DB: %s", db)
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	if c, exists := a.clients[dbNum]; exists {
		return c, nil
	}

	c := redis.NewClient(&redis.Options{
		Addr:     a.addr,
		Password: a.password,
		DB:       dbNum,
	})

	if _, err := c.Ping(context.Background()).Result(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis DB %d: %w", dbNum, err)
	}

	a.clients[dbNum] = c
	return c, nil
}

// buildKey constructs the namespaced key: "<partition>:<key>"
func buildKey(opts types.CacheOptions, key string) string {
	return fmt.Sprintf("%s:%s", opts.PartitionKey, key)
}

// Get retrieves a value from the cache.
func (a *Adapter) Get(ctx context.Context, key string, opts types.CacheOptions) (string, error) {
	c, err := a.client(opts.DB)
	if err != nil {
		return "", err
	}
	return c.Get(ctx, buildKey(opts, key)).Result()
}

// Set stores a value in the cache with the given TTL.
func (a *Adapter) Set(ctx context.Context, key string, value string, ttl time.Duration, opts types.CacheOptions) error {
	c, err := a.client(opts.DB)
	if err != nil {
		return err
	}
	return c.Set(ctx, buildKey(opts, key), value, ttl).Err()
}

// Del removes a value from the cache.
func (a *Adapter) Del(ctx context.Context, key string, opts types.CacheOptions) error {
	c, err := a.client(opts.DB)
	if err != nil {
		return err
	}
	return c.Del(ctx, buildKey(opts, key)).Err()
}

// Increment atomically increments the counter for the given key.
func (a *Adapter) Increment(ctx context.Context, key string, opts types.CacheOptions) (int64, error) {
	c, err := a.client(opts.DB)
	if err != nil {
		return 0, err
	}
	return c.Incr(ctx, buildKey(opts, key)).Result()
}

// Expire sets the TTL for the given key.
func (a *Adapter) Expire(ctx context.Context, key string, ttl time.Duration, opts types.CacheOptions) (bool, error) {
	c, err := a.client(opts.DB)
	if err != nil {
		return false, err
	}
	return c.Expire(ctx, buildKey(opts, key), ttl).Result()
}

// SetNX sets the value only if the key does not already exist.
func (a *Adapter) SetNX(ctx context.Context, key string, value string, ttl time.Duration, opts types.CacheOptions) (bool, error) {
	c, err := a.client(opts.DB)
	if err != nil {
		return false, err
	}
	return c.SetNX(ctx, buildKey(opts, key), value, ttl).Result()
}
