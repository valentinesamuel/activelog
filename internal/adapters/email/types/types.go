package types

import "context"

// SendEmailInput holds all the fields needed to send an email.
type SendEmailInput struct {
	To       string
	From     string
	Subject  string
	HTMLBody string
	TextBody string
}

// EmailProvider is the interface all email backends must implement.
type EmailProvider interface {
	Send(ctx context.Context, input SendEmailInput) error
}
