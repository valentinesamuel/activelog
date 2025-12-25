package config

import "os"

type Config struct {
	DatabaseUrl string
	ServerPort  string
}

func Load() *Config {
	return &Config{
		DatabaseUrl: getEnv("DATABASE_URL", "postgres://activelog_user:your_secure_password@localhost/activelog?sslmode=disable"),
		ServerPort:  getEnv("PORT", "8080"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
