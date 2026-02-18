package di

import (
	"log"

	"github.com/valentinesamuel/activelog/internal/config"
	"github.com/valentinesamuel/activelog/internal/container"
	"github.com/valentinesamuel/activelog/internal/email/noop"
	"github.com/valentinesamuel/activelog/internal/email/smtp"
	"github.com/valentinesamuel/activelog/internal/email/types"
)

// RegisterEmail registers the email provider in the DI container.
func RegisterEmail(c *container.Container) {
	c.Register(EmailProviderKey, func(c *container.Container) (interface{}, error) {
		return createProvider(), nil
	})
}

// createProvider selects an email backend based on EMAIL_PROVIDER env var.
func createProvider() types.EmailProvider {
	switch config.Email.Provider {
	case "smtp":
		provider, err := smtp.New()
		if err != nil {
			log.Printf("Warning: Failed to initialize SMTP provider: %v. Email operations will fail.", err)
			return nil
		}
		log.Printf("Email provider initialized: smtp (host: %s)", config.Email.SMTP.Host)
		return provider

	default:
		log.Printf("Email provider initialized: noop")
		return noop.New()
	}
}
