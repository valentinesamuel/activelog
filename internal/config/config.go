package config

import "os"

type Config struct {
	DatabaseUrl      string
	ServerPort       string
	JWTSecret        string
	EnableQueryLogging bool
}

func Load() *Config {
	return &Config{
		DatabaseUrl:      GetEnv("DATABASE_URL", "postgres://activelog_user:activelog@localhost:5444/activelog?sslmode=disable"),
		ServerPort:       GetEnv("PORT", "8080"),
		JWTSecret:        GetEnv("JWT_SECRET", "secret"),
		EnableQueryLogging: GetEnv("ENABLE_QUERY_LOGGING", "true") == "true",
	}
}

func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
