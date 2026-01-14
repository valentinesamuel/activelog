# Broker Pattern - Use Case Orchestrator (Like kuja_user_ms)

## What is the Broker Pattern?

The Broker in `kuja_user_ms` is **NOT a message queue** (like RabbitMQ/Kafka). Instead, it's a **use case orchestrator** that:

1. ✅ Runs multiple use cases **sequentially**
2. ✅ Within a **single database transaction**
3. ✅ Each use case receives **results from previous use cases**
4. ✅ If **any use case fails**, entire transaction **rolls back**
5. ✅ Supports **timeout** and **isolation levels**

This is perfect for complex business operations that need **atomicity**.

---

## kuja_user_ms Broker Flow

```typescript
// TypeScript example from kuja_user_ms
await broker.runUsecases(
  [
    createUserUseCase,      // Step 1: Create user
    sendEmailUseCase,       // Step 2: Send welcome email (uses user from step 1)
    createSessionUseCase,   // Step 3: Create session (uses user from step 1)
  ],
  { email: 'user@example.com', password: 'hash' }, // Initial input
);

// Flow:
// 1. Start transaction
// 2. createUserUseCase.execute(tem, { email, password }) → returns { userId: 123 }
// 3. sendEmailUseCase.execute(tem, { email, password, userId: 123 }) → returns { emailSent: true }
// 4. createSessionUseCase.execute(tem, { email, password, userId: 123, emailSent: true }) → returns { sessionId: 'abc' }
// 5. Commit transaction
// 6. Return final result: { userId: 123, emailSent: true, sessionId: 'abc' }
```

**Key Point:** All use cases run in ONE transaction. If step 2 or 3 fails, the user creation from step 1 is rolled back!

---

## Go Implementation

### Directory Structure

```
internal/
├── domain/
│   └── ...
│
├── application/
│   ├── broker/                    # NEW: Broker package
│   │   ├── broker.go              # Orchestrator
│   │   ├── usecase.go             # Use case interface
│   │   └── context.go             # Execution context
│   │
│   └── activity/
│       └── usecases/
│           ├── create_activity.go # Individual use cases
│           └── send_notification.go
│
└── infrastructure/
    └── ...
```

---

## Core Broker Implementation

### Use Case Interface (`internal/application/broker/usecase.go`)

```go
package broker

import (
	"context"
	"database/sql"
)

// UseCase is the interface that all use cases must implement
// Similar to kuja_user_ms's Usecase abstract class
type UseCase interface {
	// Execute runs the use case with the transactional database
	// and accumulated results from previous use cases
	Execute(ctx context.Context, tx *sql.Tx, input map[string]interface{}) (map[string]interface{}, error)
}

// UseCaseFunc is a functional wrapper for simple use cases
type UseCaseFunc func(ctx context.Context, tx *sql.Tx, input map[string]interface{}) (map[string]interface{}, error)

func (f UseCaseFunc) Execute(ctx context.Context, tx *sql.Tx, input map[string]interface{}) (map[string]interface{}, error) {
	return f(ctx, tx, input)
}
```

### Execution Context (`internal/application/broker/context.go`)

```go
package broker

import (
	"context"
	"database/sql"
	"time"
)

// ExecutionContext holds the context for use case execution
type ExecutionContext struct {
	Context        context.Context
	Transaction    *sql.Tx
	Results        map[string]interface{} // Accumulated results
	StartTime      time.Time
	IsolationLevel sql.IsolationLevel
}

// NewExecutionContext creates a new execution context
func NewExecutionContext(ctx context.Context, tx *sql.Tx, initialInput map[string]interface{}) *ExecutionContext {
	return &ExecutionContext{
		Context:     ctx,
		Transaction: tx,
		Results:     make(map[string]interface{}),
		StartTime:   time.Now(),
	}
}

// AddResults merges new results into accumulated results
func (ec *ExecutionContext) AddResults(newResults map[string]interface{}) {
	for key, value := range newResults {
		ec.Results[key] = value
	}
}

// GetAllResults returns all accumulated results
func (ec *ExecutionContext) GetAllResults() map[string]interface{} {
	return ec.Results
}
```

### Broker Orchestrator (`internal/application/broker/broker.go`)

