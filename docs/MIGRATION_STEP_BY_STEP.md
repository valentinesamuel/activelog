# Step-by-Step Migration Guide

## Getting Started - Migrate Activity Feature First

This guide walks you through migrating the **Activity** feature as a proof of concept. Once this works, you can apply the same pattern to User, Stats, and Tag features.

---

## Step 1: Create New Directory Structure (15 minutes)

```bash
cd /Users/valentinesamuel/Desktop/projects/go-projects/activelog

# Create domain layer
mkdir -p internal/domain/activity
mkdir -p internal/domain/user
mkdir -p internal/domain/tag
mkdir -p internal/domain/shared

# Create application layer
mkdir -p internal/application/activity/{usecases,dto,ports}
mkdir -p internal/application/stats/{usecases,dto}
mkdir -p internal/application/user/{usecases,dto}

# Create infrastructure layer
mkdir -p internal/infrastructure/persistence/postgres
mkdir -p internal/infrastructure/persistence/memory
mkdir -p internal/infrastructure/logging
mkdir -p internal/infrastructure/config

# Create interfaces layer
mkdir -p internal/interfaces/http/{activity,stats,user,health,middleware}
mkdir -p internal/interfaces/validator

# Create shared layer
mkdir -p internal/shared/{errors,utils,constants}

# Create test structure
mkdir -p tests/{unit,integration,e2e,benchmark,testhelpers,testdata}
mkdir -p tests/unit/{domain,application,infrastructure}
mkdir -p tests/unit/domain/{activity,user,tag}
mkdir -p tests/unit/application/{activity,stats,user}
mkdir -p tests/integration/{activity,stats,user}
mkdir -p tests/e2e/{activity,stats,user}
mkdir -p tests/benchmark/{activity,stats}
mkdir -p tests/testhelpers/{database,mocks,builders,assertions}
mkdir -p tests/testdata/{fixtures,golden}

echo "✅ Directory structure created!"
```

---

## Step 2: Create Domain Layer for Activity (30 minutes)

### 2.1 Create Activity Entity

```bash
# Create the entity file
touch internal/domain/activity/entity.go
```

Add this content:

```go
// internal/domain/activity/entity.go
package activity

import (
	"errors"
	"time"
)

// Activity represents the core business entity
type Activity struct {
	ID              int64
	UserID          int
	Type            ActivityType
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

// ActivityType is a value object
type ActivityType string

const (
	Running    ActivityType = "running"
	Cycling    ActivityType = "cycling"
	Swimming   ActivityType = "swimming"
	Gym        ActivityType = "gym"
	Walking    ActivityType = "walking"
	JumpRope   ActivityType = "jump_rope"
	Basketball ActivityType = "basketball"
)

// Tag represents activity tags
type Tag struct {
	ID   int
	Name string
}

// Domain errors
var (
	ErrActivityTitleRequired    = errors.New("activity title is required")
	ErrInvalidActivityType      = errors.New("invalid activity type")
	ErrInvalidDuration          = errors.New("duration must be positive")
	ErrActivityNotFound         = errors.New("activity not found")
	ErrUnauthorizedAccess       = errors.New("unauthorized to access this activity")
)

// Validate enforces business rules
func (a *Activity) Validate() error {
	if a.Title == "" {
		return ErrActivityTitleRequired
	}

	if a.DurationMinutes < 0 {
		return ErrInvalidDuration
	}

	if !a.Type.IsValid() {
		return ErrInvalidActivityType
	}

	return nil
}

// IsValid checks if activity type is valid
func (t ActivityType) IsValid() bool {
	validTypes := []ActivityType{
		Running, Cycling, Swimming, Gym,
		Walking, JumpRope, Basketball,
	}

	for _, valid := range validTypes {
		if t == valid {
			return true
		}
	}
	return false
}

// CanBeAccessedBy checks if user can access this activity
func (a *Activity) CanBeAccessedBy(userID int) bool {
	return a.UserID == userID
}
```

