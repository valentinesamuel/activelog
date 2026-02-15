# SetupTestDB Usage Guide

## Purpose
`SetupTestDB` creates a real PostgreSQL database running in a Docker container for integration tests. Each test gets a fresh, isolated database with all migrations applied.

## Function Signature

```go
func SetupTestDB(t *testing.T) (*sql.DB, func())
```

**Returns:**
1. `*sql.DB` - Connection to the test database (fully migrated and ready to use)
2. `func()` - Cleanup function that **must** be called with `defer cleanup()` to stop the container and close the connection

## Usage Example

```go
package repository_test

import (
    "testing"
    "github.com/valentinesamuel/activelog/internal/repository"
    "github.com/valentinesamuel/activelog/internal/repository/testhelpers"
)

func TestActivityRepository_Create(t *testing.T) {
    // Setup: Create test database with migrations
    db, cleanup := testhelpers.SetupTestDB(t)
    defer cleanup()  // CRITICAL: Always call cleanup!

    // Create repository instance
    repo := repository.NewActivityRepository(db)

    // Test your repository methods
    activity := &repository.Activity{
        UserID:           1,
        ActivityType:     "running",
        Title:            "Morning Run",
        DurationMinutes:  30,
        DistanceKm:       5.0,
    }

    err := repo.Create(activity)
    if err != nil {
        t.Fatalf("Failed to create activity: %v", err)
    }

    // Verify it was created
    retrieved, err := repo.GetByID(activity.ID)
    if err != nil {
        t.Fatalf("Failed to retrieve activity: %v", err)
    }

    if retrieved.Title != "Morning Run" {
        t.Errorf("Expected title 'Morning Run', got '%s'", retrieved.Title)
    }
}
```

## What It Does

1. **Starts a PostgreSQL container** using testcontainers
2. **Waits for database** to be ready to accept connections
3. **Runs all migrations** from the `migrations/` folder (in order: `000001_*.up.sql`, `000002_*.up.sql`, etc.)
4. **Returns connection** ready for testing
5. **Cleanup function** stops the container and closes the database connection

## Important Notes

### ⚠️ Always Call Cleanup
```go
db, cleanup := testhelpers.SetupTestDB(t)
defer cleanup()  // MUST call this or container will keep running!
```

### ✅ Fresh Database Per Test
Each call to `SetupTestDB` creates a new, isolated database container. No shared state between tests.

### 🐳 Requires Docker
- Docker daemon must be running
- Network connectivity to pull images (first time only)
- Images used: `postgres:latest`, `testcontainers/ryuk:0.13.0`

### 📊 Schema is Fully Migrated
The database includes all tables from your migrations:
- `users` table (with password_hash column)
- `activities` table
- `tags` table
- `activity_tags` junction table
- All indexes

### 🚀 Use for Integration Tests
This is for **integration tests** that test against a real database. For unit tests, use mocks.

```go
// Integration test - uses real database
func TestActivityRepository_Integration(t *testing.T) {
    db, cleanup := testhelpers.SetupTestDB(t)
    defer cleanup()
    // ... test with real database ...
}

// Unit test - uses mock
func TestActivityService_Unit(t *testing.T) {
    mockRepo := &mocks.MockActivityRepository{}
    // ... test with mock ...
}
```

## Troubleshooting

### Container fails to start
- Ensure Docker is running: `docker ps`
- Check Docker has enough resources (memory, disk)

### Network timeout errors
- Check internet connectivity
- Pull images manually first:
  ```bash
  docker pull postgres:latest
  docker pull testcontainers/ryuk:0.13.0
  ```

### Migration errors
- Verify migration files exist in `migrations/` folder
- Check migration SQL syntax
- Ensure migrations are numbered sequentially (000001, 000002, etc.)

### Port conflicts
- Testcontainers uses random ports automatically
- No conflicts with local postgres on port 5432

## Performance

Integration tests are slower than unit tests:
- Container startup: ~2-5 seconds
- Migration execution: ~100-500ms
- Test execution: varies

**Recommendation:** Run integration tests separately from unit tests:
```bash
# Unit tests only (fast)
go test -short ./...

# Integration tests only (slower)
go test -run Integration ./...
```
