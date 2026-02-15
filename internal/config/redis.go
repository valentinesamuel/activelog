package config

type CacheConfigType struct {
	Provider string
	Redis    RedisConfigType
	// Memcached MemcachedConfig
}

type RedisConfigType struct {
	Address  string
	DB       int
	Password string
}

var Cache *CacheConfigType

func loadCache() *CacheConfigType {
	return &CacheConfigType{
		Provider: GetEnv("CACHE_PROVIDER", "redis"),
		Redis: RedisConfigType{
			Address:  GetEnv("REDIS_ADDRESS", "localhost"),
			DB:       GetEnvInt("REDIS_DB", 0),
			Password: GetEnv("REDIS_PASSWORD", ""),
		},
	}
}
