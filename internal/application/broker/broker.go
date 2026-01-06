package broker

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"
)

// UseCase is the interface that all use cases must implement
// This is the Go equivalent of kuja_user_ms's Usecase abstract class
type UseCase interface {
	Execute(ctx context.Context, tx *sql.Tx, input map[string]interface{}) (map[string]interface{}, error)
}

// UseCaseFunc is a functional wrapper for simple use cases
type UseCaseFunc func(ctx context.Context, tx *sql.Tx, input map[string]interface{}) (map[string]interface{}, error)

func (f UseCaseFunc) Execute(ctx context.Context, tx *sql.Tx, input map[string]interface{}) (map[string]interface{}, error) {
	return f(ctx, tx, input)
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

// RunUseCases executes multiple use cases in a single transaction
// This is the equivalent of kuja_user_ms's broker.runUsecases()
func (b *Broker) RunUseCases(
	ctx context.Context,
	useCases []UseCase,
	initialInput map[string]interface{},
	opts ...Option,
) (map[string]interface{}, error) {
	// Validate inputs
	if len(useCases) == 0 {
		return nil, fmt.Errorf("at least one use case must be provided")
	}

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

	// Execute with timeout
	resultChan := make(chan executionResult, 1)
	go func() {
		result, err := b.executeTransaction(timeoutCtx, useCases, initialInput, config)
		resultChan <- executionResult{result: result, err: err}
	}()

	select {
	case <-timeoutCtx.Done():
		return nil, fmt.Errorf("transaction timed out after %v", config.timeout)
	case result := <-resultChan:
		return result.result, result.err
	}
}

type executionResult struct {
	result map[string]interface{}
	err    error
}

// executeTransaction runs all use cases within a single transaction
func (b *Broker) executeTransaction(
	ctx context.Context,
	useCases []UseCase,
	initialInput map[string]interface{},
	config *executionConfig,
) (map[string]interface{}, error) {
	// Begin transaction with isolation level
	tx, err := b.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: config.isolationLevel,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Ensure transaction is closed
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p) // Re-throw panic after rollback
		}
	}()

	// Run all use cases
	results, err := b.runAllUseCases(ctx, tx, useCases, initialInput)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			b.logger.Printf("failed to rollback transaction: %v", rbErr)
		}
		return nil, err
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return results, nil
}

// runAllUseCases executes use cases sequentially
func (b *Broker) runAllUseCases(
	ctx context.Context,
	tx *sql.Tx,
	useCases []UseCase,
	initialInput map[string]interface{},
) (map[string]interface{}, error) {
	b.logger.Printf("Starting transaction with %d use cases", len(useCases))

	// Initialize results with initial input
	accumulatedResults := make(map[string]interface{})
	for k, v := range initialInput {
		accumulatedResults[k] = v
	}

	// Execute each use case sequentially
	for i, useCase := range useCases {
		useCaseName := fmt.Sprintf("UseCase_%d", i+1)
		b.logger.Printf("Executing use case %d/%d: %s", i+1, len(useCases), useCaseName)

		startTime := time.Now()

		// Execute use case with accumulated results
		result, err := useCase.Execute(ctx, tx, accumulatedResults)
		if err != nil {
			duration := time.Since(startTime)
			b.logger.Printf("Use case %s failed after %v: %v", useCaseName, duration, err)
			return nil, fmt.Errorf("use case %s failed: %w", useCaseName, err)
		}

		duration := time.Since(startTime)
		b.logger.Printf("Use case %s completed in %v", useCaseName, duration)

		// Merge results (output of this use case becomes input to next)
		for k, v := range result {
			accumulatedResults[k] = v
		}
	}

	// Remove initial input from final results (like kuja_user_ms does)
	finalResults := make(map[string]interface{})
	for k, v := range accumulatedResults {
		if _, exists := initialInput[k]; !exists {
			finalResults[k] = v
		}
	}

	b.logger.Printf("Transaction completed successfully")
	return finalResults, nil
}

// WithLogger sets custom logger
func (b *Broker) WithLogger(logger *log.Logger) *Broker {
	b.logger = logger
	return b
}
