package noop

import (
	"context"
	"log"

	"github.com/valentinesamuel/activelog/internal/adapters/email/types"
)

// Provider is a no-op email backend that logs instead of sending.
// Suitable for development and testing.
type Provider struct{}

// New creates a noop Provider.
func New() *Provider {
	return &Provider{}
}

// Send logs the email details without actually sending anything.
func (p *Provider) Send(_ context.Context, input types.SendEmailInput) error {
	log.Printf("[email:noop] to=%s subject=%q (not sent)", input.To, input.Subject)
	return nil
}
