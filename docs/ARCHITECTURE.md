# ActiveLog Architecture Documentation

**Last Updated:** 2026-01-07
**Version:** 1.0.0 with Service Layer

## Table of Contents
1. [Overview](#overview)
2. [Architectural Principles](#architectural-principles)
3. [Layer Breakdown](#layer-breakdown)
4. [Request Flow](#request-flow)
5. [Service Layer Design](#service-layer-design)
6. [Transaction Management](#transaction-management)
7. [Dependency Injection](#dependency-injection)
8. [Design Patterns](#design-patterns)
9. [Trade-offs and Decisions](#trade-offs-and-decisions)

## Overview

ActiveLog follows **Clean Architecture** principles with **Hexagonal Architecture** influences, organized into distinct layers with clear boundaries and dependencies flowing inward.

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        HTTP Layer                                 │
│  (Handlers, Middleware, Routing)                                 │
└──────────────────────┬──────────────────────────────────────────┘
                       │
┌──────────────────────┴──────────────────────────────────────────┐
│                     Broker Pattern                                │
│  (Use Case Orchestration, Transaction Management)                │
└──────────────────────┬──────────────────────────────────────────┘
                       │
┌──────────────────────┴──────────────────────────────────────────┐
│                    Use Case Layer                                 │
│  (Application Logic, Flow Control)                               │
└──────────────────────┬──────────────────────────────────────────┘
                       │
          ┌────────────┴──────────────┐
          │                           │
┌─────────┴─────────┐       ┌─────────┴─────────┐
│  Service Layer    │       │  Direct Repository │
│ (Business Logic)  │       │  (Simple reads)    │
└─────────┬─────────┘       └─────────┬─────────┘
          │                           │
          └────────────┬──────────────┘
                       │
┌──────────────────────┴──────────────────────────────────────────┐
│                   Repository Layer                                │
│  (Data Access, Database Operations)                              │
└──────────────────────┬──────────────────────────────────────────┘
                       │
┌──────────────────────┴──────────────────────────────────────────┐
│                      Database                                     │
│  (PostgreSQL)                                                     │
└─────────────────────────────────────────────────────────────────┘
```

## Architectural Principles

### 1. Dependency Inversion
- **High-level modules don't depend on low-level modules**
- Both depend on abstractions (interfaces)
- Example: Use cases depend on `service.ActivityServiceInterface`, not concrete service

### 2. Single Responsibility
- Each layer has ONE reason to change
- **Handlers:** HTTP concerns only
- **Broker:** Transaction and orchestration only
- **Use Cases:** Application flow only
- **Services:** Business logic only
- **Repositories:** Data access only

### 3. Flexibility with Fallbacks
- **Preferred path:** Use cases → Services → Repositories
- **Fallback path:** Use cases → Repositories (for simple operations)
- Allows optimization without sacrificing architecture

### 4. Interface Segregation
- Small, focused interfaces
- Example: `ActivityServiceInterface` vs. `StatsServiceInterface`
- Easy to mock and test

## Layer Breakdown

### Layer 1: HTTP Layer (Entry Point)
**Location:** `/internal/handlers`

**Responsibilities:**
- Parse HTTP requests
- Call broker with appropriate use cases
- Format HTTP responses
- Handle HTTP-specific concerns (status codes, headers)

**Example:**
```go
func (h *ActivityHandlerV2) CreateActivity(w http.ResponseWriter, r *http.Request) {
    var req models.CreateActivityRequest
    json.NewDecoder(r.Body).Decode(&req)

    result, err := h.broker.RunUseCases(
        ctx,
        []broker.UseCase{h.createActivityUC},
        map[string]interface{}{
            "user_id": 1,
            "request": &req,
        },
    )

    response.SendJSON(w, http.StatusCreated, result["activity"])
}
```

### Layer 2: Broker Pattern (Orchestration)
**Location:** `/internal/application/broker`

**Responsibilities:**
- Orchestrate multiple use cases
- Manage transaction boundaries
- Handle transaction commits/rollbacks
- Chain use case results
- Detect transactional vs. non-transactional use cases

**Key Features:**
- **Transaction Boundary Breaking:** UC1(tx) → UC2(non-tx) → UC3(tx)
- **Result Chaining:** Output of UC1 → Input of UC2
- **Automatic Cleanup:** Rollback on error, commit on success

**Example:**
```go
result, err := broker.RunUseCases(
    ctx,
    []broker.UseCase{createUC, getUC, updateUC},
    input,
)
// Broker automatically:
// 1. Starts transaction for createUC
// 2. Commits before getUC (non-transactional)
// 3. Starts new transaction for updateUC
// 4. Commits final transaction
```

### Layer 3: Use Case Layer (Application Logic)
**Location:** `/internal/application/*/usecases`

**Responsibilities:**
- Coordinate application flow
- Decide whether to use service or repository for each operation
- Map between handler input and service/repository calls
- Return results to broker

**Design:** Flexible use cases with runtime decision-making
```go
type CreateActivityUseCase struct {
    service service.ActivityServiceInterface      // For operations requiring business logic
    repo    repository.ActivityRepositoryInterface // For simple operations or when service not needed
}

// Single constructor that provides both dependencies
// Use case decides at runtime which one to use based on operation needs
func NewCreateActivityUseCase(
    svc service.ActivityServiceInterface,
    repo repository.ActivityRepositoryInterface,
) *CreateActivityUseCase {
    return &CreateActivityUseCase{
        service: svc,
        repo:    repo,
    }
}

// Execute method decides which dependency to use
func (uc *CreateActivityUseCase) Execute(...) {
    // DECISION: Use service for create operations because we need business logic validation
    activity, err := uc.service.CreateActivity(ctx, tx, userID, req)

    // Example: Could also use repo directly if needed
    // tags, _ := uc.repo.GetActivitiesWithTags(ctx, userID, filters)
}
```

**Decision Guidelines:**
- **Use Service when:**
  - Complex business logic needed
  - Validation required
  - Multiple repositories must be coordinated
  - Ownership verification required
  - Business policy enforcement needed

- **Use Repository directly when:**
  - Simple CRUD operations
  - No business logic needed
  - Performance-critical reads
  - Straightforward data retrieval

- **Use Both when:**
  - Service for main operation
  - Repository for related/supplementary queries
  - Service for business logic + repo for optimization

### Layer 4: Service Layer (Business Logic)
**Location:** `/internal/service`

**Responsibilities:**
- Encapsulate business rules
- Validate business constraints
- Coordinate multiple repositories
- Implement complex domain logic

**Example - ActivityService:**
```go
func (s *ActivityService) CreateActivity(
    ctx context.Context,
    tx repository.TxConn,
    userID int,
    req *models.CreateActivityRequest,
) (*models.Activity, error) {
    // Business Rule 1: Activity date cannot be in the future
    if req.ActivityDate.After(time.Now()) {
        return nil, fmt.Errorf("activity date cannot be in the future")
    }

    // Business Rule 2: Duration must be reasonable
    if req.DurationMinutes > 1440 {
        return nil, fmt.Errorf("duration cannot exceed 24 hours")
    }

    // Business Rule 3: Distance must be positive
    if req.DistanceKm < 0 {
        return nil, fmt.Errorf("distance must be positive")
    }

    // Create activity
    activity := &models.Activity{/* ... */}
    if err := s.activityRepo.Create(ctx, tx, activity); err != nil {
        return nil, err
    }

    return activity, nil
}
```

**Key Benefits:**
- Business logic is **testable in isolation**
- Rules are **centralized and reusable**
- Changes to business logic **don't affect use cases or repositories**

### Layer 5: Repository Layer (Data Access)
**Location:** `/internal/repository`

**Responsibilities:**
- Execute database queries
- Map between database rows and domain models
- Handle database-specific concerns (connection, transactions)
- NO business logic

**Example:**
```go
func (r *ActivityRepository) Create(
    ctx context.Context,
    tx TxConn,
    activity *models.Activity,
) error {
    query := `INSERT INTO activities (...) VALUES ($1, $2, ...) RETURNING id`
    return tx.QueryRowContext(ctx, query, activity.UserID, activity.Title, ...).
        Scan(&activity.ID)
}
```

## Request Flow

### Write Operation Flow (Create Activity)

```
Client Request
    ↓
HTTP Handler (activity_v2.go)
    ├─ Parse JSON request
    ├─ Extract user_id from auth context
    └─ Call broker.RunUseCases()
        ↓
Broker Pattern (broker.go)
    ├─ Detect use case needs transaction (RequiresTransaction() = true)
    ├─ Begin transaction
    └─ Execute CreateActivityUseCase
        ↓
Use Case (create_activity.go)
    ├─ Extract input from map
    └─ Call service.CreateActivity()
        ↓
Service Layer (activity_service.go)
    ├─ Validate business rules
    │   ├─ Date not in future
    │   ├─ Duration reasonable
    │   └─ Distance positive
    ├─ Build activity entity
    └─ Call repo.Create()
        ↓
Repository Layer (activity_repository.go)
    ├─ Execute SQL INSERT
    ├─ Scan returned ID
    └─ Return activity
        ↓
Service → Use Case → Broker
    ├─ Commit transaction
    └─ Return result
        ↓
Handler → Client
    └─ Send JSON response (201 Created)
```

### Read Operation Flow (Get Activity)

```
Client Request
    ↓
HTTP Handler
    └─ Call broker.RunUseCases()
        ↓
Broker Pattern
    ├─ Detect use case is non-transactional (no RequiresTransaction() method)
    └─ Execute GetActivityUseCase (NO transaction)
        ↓
Use Case
    └─ Call repo.GetByID() directly (bypasses service for simple read)
        ↓
Repository
    ├─ Execute SQL SELECT
    └─ Return activity
        ↓
Use Case → Broker → Handler → Client
```

**Why bypass service for reads?**
- Simple reads have no business logic
- Reduces unnecessary abstraction layers
- Improves performance (~8% faster)
- Flexibility to add service later if needed

### Mixed Transaction Flow (Complex Operation)

```
Broker receives: [CreateUC (tx), GetUC (non-tx), UpdateUC (tx)]
    ↓
Step 1: CreateActivityUseCase
    ├─ Begin Transaction 1
    ├─ Service validates & creates
    ├─ Commit Transaction 1
    └─ Result: {"activity_id": 123}
        ↓
Step 2: GetActivityUseCase (non-transactional)
    ├─ NO transaction
    ├─ Repository fetches activity
    └─ Result: {"activity": {...}}
        ↓
Step 3: UpdateActivityUseCase
    ├─ Begin Transaction 2
    ├─ Service validates & updates
    ├─ Commit Transaction 2
    └─ Result: {"activity": {...}, "updated": true}
        ↓
Final result: Merged results from all use cases
```

## Service Layer Design

### When to Use Service Layer

**✅ Use Service Layer For:**
- Write operations (Create, Update, Delete)
- Complex business logic
- Multi-repository coordination
- Business rule validation
- Operations with side effects

**❌ Skip Service Layer For:**
- Simple reads (GetByID, List)
- Operations with no business logic
- Performance-critical queries
- Aggregations already handled by database

### Service Interface Design

```go
// Good: Focused, single-purpose interface
type ActivityServiceInterface interface {
    CreateActivity(...) (*models.Activity, error)
    UpdateActivity(...) (*models.Activity, error)
    DeleteActivity(...) error
}

// Bad: God interface with too many responsibilities
type MegaServiceInterface interface {
    // Activities
    CreateActivity(...)
    UpdateActivity(...)
    // Users
    CreateUser(...)
    // Stats
    GetStats(...)
    // Notifications
    SendEmail(...)
}
```

### Service Implementation Pattern

```go
type ActivityService struct {
    activityRepo repository.ActivityRepositoryInterface
    tagRepo      repository.TagRepositoryInterface
    // Add more repositories as needed
}

func NewActivityService(
    activityRepo repository.ActivityRepositoryInterface,
    tagRepo repository.TagRepositoryInterface,
) *ActivityService {
    return &ActivityService{
        activityRepo: activityRepo,
        tagRepo:      tagRepo,
    }
}

func (s *ActivityService) CreateActivity(...) (*models.Activity, error) {
    // 1. Validate business rules
    // 2. Coordinate multiple repositories if needed
    // 3. Apply business logic
    // 4. Return result
}
```

## Transaction Management

### Transaction Marker Interface

Use cases declare transaction requirements via optional method:

```go
type TransactionalUseCase interface {
    RequiresTransaction() bool
}

// Transactional use case
func (uc *CreateActivityUseCase) RequiresTransaction() bool {
    return true
}

// Non-transactional use case (no method = defaults to false)
type GetActivityUseCase struct {}
// No RequiresTransaction() method
```

### Transaction Lifecycle

```go
// Broker detects transaction needs
for _, useCase := range useCases {
    needsTx := false
    if txUC, ok := useCase.(TransactionalUseCase); ok {
        needsTx = txUC.RequiresTransaction()
    }

    if needsTx && activeTx == nil {
        // Start new transaction
        activeTx, _ = broker.db.BeginTx(ctx, ...)
    } else if !needsTx && activeTx != nil {
        // Boundary break: commit active transaction
        activeTx.Commit()
        activeTx = nil
    }

    // Execute use case
    result, err := useCase.Execute(ctx, activeTx, input)

    // Handle errors
    if err != nil {
        if activeTx != nil {
            activeTx.Rollback()
        }
        return nil, err
    }
}

// Commit final transaction if exists
if activeTx != nil {
    activeTx.Commit()
}
```

### Transaction Best Practices

1. **Keep transactions short** - Only write operations
2. **Don't mix transactions with external calls** - API calls, email, etc.
3. **Use transaction boundary breaking** - For mixed read/write chains
4. **Let broker manage transactions** - Don't manage transactions manually in use cases

## Dependency Injection

### DI Container Pattern

**Location:** `/cmd/api/container.go`

ActiveLog uses a **manual factory pattern** for dependency injection:

```go
func setupContainer(db repository.DBConn) *container.Container {
    c := container.New()

    // 1. Register repositories
    c.Register("activityRepo", func(c *container.Container) (interface{}, error) {
        db := c.MustResolve("db").(repository.DBConn)
        tagRepo := c.MustResolve("tagRepo").(*repository.TagRepository)
        return repository.NewActivityRepository(db, tagRepo), nil
    })

    // 2. Register services
    c.Register("activityService", func(c *container.Container) (interface{}, error) {
        activityRepo := c.MustResolve("activityRepo").(repository.ActivityRepositoryInterface)
        tagRepo := c.MustResolve("tagRepo").(repository.TagRepositoryInterface)
        return service.NewActivityService(activityRepo, tagRepo), nil
    })

    // 3. Register use cases (inject both service and repository)
    c.Register("createActivityUC", func(c *container.Container) (interface{}, error) {
        svc := c.MustResolve("activityService").(service.ActivityServiceInterface)
        repo := c.MustResolve("activityRepo").(repository.ActivityRepositoryInterface)
        return usecases.NewCreateActivityUseCase(svc, repo), nil
    })

    // 4. Register handlers
    c.Register("activityHandler", func(c *container.Container) (interface{}, error) {
        broker := c.MustResolve("broker").(*broker.Broker)
        repo := c.MustResolve("activityRepo").(repository.ActivityRepositoryInterface)
        createUC := c.MustResolve("createActivityUC").(broker.UseCase)
        // ... more use cases
        return handlers.NewActivityHandlerV2(broker, repo, createUC, ...), nil
    })

    return c
}
```

### Registration Order

**CRITICAL:** Dependencies must be registered before dependents!

```
1. Core (DB, Logger)
2. Repositories
3. Services
4. Broker
5. Use Cases
6. Handlers
```

Wrong order causes panic: `dependency not found`

### Singleton vs. Transient

```go
// Singleton: Created once, reused forever
c.RegisterSingleton("db", db)
c.RegisterSingleton("broker", broker.NewBroker(db))

// Transient: Created every time it's resolved
c.Register("createActivityUC", func(c *container.Container) (interface{}, error) {
    return usecases.NewCreateActivityUseCase(...), nil
})
```

**Rule of Thumb:**
- **Singleton:** Stateless services, repositories, broker
- **Transient:** Rarely used (most dependencies are singletons in practice)

## Design Patterns

### 1. Broker Pattern (Use Case Orchestrator)
**Purpose:** Coordinate multiple use cases within transactions

**Benefits:**
- Centralized transaction management
- Use case composition
- Result chaining
- Consistent error handling

### 2. Service Layer Pattern
**Purpose:** Encapsulate business logic separate from application flow

**Benefits:**
- Testable business logic
- Reusable across use cases
- Clear separation of concerns

### 3. Repository Pattern
**Purpose:** Abstract data access from business logic

**Benefits:**
- Easy to swap databases
- Testable with mock repositories
- Clear data access layer

### 4. Dependency Injection
**Purpose:** Invert dependencies, enable testing, reduce coupling

**Benefits:**
- Easy to test (inject mocks)
- Flexible (swap implementations)
- Clear dependency graph

### 5. Interface Segregation
**Purpose:** Small, focused interfaces instead of god interfaces

**Benefits:**
- Easy to implement
- Easy to mock
- Clear contracts

## Trade-offs and Decisions

### Decision 1: Service Layer (Optional, But Recommended)

**Options Considered:**
1. **No service layer** - Use cases call repositories directly
2. **Mandatory service layer** - All use cases must use services
3. **Optional service layer** (CHOSEN) - Use cases can choose

**Why Optional:**
- Flexibility for simple operations
- Performance optimization when needed
- Gradual migration path
- Not all operations need business logic

**Recommendation:** Use services for writes, skip for simple reads

### Decision 2: Map vs. Struct for Use Case Input

**Chosen:** `map[string]interface{}`

**Pros:**
- Flexible (can add fields without changing signatures)
- Easy to chain results
- Works with broker pattern

**Cons:**
- Runtime type assertions
- No compile-time safety
- More allocations

**Alternative Considered:** Typed structs (rejected for flexibility)

### Decision 3: Transaction Marker Interface vs. Annotation

**Chosen:** Optional `RequiresTransaction() bool` method

**Pros:**
- Defaults to non-transactional (performance)
- Explicit declaration
- Easy to implement

**Cons:**
- Not enforced at compile time
- Easy to forget

**Alternative Considered:** Annotations (not available in Go)

### Decision 4: DI Container vs. Wire/Fx

**Chosen:** Manual factory pattern

**Pros:**
- No external dependencies
- Full control
- Easy to understand
- Predictable behavior

**Cons:**
- More boilerplate
- No compile-time safety
- Manual dependency graph

**Alternative Considered:** Wire (rejected for simplicity), Fx (rejected for complexity)

## Performance Characteristics

See [PERFORMANCE_BENCHMARKS.md](./PERFORMANCE_BENCHMARKS.md) for detailed metrics.

**Summary:**
- **Broker overhead:** ~10% latency increase
- **Service layer:** ~8% latency increase
- **DI container:** <1% overhead
- **Total overhead:** ~18-20% vs. direct repository calls
- **Verdict:** Acceptable trade-off for maintainability

## Testing Strategy

### Unit Testing

```go
// Test service in isolation
func TestActivityService_CreateActivity(t *testing.T) {
    mockRepo := &mockActivityRepo{}
    svc := service.NewActivityService(mockRepo, ...)

    activity, err := svc.CreateActivity(ctx, nil, 1, req)

    assert.NoError(t, err)
    assert.NotNil(t, activity)
}

// Test use case with mock service
func TestCreateActivityUseCase(t *testing.T) {
    mockService := &mockActivityService{}
    uc := usecases.NewCreateActivityUseCase(mockService)

    result, err := uc.Execute(ctx, nil, input)

    assert.NoError(t, err)
}
```

### Integration Testing

```go
// Test full stack with real database
func TestIntegration_CreateActivity(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()

    // Wire up real dependencies
    repo := repository.NewActivityRepository(db, ...)
    svc := service.NewActivityService(repo, ...)
    uc := usecases.NewCreateActivityUseCase(svc)
    broker := broker.NewBroker(db.GetRawDB())

    // Test end-to-end
    result, err := broker.RunUseCases(ctx, []broker.UseCase{uc}, input)

    // Verify in database
    var count int
    db.QueryRow("SELECT COUNT(*) FROM activities").Scan(&count)
    assert.Equal(t, 1, count)
}
```

## Migration Path

### Migrating Existing Code to Service Layer

**Step 1:** Create service interface
```go
type ActivityServiceInterface interface {
    CreateActivity(...) (*models.Activity, error)
}
```

**Step 2:** Implement service
```go
type ActivityService struct {
    activityRepo repository.ActivityRepositoryInterface
}

func (s *ActivityService) CreateActivity(...) (*models.Activity, error) {
    // Move business logic from use case here
    // Validate
    // Create
    return activity, nil
}
```

**Step 3:** Update use case constructor to accept both service and repository
```go
// Before (old pattern - either service OR repo)
uc := usecases.NewCreateActivityUseCase(svc)
// OR
uc := usecases.NewCreateActivityUseCaseWithRepo(repo)

// After (new pattern - both service AND repo)
uc := usecases.NewCreateActivityUseCase(svc, repo)
```

**Step 4:** Use case decides which dependency to use at runtime
```go
// In the use case Execute method
func (uc *CreateActivityUseCase) Execute(...) {
    // DECISION: Use service for business logic validation
    activity, err := uc.service.CreateActivity(ctx, tx, userID, req)

    // Could also use repo directly if needed for supplementary queries
    // tags, _ := uc.repo.GetActivitiesWithTags(ctx, userID, filters)
}
```

## Conclusion

ActiveLog's architecture prioritizes:
1. **Maintainability** - Clear layers, single responsibilities
2. **Testability** - Easy to mock, isolated testing
3. **Flexibility** - Use cases receive both service and repository, deciding at runtime which to use based on operation needs
4. **Performance** - Acceptable overhead (<20%) for significant benefits

The architecture is **production-ready** and **scales well** for the anticipated load.

---

**Questions or Concerns?** See [BROKER_IMPLEMENTATION_SUMMARY.md](./BROKER_IMPLEMENTATION_SUMMARY.md) for broker pattern details.
