# Broker Orchestrator - Usage Examples

## Quick Start

The broker orchestrator allows you to execute multiple use cases in a single atomic transaction, with automatic result chaining between use cases.

## Basic Usage

### 1. Define Your Use Cases

Each use case must implement the `UseCase` interface:

```go
type UseCase interface {
    Execute(ctx context.Context, tx *sql.Tx, input map[string]interface{}) (map[string]interface{}, error)
}
```

Example use case:

```go
type CreateActivityUseCase struct {
    activityRepo repository.ActivityRepository
}

func (uc *CreateActivityUseCase) Execute(
    ctx context.Context,
    tx *sql.Tx,
    input map[string]interface{},
) (map[string]interface{}, error) {
    // Extract input
    userID := input["user_id"].(int)
    title := input["title"].(string)

    // Create activity
    activity := &models.Activity{
        UserID: userID,
        Title:  title,
    }

    err := uc.activityRepo.Create(ctx, tx, activity)
    if err != nil {
        return nil, err
    }

    // Return output for next use case
    return map[string]interface{}{
        "activity_id": activity.ID,
    }, nil
}
```

### 2. Create Broker Instance

```go
import "github.com/valentinesamuel/activelog/internal/application/broker"

db := // your *sql.DB connection
brokerInstance := broker.NewBroker(db)
```

### 3. Execute Use Cases

```go
// Define use cases
createActivity := &CreateActivityUseCase{activityRepo: activityRepo}
updateStats := &UpdateUserStatsUseCase{statsRepo: statsRepo}
sendNotification := &SendNotificationUseCase{notificationService: notificationSvc}

// Execute in single transaction
result, err := brokerInstance.RunUseCases(
    ctx,
    []broker.UseCase{
        createActivity,
        updateStats,
        sendNotification,
    },
    map[string]interface{}{
        "user_id": 123,
        "title":   "Morning Run",
    },
)

if err != nil {
    // All use cases rolled back
    log.Printf("transaction failed: %v", err)
    return err
}

// Success! All use cases committed
log.Printf("activity created: %v", result["activity_id"])
```

## Advanced Examples

### Example 1: Activity Creation with Tags and Stats

```go
type CreateActivityWithTagsHandler struct {
    broker             *broker.Broker
    createActivityUC   broker.UseCase
    attachTagsUC       broker.UseCase
    updateStatsUC      broker.UseCase
}

func (h *CreateActivityWithTagsHandler) Handle(
    ctx context.Context,
    userID int,
    title string,
    tags []string,
) error {
    result, err := h.broker.RunUseCases(
        ctx,
        []broker.UseCase{
            h.createActivityUC,
            h.attachTagsUC,
            h.updateStatsUC,
        },
        map[string]interface{}{
            "user_id": userID,
            "title":   title,
            "tags":    tags,
        },
    )

    if err != nil {
        return fmt.Errorf("failed to create activity: %w", err)
    }

    log.Printf("Created activity %d with %d tags",
        result["activity_id"],
        result["tags_attached"])

    return nil
}
```

### Example 2: With Custom Timeout and Isolation Level

```go
result, err := brokerInstance.RunUseCases(
    ctx,
    useCases,
    initialInput,
    broker.WithTimeout(5*time.Second),
    broker.WithIsolationLevel(sql.LevelSerializable),
)
```

### Example 3: Using UseCaseFunc for Simple Operations

For simple use cases, you can use `UseCaseFunc` instead of creating a struct:

```go
validateUser := broker.UseCaseFunc(func(
    ctx context.Context,
    tx *sql.Tx,
    input map[string]interface{},
) (map[string]interface{}, error) {
    userID := input["user_id"].(int)

    var exists bool
    err := tx.QueryRowContext(ctx,
        "SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)",
        userID,
    ).Scan(&exists)

    if err != nil {
        return nil, err
    }

    if !exists {
        return nil, fmt.Errorf("user %d not found", userID)
    }

    return map[string]interface{}{"user_validated": true}, nil
})

// Use it in the broker
result, err := brokerInstance.RunUseCases(
    ctx,
    []broker.UseCase{
        validateUser,
        createActivity,
        updateStats,
    },
    map[string]interface{}{"user_id": 123},
)
```

## Result Chaining

The broker automatically chains results between use cases:

