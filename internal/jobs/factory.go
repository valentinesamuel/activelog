package jobs

import (
	"context"
	"fmt"

	"github.com/valentinesamuel/activelog/internal/queue/types"
)

// HandlerFunc is the signature every job handler must implement.
type HandlerFunc func(ctx context.Context, payload types.JobPayload) error

// HandlerFactory routes incoming jobs to the correct handler based on EventType.
type HandlerFactory struct {
	handlers map[types.EventType]HandlerFunc
}

// NewHandlerFactory creates an empty HandlerFactory.
func NewHandlerFactory() *HandlerFactory {
	return &HandlerFactory{
		handlers: make(map[types.EventType]HandlerFunc),
	}
}

// Register associates an EventType with a handler function.
func (f *HandlerFactory) Register(event types.EventType, handler HandlerFunc) {
	f.handlers[event] = handler
}

// Dispatch finds the handler for payload.Event and calls it.
func (f *HandlerFactory) Dispatch(ctx context.Context, payload types.JobPayload) error {
	handler, ok := f.handlers[payload.Event]
	if !ok {
		return fmt.Errorf("factory: no handler registered for event %q", payload.Event)
	}
	return handler(ctx, payload)
}
