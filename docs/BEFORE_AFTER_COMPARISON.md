# Before & After Comparison - Visual Guide

## Architecture Comparison

### Current Architecture (Layer-Based) ❌

```
┌─────────────────────────────────────────┐
│          HTTP Handlers                  │  Everything by technical layer
│  (All handlers mixed together)          │
├─────────────────────────────────────────┤
│          Repository Layer                │
│  (All repositories mixed)                │
├─────────────────────────────────────────┤
│          Models Layer                    │
│  (All models mixed)                      │
├─────────────────────────────────────────┤
│          Database                        │
└─────────────────────────────────────────┘

Problems:
- Hard to find related code
- Changes affect many files
- No clear feature boundaries
- Tests scattered everywhere
```

### Proposed Architecture (Feature-Based DDD) ✅

```
┌─────────────────────────────────────────────────────┐
│  Interfaces Layer (HTTP/API)                        │
│  ├── activity/  ├── stats/  ├── user/              │
├─────────────────────────────────────────────────────┤
│  Application Layer (Use Cases)                      │
│  ├── activity/  ├── stats/  ├── user/              │
├─────────────────────────────────────────────────────┤
│  Domain Layer (Business Rules) - NO DEPENDENCIES    │
│  ├── activity/  ├── stats/  ├── user/              │
├─────────────────────────────────────────────────────┤
│  Infrastructure Layer (DB, Cache, External)         │
│  └── persistence/ logging/ config/                  │
└─────────────────────────────────────────────────────┘

Benefits:
✅ Features are self-contained
✅ Clear dependency flow
✅ Easy to find related code
✅ Tests organized by purpose
```

---

## File Structure Comparison

### BEFORE (Current - Messy)

```
internal/
├── handlers/
│   ├── activity.go              ← HTTP logic
│   ├── activity_test.go         ← Tests mixed with code
│   ├── stats.go
│   ├── stats_handler_test.go
│   ├── user.go
│   └── user_test.go
│
├── repository/
│   ├── activity_repository.go   ← DB logic in different folder
│   ├── activity_repository_test.go
│   ├── activity_repository_bench_test.go
│   ├── integration_test.go      ← Generic integration tests
│   ├── user_repository.go
│   ├── stats_repository.go
│   └── testhelpers/
│       └── container.go         ← Shared helpers
│
└── models/
    ├── activity.go              ← Data model in yet another folder
    ├── user.go
    └── tag.go

Problems:
❌ Activity code spread across 3 directories
❌ Tests mixed with implementation
❌ No clear feature boundaries
❌ Hard to test in isolation
```

### AFTER (Proposed - Clean)

```
internal/
├── domain/                       ← Core business logic
│   ├── activity/
│   │   ├── entity.go            ← Business entity
│   │   ├── repository.go        ← Interface (port)
│   │   ├── service.go           ← Domain logic
│   │   └── errors.go            ← Domain errors
│   ├── user/
│   │   ├── entity.go
│   │   ├── repository.go
│   │   └── errors.go
│   └── tag/
│       └── entity.go
│
├── application/                  ← Use cases
│   ├── activity/
│   │   ├── usecases/
│   │   │   ├── create_activity.go
│   │   │   ├── get_activity.go
│   │   │   └── list_activities.go
│   │   └── dto/
│   │       ├── requests.go
│   │       └── responses.go
│   └── stats/
│       └── usecases/
│
├── infrastructure/               ← External implementations
│   ├── persistence/
│   │   └── postgres/
│   │       ├── activity_repository.go
│   │       ├── user_repository.go
│   │       └── connection.go
│   └── logging/
│
└── interfaces/                   ← HTTP/API layer
    └── http/
        ├── activity/
        │   ├── handler.go
        │   └── routes.go
        ├── stats/
        └── middleware/

tests/                            ← ALL tests separate
├── unit/
│   ├── domain/activity/
│   ├── application/activity/
│   └── infrastructure/
├── integration/activity/
├── e2e/activity/
└── benchmark/activity/

Benefits:
✅ All activity code together
✅ Clear layer separation
✅ Tests organized by type
✅ Easy to navigate
```

---

## Code Example Comparison

### Finding Activity Creation Logic

#### BEFORE: Scattered Across Files

