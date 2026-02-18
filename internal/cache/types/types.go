package types

import (
	"context"
	"time"
)

type CacheProvider interface {
	Get(key string) (string, error)
	Set(key string, value string, ttl time.Duration) error
	Del(key string) error
	Increment(key string) (int64, error)
	Expire(key string, ttl time.Duration) (bool, error)
}

// --- Multi-DB Adapter types ---

type CacheDBName string

const (
	CacheDBActivityData CacheDBName = "ACTIVITY_DATA"
	CacheDBStats        CacheDBName = "STATS"
	CacheDBRateLimits   CacheDBName = "RATE_LIMITS"
)

type CachePartition string

const (
	CachePartitionActivities        CachePartition = "activities"
	CachePartitionStats             CachePartition = "stats"
	CachePartitionRateLimitConfig   CachePartition = "ratelimit:config"
	CachePartitionRateLimitCounters CachePartition = "ratelimit:counters"
)

// CacheOptions is required on every CacheAdapter call.
// DB and PartitionKey are always mandatory — no defaults.
type CacheOptions struct {
	DB           CacheDBName
	PartitionKey CachePartition
}

// CacheAdapter is the high-level interface for general caching.
type CacheAdapter interface {
	Get(ctx context.Context, key string, opts CacheOptions) (string, error)
	Set(ctx context.Context, key string, value string, ttl time.Duration, opts CacheOptions) error
	Del(ctx context.Context, key string, opts CacheOptions) error
}

// RateLimitCacheProvider is the dedicated interface for rate limiter counter operations.
type RateLimitCacheProvider interface {
	Increment(ctx context.Context, key string, opts CacheOptions) (int64, error)
	Expire(ctx context.Context, key string, ttl time.Duration, opts CacheOptions) (bool, error)
	SetNX(ctx context.Context, key string, value string, ttl time.Duration, opts CacheOptions) (bool, error)
}
