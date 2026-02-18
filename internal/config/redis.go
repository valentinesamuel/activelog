package config

type CacheDBNumbers struct {
	ActivityData int // REDIS_DB_ACTIVITY_DATA (default 1)
	Stats        int // REDIS_DB_STATS         (default 2)
	RateLimits   int // REDIS_DB_RATE_LIMITS   (default 3)
}

type CacheConfigType struct {
	Provider string
	Redis    RedisConfigType
	DBs      CacheDBNumbers
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
		DBs: CacheDBNumbers{
			ActivityData: GetEnvInt("REDIS_DB_ACTIVITY_DATA", 1),
			Stats:        GetEnvInt("REDIS_DB_STATS", 2),
			RateLimits:   GetEnvInt("REDIS_DB_RATE_LIMITS", 3),
		},
	}
}
