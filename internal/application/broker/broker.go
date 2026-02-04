package broker

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"
)

// TransactionalUseCase is a marker interface for use cases that require transactions
// Use cases that implement this interface can declare whether they need to run in a transaction
//
// Default behavior: Use cases WITHOUT this method are assumed to be NON-TRANSACTIONAL
// This optimizes read-heavy operations which don't need transaction overhead
type TransactionalUseCase interface {
	RequiresTransaction() bool
}

// Broker orchestrates multiple use cases in a single transaction
// Inspired by kuja_user_ms broker pattern
type Broker struct {
	db                    *sql.DB
	defaultTimeout        time.Duration
	defaultIsolationLevel sql.IsolationLevel
	logger                *log.Logger
}

// NewBroker creates a new broker instance
func NewBroker(db *sql.DB) *Broker {
	return &Broker{
		db:                    db,
		defaultTimeout:        60 * time.Second,
		defaultIsolationLevel: sql.LevelReadCommitted,
		logger:                log.Default(),
	}
}

// Option configures broker execution
type Option func(*executionConfig)

type executionConfig struct {
	timeout        time.Duration
	isolationLevel sql.IsolationLevel
}

// WithTimeout sets execution timeout
func WithTimeout(d time.Duration) Option {
	return func(c *executionConfig) {
		c.timeout = d
	}
}

// WithIsolationLevel sets transaction isolation level
func WithIsolationLevel(level sql.IsolationLevel) Option {
	return func(c *executionConfig) {
		c.isolationLevel = level
	}
}

// WithLogger sets custom logger
func (b *Broker) WithLogger(logger *log.Logger) *Broker {
	b.logger = logger
	return b
}

// RunUseCase executes a single typed use case with full type safety.
// This is the preferred method for executing use cases as it provides
// compile-time type checking for both input and output.
//
// Example usage:
//
//	result, err := broker.RunUseCase(b, ctx, listActivitiesUC,
//	    usecases.ListActivitiesInput{UserID: 1, QueryOptions: opts})
//	// result is ListActivitiesOutput - no type assertion needed
//	activities := result.Result
//
// For use cases that require transactions, the function checks if the use case
// implements TransactionalTypedUseCase and handles transaction management automatically.
func RunUseCase[I, O any](
	b *Broker,
	ctx context.Context,
	uc TypedUseCase[I, O],
	input I,
	opts ...Option,
) (O, error) {
	var zero O

	// Apply options
	config := &executionConfig{
		timeout:        b.defaultTimeout,
		isolationLevel: b.defaultIsolationLevel,
	}
	for _, opt := range opts {
		opt(config)
	}

	// Create timeout context
	timeoutCtx, cancel := context.WithTimeout(ctx, config.timeout)
	defer cancel()

	// Check if use case requires transaction
	needsTx := false
	if txUC, ok := any(uc).(TransactionalUseCase); ok {
		needsTx = txUC.RequiresTransaction()
	}

	// Execute with timeout
	type result struct {
		output O
		err    error
	}
	resultChan := make(chan result, 1)

	go func() {
		var tx *sql.Tx
		var err error

		// Start transaction if needed
		if needsTx {
			tx, err = b.db.BeginTx(timeoutCtx, &sql.TxOptions{
				Isolation: config.isolationLevel,
			})
			if err != nil {
				resultChan <- result{zero, fmt.Errorf("failed to begin transaction: %w", err)}
				return
			}
		}

		// Execute use case
		output, err := uc.Execute(timeoutCtx, tx, input)
		if err != nil {
			if tx != nil {
				if rbErr := tx.Rollback(); rbErr != nil {
					b.logger.Printf("failed to rollback transaction: %v", rbErr)
				}
			}
			resultChan <- result{zero, err}
			return
		}

		// Commit transaction if needed
		if tx != nil {
			if err := tx.Commit(); err != nil {
				resultChan <- result{zero, fmt.Errorf("failed to commit transaction: %w", err)}
				return
			}
		}

		resultChan <- result{output, nil}
	}()

	select {
	case <-timeoutCtx.Done():
		return zero, fmt.Errorf("use case timed out after %v", config.timeout)
	case res := <-resultChan:
		return res.output, res.err
	}
}
