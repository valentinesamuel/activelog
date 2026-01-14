# Integration Tests Guide - Week 12 Task 3

## Overview

This guide explains how to write and run integration tests using **testcontainers** for real database testing. Integration tests verify that your code works correctly with actual database operations, constraints, and transactions.

## What We're Testing

### ✅ Task 3 Requirements (Week 12)

1. **Full transaction flow** - Create activity with tags atomically
2. **Concurrent insertions** - Multiple goroutines inserting data safely
3. **Foreign key constraints** - Database enforces referential integrity
4. **Unique constraint violations** - Database rejects duplicate data
5. **Verify actual database state** - Check data after operations

## Test File Structure

```
internal/repository/integration_test.go
├── Helper Functions (reusable test utilities)
│   ├── createIntegrationTestUser()
│   ├── createIntegrationTestActivity()
│   ├── verifyActivityExists()
│   ├── verifyTagExists()
│   ├── verifyActivityTagLink()
│   └── countActivities()
│
└── Integration Tests
    ├── TestIntegration_CreateActivityWithTags
    ├── TestIntegration_CreateActivityWithTags_RollbackOnError
    ├── TestIntegration_ConcurrentInsertions
    ├── TestIntegration_ForeignKeyConstraint
    ├── TestIntegration_UniqueConstraintViolations
    └── TestIntegration_ComplexQueryWithJoins
```

## Test Breakdown

### 1. Full Transaction Flow

**Test:** `TestIntegration_CreateActivityWithTags`

**What it tests:**
- Creates an activity with 3 tags in a single transaction
- Verifies activity is created in database
- Verifies tags are created in database
- Verifies activity-tag links exist
- Verifies data can be retrieved

**Why it's important:**
- Tests ACID transaction properties
- Ensures all-or-nothing behavior
- Validates multi-table operations

**Example:**
```go
activity := &models.Activity{
    UserID: userID,
    Title: "Morning Run",
}
tags := []*models.Tag{
    {Name: "outdoor"},
    {Name: "cardio"},
}

err := activityRepo.CreateWithTags(ctx, activity, tags)
// Verify: activity exists, tags exist, links exist
```

### 2. Transaction Rollback on Error

**Test:** `TestIntegration_CreateActivityWithTags_RollbackOnError`

**What it tests:**
- Attempts to create activity with invalid tag (too long)
- Verifies transaction fails
- Verifies **NO** data was committed (rollback worked)
- Even valid tag should not exist

**Why it's important:**
- Tests transaction rollback
- Ensures data consistency
- Validates error handling

**Example:**
```go
tags := []*models.Tag{
    {Name: "validtag"},
    {Name: "this_is_way_too_long..."}, // Violates VARCHAR(50)
}

err := activityRepo.CreateWithTags(ctx, activity, tags)
// Should fail, and "validtag" should NOT exist (rolled back)
```

### 3. Concurrent Insertions

**Test:** `TestIntegration_ConcurrentInsertions`

**What it tests:**
- Launches 10 goroutines simultaneously
- Each goroutine inserts an activity
- Verifies all 10 activities were created
- Verifies all IDs are unique (no race conditions)

**Why it's important:**
- Tests thread safety
- Validates database concurrency handling
- Ensures no data corruption

**Example:**
```go
var wg sync.WaitGroup
for i := 0; i < 10; i++ {
    go func(index int) {
        defer wg.Done()
        // Create activity...
    }(i)
}
wg.Wait()
// Verify: 10 activities exist, all unique IDs
```

### 4. Foreign Key Constraints

**Test:** `TestIntegration_ForeignKeyConstraint`

**What it tests:**
- **Test 1:** Cannot create activity for non-existent user
- **Test 2:** Deleting user CASCADE deletes their activities

**Why it's important:**
- Validates referential integrity
- Tests CASCADE DELETE behavior
- Ensures orphan data prevention