```go
// 1. Model definition (internal/models/activity.go)
type Activity struct {
    ID int64
    Title string
    // ...
}

// 2. Repository (internal/repository/activity_repository.go)
func (r *Repo) Create(ctx, tx, activity) error {
    // DB logic mixed with validation
    if activity.Title == "" {
        return errors.New("title required")
    }
    // Insert...
}

// 3. Handler (internal/handlers/activity.go)
func (h *Handler) CreateActivity(w, r) {
    // Parse + validation + business logic + DB + response
    // Everything in one function!
}

Problem: Have to open 3 files to understand activity creation
```

#### AFTER: Organized by Responsibility

```go
// 1. Domain entity (internal/domain/activity/entity.go)
type Activity struct {
    ID int64
    Title string
}

func (a *Activity) Validate() error {
    if a.Title == "" {
        return ErrActivityTitleRequired
    }
    return nil
}

// 2. Use case (internal/application/activity/usecases/create_activity.go)
func (uc *CreateActivityUseCase) Execute(ctx, req) (*Response, error) {
    // 1. Convert DTO to domain
    activity := toEntity(req)

    // 2. Validate business rules
    if err := activity.Validate(); err != nil {
        return nil, err
    }

    // 3. Persist
    return uc.repo.Create(ctx, activity)
}

// 3. Repository (internal/infrastructure/postgres/activity_repository.go)
func (r *Repo) Create(ctx, activity) error {
    // Pure DB logic, no validation
    return r.db.QueryRow(...)
}

// 4. Handler (internal/interfaces/http/activity/handler.go)
func (h *Handler) CreateActivity(w, r) {
    // Parse request
    var req dto.CreateActivityRequest
    json.Decode(r.Body, &req)

    // Call use case
    resp, err := h.useCase.Execute(r.Context(), req)

    // Return response
    json.Encode(w, resp)
}

Benefit: Each file has ONE clear responsibility
```

---

## Test Organization Comparison

### BEFORE: Tests Everywhere

```
internal/repository/
├── activity_repository.go
├── activity_repository_test.go      ← Unit tests
├── activity_repository_bench_test.go← Benchmarks
├── integration_test.go              ← Integration tests (generic)
└── testhelpers/
    └── container.go

internal/handlers/
├── activity.go
└── activity_test.go                 ← Handler tests

Problems:
❌ Tests mixed with production code
❌ Hard to run only unit tests
❌ Hard to run only integration tests
❌ No clear test categorization
```

### AFTER: Tests by Purpose

```
tests/
├── unit/                            ← Fast, isolated tests
│   ├── domain/
│   │   └── activity/
│   │       ├── entity_test.go       ← Business logic tests
│   │       └── service_test.go
│   ├── application/
│   │   └── activity/
│   │       └── create_activity_test.go ← Use case tests (with mocks)
│   └── infrastructure/
│       └── postgres/
│           └── activity_repository_test.go ← DB tests
│
├── integration/                     ← Multiple components
│   └── activity/
│       ├── create_activity_integration_test.go
│       └── activity_with_tags_test.go
│
├── e2e/                            ← Full API tests
│   └── activity/
│       └── create_activity_api_test.go
│
├── benchmark/                       ← Performance tests
│   └── activity/
│       └── repository_benchmark_test.go
│
└── testhelpers/                     ← Shared test utilities
    ├── database/
    ├── mocks/
    ├── builders/
    └── assertions/

Benefits:
✅ Clear test categorization
✅ Run fast tests during dev: `make test-unit`
✅ Run slow tests in CI: `make test-integration`
✅ Tests don't pollute production code
```

---

## Running Tests Comparison

### BEFORE: Limited Test Control

```bash
# Run all tests (slow, runs everything)
go test ./...

# Run tests for repository
go test ./internal/repository/...

# No way to run only:
# - Unit tests
# - Integration tests
# - Benchmarks separately

Problem: Always run slow tests, even for quick feedback
```

### AFTER: Granular Test Control

