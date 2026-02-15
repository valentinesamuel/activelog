# Test Restructuring Plan - Clean Architecture & DDD

## Current Problems

### 1. **Layer-Based Structure (Anemic)**
```
internal/
├── handlers/        # All HTTP handlers together
├── repository/      # All repositories together
├── models/          # All models together
└── tests scattered everywhere
```

**Issues:**
- Hard to understand which code belongs to which feature
- Tests are mixed with implementation
- Changes to one feature touch many directories
- Difficult to isolate and test domains independently

### 2. **Current Test Organization**
```
internal/repository/
├── activity_repository.go
├── activity_repository_test.go          # Unit tests mixed with code
├── activity_repository_bench_test.go    # Benchmarks mixed with code
├── integration_test.go                  # Generic integration tests
└── testhelpers/                         # Shared but not domain-specific
```

**Issues:**
- Tests are jumbled together
- No clear separation of test types
- Hard to find tests for specific features
- Test helpers are global, not domain-specific

---

## Proposed Structure - Feature-Based DDD

Following `kuja_user_ms` patterns adapted for Go:

```
activelog/
├── cmd/
│   └── api/
│       └── main.go                      # Application entry point
│
├── internal/
│   ├── domain/                          # 📦 Core Domain Logic (NO DEPENDENCIES)
│   │   ├── activity/
│   │   │   ├── entity.go                # Activity entity (business rules)
│   │   │   ├── repository.go            # Repository interface (port)
│   │   │   ├── service.go               # Domain service (business logic)
│   │   │   ├── errors.go                # Domain-specific errors
│   │   │   └── value_objects.go         # Value objects (ActivityType, etc.)
│   │   │
│   │   ├── user/
│   │   │   ├── entity.go
│   │   │   ├── repository.go
│   │   │   ├── service.go
│   │   │   └── errors.go
│   │   │
│   │   ├── tag/
│   │   │   ├── entity.go
│   │   │   ├── repository.go
│   │   │   └── service.go
│   │   │
│   │   └── shared/
│   │       ├── base_entity.go
│   │       └── types.go
│   │
│   ├── application/                     # 🎯 Use Cases / Application Logic
│   │   ├── activity/
│   │   │   ├── usecases/
│   │   │   │   ├── create_activity.go
│   │   │   │   ├── get_activity.go
│   │   │   │   ├── list_activities.go
│   │   │   │   └── delete_activity.go
│   │   │   ├── dto/
│   │   │   │   ├── create_activity_request.go
│   │   │   │   ├── create_activity_response.go
│   │   │   │   └── filters.go
│   │   │   └── ports/
│   │   │       └── activity_service.go  # Application service interface
│   │   │
│   │   ├── stats/
│   │   │   ├── usecases/
│   │   │   │   ├── get_weekly_stats.go
│   │   │   │   └── get_monthly_stats.go
│   │   │   └── dto/
│   │   │       └── stats_response.go
│   │   │
│   │   └── user/
│   │       ├── usecases/
│   │       │   ├── create_user.go
│   │       │   └── get_user.go
│   │       └── dto/
│   │           └── user_request.go
│   │
│   ├── infrastructure/                  # 🔌 External Adapters (Implementations)
│   │   ├── persistence/
│   │   │   ├── postgres/
│   │   │   │   ├── activity_repository.go
│   │   │   │   ├── user_repository.go
│   │   │   │   ├── tag_repository.go
│   │   │   │   ├── stats_repository.go
│   │   │   │   ├── migrations/
│   │   │   │   └── connection.go
│   │   │   └── memory/                  # In-memory implementations for testing
│   │   │       ├── activity_repository_memory.go
│   │   │       └── user_repository_memory.go
│   │   │
│   │   ├── logging/
│   │   │   └── logger.go
│   │   │
│   │   └── config/
│   │       └── config.go
│   │
│   ├── interfaces/                      # 🌐 HTTP/API Layer
│   │   ├── http/
│   │   │   ├── activity/
│   │   │   │   ├── handler.go
│   │   │   │   └── routes.go
│   │   │   ├── stats/
│   │   │   │   ├── handler.go
│   │   │   │   └── routes.go
│   │   │   ├── user/
│   │   │   │   ├── handler.go
│   │   │   │   └── routes.go
│   │   │   ├── health/
│   │   │   │   └── handler.go
│   │   │   └── middleware/
│   │   │       ├── auth.go
│   │   │       ├── cors.go
│   │   │       └── logger.go
│   │   │
│   │   └── validator/
│   │       └── validator.go
│   │
│   └── shared/                          # 🔧 Shared Utilities
│       ├── errors/
│       │   └── errors.go
│       ├── utils/
│       │   └── time.go
│       └── constants/
│           └── constants.go
│
└── tests/                               # 📋 ALL TESTS SEPARATED
    ├── unit/                            # Fast, isolated tests
    │   ├── domain/
    │   │   ├── activity/
    │   │   │   ├── entity_test.go
    │   │   │   ├── service_test.go
    │   │   │   └── value_objects_test.go
    │   │   ├── user/
    │   │   │   ├── entity_test.go
    │   │   │   └── service_test.go
    │   │   └── tag/
    │   │       └── entity_test.go
    │   │
    │   ├── application/
    │   │   ├── activity/
    │   │   │   ├── create_activity_test.go
    │   │   │   ├── get_activity_test.go
    │   │   │   └── list_activities_test.go
    │   │   └── stats/
    │   │       ├── get_weekly_stats_test.go
    │   │       └── get_monthly_stats_test.go
    │   │
    │   └── infrastructure/
    │       └── persistence/
    │           └── postgres/
    │               ├── activity_repository_test.go
    │               └── user_repository_test.go
    │
    ├── integration/                     # Database + multiple layers
    │   ├── activity/
    │   │   ├── create_activity_integration_test.go
    │   │   ├── activity_with_tags_test.go
    │   │   └── transactions_test.go
    │   ├── stats/
    │   │   └── stats_queries_test.go
    │   └── fixtures/
    │       ├── activity_fixtures.go
    │       └── user_fixtures.go
    │
    ├── e2e/                             # End-to-end API tests
    │   ├── activity/
    │   │   ├── create_activity_e2e_test.go
    │   │   └── list_activities_e2e_test.go
    │   ├── stats/
    │   │   └── stats_api_e2e_test.go
    │   └── testserver/
    │       └── server.go                # Test HTTP server setup
    │
    ├── benchmark/                       # Performance tests
    │   ├── activity/
    │   │   ├── create_benchmark_test.go
    │   │   ├── query_benchmark_test.go
    │   │   └── n_plus_one_benchmark_test.go
    │   └── shared/
    │       └── benchmark_helpers.go
    │
    ├── testhelpers/                     # Shared test utilities
    │   ├── database/
    │   │   ├── container.go             # Testcontainers setup
    │   │   ├── seeds.go                 # Database seeding
    │   │   └── cleanup.go               # Database cleanup
    │   ├── mocks/
    │   │   ├── activity_repository_mock.go
    │   │   ├── user_repository_mock.go
    │   │   └── tag_repository_mock.go
    │   ├── builders/                    # Test data builders
    │   │   ├── activity_builder.go
    │   │   ├── user_builder.go
    │   │   └── tag_builder.go
    │   └── assertions/
    │       └── custom_assertions.go
    │
    └── testdata/                        # Static test data
        ├── fixtures/
        │   ├── activities.json
        │   └── users.json
        └── golden/                      # Golden file testing
            └── expected_responses.json
```

