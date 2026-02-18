package config

// EmailConfigType holds email provider configuration
type EmailConfigType struct {
	Provider string
	From     string
	SMTP     SMTPConfigType
}

// SMTPConfigType holds SMTP server configuration
type SMTPConfigType struct {
	Host string
	Port int
	User string
	Pass string
}

// Email is the global email configuration instance
var Email *EmailConfigType

// loadEmail loads email configuration from environment variables
func loadEmail() *EmailConfigType {
	return &EmailConfigType{
		Provider: GetEnv("EMAIL_PROVIDER", "noop"),
		From:     GetEnv("EMAIL_FROM", "noreply@activelog.app"),
		SMTP: SMTPConfigType{
			Host: GetEnv("SMTP_HOST", "localhost"),
			Port: GetEnvInt("SMTP_PORT", 587),
			User: GetEnv("SMTP_USER", ""),
			Pass: GetEnv("SMTP_PASS", ""),
		},
	}
}