```bash
# Fast unit tests (< 1 second)
make test-unit
go test ./tests/unit/... -short

# Integration tests only (with DB)
make test-integration
go test ./tests/integration/...

# E2E tests only
make test-e2e
go test ./tests/e2e/...

# Benchmarks only
make test-bench
go test ./tests/benchmark/... -bench=.

# All tests for one feature
make test-activity
go test ./tests/.../activity/...

# Domain layer only
make test-domain
go test ./tests/unit/domain/...

Benefit: Run exactly what you need!
```

---

## Real Workflow Comparison

### Scenario: Adding a New Activity Feature

#### BEFORE: Touch Many Files

```
1. Update internal/models/activity.go         (entity)
2. Update internal/repository/activity_*.go   (DB logic)
3. Update internal/handlers/activity.go       (HTTP logic)
4. Add tests in internal/repository/*_test.go
5. Add tests in internal/handlers/*_test.go

Result: Touched 5+ files across 3 directories
```

#### AFTER: Feature-Focused

```
1. Update internal/domain/activity/entity.go  (business rules)
2. Add internal/application/activity/usecases/new_feature.go
3. Update internal/infrastructure/postgres/activity_repository.go
4. Update internal/interfaces/http/activity/handler.go
5. Add tests in tests/unit/domain/activity/
6. Add tests in tests/integration/activity/

Result: All changes in related locations, clear separation
```

---

## Dependency Flow Comparison

### BEFORE: Circular Dependencies Risk

```
handlers ←→ repository ←→ models

Problem: Easy to create circular dependencies
```

### AFTER: Clean One-Way Dependencies

```
Interfaces (HTTP)
    ↓
Application (Use Cases)
    ↓
Domain (Entities & Interfaces)
    ↑
Infrastructure (Implementations)

Rule: Inner layers never depend on outer layers
```

---

## Navigation Comparison

### Finding Activity Code

#### BEFORE

```
"Where is activity creation logic?"
→ Check internal/handlers/activity.go
→ Check internal/repository/activity_repository.go
→ Check internal/models/activity.go
→ Open 3 files minimum

Time: ~2-3 minutes
```

#### AFTER

```
"Where is activity creation logic?"
→ Domain: internal/domain/activity/entity.go
→ Use case: internal/application/activity/usecases/create_activity.go
→ Clear separation, obvious location

Time: ~30 seconds
```

### Finding Tests

#### BEFORE

```
"Where are activity tests?"
→ internal/repository/activity_repository_test.go?
→ internal/handlers/activity_test.go?
→ What about integration tests?
→ Scattered across multiple locations

Confusion: Which test is which?
```

#### AFTER

```
"Where are activity tests?"
→ Unit tests: tests/unit/domain/activity/
→ Integration: tests/integration/activity/
→ E2E: tests/e2e/activity/
→ Benchmarks: tests/benchmark/activity/

Clear: Test type is obvious from path
```

---

## Summary Table

| Aspect | Before (Layer-Based) | After (Feature-Based DDD) |
|--------|---------------------|---------------------------|
| **Organization** | By technical layer | By feature/domain |
| **Finding code** | Search 3+ directories | One feature folder |
| **Dependencies** | Unclear, risk of cycles | Clear, one-way flow |
| **Tests** | Mixed with code | Separated by type |
| **Running tests** | All or nothing | Granular control |
| **Onboarding** | Confusing | Self-explanatory |
| **Changes** | Touch many files | Focused on feature |
| **Testability** | Hard to isolate | Easy to mock/test |
| **Scalability** | Gets messier | Scales linearly |

---

## Quick Decision Guide

**Stay with current structure if:**
- ❌ Project is very small (< 5 files)
- ❌ No time for migration
- ❌ Team unfamiliar with DDD

**Migrate to new structure if:**
- ✅ Project growing (10+ files already)
- ✅ Adding new features regularly
- ✅ Multiple developers
- ✅ Want better testability
- ✅ Want to follow Go best practices
- ✅ Inspired by kuja_user_ms structure

---

## Migration Path

```
Week 1: Create structure + Migrate Activity feature
        ↓
Week 2: Migrate User + Stats features
        ↓
Week 3: Migrate remaining features + Update tests
        ↓
Week 4: Remove old structure + Update CI/CD
        ↓
Done: Clean, maintainable codebase!
```

---

**Key Takeaway:** The new structure mirrors how you think about features, not technical layers. Just like `kuja_user_ms` does for TypeScript/NestJS!
