package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// EnvVar represents an environment variable with validation rules
type EnvVar struct {
	Key          string
	Required     bool
	DefaultValue string
	Type         string // "string", "int", "bool"
	ValidValues  []string
}

// Schema defines all environment variables and their validation rules
var Schema = []EnvVar{
	// Common
	{Key: "PORT", Required: false, DefaultValue: "8080", Type: "int"},
	{Key: "APP_NAME", Required: false, DefaultValue: "ActiveLog", Type: "string"},
	{Key: "NODE_ENV", Required: false, DefaultValue: "development", Type: "string", ValidValues: []string{"development", "staging", "production"}},
	{Key: "JWT_SECRET", Required: true, Type: "string"},
	{Key: "ENABLE_QUERY_LOGGING", Required: false, DefaultValue: "true", Type: "bool"},

	// Database
	{Key: "DATABASE_URL", Required: true, Type: "string"},

	// Storage
	{Key: "STORAGE_PROVIDER", Required: false, DefaultValue: "s3", Type: "string", ValidValues: []string{"s3", "local", "supabase", "azure"}},

	// Email
	{Key: "EMAIL_PROVIDER", Required: false, DefaultValue: "noop", Type: "string", ValidValues: []string{"smtp", "noop"}},
	{Key: "EMAIL_FROM", Required: false, DefaultValue: "noreply@activelog.app", Type: "string"},
	{Key: "SMTP_HOST", Required: false, DefaultValue: "localhost", Type: "string"},
	{Key: "SMTP_PORT", Required: false, DefaultValue: "587", Type: "int"},
	{Key: "SMTP_USER", Required: false, DefaultValue: "", Type: "string"},
	{Key: "SMTP_PASS", Required: false, DefaultValue: "", Type: "string"},

	// Webhook
	{Key: "WEBHOOK_PROVIDER", Required: false, DefaultValue: "memory", Type: "string", ValidValues: []string{"memory", "redis", "nats"}},
	{Key: "WEBHOOK_STREAM_MAX_LEN", Required: false, DefaultValue: "10000", Type: "int"},
	{Key: "WEBHOOK_RETRY_POLL_SECONDS", Required: false, DefaultValue: "30", Type: "int"},
	{Key: "NATS_URL", Required: false, DefaultValue: "nats://localhost:4222", Type: "string"},

	// AWS S3
	{Key: "AWS_S3_BUCKET", Required: false, DefaultValue: "", Type: "string"},
	{Key: "AWS_REGION", Required: false, DefaultValue: "us-east-1", Type: "string"},
	{Key: "AWS_ACCESS_KEY_ID", Required: false, DefaultValue: "", Type: "string"},
	{Key: "AWS_SECRET_ACCESS_KEY", Required: false, DefaultValue: "", Type: "string"},
	{Key: "AWS_S3_ENDPOINT", Required: false, DefaultValue: "", Type: "string"},
	{Key: "AWS_S3_PATH_STYLE", Required: false, DefaultValue: "false", Type: "bool"},
}

// ValidationError represents a configuration validation error
type ValidationError struct {
	Key     string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Key, e.Message)
}

// ValidationErrors holds multiple validation errors
type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	var msgs []string
	for _, err := range e {
		msgs = append(msgs, err.Error())
	}
	return fmt.Sprintf("configuration validation failed:\n  - %s", strings.Join(msgs, "\n  - "))
}

// ValidateSchema validates all environment variables against the schema
func ValidateSchema() error {
	var errors ValidationErrors

	for _, env := range Schema {
		value := os.Getenv(env.Key)

		// Check required
		if env.Required && value == "" {
			errors = append(errors, ValidationError{
				Key:     env.Key,
				Message: "required but not set",
			})
			continue
		}

		// Use default if not set
		if value == "" {
			value = env.DefaultValue
		}

		// Skip further validation if empty and not required
		if value == "" {
			continue
		}

		// Validate type
		switch env.Type {
		case "int":
			if _, err := strconv.Atoi(value); err != nil {
				errors = append(errors, ValidationError{
					Key:     env.Key,
					Message: fmt.Sprintf("must be an integer, got '%s'", value),
				})
			}
		case "bool":
			if value != "true" && value != "false" {
				errors = append(errors, ValidationError{
					Key:     env.Key,
					Message: fmt.Sprintf("must be 'true' or 'false', got '%s'", value),
				})
			}
		}

		// Validate allowed values
		if len(env.ValidValues) > 0 {
			valid := false
			for _, v := range env.ValidValues {
				if value == v {
					valid = true
					break
				}
			}
			if !valid {
				errors = append(errors, ValidationError{
					Key:     env.Key,
					Message: fmt.Sprintf("must be one of [%s], got '%s'", strings.Join(env.ValidValues, ", "), value),
				})
			}
		}
	}

	if len(errors) > 0 {
		return errors
	}
	return nil
}

// GetEnv retrieves an environment variable with a default fallback
func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetEnvInt retrieves an environment variable as int with a default fallback
func GetEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return defaultValue
}

// GetEnvBool retrieves an environment variable as bool with a default fallback
func GetEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return value == "true"
	}
	return defaultValue
}
