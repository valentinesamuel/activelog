package smtp

import (
	"context"
	"fmt"

	"gopkg.in/gomail.v2"

	"github.com/valentinesamuel/activelog/internal/config"
	"github.com/valentinesamuel/activelog/internal/email/types"
)

// Provider sends emails via SMTP using gomail.
type Provider struct {
	dialer *gomail.Dialer
	from   string
}

// New creates an SMTP Provider from the global email config.
func New() (*Provider, error) {
	cfg := config.Email
	if cfg.SMTP.Host == "" {
		return nil, fmt.Errorf("smtp: SMTP_HOST is required")
	}

	d := gomail.NewDialer(cfg.SMTP.Host, cfg.SMTP.Port, cfg.SMTP.User, cfg.SMTP.Pass)
	return &Provider{dialer: d, from: cfg.From}, nil
}

// Send builds and dispatches a single email message.
func (p *Provider) Send(_ context.Context, input types.SendEmailInput) error {
	from := input.From
	if from == "" {
		from = p.from
	}

	m := gomail.NewMessage()
	m.SetHeader("From", from)
	m.SetHeader("To", input.To)
	m.SetHeader("Subject", input.Subject)

	if input.HTMLBody != "" {
		m.SetBody("text/html", input.HTMLBody)
	}
	if input.TextBody != "" {
		if input.HTMLBody != "" {
			m.AddAlternative("text/plain", input.TextBody)
		} else {
			m.SetBody("text/plain", input.TextBody)
		}
	}

	if err := p.dialer.DialAndSend(m); err != nil {
		return fmt.Errorf("smtp: send: %w", err)
	}
	return nil
}
