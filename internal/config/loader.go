package config

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
)

// MustLoad loads all configuration from environment variables
// It validates the configuration and panics if validation fails
// Call this at application startup
func MustLoad() {
	if err := LoadWithValidation(); err != nil {
		log.Fatalf("Configuration error: %v", err)
	}
}

// LoadWithValidation loads and validates all configuration
// Returns an error if validation fails
func LoadWithValidation() error {
	// Load .env file if it exists (ignore error if not found)
	_ = godotenv.Load()

	// Validate schema
	if err := ValidateSchema(); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	// Load all config modules
	Common = loadCommon()
	Database = loadDatabase()
	Storage = loadStorage()

	return nil
}

// Load loads configuration without validation
// Prefer using MustLoad() or LoadWithValidation() for new code
func Load() {
	// Load .env file if it exists
	_ = godotenv.Load()

	// Load all config modules without validation
	Common = loadCommon()
	Database = loadDatabase()
	Storage = loadStorage()
}