```go
package broker

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"
)

// Broker orchestrates multiple use cases in a single transaction
// Inspired by kuja_user_ms broker pattern
type Broker struct {
	db                   *sql.DB
	defaultTimeout       time.Duration
	defaultIsolationLevel sql.IsolationLevel
	logger               *log.Logger
}

// NewBroker creates a new broker instance
func NewBroker(db *sql.DB) *Broker {
	return &Broker{
		db:                   db,
		defaultTimeout:       60 * time.Second,
		defaultIsolationLevel: sql.LevelReadCommitted,
		logger:               log.Default(),
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
```

---

## Use Case Examples

### Simple Use Case (`internal/application/activity/usecases/create_activity.go`)

```go
package usecases

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/valentinesamuel/activelog/internal/application/broker"
	"github.com/valentinesamuel/activelog/internal/domain/activity"
)

// CreateActivityUseCase creates a new activity
// Implements broker.UseCase interface
type CreateActivityUseCase struct {
	// Dependencies can be injected here if needed
}

func NewCreateActivityUseCase() *CreateActivityUseCase {
	return &CreateActivityUseCase{}
}

// Execute implements broker.UseCase
func (uc *CreateActivityUseCase) Execute(
	ctx context.Context,
	tx *sql.Tx,
	input map[string]interface{},
) (map[string]interface{}, error) {
	// Extract input parameters
	userID, ok := input["user_id"].(int)
	if !ok {
		return nil, fmt.Errorf("user_id is required")
	}

	title, ok := input["title"].(string)
	if !ok {
		return nil, fmt.Errorf("title is required")
	}

	// Create activity using the transactional connection
	query := `
		INSERT INTO activities (user_id, activity_type, title, activity_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	var activityID int64
	err := tx.QueryRowContext(ctx, query,
		userID,
		input["type"],
		title,
		time.Now(),
		time.Now(),
		time.Now(),
	).Scan(&activityID)

	if err != nil {
		return nil, fmt.Errorf("failed to create activity: %w", err)
	}

	// Return results that will be passed to next use case
	return map[string]interface{}{
		"activity_id": activityID,
		"title":       title,
	}, nil
}
```

### Chained Use Case (`internal/application/activity/usecases/send_notification.go`)

```go
package usecases

import (
	"context"
	"database/sql"
	"fmt"
	"log"
)

// SendNotificationUseCase sends notification about created activity
// This use case depends on output from CreateActivityUseCase
type SendNotificationUseCase struct {
	// Could inject notification service here
}

func NewSendNotificationUseCase() *SendNotificationUseCase {
	return &SendNotificationUseCase{}
}

