# How to Enable Query Logging in ActiveLog

## Method 1: Application-Level Logging (Recommended for Development)

### Step 1: Update your database initialization in `cmd/api/main.go`

```go
package main

import (
    "database/sql"
    "log"
    "os"

    _ "github.com/lib/pq"
    "github.com/valentinesamuel/activelog/pkg/database"
)

func main() {
    // Connect to database
    db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // 🔥 Enable query logging in development
    var dbConn interface{} = db
    if os.Getenv("ENVIRONMENT") == "development" {
        logger := log.New(os.Stdout, "[SQL] ", log.LstdFlags)
        dbConn = database.NewLoggingDB(db, logger)
    }

    // Pass dbConn to your repositories
    // activityRepo := repository.NewActivityRepository(dbConn.(*database.LoggingDB))

    // ... rest of your app setup
}
```

### Step 2: Update repository to accept the logging wrapper

**Option A: Use interface (recommended)**

```go
// internal/repository/activity_repository.go
package repository

import (
    "context"
    "database/sql"
)

// DBConn interface that both *sql.DB and *database.LoggingDB implement
type DBConn interface {
    QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
    QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
    ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

type ActivityRepository struct {
    db DBConn  // Changed from *sql.DB
}

func NewActivityRepository(db DBConn) *ActivityRepository {
    return &ActivityRepository{db: db}
}

// All your methods stay the same!
func (r *ActivityRepository) GetByID(ctx context.Context, id int) (*Activity, error) {
    // This will be logged automatically if using LoggingDB
    row := r.db.QueryRowContext(ctx, "SELECT * FROM activities WHERE id = $1", id)
    // ...
}
```

**Option B: Type assertion (simpler but less flexible)**

```go
type ActivityRepository struct {
    db *sql.DB
}

// In main.go, just type assert when needed:
loggedDB := database.NewLoggingDB(db, logger)
activityRepo := repository.NewActivityRepository(loggedDB.DB) // Use the underlying DB
```

### Step 3: See the logs!

Run your app and you'll see output like:

```
[SQL] 2024/01/03 15:30:45 ✅ [QUERY ROW] ⚡ 2.5ms | SELECT * FROM activities WHERE id = $1 | args: [1]
[SQL] 2024/01/03 15:30:45 ✅ [EXEC] ⚡ 5.1ms | INSERT INTO activities (user_id, activity_type...) | args: [1 running 30]
[SQL] 2024/01/03 15:30:45 ✅ BEGIN TRANSACTION (took 1.2ms)
[SQL] 2024/01/03 15:30:45 ✅ [TX EXEC] ⚡ 3.8ms | INSERT INTO tags (name) VALUES ($1) | args: [morning]
[SQL] 2024/01/03 15:30:45 ✅ COMMIT (took 2.1ms)
```

---

## Method 2: PostgreSQL-Level Logging (Great for debugging)

### Enable query logging in PostgreSQL config

#### On macOS (Homebrew PostgreSQL):

```bash
# 1. Find your postgresql.conf
psql -U postgres -c "SHOW config_file"
# Output: /opt/homebrew/var/postgresql@15/postgresql.conf

# 2. Edit the file
code /opt/homebrew/var/postgresql@15/postgresql.conf

# 3. Add these lines:
log_statement = 'all'                    # Log all queries
log_duration = on                        # Show query execution time
log_line_prefix = '%t [%p] %u@%d '      # Add timestamp and user info
logging_collector = on                   # Enable log collection
log_directory = 'log'                    # Log directory
log_filename = 'postgresql-%Y-%m-%d.log' # Log file pattern

# 4. Restart PostgreSQL
brew services restart postgresql@15

# 5. Tail the logs
tail -f /opt/homebrew/var/postgresql@15/log/postgresql-$(date +%Y-%m-%d).log
```

#### On Linux:

```bash
# 1. Find config
sudo -u postgres psql -c "SHOW config_file"
# Output: /etc/postgresql/15/main/postgresql.conf

# 2. Edit
sudo nano /etc/postgresql/15/main/postgresql.conf

# 3. Add the same lines as above

# 4. Restart
sudo systemctl restart postgresql

# 5. Tail logs
sudo tail -f /var/log/postgresql/postgresql-15-main.log
```

