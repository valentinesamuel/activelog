# MONTH 3: Advanced Database & Testing

**Weeks:** 9-12
**Phase:** Database Mastery & Quality Assurance
**Theme:** Build confidence through testing and advanced database patterns

---

## Overview

This month focuses on two critical pillars of backend development: advanced database operations and comprehensive testing. You'll learn how to handle complex data relationships, write efficient queries, and achieve 70%+ test coverage. By the end, you'll be confident writing production-grade database code and tests.

---

## Learning Path

### Week 9: Database Transactions + N+1 Query Problem (30 min)
- ACID transaction guarantees
- BEGIN, COMMIT, ROLLBACK
- Transaction isolation levels
- **NEW:** N+1 query problem detection and solutions

### Week 10: Complex Queries + Joins + Graceful Shutdown (45 min)
- SQL JOINs (INNER, LEFT, RIGHT)
- Aggregate functions (COUNT, SUM, AVG)
- GROUP BY and HAVING
- **NEW:** Graceful shutdown with signal handling

### Week 11: Table-Driven Tests + Mocking + Mock Generation Tools (30 min)
- Table-driven test pattern (Go idiom)
- Testing with testify/assert
- Mock repositories for handlers
- **NEW:** Mock generation with mockgen/gomock

### Week 12: Benchmarking + Optimization + Testcontainers (45 min)
- Benchmark functions in Go
- Profiling with pprof
- Query optimization techniques
- **NEW:** Integration testing with testcontainers

---

## Key Concepts

- **ACID transactions**
  - Atomicity: All or nothing
  - Consistency: Valid state transitions
  - Isolation: Concurrent safety
  - Durability: Permanent once committed

- **Many-to-many relationships**
  - Junction tables
  - Tag system implementation
  - Efficient querying patterns

- 🔴 **N+1 query problem detection and solutions**
  - Identifying N+1 queries
  - Using JOINs to solve
  - Eager loading vs lazy loading
  - Query performance monitoring

- **Table-driven test pattern**
  - Go testing idiom
  - Reduce test code duplication
  - Improve test readability

- **Mock repositories**
  - Isolate handler logic
  - Test without database
  - Predictable test data

- 🔴 **Graceful shutdown with signal handling**
  - SIGTERM/SIGINT handling
  - Connection cleanup
  - Request draining
  - Timeout management

- 🔴 **Mock generation (mockgen, gomock)**
  - Auto-generate mocks from interfaces
  - Reduce boilerplate
  - Type-safe mocking

- **Query profiling**
  - EXPLAIN ANALYZE
  - Slow query logs
  - Index optimization

- 🔴 **Integration testing with testcontainers**
  - Real database in tests
  - Docker-based test isolation
  - Reproducible test environments

---

## Database Additions

```sql
-- Tags system (many-to-many relationship)
CREATE TABLE tags (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE activity_tags (
    activity_id INTEGER REFERENCES activities(id) ON DELETE CASCADE,
    tag_id INTEGER REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (activity_id, tag_id)
);

-- Indexes for performance
CREATE INDEX idx_activities_user_date ON activities(user_id, activity_date);
CREATE INDEX idx_activities_type ON activities(activity_type);
CREATE INDEX idx_activity_tags_activity ON activity_tags(activity_id);
CREATE INDEX idx_activity_tags_tag ON activity_tags(tag_id);
CREATE INDEX idx_tags_name ON tags(name);
```

---

## Testing Goals

- ✅ **70%+ code coverage** (run `go test -cover ./...`)
- ✅ **Table-driven tests for all repos** (consistent pattern)
- ✅ **Mock testing for handlers** (isolate business logic)
- ✅ **Benchmark critical paths** (identify bottlenecks)
- ✅ **Integration tests with real database** (testcontainers)

---

## Code Examples