```go
// UseCase 1: Create activity
func (uc *CreateActivityUseCase) Execute(ctx, tx, input) (map[string]interface{}, error) {
    // Uses: user_id, title from initial input
    // Returns: activity_id
    return map[string]interface{}{
        "activity_id": 456,
    }, nil
}

// UseCase 2: Attach tags
func (uc *AttachTagsUseCase) Execute(ctx, tx, input) (map[string]interface{}, error) {
    // Receives: user_id, title (from initial), activity_id (from UseCase 1)
    activityID := input["activity_id"].(int64)
    tags := input["tags"].([]string)

    // Attach tags...

    return map[string]interface{}{
        "tags_attached": len(tags),
    }, nil
}

// UseCase 3: Update stats
func (uc *UpdateStatsUseCase) Execute(ctx, tx, input) (map[string]interface{}, error) {
    // Receives: user_id, title, activity_id, tags_attached
    userID := input["user_id"].(int)

    // Update stats...

    return map[string]interface{}{
        "stats_updated": true,
    }, nil
}
```

**Final result excludes initial input:**
```go
// result = {
//     "activity_id": 456,
//     "tags_attached": 2,
//     "stats_updated": true,
// }
// Note: user_id and title are NOT in final result
```

## Error Handling

### Automatic Rollback

If any use case fails, the entire transaction is rolled back:

```go
result, err := brokerInstance.RunUseCases(ctx, useCases, input)
if err != nil {
    // Transaction automatically rolled back
    // No partial changes committed

    // Check which use case failed
    if strings.Contains(err.Error(), "UseCase_2") {
        log.Printf("second use case failed")
    }

    return err
}
```

### Use Case Error Handling

```go
func (uc *CreateActivityUseCase) Execute(ctx, tx, input) (map[string]interface{}, error) {
    activity := &models.Activity{...}

    err := uc.repo.Create(ctx, tx, activity)
    if err != nil {
        // Return error - transaction will be rolled back
        return nil, fmt.Errorf("failed to create activity: %w", err)
    }

    return map[string]interface{}{"activity_id": activity.ID}, nil
}
```

## Integration with HTTP Handlers

### Handler Example