// Execute implements broker.UseCase
func (uc *SendNotificationUseCase) Execute(
	ctx context.Context,
	tx *sql.Tx,
	input map[string]interface{},
) (map[string]interface{}, error) {
	// This use case receives output from previous use case
	activityID, ok := input["activity_id"].(int64)
	if !ok {
		return nil, fmt.Errorf("activity_id not found in input")
	}

	userID, ok := input["user_id"].(int)
	if !ok {
		return nil, fmt.Errorf("user_id not found in input")
	}

	title := input["title"].(string)

	// Send notification (simplified)
	log.Printf("Sending notification: Activity '%s' (ID: %d) created for user %d", title, activityID, userID)

	// Could insert notification record in database using tx
	query := `
		INSERT INTO notifications (user_id, message, created_at)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	var notificationID int64
	err := tx.QueryRowContext(ctx, query,
		userID,
		fmt.Sprintf("Activity '%s' created successfully", title),
		time.Now(),
	).Scan(&notificationID)

	if err != nil {
		return nil, fmt.Errorf("failed to create notification: %w", err)
	}

	// Return results
	return map[string]interface{}{
		"notification_id":   notificationID,
		"notification_sent": true,
	}, nil
}
```

### Update Stats Use Case

```go
package usecases

import (
	"context"
	"database/sql"
	"fmt"
)

// UpdateStatsUseCase updates user statistics after activity creation
type UpdateStatsUseCase struct{}

func NewUpdateStatsUseCase() *UpdateStatsUseCase {
	return &UpdateStatsUseCase{}
}

func (uc *UpdateStatsUseCase) Execute(
	ctx context.Context,
	tx *sql.Tx,
	input map[string]interface{},
) (map[string]interface{}, error) {
	userID, ok := input["user_id"].(int)
	if !ok {
		return nil, fmt.Errorf("user_id not found")
	}

	// Update user's activity count
	query := `
		UPDATE user_stats
		SET total_activities = total_activities + 1,
		    updated_at = NOW()
		WHERE user_id = $1
	`

	_, err := tx.ExecContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to update stats: %w", err)
	}

	return map[string]interface{}{
		"stats_updated": true,
	}, nil
}
```

---

## Using the Broker

### HTTP Handler Example

```go
package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/valentinesamuel/activelog/internal/application/broker"
	"github.com/valentinesamuel/activelog/internal/application/activity/usecases"
)

type ActivityHandler struct {
	broker *broker.Broker

	// Use cases
	createActivity   *usecases.CreateActivityUseCase
	sendNotification *usecases.SendNotificationUseCase
	updateStats      *usecases.UpdateStatsUseCase
}

func NewActivityHandler(
	broker *broker.Broker,
	createActivity *usecases.CreateActivityUseCase,
	sendNotification *usecases.SendNotificationUseCase,
	updateStats *usecases.UpdateStatsUseCase,
) *ActivityHandler {
	return &ActivityHandler{
		broker:           broker,
		createActivity:   createActivity,
		sendNotification: sendNotification,
		updateStats:      updateStats,
	}
}

func (h *ActivityHandler) CreateActivity(w http.ResponseWriter, r *http.Request) {
	// Parse request
	var req struct {
		UserID int    `json:"user_id"`
		Type   string `json:"type"`
		Title  string `json:"title"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Orchestrate multiple use cases in ONE transaction
	// This is the key pattern from kuja_user_ms!
	results, err := h.broker.RunUseCases(
		r.Context(),
		[]broker.UseCase{
			h.createActivity,   // Step 1: Create activity
			h.sendNotification, // Step 2: Send notification (uses activity_id from step 1)
			h.updateStats,      // Step 3: Update stats (uses user_id)
		},
		map[string]interface{}{
			"user_id": req.UserID,
			"type":    req.Type,
			"title":   req.Title,
		},
		broker.WithTimeout(30 * time.Second),
	)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Results contain output from all use cases
	// {
	//   "activity_id": 123,
	//   "title": "Morning Run",
	//   "notification_id": 456,
	//   "notification_sent": true,
	//   "stats_updated": true
	// }

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}
```

---

## Main Setup

```go
package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/valentinesamuel/activelog/internal/application/broker"
	"github.com/valentinesamuel/activelog/internal/application/activity/usecases"
	"github.com/valentinesamuel/activelog/internal/interfaces/http/handlers"
)

func main() {
	// Setup database
	db, err := sql.Open("postgres", "connection_string")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create broker
	brk := broker.NewBroker(db)

	// Create use cases
	createActivity := usecases.NewCreateActivityUseCase()
	sendNotification := usecases.NewSendNotificationUseCase()
	updateStats := usecases.NewUpdateStatsUseCase()

	// Create handler with broker
	activityHandler := handlers.NewActivityHandler(
		brk,
		createActivity,
		sendNotification,
		updateStats,
	)

	// Setup routes
	http.HandleFunc("/activities", activityHandler.CreateActivity)

	// Start server
	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
```

---

## Key Benefits

### 1. Atomicity
```go
// All these operations succeed together OR fail together
results, err := broker.RunUseCases(ctx, []UseCase{
    createUser,      // If this succeeds but...
    sendWelcomeEmail, // ...this fails, then...
    createSession,   // ...user creation is ROLLED BACK!
})
```

### 2. Result Chaining
```go
// Use case 1 output → Use case 2 input
CreateActivity → { activity_id: 123 }
                        ↓
SendNotification receives { user_id: 1, activity_id: 123 }
                        ↓
UpdateStats receives { user_id: 1, activity_id: 123, notification_sent: true }
```

### 3. Clean Separation
```go
// Each use case is independent and testable
type CreateActivityUseCase struct {}
type SendNotificationUseCase struct {}
type UpdateStatsUseCase struct {}

// But broker orchestrates them atomically
```

### 4. Transaction Management
```go
// No manual transaction handling in use cases!
// Broker handles: BEGIN, COMMIT, ROLLBACK

// Use cases just receive *sql.Tx and use it
```

---

## Testing the Broker

### Unit Test (`tests/unit/broker/broker_test.go`)

```go
package broker_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/valentinesamuel/activelog/internal/application/broker"
	"github.com/valentinesamuel/activelog/tests/testhelpers/database"
)

