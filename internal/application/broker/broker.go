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

// TransactionalUseCase is an optional marker interface for use cases that require transactions
// Use cases that implement this interface can declare whether they need to run in a transaction
// Similar to kuja_user_ms's transactional decorator pattern
//
// Default behavior: Use cases WITHOUT this method are assumed to be NON-TRANSACTIONAL
// This optimizes read-heavy operations which don't need transaction overhead
//
// Example usage:
//   type CreateActivityUseCase struct { ... }
//   func (uc *CreateActivityUseCase) RequiresTransaction() bool { return true }  // Write operation
//
//   type GetActivityUseCase struct { ... }
//   // No RequiresTransaction() method = non-transactional (read operation)
type TransactionalUseCase interface {
	RequiresTransaction() bool
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

// executeTransaction runs all use cases with conditional transaction management
// Supports transaction boundary breaking: UC1(tx) → UC2(tx) → UC3(non-tx) → UC4(tx)
func (b *Broker) executeTransaction(
	ctx context.Context,
	useCases []UseCase,
	initialInput map[string]interface{},
	config *executionConfig,
) (map[string]interface{}, error) {
	// Run all use cases with conditional transaction management
	// Transaction creation/commit is now handled inside runAllUseCases
	return b.runAllUseCases(ctx, useCases, initialInput, config)
}

// runAllUseCases executes use cases sequentially with conditional transaction management
// Supports transaction boundary breaking:
//   - UC1(tx) → UC2(tx) → UC3(non-tx) → UC4(tx)
//   - Transaction commits before UC3, new transaction starts for UC4
func (b *Broker) runAllUseCases(
	ctx context.Context,
	useCases []UseCase,
	initialInput map[string]interface{},
	config *executionConfig,
) (map[string]interface{}, error) {
	b.logger.Printf("Starting use case chain with %d use cases", len(useCases))

	// Initialize results with initial input
	accumulatedResults := make(map[string]interface{})
	for k, v := range initialInput {
		accumulatedResults[k] = v
	}

	var activeTx *sql.Tx = nil

	// Cleanup: rollback any uncommitted transaction on panic
	defer func() {
		if activeTx != nil {
			if rbErr := activeTx.Rollback(); rbErr != nil {
				b.logger.Printf("failed to rollback transaction during panic recovery: %v", rbErr)
			}
		}
		if p := recover(); p != nil {
			panic(p) // Re-throw panic after cleanup
		}
	}()

	// Execute each use case sequentially
	for i, useCase := range useCases {
		useCaseName := fmt.Sprintf("UseCase_%d_%T", i+1, useCase)

		// Check if use case requires transaction via marker interface
		needsTx := false
		if txUC, ok := useCase.(TransactionalUseCase); ok {
			needsTx = txUC.RequiresTransaction()
		}
		// Default: non-transactional (if marker interface not implemented)

		// Transaction boundary management
		if needsTx && activeTx == nil {
			// Start new transaction
			b.logger.Printf("%s requires transaction, starting new transaction", useCaseName)
			newTx, err := b.db.BeginTx(ctx, &sql.TxOptions{
				Isolation: config.isolationLevel,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to begin transaction: %w", err)
			}
			activeTx = newTx
		} else if !needsTx && activeTx != nil {
			// TRANSACTION BOUNDARY BREAK: Commit active transaction
			b.logger.Printf("%s does not require transaction, committing active transaction", useCaseName)
			if err := activeTx.Commit(); err != nil {
				return nil, fmt.Errorf("failed to commit transaction at boundary: %w", err)
			}
			activeTx = nil
		}

		// Execute use case
		b.logger.Printf("Executing use case %d/%d: %s (tx: %v)", i+1, len(useCases), useCaseName, activeTx != nil)
		startTime := time.Now()

		result, err := useCase.Execute(ctx, activeTx, accumulatedResults)
		if err != nil {
			duration := time.Since(startTime)
			b.logger.Printf("%s failed after %v: %v", useCaseName, duration, err)

			// Rollback active transaction if exists
			if activeTx != nil {
				if rbErr := activeTx.Rollback(); rbErr != nil {
					b.logger.Printf("failed to rollback transaction: %v", rbErr)
				}
				activeTx = nil
			}

			return nil, fmt.Errorf("%s failed: %w", useCaseName, err)
		}

		duration := time.Since(startTime)
		b.logger.Printf("%s completed in %v", useCaseName, duration)

		// Merge results (output of this use case becomes input to next)
		for k, v := range result {
			accumulatedResults[k] = v
		}
	}

	// Commit final transaction if exists
	if activeTx != nil {
		b.logger.Printf("Committing final transaction")
		if err := activeTx.Commit(); err != nil {
			return nil, fmt.Errorf("failed to commit final transaction: %w", err)
		}
		activeTx = nil
	}

	// Remove initial input from final results (like kuja_user_ms does)
	finalResults := make(map[string]interface{})
	for k, v := range accumulatedResults {
		if _, exists := initialInput[k]; !exists {
			finalResults[k] = v
		}
	}

	b.logger.Printf("Use case chain completed successfully")
	return finalResults, nil
}

// WithLogger sets custom logger
func (b *Broker) WithLogger(logger *log.Logger) *Broker {
	b.logger = logger
	return b
}
