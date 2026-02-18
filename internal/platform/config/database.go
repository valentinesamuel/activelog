package config

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	URL            string
	EnableLogging  bool
	MaxConnections int
	MaxIdleConns   int
}

// Database is the global database configuration instance
var Database *DatabaseConfig

// loadDatabase loads database configuration from environment variables
func loadDatabase() *DatabaseConfig {
	return &DatabaseConfig{
		URL:            GetEnv("DATABASE_URL", "postgres://activelog_user:activelog@localhost:5444/activelog?sslmode=disable"),
		EnableLogging:  GetEnvBool("ENABLE_QUERY_LOGGING", true),
		MaxConnections: GetEnvInt("DATABASE_MAX_CONNECTIONS", 25),
		MaxIdleConns:   GetEnvInt("DATABASE_MAX_IDLE_CONNECTIONS", 5),
	}
}