### 2.2 Create Repository Interface (Port)

```bash
touch internal/domain/activity/repository.go
```

```go
// internal/domain/activity/repository.go
package activity

import "context"

// Repository defines the contract for activity persistence
// This is a PORT in hexagonal architecture
type Repository interface {
	Create(ctx context.Context, activity *Activity) error
	GetByID(ctx context.Context, id int64) (*Activity, error)
	ListByUser(ctx context.Context, userID int, filters Filters) ([]*Activity, error)
	Update(ctx context.Context, activity *Activity) error
	Delete(ctx context.Context, id int64) error
	CreateWithTags(ctx context.Context, activity *Activity, tags []Tag) error
	GetActivitiesWithTags(ctx context.Context, userID int, filters Filters) ([]*Activity, error)
}

// Filters for querying activities
type Filters struct {
	ActivityType ActivityType
	StartDate    *time.Time
	EndDate      *time.Time
	Limit        int
	Offset       int
}
```

### 2.3 Create Domain Service (optional, for complex logic)

```bash
touch internal/domain/activity/service.go
```

```go
// internal/domain/activity/service.go
package activity

// Service handles complex domain logic that doesn't fit in entities
type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// CalculateCalories is an example of domain logic
func (s *Service) CalculateCalories(activity *Activity) int {
	// Complex business rules for calorie calculation
	// This is just an example - implement your actual logic
	baseRate := map[ActivityType]float64{
		Running:    10.0,
		Cycling:    8.0,
		Swimming:   12.0,
		Gym:        6.0,
		Walking:    4.0,
		JumpRope:   11.0,
		Basketball: 9.0,
	}

	rate, exists := baseRate[activity.Type]
	if !exists {
		rate = 5.0 // default
	}

	return int(float64(activity.DurationMinutes) * rate)
}
```

---

## Step 3: Move Repository Implementation to Infrastructure (20 minutes)

### 3.1 Copy existing repository to new location

```bash
# Copy the implementation
cp internal/repository/activity_repository.go \
   internal/infrastructure/persistence/postgres/activity_repository.go

# Update package name in the file
sed -i '' 's/^package repository$/package postgres/' \
   internal/infrastructure/persistence/postgres/activity_repository.go
```

### 3.2 Update imports and make it implement the interface

Edit `internal/infrastructure/persistence/postgres/activity_repository.go`:

```go
package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/valentinesamuel/activelog/internal/domain/activity"
	"github.com/valentinesamuel/activelog/internal/database"
)

type ActivityRepository struct {
	db      *database.LoggingDB
	tagRepo *TagRepository
}

func NewActivityRepository(db *database.LoggingDB, tagRepo *TagRepository) *ActivityRepository {
	return &ActivityRepository{
		db:      db,
		tagRepo: tagRepo,
	}
}

// Create implements activity.Repository
func (ar *ActivityRepository) Create(ctx context.Context, act *activity.Activity) error {
	query := `
		INSERT INTO activities
		(user_id, activity_type, title, description, duration_minutes, distance_km, calories_burned, notes, activity_date)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at
	`

	row := ar.db.QueryRowContext(ctx, query,
		act.UserID,
		act.Type,
		act.Title,
		act.Description,
		act.DurationMinutes,
		act.DistanceKm,
		act.CaloriesBurned,
		act.Notes,
		act.ActivityDate,
	)

	return row.Scan(&act.ID, &act.CreatedAt, &act.UpdatedAt)
}

// ... rest of the implementation
```

---

## Step 4: Create Application Layer - Use Cases (30 minutes)

### 4.1 Create DTOs

```bash
touch internal/application/activity/dto/requests.go
touch internal/application/activity/dto/responses.go
```