---

## Key Improvements

### 1. **Feature-Based Organization** (like kuja_user_ms)

**Before:**
```go
// Everything scattered
internal/handlers/activity.go
internal/repository/activity_repository.go
internal/models/activity.go
```

**After:**
```go
// Everything related to activities is together
internal/domain/activity/           # Core logic
internal/application/activity/      # Use cases
internal/infrastructure/postgres/   # Implementation
tests/unit/domain/activity/         # Tests
```

### 2. **Test Separation by Type**

```
tests/
├── unit/           # Fast, no DB (< 1ms per test)
├── integration/    # DB + multiple components (< 100ms)
├── e2e/            # Full API tests (< 1s)
└── benchmark/      # Performance tests
```

**Benefits:**
- Run fast tests during development
- Run slow tests in CI only
- Clear test purpose from directory

### 3. **Clean Architecture Layers**

```
┌─────────────────────────────────────┐
│   interfaces/ (HTTP Handlers)      │  ← Outermost layer
├─────────────────────────────────────┤
│   application/ (Use Cases)         │  ← Application logic
├─────────────────────────────────────┤
│   domain/ (Entities & Interfaces)  │  ← Core business rules
├─────────────────────────────────────┤
│   infrastructure/ (DB, Cache, etc) │  ← External adapters
└─────────────────────────────────────┘

Dependency Rule: Inner layers never depend on outer layers
```

