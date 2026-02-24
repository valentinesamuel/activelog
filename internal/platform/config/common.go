package config

// CommonConfig holds common application configuration
type CommonConfig struct {
	Port               int
	AppName            string
	Environment        string
	IsDevelopment      bool
	EnableQueryLogging bool
	Auth               AuthConfig
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	JWTSecret string
}

// Common is the global common configuration instance
var Common *CommonConfig

// loadCommon loads common configuration from environment variables
func loadCommon() *CommonConfig {
	env := GetEnv("NODE_ENV", "development")

	return &CommonConfig{
		Port:               GetEnvInt("PORT", 8080),
		AppName:            GetEnv("APP_NAME", "ActiveLog"),
		Environment:        env,
		IsDevelopment:      env == "development",
		EnableQueryLogging: GetEnvBool("ENABLE_QUERY_LOGGING", true),
		Auth: AuthConfig{
			JWTSecret: GetEnv("JWT_SECRET", ""),
		},
	}
}
