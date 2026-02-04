package types

import "time"

type CacheProvider interface {
	Get(key string) (string, error)
	Set(key string, value string, ttl time.Duration) error
	Del(key string) error
	Increment(key string) (int64, error)
	Expire(key string, ttl time.Duration) (bool, error)
}