**Example:**
```go
// Test 1: Invalid foreign key
activity := &models.Activity{
    UserID: 99999, // Doesn't exist
}
err := repo.Create(ctx, nil, activity)
// Should fail with foreign key violation

// Test 2: CASCADE DELETE
db.Exec("DELETE FROM users WHERE id = $1", userID)
// Activities should be automatically deleted
```

### 5. Unique Constraint Violations

**Test:** `TestIntegration_UniqueConstraintViolations`

**What it tests:**
- **Test 1:** Cannot create user with duplicate email
- **Test 2:** Cannot create user with duplicate username
- **Test 3:** Tags with same name reuse existing tag (ON CONFLICT)

**Why it's important:**
- Validates unique constraints work
- Tests ON CONFLICT DO UPDATE behavior
- Ensures data uniqueness

**Example:**
```go
// Create first user
createIntegrationTestUser(t, db, "test@example.com", "user1")

// Try duplicate email (should fail)
db.QueryRow("INSERT INTO users (email, ...) VALUES ($1, ...)",
    "test@example.com", ...)
// Should fail with unique constraint violation
```

### 6. Complex Queries with JOINs

**Test:** `TestIntegration_ComplexQueryWithJoins`

**What it tests:**
- Creates activities with varying numbers of tags (2, 1, 0)
- Queries activities with LEFT JOINs
- Verifies tag counts are correct for each activity

**Why it's important:**
- Tests complex SQL queries
- Validates JOIN logic
- Ensures N+1 query solution works

**Example:**
```go
// Activity 1: 2 tags
// Activity 2: 1 tag
// Activity 3: 0 tags

activities, _ := repo.GetActivitiesWithTags(ctx, userID, filters)
// Verify each activity has correct number of tags
```

## Helper Functions Explained

**Note:** Helper functions are prefixed with `Integration` to avoid conflicts with existing test helpers in `activity_repository_test.go`.

### `createIntegrationTestUser(t, db, email, username) int`
Creates a user for testing and returns the user ID.

**Usage:**
```go
userID := createIntegrationTestUser(t, db, "test@example.com", "testuser")
```

### `createIntegrationTestActivity(t, repo, userID, type, title) *models.Activity`
Creates an activity with default values.

**Usage:**
```go
activity := createIntegrationTestActivity(t, repo, userID, "running", "Morning Run")
```

### `verifyActivityExists(t, db, activityID) bool`
Checks if an activity exists in the database.

**Usage:**
```go
if !verifyActivityExists(t, db, activity.ID) {
    t.Fatal("Activity doesn't exist")
}
```

### `verifyTagExists(t, db, tagName) (bool, int)`
Checks if a tag exists and returns its ID.

**Usage:**
```go
exists, tagID := verifyTagExists(t, db, "outdoor")
```

### `verifyActivityTagLink(t, db, activityID, tagID) bool`
Checks if activity-tag link exists.

**Usage:**
```go
if !verifyActivityTagLink(t, db, activity.ID, tagID) {
    t.Fatal("Link missing")
}
```

### `countActivities(t, db, userID) int`
Counts total activities for a user.

**Usage:**
```go
count := countActivities(t, db, userID)
```

## Running the Tests

### Run All Integration Tests
```bash
go test -v ./internal/repository -run Integration
```

### Run Specific Integration Test
```bash
go test -v ./internal/repository -run TestIntegration_CreateActivityWithTags
```

### Run with Coverage
```bash
go test -v -coverprofile=coverage.out ./internal/repository -run Integration
go tool cover -html=coverage.out
```

### Run in Parallel (faster)
```bash
go test -v -parallel=4 ./internal/repository -run Integration
```

## Prerequisites

1. **Docker must be running**
   ```bash
   docker ps
   ```

2. **Pre-pull images (optional, speeds up tests)**
   ```bash
   docker pull postgres:latest
   docker pull testcontainers/ryuk:0.13.0
   ```

## How Integration Tests Work