### 4. **Test Helpers Organization**

**Before:**
```
internal/repository/testhelpers/container.go  # Generic
```

**After:**
```
tests/testhelpers/
├── database/       # DB-specific helpers
├── mocks/          # Generated mocks
├── builders/       # Test data builders (Factory pattern)
└── assertions/     # Custom assertions
```

---

## Migration Strategy

### Phase 1: Create New Structure (Week 1)

```bash
# 1. Create new directory structure
mkdir -p internal/{domain,application,infrastructure,interfaces,shared}
mkdir -p tests/{unit,integration,e2e,benchmark,testhelpers,testdata}

# 2. Move domain entities
mv internal/models/activity.go internal/domain/activity/entity.go
mv internal/models/user.go internal/domain/user/entity.go
mv internal/models/tag.go internal/domain/tag/entity.go

# 3. Extract repository interfaces
# Create internal/domain/activity/repository.go with interface
# Create internal/domain/user/repository.go with interface

# 4. Move implementations
mv internal/repository/*_repository.go internal/infrastructure/persistence/postgres/
```

### Phase 2: Reorganize Tests (Week 1-2)

```bash
# 1. Move unit tests
mv internal/repository/activity_repository_test.go \
   tests/unit/infrastructure/persistence/postgres/activity_repository_test.go

# 2. Move integration tests
mv internal/repository/integration_test.go \
   tests/integration/activity/transactions_test.go

# 3. Move benchmarks
mv internal/repository/activity_repository_bench_test.go \
   tests/benchmark/activity/repository_benchmark_test.go

# 4. Move test helpers
mv internal/repository/testhelpers/* tests/testhelpers/database/
```

### Phase 3: Extract Use Cases (Week 2)

```bash
# 1. Extract use cases from handlers
# Before: internal/handlers/activity.go (everything in one file)
# After: internal/application/activity/usecases/create_activity.go
```

### Phase 4: Update Imports (Week 2-3)

```bash
# Update all imports to reflect new structure
# Use IDE refactoring or:
find . -name "*.go" -exec sed -i '' 's/old/new/g' {} +
```

---

## Example: Activity Feature Restructured

### Domain Layer (`internal/domain/activity/`)

```go
// entity.go - Pure business entity
package activity

import "time"

type Activity struct {
    ID              int64
    UserID          int
    Type            ActivityType    // Value object
    Title           string
    Description     string
    DurationMinutes int
    DistanceKm      float64
    CaloriesBurned  int
    Notes           string
    ActivityDate    time.Time
    Tags            []Tag
    CreatedAt       time.Time
    UpdatedAt       time.Time
}

// Business rule: Activity title is required
func (a *Activity) Validate() error {
    if a.Title == "" {
        return ErrActivityTitleRequired
    }
    return nil
}

// repository.go - Interface (port)
package activity

import "context"

type Repository interface {
    Create(ctx context.Context, activity *Activity) error
    GetByID(ctx context.Context, id int64) (*Activity, error)
    ListByUser(ctx context.Context, userID int, filters Filters) ([]*Activity, error)
    Delete(ctx context.Context, id int64) error
}

// service.go - Domain service (complex business logic)
package activity

type Service struct {
    repo Repository
}

func NewService(repo Repository) *Service {
    return &Service{repo: repo}
}

// Complex business logic that doesn't fit in entity
func (s *Service) CalculateCaloriesForActivity(activity *Activity) int {
    // Business rules for calorie calculation
    return 0 // Simplified
}
```