```go
type ActivityHandler struct {
    broker           *broker.Broker
    createActivityUC broker.UseCase
    attachTagsUC     broker.UseCase
    updateStatsUC    broker.UseCase
}

func (h *ActivityHandler) CreateActivity(w http.ResponseWriter, r *http.Request) {
    var req struct {
        UserID int      `json:"user_id"`
        Title  string   `json:"title"`
        Tags   []string `json:"tags"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Execute use cases in single transaction
    result, err := h.broker.RunUseCases(
        r.Context(),
        []broker.UseCase{
            h.createActivityUC,
            h.attachTagsUC,
            h.updateStatsUC,
        },
        map[string]interface{}{
            "user_id": req.UserID,
            "title":   req.Title,
            "tags":    req.Tags,
        },
        broker.WithTimeout(5*time.Second), // HTTP request timeout
    )

    if err != nil {
        log.Printf("failed to create activity: %v", err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    // Return success response
    json.NewEncoder(w).Encode(map[string]interface{}{
        "success":     true,
        "activity_id": result["activity_id"],
        "message":     "Activity created successfully",
    })
}
```

## Testing

### Testing Individual Use Cases

```go
func TestCreateActivityUseCase(t *testing.T) {
    mockRepo := &MockActivityRepository{}
    useCase := &CreateActivityUseCase{activityRepo: mockRepo}

    input := map[string]interface{}{
        "user_id": 123,
        "title":   "Test Activity",
    }

    result, err := useCase.Execute(context.Background(), nil, input)

    assert.NoError(t, err)
    assert.NotNil(t, result["activity_id"])
}
```

### Testing Broker Orchestration

```go
func TestActivityCreationFlow(t *testing.T) {
    // Setup test database
    db := setupTestDB(t)
    defer db.Close()

    broker := broker.NewBroker(db)

    // Create real or mock use cases
    createUC := &CreateActivityUseCase{...}
    tagsUC := &AttachTagsUseCase{...}
    statsUC := &UpdateStatsUseCase{...}

    result, err := broker.RunUseCases(
        context.Background(),
        []broker.UseCase{createUC, tagsUC, statsUC},
        map[string]interface{}{
            "user_id": 1,
            "title":   "Test",
            "tags":    []string{"running"},
        },
    )

    assert.NoError(t, err)
    assert.NotNil(t, result["activity_id"])
    assert.Equal(t, 1, result["tags_attached"])
}
```

## Performance Characteristics

Based on benchmarks (Apple M3 Pro):

- **Single Use Case**: ~160 μs/op, 2.4 KB/op, 30 allocations
- **5 Use Cases**: ~160 μs/op, 2.6 KB/op, 41 allocations

The overhead is minimal and scales linearly with the number of use cases.

## Best Practices

### 1. Keep Use Cases Focused

```go
// Good: Single responsibility
type CreateActivityUseCase struct { ... }
type AttachTagsUseCase struct { ... }
type UpdateStatsUseCase struct { ... }

// Bad: God use case
type CreateActivityWithEverythingUseCase struct { ... }
```

### 2. Return Useful Data

```go
// Good: Return data needed by next use cases
return map[string]interface{}{
    "activity_id": activity.ID,
    "created_at":  activity.CreatedAt,
}, nil

// Bad: Return everything including input
return map[string]interface{}{
    "user_id":     input["user_id"],  // Already in input!
    "activity_id": activity.ID,
}, nil
```

### 3. Use Custom Logger

```go
// Production: Use your logger
logger := log.New(os.Stdout, "[BROKER] ", log.LstdFlags)
broker := broker.NewBroker(db).WithLogger(logger)

// Testing: Use silent logger
logger := log.New(io.Discard, "", 0)
broker := broker.NewBroker(db).WithLogger(logger)
```

### 4. Set Appropriate Timeouts

```go
// Short operations
broker.RunUseCases(ctx, useCases, input,
    broker.WithTimeout(2*time.Second))

// Long operations (reports, batch processing)
broker.RunUseCases(ctx, useCases, input,
    broker.WithTimeout(30*time.Second))

// Critical operations with high consistency requirements
broker.RunUseCases(ctx, useCases, input,
    broker.WithTimeout(5*time.Second),
    broker.WithIsolationLevel(sql.LevelSerializable))
```

### 5. Handle Context Cancellation

```go
result, err := broker.RunUseCases(ctx, useCases, input)
if err != nil {
    if ctx.Err() == context.Canceled {
        log.Println("operation canceled by user")
        return nil
    }
    if ctx.Err() == context.DeadlineExceeded {
        log.Println("operation timed out")
        return nil
    }
    // Other errors
    return err
}
```

## Common Patterns

### Pattern 1: Validation → Action → Notification

```go
validateInput := broker.UseCaseFunc(func(ctx, tx, input) (map[string]interface{}, error) {
    // Validate input data
    return map[string]interface{}{"validated": true}, nil
})

performAction := &CreateActivityUseCase{...}

sendNotification := broker.UseCaseFunc(func(ctx, tx, input) (map[string]interface{}, error) {
    // Send notification
    return map[string]interface{}{"notified": true}, nil
})

broker.RunUseCases(ctx, []broker.UseCase{
    validateInput,
    performAction,
    sendNotification,
}, input)
```

### Pattern 2: Create → Associate → Update Stats

```go
createEntity := &CreateActivityUseCase{...}
associateRelations := &AttachTagsUseCase{...}
updateMetrics := &UpdateStatsUseCase{...}

broker.RunUseCases(ctx, []broker.UseCase{
    createEntity,
    associateRelations,
    updateMetrics,
}, input)
```

### Pattern 3: Saga Pattern (Compensating Transactions)

For complex workflows where you might need to compensate:

```go
// Not directly supported - each use case should handle its own rollback logic
// The broker handles transaction rollback, not business-level compensation
```

## Migration from Direct Transaction Handling

### Before (Manual Transaction Management)

```go
func (h *Handler) CreateActivity(ctx context.Context, req Request) error {
    tx, err := h.db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback() // Will be no-op if committed

    // Create activity
    activity, err := h.activityRepo.Create(ctx, tx, req.Activity)
    if err != nil {
        return err
    }

    // Attach tags
    if err := h.tagRepo.AttachTags(ctx, tx, activity.ID, req.Tags); err != nil {
        return err
    }

    // Update stats
    if err := h.statsRepo.Update(ctx, tx, req.UserID); err != nil {
        return err
    }

    return tx.Commit()
}
```

### After (Broker Pattern)

```go
func (h *Handler) CreateActivity(ctx context.Context, req Request) error {
    result, err := h.broker.RunUseCases(
        ctx,
        []broker.UseCase{
            h.createActivityUC,
            h.attachTagsUC,
            h.updateStatsUC,
        },
        map[string]interface{}{
            "user_id": req.UserID,
            "title":   req.Activity.Title,
            "tags":    req.Tags,
        },
    )

    if err != nil {
        return err
    }

    log.Printf("Created activity: %v", result["activity_id"])
    return nil
}
```

## Summary

The broker orchestrator provides:

- ✅ **Atomic transactions** - All or nothing execution
- ✅ **Result chaining** - Automatic data flow between use cases
- ✅ **Error handling** - Automatic rollback on failure
- ✅ **Timeout support** - Configurable execution timeouts
- ✅ **Clean separation** - Each use case has single responsibility
- ✅ **Testability** - Easy to test use cases in isolation
- ✅ **Performance** - Minimal overhead (~160μs per execution)

For more details, see:
- `internal/application/broker/broker.go` - Implementation
- `internal/application/broker/broker_test.go` - Test examples
- `docs/BROKER_ORCHESTRATOR_PATTERN.md` - Pattern explanation