```
1. TestIntegration_XXX starts
   ↓
2. testhelpers.SetupTestDB(t)
   ↓
3. Starts postgres container in Docker
   ↓
4. Runs migrations (creates schema)
   ↓
5. Returns db connection + cleanup function
   ↓
6. Test runs against real database
   ↓
7. defer cleanup() called
   ↓
8. Container stopped and removed
```

## Best Practices

### ✅ DO

- **Always call cleanup**: `defer cleanup()`
- **Use t.Helper()**: Mark helper functions with `t.Helper()`
- **Test actual database state**: Use SQL queries to verify
- **Test error cases**: Not just happy paths
- **Use subtests**: `t.Run("TestName", func(t *testing.T) {...})`

### ❌ DON'T

- **Don't share database between tests**: Each test gets fresh DB
- **Don't skip cleanup**: Will leave containers running
- **Don't assume test order**: Tests may run in parallel
- **Don't hardcode IDs**: Use returned IDs from inserts
- **Don't test implementation details**: Test behavior, not code

## Common Patterns

### Pattern 1: Setup → Execute → Verify

```go
func TestIntegration_Something(t *testing.T) {
    // Setup
    db, cleanup := testhelpers.SetupTestDB(t)
    defer cleanup()
    userID := createIntegrationTestUser(t, db, "test@example.com", "user")

    // Execute
    err := repo.SomeMethod(data)

    // Verify
    if err != nil {
        t.Fatalf("SomeMethod failed: %v", err)
    }
    verifyDataExists(t, db, expectedData)
}
```

### Pattern 2: Test Error Cases

```go
func TestIntegration_ErrorCase(t *testing.T) {
    db, cleanup := testhelpers.SetupTestDB(t)
    defer cleanup()

    // Execute with invalid data
    err := repo.Create(invalidData)

    // Verify error occurred
    if err == nil {
        t.Fatal("Expected error, got nil")
    }

    // Verify database state unchanged
    count := countRecords(t, db)
    if count != 0 {
        t.Errorf("Expected 0 records after error, got %d", count)
    }
}
```

### Pattern 3: Concurrent Operations

```go
func TestIntegration_Concurrent(t *testing.T) {
    db, cleanup := testhelpers.SetupTestDB(t)
    defer cleanup()

    var wg sync.WaitGroup
    errChan := make(chan error, numGoroutines)

    for i := 0; i < numGoroutines; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            if err := doWork(); err != nil {
                errChan <- err
            }
        }()
    }

    wg.Wait()
    close(errChan)

    for err := range errChan {
        t.Errorf("Concurrent operation failed: %v", err)
    }
}
```

## Troubleshooting

### Container startup fails
**Problem:** `Failed to start postgres container`
**Solution:**
- Check Docker is running: `docker ps`
- Check network connectivity
- Pull image manually: `docker pull postgres:latest`

### Migrations not found
**Problem:** `no migration files found`
**Solution:**
- Verify migrations folder exists: `ls migrations/`
- Check migrations have `.up.sql` extension
- Verify working directory in test

### Tests are slow
**Problem:** Tests take 30+ seconds
**Solution:**
- Pre-pull Docker images
- Run tests in parallel: `-parallel=4`
- Use `-short` flag to skip integration tests for quick runs

### Port conflicts
**Problem:** `port already in use`
**Solution:**
- Testcontainers uses random ports automatically
- Kill existing containers: `docker ps -a | grep postgres`

## Performance Notes

Integration tests are **slower** than unit tests:
- Container startup: ~2-5 seconds
- Migration execution: ~100-500ms
- Test execution: varies

**Recommendation:**
```bash
# Fast (unit tests only)
go test -short ./...

# Slow (all tests including integration)
go test ./...

# Integration tests only
go test -run Integration ./...
```

## Summary

Integration tests using testcontainers provide:
- ✅ **Real database testing** (not mocks)
- ✅ **Isolated test environment** (fresh DB per test)
- ✅ **Automatic cleanup** (containers removed after test)
- ✅ **CI/CD compatible** (works in GitHub Actions)
- ✅ **Confidence in production code** (tests actual database behavior)

You've successfully completed Week 12 Task 3! 🎉