```go
// internal/application/activity/dto/requests.go
package dto

import "time"

type CreateActivityRequest struct {
	UserID          int       `json:"user_id" validate:"required"`
	Type            string    `json:"type" validate:"required"`
	Title           string    `json:"title" validate:"required"`
	Description     string    `json:"description"`
	DurationMinutes int       `json:"duration_minutes" validate:"required,min=1"`
	DistanceKm      float64   `json:"distance_km"`
	CaloriesBurned  int       `json:"calories_burned"`
	Notes           string    `json:"notes"`
	ActivityDate    time.Time `json:"activity_date" validate:"required"`
}

type UpdateActivityRequest struct {
	ID              int64     `json:"id" validate:"required"`
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	DurationMinutes int       `json:"duration_minutes"`
	DistanceKm      float64   `json:"distance_km"`
	CaloriesBurned  int       `json:"calories_burned"`
	Notes           string    `json:"notes"`
}

type ListActivitiesRequest struct {
	UserID       int        `json:"user_id" validate:"required"`
	ActivityType string     `json:"activity_type"`
	StartDate    *time.Time `json:"start_date"`
	EndDate      *time.Time `json:"end_date"`
	Limit        int        `json:"limit" validate:"min=1,max=100"`
	Offset       int        `json:"offset" validate:"min=0"`
}
```

```go
// internal/application/activity/dto/responses.go
package dto

import "time"

type ActivityResponse struct {
	ID              int64     `json:"id"`
	UserID          int       `json:"user_id"`
	Type            string    `json:"type"`
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	DurationMinutes int       `json:"duration_minutes"`
	DistanceKm      float64   `json:"distance_km"`
	CaloriesBurned  int       `json:"calories_burned"`
	Notes           string    `json:"notes"`
	ActivityDate    time.Time `json:"activity_date"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type ListActivitiesResponse struct {
	Activities []*ActivityResponse `json:"activities"`
	Total      int                 `json:"total"`
	Limit      int                 `json:"limit"`
	Offset     int                 `json:"offset"`
}
```

### 4.2 Create Use Cases

```bash
touch internal/application/activity/usecases/create_activity.go
```

```go
// internal/application/activity/usecases/create_activity.go
package usecases

import (
	"context"

	"github.com/valentinesamuel/activelog/internal/application/activity/dto"
	"github.com/valentinesamuel/activelog/internal/domain/activity"
)

type CreateActivityUseCase struct {
	activityRepo activity.Repository
}

func NewCreateActivityUseCase(activityRepo activity.Repository) *CreateActivityUseCase {
	return &CreateActivityUseCase{
		activityRepo: activityRepo,
	}
}

func (uc *CreateActivityUseCase) Execute(
	ctx context.Context,
	req dto.CreateActivityRequest,
) (*dto.ActivityResponse, error) {
	// 1. Convert DTO to domain entity
	act := &activity.Activity{
		UserID:          req.UserID,
		Type:            activity.ActivityType(req.Type),
		Title:           req.Title,
		Description:     req.Description,
		DurationMinutes: req.DurationMinutes,
		DistanceKm:      req.DistanceKm,
		CaloriesBurned:  req.CaloriesBurned,
		Notes:           req.Notes,
		ActivityDate:    req.ActivityDate,
	}

	// 2. Validate business rules
	if err := act.Validate(); err != nil {
		return nil, err
	}

	// 3. Persist
	if err := uc.activityRepo.Create(ctx, act); err != nil {
		return nil, err
	}

	// 4. Convert to response DTO
	return &dto.ActivityResponse{
		ID:              act.ID,
		UserID:          act.UserID,
		Type:            string(act.Type),
		Title:           act.Title,
		Description:     act.Description,
		DurationMinutes: act.DurationMinutes,
		DistanceKm:      act.DistanceKm,
		CaloriesBurned:  act.CaloriesBurned,
		Notes:           act.Notes,
		ActivityDate:    act.ActivityDate,
		CreatedAt:       act.CreatedAt,
		UpdatedAt:       act.UpdatedAt,
	}, nil
}
```

---

## Step 5: Move Tests to New Structure (45 minutes)

### 5.1 Move Unit Tests

```bash
# Create unit test for domain entity
touch tests/unit/domain/activity/entity_test.go
```

```go
// tests/unit/domain/activity/entity_test.go
package activity_test