#### Output looks like:

```
2024-01-03 15:30:45.123 EST [12345] activelog_user@activelog_dev LOG:  duration: 2.456 ms  statement: SELECT * FROM activities WHERE user_id = $1
2024-01-03 15:30:45.130 EST [12345] activelog_user@activelog_dev LOG:  duration: 5.123 ms  statement: INSERT INTO activities...
```

---

## Method 3: Quick & Dirty - Direct Printf Logging

Just add this to your repository methods temporarily:

```go
func (r *ActivityRepository) GetByID(ctx context.Context, id int) (*Activity, error) {
    query := "SELECT * FROM activities WHERE id = $1"

    // 🔥 Quick debug logging
    log.Printf("[QUERY] %s | args: %v", query, id)

    row := r.db.QueryRowContext(ctx, query, id)
    // ...
}
```

---

## Testing N+1 Query Problem

With logging enabled, you can verify your N+1 fix:

```go
// BAD: This will show N+1 queries in logs
func GetActivitiesWithTags_BAD(userID int) {
    // 1 query
    activities := getActivities(userID)

    // N queries (one per activity)
    for _, activity := range activities {
        tags := getTags(activity.ID)  // You'll see this query repeated N times!
        activity.Tags = tags
    }
}

// GOOD: This will show only 1 query in logs
func GetActivitiesWithTags_GOOD(userID int) {
    // 1 query with JOIN - you'll see ONE complex query
    activities := getActivitiesWithTagsInOneQuery(userID)
}
```

### Example log output comparison:

**N+1 Problem (BAD):**
```
[SQL] ✅ SELECT * FROM activities WHERE user_id = $1
[SQL] ✅ SELECT name FROM tags WHERE activity_id = $1  -- Query 1
[SQL] ✅ SELECT name FROM tags WHERE activity_id = $1  -- Query 2
[SQL] ✅ SELECT name FROM tags WHERE activity_id = $1  -- Query 3
... (repeated N times!)
```

**Optimized (GOOD):**
```
[SQL] ✅ SELECT activities.*, tags.name FROM activities LEFT JOIN activity_tags LEFT JOIN tags...
```

---

## Best Practice: Environment-Based Logging

```go
// config/config.go
type Config struct {
    EnableQueryLogging bool
}

func Load() *Config {
    return &Config{
        EnableQueryLogging: os.Getenv("ENABLE_QUERY_LOGGING") == "true",
    }
}

// main.go
if config.EnableQueryLogging {
    loggedDB := database.NewLoggingDB(db, logger)
    db = loggedDB
}
```

Then in your `.env`:

```bash
# Development
ENABLE_QUERY_LOGGING=true

# Production (disable for performance)
ENABLE_QUERY_LOGGING=false
```

---

## Performance Impact

⚠️ **Warning:** Query logging adds overhead:
- Application-level logging: ~0.1-0.5ms per query
- PostgreSQL logging: ~0.01-0.1ms per query

**Recommendation:**
- ✅ Use in development
- ✅ Use temporarily in staging to debug
- ❌ Avoid in production (use only for critical debugging)

---

## Alternative: Use `sqldblogger` Package

If you don't want to maintain your own logger:

```bash
go get github.com/simukti/sqldb-logger
```

```go
import (
    "github.com/simukti/sqldb-logger"
    "github.com/simukti/sqldb-logger/logadapter/zerologadapter"
)

db, _ := sql.Open("postgres", connStr)
db = sqldblogger.OpenDriver(connStr, db.Driver(), zerologadapter.New(logger))
```

---

## Summary

**For your use case (verifying N+1 fix):**

1. ✅ Use **Method 1** (Application-Level) - easiest to see query patterns
2. Run your app
3. Look for repeated queries with same pattern
4. Compare before/after optimizing

**Quick start:**
```bash
# 1. Use the logger I created
# 2. Wrap your db in main.go
# 3. Run your app
# 4. Watch the logs!
```

That's it! You'll now see every query that hits your database, just like NestJS. 🎉