### Database Transactions
```go
func (r *ActivityRepository) CreateWithTags(ctx context.Context, activity *models.Activity, tags []string) error {
    // Start transaction
    tx, err := r.db.BeginTx(ctx, nil)
    if err != nil {
        return fmt.Errorf("begin transaction: %w", err)
    }
    defer tx.Rollback() // Rollback if not committed

    // Insert activity
    err = tx.QueryRowContext(ctx, `
        INSERT INTO activities (user_id, activity_type, duration_minutes, distance_km, notes, activity_date)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id, created_at, updated_at
    `, activity.UserID, activity.Type, activity.Duration, activity.Distance, activity.Notes, activity.Date).
        Scan(&activity.ID, &activity.CreatedAt, &activity.UpdatedAt)
    if err != nil {
        return fmt.Errorf("insert activity: %w", err)
    }

    // Insert tags
    for _, tagName := range tags {
        var tagID int

        // Get or create tag
        err = tx.QueryRowContext(ctx, `
            INSERT INTO tags (name) VALUES ($1)
            ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
            RETURNING id
        `, tagName).Scan(&tagID)
        if err != nil {
            return fmt.Errorf("insert tag: %w", err)
        }

        // Link activity to tag
        _, err = tx.ExecContext(ctx, `
            INSERT INTO activity_tags (activity_id, tag_id) VALUES ($1, $2)
        `, activity.ID, tagID)
        if err != nil {
            return fmt.Errorf("link tag: %w", err)
        }
    }

    // Commit transaction
    if err = tx.Commit(); err != nil {
        return fmt.Errorf("commit transaction: %w", err)
    }

    return nil
}
```

### 🔴 N+1 Query Problem Solution
```go
// ❌ BAD: N+1 Query Problem
func (r *ActivityRepository) GetActivitiesWithTags_BAD(ctx context.Context, userID int) ([]*models.Activity, error) {
    // 1 query to get activities
    activities, err := r.GetByUserID(ctx, userID)
    if err != nil {
        return nil, err
    }

    // N queries (one per activity) to get tags
    for _, activity := range activities {
        tags, err := r.GetTagsForActivity(ctx, activity.ID)
        if err != nil {
            return nil, err
        }
        activity.Tags = tags
    }

    return activities, nil
}

// ✅ GOOD: Single JOIN Query
func (r *ActivityRepository) GetActivitiesWithTags(ctx context.Context, userID int) ([]*models.Activity, error) {
    query := `
        SELECT
            a.id, a.user_id, a.activity_type, a.duration_minutes, a.distance_km,
            a.notes, a.activity_date, a.created_at, a.updated_at,
            t.id as tag_id, t.name as tag_name
        FROM activities a
        LEFT JOIN activity_tags at ON a.id = at.activity_id
        LEFT JOIN tags t ON at.tag_id = t.id
        WHERE a.user_id = $1
        ORDER BY a.activity_date DESC
    `

    rows, err := r.db.QueryContext(ctx, query, userID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    // Map to handle multiple tags per activity
    activityMap := make(map[int]*models.Activity)

    for rows.Next() {
        var (
            a                  models.Activity
            tagID              sql.NullInt64
            tagName            sql.NullString
        )

        err := rows.Scan(
            &a.ID, &a.UserID, &a.Type, &a.Duration, &a.Distance,
            &a.Notes, &a.Date, &a.CreatedAt, &a.UpdatedAt,
            &tagID, &tagName,
        )
        if err != nil {
            return nil, err
        }

        // Check if activity already in map
        if _, exists := activityMap[a.ID]; !exists {
            activityMap[a.ID] = &a
            activityMap[a.ID].Tags = []string{}
        }

        // Add tag if present
        if tagID.Valid && tagName.Valid {
            activityMap[a.ID].Tags = append(activityMap[a.ID].Tags, tagName.String)
        }
    }

    // Convert map to slice
    activities := make([]*models.Activity, 0, len(activityMap))
    for _, activity := range activityMap {
        activities = append(activities, activity)
    }

    return activities, nil
}
```

### 🔴 Graceful Shutdown
```go
func main() {
    // ... setup code ...

    srv := &http.Server{
        Addr:    ":8080",
        Handler: router,
    }

    // Channel to listen for interrupt signals
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

    // Start server in goroutine
    go func() {
        log.Println("Server starting on :8080")
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("Server error: %v", err)
        }
    }()

    // Wait for interrupt signal
    <-quit
    log.Println("Shutting down server...")

    // Create shutdown context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    // Attempt graceful shutdown
    if err := srv.Shutdown(ctx); err != nil {
        log.Printf("Server forced to shutdown: %v", err)
    }

    // Close database connections
    if err := db.Close(); err != nil {
        log.Printf("Error closing database: %v", err)
    }

    log.Println("Server exited")
}
```

### Table-Driven Tests
```go
func TestActivityRepository_Create(t *testing.T) {
    tests := []struct {
        name    string
        activity models.Activity
        wantErr bool
        errType error
    }{
        {
            name: "valid activity",
            activity: models.Activity{
                UserID:   1,
                Type:     "running",
                Duration: 30,
                Distance: 5.0,
                Date:     time.Now(),
            },
            wantErr: false,
        },
        {
            name: "missing user_id",
            activity: models.Activity{
                Type:     "running",
                Duration: 30,
            },
            wantErr: true,
            errType: sql.ErrNoRows,
        },
        {
            name: "negative duration",
            activity: models.Activity{
                UserID:   1,
                Type:     "running",
                Duration: -10,
            },
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := repo.Create(context.Background(), &tt.activity)

            if tt.wantErr {
                assert.Error(t, err)
                if tt.errType != nil {
                    assert.ErrorIs(t, err, tt.errType)
                }
            } else {
                assert.NoError(t, err)
                assert.NotZero(t, tt.activity.ID)
            }
        })
    }
}
```