### Application Layer (`internal/application/activity/`)

```go
// usecases/create_activity.go
package usecases

import (
    "context"
    "github.com/valentinesamuel/activelog/internal/domain/activity"
)

type CreateActivityUseCase struct {
    activityRepo activity.Repository
    userRepo     user.Repository
}

func NewCreateActivityUseCase(
    activityRepo activity.Repository,
    userRepo user.Repository,
) *CreateActivityUseCase {
    return &CreateActivityUseCase{
        activityRepo: activityRepo,
        userRepo:     userRepo,
    }
}

func (uc *CreateActivityUseCase) Execute(
    ctx context.Context,
    req CreateActivityRequest,
) (*CreateActivityResponse, error) {
    // 1. Validate user exists
    _, err := uc.userRepo.GetByID(ctx, req.UserID)
    if err != nil {
        return nil, err
    }

    // 2. Create domain entity
    act := &activity.Activity{
        UserID:          req.UserID,
        Type:            activity.ActivityType(req.Type),
        Title:           req.Title,
        Description:     req.Description,
        DurationMinutes: req.DurationMinutes,
    }

    // 3. Validate business rules
    if err := act.Validate(); err != nil {
        return nil, err
    }

    // 4. Persist
    if err := uc.activityRepo.Create(ctx, act); err != nil {
        return nil, err
    }

    // 5. Return DTO
    return &CreateActivityResponse{
        ID:        act.ID,
        Title:     act.Title,
        CreatedAt: act.CreatedAt,
    }, nil
}
```

### Infrastructure Layer (`internal/infrastructure/persistence/postgres/`)

```go
// activity_repository.go - Concrete implementation
package postgres

import (
    "context"
    "github.com/valentinesamuel/activelog/internal/domain/activity"
)

type ActivityRepository struct {
    db *database.LoggingDB
}

// Implements activity.Repository interface
func (r *ActivityRepository) Create(
    ctx context.Context,
    act *activity.Activity,
) error {
    query := `INSERT INTO activities (...) VALUES (...)`
    return r.db.QueryRow(query, ...).Scan(&act.ID)
}
```

### Interface Layer (`internal/interfaces/http/activity/`)

```go
// handler.go - HTTP adapter
package activity

import (
    "net/http"
    "github.com/valentinesamuel/activelog/internal/application/activity/usecases"
)

type Handler struct {
    createUseCase *usecases.CreateActivityUseCase
}

func NewHandler(createUseCase *usecases.CreateActivityUseCase) *Handler {
    return &Handler{createUseCase: createUseCase}
}

func (h *Handler) CreateActivity(w http.ResponseWriter, r *http.Request) {
    // 1. Parse request
    var req usecases.CreateActivityRequest
    json.NewDecoder(r.Body).Decode(&req)

    // 2. Call use case
    resp, err := h.createUseCase.Execute(r.Context(), req)
    if err != nil {
        // Handle error
        return
    }

    // 3. Return response
    json.NewEncoder(w).Encode(resp)
}
```

---

## Test Organization Examples

### Unit Test (`tests/unit/domain/activity/entity_test.go`)

```go
package activity_test

import (
    "testing"
    "github.com/valentinesamuel/activelog/internal/domain/activity"
)

func TestActivity_Validate(t *testing.T) {
    tests := []struct {
        name    string
        activity *activity.Activity
        wantErr bool
    }{
        {
            name: "valid activity",
            activity: &activity.Activity{
                Title: "Morning Run",
            },
            wantErr: false,
        },
        {
            name: "missing title",
            activity: &activity.Activity{
                Title: "",
            },
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.activity.Validate()
            if (err != nil) != tt.wantErr {
                t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### Integration Test (`tests/integration/activity/create_activity_integration_test.go`)

```go
package activity_test

import (
    "context"
    "testing"
    "github.com/valentinesamuel/activelog/internal/application/activity/usecases"
    "github.com/valentinesamuel/activelog/tests/testhelpers/database"
    "github.com/valentinesamuel/activelog/tests/testhelpers/builders"
)

