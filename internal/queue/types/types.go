package types

import (
	"context"
)

type EmailTask struct {
	To      string
	Subject string
	Body    string
}

type QueueProvider interface {
	Enqueue(ctx context.Context, input EmailTask) error
}
