package config

type RedisConfig struct {
	Address  string
	DB       int
	PASSWORD string
}

var Redis *RedisConfig

func loadRedis() *RedisConfig {
	return &RedisConfig{
		Address:  GetEnv("REDIS_ADDRESS", "localhost"),
		DB:       GetEnvInt("REDIS_DB", 0),
		PASSWORD: GetEnv("REDIS_PASSWORD", ""),
	}
}