func TestBroker_SuccessfulExecution(t *testing.T) {
	// Setup
	db, cleanup := database.SetupClean(t)
	defer cleanup()

	brk := broker.NewBroker(db)

	// Create mock use cases
	useCase1 := broker.UseCaseFunc(func(ctx context.Context, tx *sql.Tx, input map[string]interface{}) (map[string]interface{}, error) {
		return map[string]interface{}{
			"step1": "completed",
		}, nil
	})

	useCase2 := broker.UseCaseFunc(func(ctx context.Context, tx *sql.Tx, input map[string]interface{}) (map[string]interface{}, error) {
		// Should receive step1 result
		if input["step1"] != "completed" {
			t.Error("Expected step1 result")
		}
		return map[string]interface{}{
			"step2": "completed",
		}, nil
	})

	// Execute
	results, err := brk.RunUseCases(
		context.Background(),
		[]broker.UseCase{useCase1, useCase2},
		map[string]interface{}{"initial": "data"},
	)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if results["step1"] != "completed" {
		t.Error("Expected step1 result")
	}

	if results["step2"] != "completed" {
		t.Error("Expected step2 result")
	}

	// Initial input should be removed
	if _, exists := results["initial"]; exists {
		t.Error("Initial input should be removed from results")
	}
}

func TestBroker_Rollback(t *testing.T) {
	db, cleanup := database.SetupClean(t)
	defer cleanup()

	brk := broker.NewBroker(db)

	// Use case 1 succeeds
	useCase1 := broker.UseCaseFunc(func(ctx context.Context, tx *sql.Tx, input map[string]interface{}) (map[string]interface{}, error) {
		// Insert a record
		_, err := tx.ExecContext(ctx, "INSERT INTO activities (user_id, title, created_at) VALUES ($1, $2, NOW())", 1, "Test")
		return map[string]interface{}{"inserted": true}, err
	})

	// Use case 2 fails
	useCase2 := broker.UseCaseFunc(func(ctx context.Context, tx *sql.Tx, input map[string]interface{}) (map[string]interface{}, error) {
		return nil, fmt.Errorf("intentional failure")
	})

	// Execute
	_, err := brk.RunUseCases(
		context.Background(),
		[]broker.UseCase{useCase1, useCase2},
		map[string]interface{}{},
	)

	// Should fail
	if err == nil {
		t.Fatal("Expected error")
	}

	// Verify rollback - activity should NOT exist
	var count int
	db.QueryRow("SELECT COUNT(*) FROM activities").Scan(&count)
	if count != 0 {
		t.Errorf("Expected 0 activities (rollback), got %d", count)
	}
}
```

---

## Comparison: kuja_user_ms vs Go

### TypeScript (kuja_user_ms)
```typescript
await broker.runUsecases(
  [createUserUseCase, sendEmailUseCase],
  { email: 'test@example.com' }
);

// Use case:
class CreateUserUseCase extends Usecase {
  async execute(tem: EntityManager, args: any) {
    const user = await tem.save(User, args);
    return { userId: user.id };
  }
}
```

### Go (activelog)
```go
broker.RunUseCases(
  ctx,
  []UseCase{createUserUseCase, sendEmailUseCase},
  map[string]interface{}{"email": "test@example.com"},
)

// Use case:
type CreateUserUseCase struct{}

func (uc *CreateUserUseCase) Execute(
  ctx context.Context,
  tx *sql.Tx,
  input map[string]interface{},
) (map[string]interface{}, error) {
  // Insert user using tx
  var userID int
  err := tx.QueryRowContext(ctx, "INSERT INTO users...").Scan(&userID)
  return map[string]interface{}{"user_id": userID}, err
}
```

**Key Similarities:**
- ✅ Sequential execution
- ✅ Single transaction
- ✅ Result chaining
- ✅ Automatic rollback
- ✅ Timeout support

---

## When to Use Broker Pattern

### ✅ Use Broker When:
- Multiple operations must be atomic (all succeed or all fail)
- Operations depend on each other's results
- You want to avoid distributed transactions
- Business workflow spans multiple domains

### ❌ Don't Use Broker When:
- Operations can be independent (use message queue instead)
- Operations are slow and can be async
- You need to scale operations independently
- Cross-service transactions (use Saga pattern)

---

## Summary

The Broker pattern from `kuja_user_ms` is **NOT** a message queue - it's a **use case orchestrator** that ensures atomicity across multiple business operations.

**Key Takeaways:**
1. One transaction for multiple use cases
2. Sequential execution with result chaining
3. Automatic rollback on any failure
4. Perfect for complex business workflows
5. Clean separation of concerns

This pattern is now ready for use in your Go project! 🚀