### 🔴 Mock Generation with gomock
```go
// Generate mocks: go generate ./...
//go:generate mockgen -source=repository.go -destination=mocks/repository_mock.go -package=mocks

// Use in tests
func TestActivityHandler_Create(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockRepo := mocks.NewMockActivityRepository(ctrl)
    handler := NewActivityHandler(mockRepo)

    // Set expectations
    mockRepo.EXPECT().
        Create(gomock.Any(), gomock.Any()).
        Return(nil).
        Times(1)

    // Test handler
    req := httptest.NewRequest("POST", "/activities", body)
    w := httptest.NewRecorder()

    handler.Create(w, req)

    assert.Equal(t, http.StatusCreated, w.Code)
}
```

### 🔴 Integration Tests with Testcontainers
```go
import (
    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/wait"
)

func setupTestDB(t *testing.T) (*sql.DB, func()) {
    ctx := context.Background()

    req := testcontainers.ContainerRequest{
        Image:        "postgres:15",
        ExposedPorts: []string{"5432/tcp"},
        Env: map[string]string{
            "POSTGRES_DB":       "testdb",
            "POSTGRES_USER":     "test",
            "POSTGRES_PASSWORD": "test",
        },
        WaitingFor: wait.ForLog("database system is ready to accept connections"),
    }

    container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: req,
        Started:          true,
    })
    require.NoError(t, err)

    host, _ := container.Host(ctx)
    port, _ := container.MappedPort(ctx, "5432")

    dsn := fmt.Sprintf("postgres://test:test@%s:%s/testdb?sslmode=disable", host, port.Port())
    db, err := sql.Open("postgres", dsn)
    require.NoError(t, err)

    // Run migrations
    runMigrations(t, db)

    // Return cleanup function
    cleanup := func() {
        db.Close()
        container.Terminate(ctx)
    }

    return db, cleanup
}
```

### Benchmarking
```go
func BenchmarkActivityRepository_Create(b *testing.B) {
    // Setup
    db := setupTestDB(b)
    defer db.Close()
    repo := repository.NewActivityRepository(db)

    activity := &models.Activity{
        UserID:   1,
        Type:     "running",
        Duration: 30,
        Distance: 5.0,
        Date:     time.Now(),
    }

    b.ResetTimer() // Don't count setup time

    for i := 0; i < b.N; i++ {
        _ = repo.Create(context.Background(), activity)
    }
}

// Run: go test -bench=. -benchmem
```

---

## Common Pitfalls

1. **Forgetting to rollback transactions**
   - ❌ No defer tx.Rollback()
   - ✅ Always defer rollback after BeginTx

2. **N+1 queries in production**
   - ❌ Loading related data in loops
   - ✅ Use JOINs or batch loading

3. **Not testing error paths**
   - ❌ Only testing happy path
   - ✅ Test all error conditions

4. **Ignoring database indexes**
   - ❌ Slow queries on unindexed columns
   - ✅ Index foreign keys and common WHERE clauses

5. **Not handling shutdown gracefully**
   - ❌ Killing server mid-request
   - ✅ Drain connections before shutdown

---

## Testing Checklist

- [ ] Repository layer: 100% coverage
- [ ] Handler layer: 80%+ coverage with mocks
- [ ] Integration tests with testcontainers
- [ ] Benchmark critical operations
- [ ] Test transaction rollback scenarios
- [ ] Test concurrent access patterns
- [ ] Verify N+1 queries are eliminated

---

## Resources

- [Go Testing Documentation](https://golang.org/pkg/testing/)
- [testify/assert](https://pkg.go.dev/github.com/stretchr/testify/assert)
- [gomock](https://github.com/golang/mock)
- [testcontainers-go](https://golang.testcontainers.org/)
- [PostgreSQL Transactions](https://www.postgresql.org/docs/current/tutorial-transactions.html)

---

## Next Steps

After completing Month 3, you'll move to **Month 4: File Uploads & Cloud Storage**, where you'll learn:
- Local file uploads
- AWS S3 integration
- Image processing
- OpenAPI/Swagger documentation

**You now have a robust, well-tested database layer!** ✅