import (
	"testing"

	"github.com/valentinesamuel/activelog/internal/domain/activity"
)

func TestActivity_Validate(t *testing.T) {
	tests := []struct {
		name     string
		activity *activity.Activity
		wantErr  error
	}{
		{
			name: "valid activity",
			activity: &activity.Activity{
				Title:           "Morning Run",
				Type:            activity.Running,
				DurationMinutes: 30,
			},
			wantErr: nil,
		},
		{
			name: "missing title",
			activity: &activity.Activity{
				Title:           "",
				Type:            activity.Running,
				DurationMinutes: 30,
			},
			wantErr: activity.ErrActivityTitleRequired,
		},
		{
			name: "invalid duration",
			activity: &activity.Activity{
				Title:           "Test",
				Type:            activity.Running,
				DurationMinutes: -5,
			},
			wantErr: activity.ErrInvalidDuration,
		},
		{
			name: "invalid type",
			activity: &activity.Activity{
				Title:           "Test",
				Type:            activity.ActivityType("invalid"),
				DurationMinutes: 30,
			},
			wantErr: activity.ErrInvalidActivityType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.activity.Validate()
			if err != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestActivityType_IsValid(t *testing.T) {
	tests := []struct {
		name         string
		activityType activity.ActivityType
		want         bool
	}{
		{"valid running", activity.Running, true},
		{"valid cycling", activity.Cycling, true},
		{"invalid type", activity.ActivityType("invalid"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.activityType.IsValid(); got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}
```

### 5.2 Move Integration Tests

```bash
# Move integration tests
touch tests/integration/activity/create_activity_test.go
```

### 5.3 Create Test Helpers

```bash
# Move database helpers
cp internal/repository/testhelpers/container.go tests/testhelpers/database/
```

### 5.4 Create Test Data Builders

```bash
touch tests/testhelpers/builders/activity_builder.go
```

```go
// tests/testhelpers/builders/activity_builder.go
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
			UserID:          1,
			Type:            activity.Running,
			Title:           "Default Activity",
			Description:     "Default description",
			DurationMinutes: 30,
			DistanceKm:      5.0,
			CaloriesBurned:  250,
			ActivityDate:    time.Now(),
		},
	}
}

func (b *ActivityBuilder) WithUserID(userID int) *ActivityBuilder {
	b.activity.UserID = userID
	return b
}

func (b *ActivityBuilder) WithTitle(title string) *ActivityBuilder {
	b.activity.Title = title
	return b
}

func (b *ActivityBuilder) WithType(actType activity.ActivityType) *ActivityBuilder {
	b.activity.Type = actType
	return b
}

func (b *ActivityBuilder) WithDuration(minutes int) *ActivityBuilder {
	b.activity.DurationMinutes = minutes
	return b
}

func (b *ActivityBuilder) Build() *activity.Activity {
	return b.activity
}

// Usage example:
// activity := builders.NewActivityBuilder().
//     WithTitle("Morning Run").
//     WithUserID(123).
//     WithDuration(45).
//     Build()
```

---

## Step 6: Update Makefile (10 minutes)

Add these commands to your Makefile:

```makefile
# Test commands by type
.PHONY: test-unit test-integration test-e2e test-domain test-activity

## test-unit: Run unit tests (fast, no DB)
test-unit:
	@echo "Running unit tests..."
	go test ./tests/unit/... -v -short

## test-integration: Run integration tests (with DB)
test-integration:
	@echo "Running integration tests..."
	go test ./tests/integration/... -v

## test-e2e: Run end-to-end tests
test-e2e:
	@echo "Running E2E tests..."
	go test ./tests/e2e/... -v

## test-domain: Run domain layer tests only
test-domain:
	@echo "Running domain tests..."
	go test ./tests/unit/domain/... -v

## test-activity: Run all activity-related tests
test-activity:
	@echo "Running activity tests..."
	go test ./tests/unit/domain/activity/... -v
	go test ./tests/unit/application/activity/... -v
	go test ./tests/integration/activity/... -v

## test-coverage-new: Coverage for new structure
test-coverage-new:
	@echo "Running coverage for new structure..."
	go test ./tests/... -coverprofile=coverage_new.out
	go tool cover -html=coverage_new.out
```

---

## Step 7: Verify Everything Works (20 minutes)

```bash
# 1. Run domain tests
make test-domain

# 2. Run unit tests
make test-unit

# 3. Run all activity tests
make test-activity

# 4. Run full test suite
go test ./... -v
```

---

## Step 8: Update Documentation (15 minutes)

Update your README.md to reflect the new structure:

```markdown
## Project Structure

This project follows Clean Architecture and Domain-Driven Design principles:

### Domain Layer (`internal/domain/`)
Core business logic with no external dependencies:
- `activity/` - Activity domain (entities, repository interfaces, business rules)
- `user/` - User domain
- `tag/` - Tag domain

### Application Layer (`internal/application/`)
Use cases and application logic:
- `activity/usecases/` - Activity use cases
- `activity/dto/` - Data transfer objects

### Infrastructure Layer (`internal/infrastructure/`)
External adapters and implementations:
- `persistence/postgres/` - Database implementations
- `logging/` - Logging implementations

### Interface Layer (`internal/interfaces/`)
HTTP handlers and API:
- `http/activity/` - Activity HTTP handlers

### Tests (`tests/`)
All tests organized by type:
- `unit/` - Fast, isolated tests
- `integration/` - Database integration tests
- `e2e/` - End-to-end API tests
- `benchmark/` - Performance benchmarks
```

---

## Quick Comparison

### Before Migration
```bash
# Finding activity code
internal/handlers/activity.go        # HTTP logic
internal/repository/activity_*.go    # DB logic
internal/models/activity.go          # Data model
internal/repository/*_test.go        # Tests mixed with code
```

### After Migration
```bash
# Everything organized by feature and layer
internal/domain/activity/             # Business logic
internal/application/activity/        # Use cases
internal/infrastructure/postgres/     # DB implementation
tests/unit/domain/activity/           # Domain tests
tests/integration/activity/           # Integration tests
```

---

## Validation Checklist

After migration, verify:

- [ ] All tests pass: `go test ./...`
- [ ] Unit tests run fast: `make test-unit` (< 1 second)
- [ ] Integration tests work: `make test-integration`
- [ ] Benchmarks run: `make bench`
- [ ] Build succeeds: `make build`
- [ ] Server starts: `make run`
- [ ] API endpoints work correctly

---

## Common Issues & Solutions

### Issue 1: Import Cycle Detected

**Problem:**
```
import cycle not allowed
```

**Solution:**
- Domain layer should NOT import application/infrastructure layers
- Check dependency direction: Domain ← Application ← Infrastructure

### Issue 2: Tests Can't Find Packages

**Problem:**
```
cannot find package "github.com/valentinesamuel/activelog/internal/domain/activity"
```

**Solution:**
```bash
# Update go.mod
go mod tidy

# Rebuild
go build ./...
```

### Issue 3: Database Tests Failing

**Problem:**
```
tests failing after migration
```

**Solution:**
- Ensure testhelpers are moved correctly
- Update import paths in test files
- Run: `go test ./tests/testhelpers/... -v` first

---

## Next Steps

Once Activity feature is successfully migrated:

1. **Migrate User feature** (similar pattern)
2. **Migrate Stats feature**
3. **Migrate Tag feature**
4. **Remove old structure** from `internal/repository/`, `internal/handlers/`, `internal/models/`
5. **Update CI/CD** to use new test commands

---

## Getting Help

If you encounter issues:

1. Check existing tests still pass: `go test ./internal/...`
2. Verify new tests pass: `go test ./tests/...`
3. Compare with examples in this guide
4. Review dependency flow: Domain → Application → Infrastructure

---

**Remember:** Migrate one feature at a time, keep tests passing!