func TestCreateActivity_Integration(t *testing.T) {
    // Setup test database
    db, cleanup := database.Setup(t)
    defer cleanup()

    // Create repositories
    activityRepo := postgres.NewActivityRepository(db)
    userRepo := postgres.NewUserRepository(db)

    // Create use case
    useCase := usecases.NewCreateActivityUseCase(activityRepo, userRepo)

    // Setup test data using builder pattern
    user := builders.NewUserBuilder().
        WithEmail("test@example.com").
        Build()
    userRepo.Create(context.Background(), user)

    // Execute test
    req := usecases.CreateActivityRequest{
        UserID: user.ID,
        Title:  "Test Activity",
    }

    resp, err := useCase.Execute(context.Background(), req)

    // Assertions
    if err != nil {
        t.Fatalf("Expected no error, got %v", err)
    }
    if resp.ID == 0 {
        t.Error("Expected activity ID to be set")
    }
}
```

### E2E Test (`tests/e2e/activity/create_activity_e2e_test.go`)

```go
package activity_test

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    "github.com/valentinesamuel/activelog/tests/e2e/testserver"
)

func TestCreateActivityAPI_E2E(t *testing.T) {
    // Setup full test server
    server := testserver.NewTestServer(t)
    defer server.Cleanup()

    // Prepare request
    reqBody := map[string]interface{}{
        "user_id": 1,
        "title":   "Morning Run",
        "type":    "running",
    }
    body, _ := json.Marshal(reqBody)

    // Make HTTP request
    req := httptest.NewRequest("POST", "/api/activities", bytes.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()

    server.Router.ServeHTTP(w, req)

    // Assert response
    if w.Code != http.StatusCreated {
        t.Errorf("Expected status 201, got %d", w.Code)
    }

    var resp map[string]interface{}
    json.NewDecoder(w.Body).Decode(&resp)

    if resp["id"] == nil {
        t.Error("Expected ID in response")
    }
}
```

### Benchmark (`tests/benchmark/activity/create_benchmark_test.go`)

```go
package activity_test

import (
    "context"
    "testing"
    "github.com/valentinesamuel/activelog/tests/testhelpers/database"
    "github.com/valentinesamuel/activelog/tests/testhelpers/builders"
)

func BenchmarkCreateActivity(b *testing.B) {
    db, cleanup := database.Setup(b)
    defer cleanup()

    repo := postgres.NewActivityRepository(db)
    user := builders.NewUserBuilder().Build()

    b.ResetTimer()
    b.ReportAllocs()

    for i := 0; i < b.N; i++ {
        activity := builders.NewActivityBuilder().
            WithUserID(user.ID).
            Build()

        _ = repo.Create(context.Background(), activity)
    }
}
```

---

## Test Helpers Organization

### Builder Pattern (`tests/testhelpers/builders/activity_builder.go`)

```go
package builders

import (
    "time"
    "github.com/valentinesamuel/activelog/internal/domain/activity"
)

type ActivityBuilder struct {
    activity *activity.Activity
}

func NewActivityBuilder() *ActivityBuilder {
    return &ActivityBuilder{
        activity: &activity.Activity{
            Title:           "Default Activity",
            Type:            activity.Running,
            DurationMinutes: 30,
            ActivityDate:    time.Now(),
        },
    }
}

func (b *ActivityBuilder) WithTitle(title string) *ActivityBuilder {
    b.activity.Title = title
    return b
}

func (b *ActivityBuilder) WithUserID(userID int) *ActivityBuilder {
    b.activity.UserID = userID
    return b
}

func (b *ActivityBuilder) WithType(actType activity.ActivityType) *ActivityBuilder {
    b.activity.Type = actType
    return b
}

func (b *ActivityBuilder) Build() *activity.Activity {
    return b.activity
}

// Usage:
// activity := builders.NewActivityBuilder().
//     WithTitle("Morning Run").
//     WithUserID(1).
//     Build()
```

### Database Helpers (`tests/testhelpers/database/setup.go`)

```go
package database

import (
    "testing"
    "github.com/valentinesamuel/activelog/internal/infrastructure/persistence/postgres"
)

func Setup(t testing.TB) (*postgres.DB, func()) {
    db, cleanup := setupContainer(t)
    runMigrations(t, db)
    seedData(t, db)
    return db, cleanup
}

func SetupClean(t testing.TB) (*postgres.DB, func()) {
    db, cleanup := setupContainer(t)
    runMigrations(t, db)
    // No seed data
    return db, cleanup
}
```

---

## Running Tests by Type

### Makefile Commands

```makefile
# Unit tests (fast, no DB)
test-unit:
	@echo "Running unit tests..."
	go test ./tests/unit/... -v -short

# Integration tests (with DB)
test-integration:
	@echo "Running integration tests..."
	go test ./tests/integration/... -v

# E2E tests (full API)
test-e2e:
	@echo "Running E2E tests..."
	go test ./tests/e2e/... -v

# Benchmarks
test-bench:
	@echo "Running benchmarks..."
	go test ./tests/benchmark/... -bench=. -benchmem

# All tests
test-all:
	@echo "Running all tests..."
	go test ./tests/... -v

# Domain tests only
test-domain:
	@echo "Running domain tests..."
	go test ./tests/unit/domain/... -v

# Activity feature tests
test-activity:
	@echo "Running activity tests..."
	go test ./tests/unit/domain/activity/... -v
	go test ./tests/integration/activity/... -v
	go test ./tests/e2e/activity/... -v
```

---

## Benefits of This Structure

### 1. **Feature Cohesion**
```
All activity-related code is together:
- Domain: internal/domain/activity/
- Use cases: internal/application/activity/
- Tests: tests/*/activity/
```

### 2. **Clear Dependencies**
```
Domain → No dependencies on anything
Application → Depends on Domain
Infrastructure → Implements Domain interfaces
Interfaces → Depends on Application
```

### 3. **Test Independence**
```
tests/unit/ → Fast, no external dependencies
tests/integration/ → Database required
tests/e2e/ → Full stack required
tests/benchmark/ → Performance focused
```

### 4. **Easy Navigation**
```
Want activity tests? → tests/*/activity/
Want activity code? → internal/domain/activity/
Want activity use cases? → internal/application/activity/usecases/
```

### 5. **Testability**
```
Domain: Easy to test (pure functions)
Application: Test with mocks
Infrastructure: Test with real DB
Interfaces: Test with test server
```

---

## Migration Checklist

- [ ] **Week 1:** Create new directory structure
- [ ] **Week 1:** Move domain entities
- [ ] **Week 1:** Extract repository interfaces
- [ ] **Week 1:** Move repository implementations
- [ ] **Week 2:** Reorganize all tests by type
- [ ] **Week 2:** Create test helpers/builders
- [ ] **Week 2:** Extract use cases from handlers
- [ ] **Week 3:** Update all imports
- [ ] **Week 3:** Update documentation
- [ ] **Week 3:** Update CI/CD pipelines
- [ ] **Week 4:** Remove old structure
- [ ] **Week 4:** Final testing

---

## Comparison: Before vs After

### Before (Current - Layer-Based)

```
❌ Hard to find all activity-related code
❌ Tests scattered everywhere
❌ Domain logic mixed with infrastructure
❌ Hard to test in isolation
❌ Unclear dependencies

internal/
├── handlers/activity.go
├── repository/activity_repository.go
├── repository/activity_repository_test.go
└── models/activity.go
```

### After (Proposed - Feature-Based DDD)

```
✅ All activity code in one place
✅ Tests organized by type
✅ Clean separation of concerns
✅ Easy to test each layer
✅ Clear dependency flow

internal/domain/activity/
internal/application/activity/
internal/infrastructure/postgres/
tests/unit/domain/activity/
tests/integration/activity/
```

---

## Next Steps

1. **Review this plan** - Discuss with team
2. **Start small** - Migrate one feature first (e.g., Activity)
3. **Run tests continuously** - Ensure nothing breaks
4. **Update documentation** - Keep README in sync
5. **Iterate** - Adjust based on learnings

---

## References

- **Clean Architecture** by Robert C. Martin
- **Domain-Driven Design** by Eric Evans
- **Go Project Layout**: https://github.com/golang-standards/project-layout
- **NestJS Structure** (your kuja_user_ms project)

---

**Remember:** This is a gradual migration. Don't try to do everything at once!
