# MONTH 3: Advanced Database & Testing

**Weeks:** 9-12
**Phase:** Database Mastery & Quality Assurance
**Theme:** Build confidence through testing and advanced database patterns

---

## Overview

This month focuses on two critical pillars of backend development: advanced database operations and comprehensive testing. You'll learn how to handle complex data relationships, write efficient queries, and achieve 70%+ test coverage. By the end, you'll be confident writing production-grade database code and tests.

---

## API Endpoints Reference (for Postman Testing)

This section contains all API request/response examples you'll need for testing in Postman during Month 3.

### Authentication Endpoints (From Earlier Months)

**User Registration:**
- **HTTP Method:** `POST`
- **URL:** `/api/v1/auth/register`
- **Headers:**
  ```
  Content-Type: application/json
  ```
- **Request Body:**
  ```json
  {
    "username": "john_doe",
    "email": "john@example.com",
    "password": "SecurePassword123!"
  }
  ```
- **Success Response (201 Created):**
  ```json
  {
    "id": 1,
    "username": "john_doe",
    "email": "john@example.com",
    "created_at": "2024-01-15T10:00:00Z"
  }
  ```

**User Login:**
- **HTTP Method:** `POST`
- **URL:** `/api/v1/auth/login`
- **Headers:**
  ```
  Content-Type: application/json
  ```
- **Request Body:**
  ```json
  {
    "email": "john@example.com",
    "password": "SecurePassword123!"
  }
  ```
- **Success Response (200 OK):**
  ```json
  {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": 1,
      "username": "john_doe",
      "email": "john@example.com"
    }
  }
  ```

### Activity CRUD Endpoints

**Get Single Activity:**
- **HTTP Method:** `GET`
- **URL:** `/api/v1/activities/{id}`
- **Headers:**
  ```
  Authorization: Bearer <your-jwt-token>
  ```
- **Success Response (200 OK):**
  ```json
  {
    "id": 123,
    "user_id": 1,
    "activity_type": "running",
    "duration_minutes": 45,
    "distance_km": 7.5,
    "notes": "Morning run in the park",
    "activity_date": "2024-01-15T06:30:00Z",
    "tags": ["morning", "outdoor", "cardio"],
    "created_at": "2024-01-15T06:35:22Z",
    "updated_at": "2024-01-15T06:35:22Z"
  }
  ```
- **Error Response (404 Not Found):**
  ```json
  {
    "error": "not found",
    "message": "activity not found"
  }
  ```

**Update Activity:**
- **HTTP Method:** `PATCH`
- **URL:** `/api/v1/activities/{id}`
- **Headers:**
  ```
  Content-Type: application/json
  Authorization: Bearer <your-jwt-token>
  ```
- **Request Body (partial update):**
  ```json
  {
    "duration_minutes": 50,
    "notes": "Updated: Morning run in the park with sprints",
    "tags": ["morning", "outdoor", "cardio", "sprints"]
  }
  ```
- **Success Response (200 OK):**
  ```json
  {
    "id": 123,
    "user_id": 1,
    "activity_type": "running",
    "duration_minutes": 50,
    "distance_km": 7.5,
    "notes": "Updated: Morning run in the park with sprints",
    "activity_date": "2024-01-15T06:30:00Z",
    "tags": ["morning", "outdoor", "cardio", "sprints"],
    "created_at": "2024-01-15T06:35:22Z",
    "updated_at": "2024-01-15T10:20:15Z"
  }
  ```

**Delete Activity:**
- **HTTP Method:** `DELETE`
- **URL:** `/api/v1/activities/{id}`
- **Headers:**
  ```
  Authorization: Bearer <your-jwt-token>
  ```
- **Success Response (204 No Content):**
  ```
  (empty body)
  ```
- **Error Response (404 Not Found):**
  ```json
  {
    "error": "not found",
    "message": "activity not found"
  }
  ```
- **Error Response (403 Forbidden):**
  ```json
  {
    "error": "forbidden",
    "message": "you can only delete your own activities"
  }
  ```

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

# WEEKLY TASK BREAKDOWNS

## Week 9: Database Transactions + N+1 Query Problem

### 📋 Implementation Tasks

**Task 1: Create Database Migration for Tags** (20 min)
- [X] Create migration file `migrations/003_create_tags.up.sql`
- [X] Add tags table schema
- [X] Add activity_tags junction table
- [X] Create indexes for performance (user_date, activity_type, tag lookups)
- [X] Create corresponding down migration `003_create_tags.down.sql`
- [X] Run migration: `migrate -path migrations -database "postgres://..." up`

**Task 2: Update Activity Model** (15 min)
- [X] Open `internal/models/activity.go`
- [X] Add `Tags []string` field to Activity struct
- [X] Add JSON tag: `json:"tags,omitempty"`
- [X] Update any existing test fixtures to include empty tags slice

**Task 3: Create Tag Repository Methods** (45 min)
- [X] Create `internal/repository/tag_repository.go`
- [X] Implement `GetOrCreateTag(ctx context.Context, name string) (int, error)`
  - **Purpose:** Ensure a tag exists in the database and get its ID. If the tag doesn't exist, create it. This prevents duplicate tags with the same name.
  - **Returns:** `(int, error)` - The tag's database ID (for linking to activities), or an error if database operation fails.
  - **Logic:** Query tags table for existing tag with given name. If found, return its ID. If not found, INSERT new tag and return the generated ID. Use a single query with INSERT ... ON CONFLICT to make it atomic.
- [X] Implement `GetTagsForActivity(ctx context.Context, activityID int) ([]string, error)`
  - **Purpose:** Retrieve all tag names associated with a specific activity. Used when displaying a single activity's details.
  - **Returns:** `([]string, error)` - Slice of tag names like `["outdoor", "cardio", "morning"]`. Empty slice `[]` if activity has no tags (not an error).
  - **Logic:** JOIN activity_tags with tags table WHERE activity_id matches. Return slice of tag names (not IDs). Return empty slice if activity has no tags (not an error).
- [X] Implement `LinkActivityTag(ctx context.Context, activityID, tagID int) error`
  - **Purpose:** Create the many-to-many relationship between an activity and a tag. This links one activity to one tag.
  - **Returns:** `error` - nil on success, error if the link already exists or if activityID/tagID is invalid.
  - **Logic:** INSERT into activity_tags table with the given activityID and tagID. The primary key constraint prevents duplicate links. Return error if foreign key constraint fails (invalid activity or tag ID).
- [X] Handle duplicate tag names (use INSERT ... ON CONFLICT)

**Task 4: Implement CreateWithTags Using Transactions** (60 min)
- [X] Add method to ActivityRepository: `CreateWithTags(ctx, activity, tags) error`
  - **Purpose:** Create a new activity and associate it with multiple tags in a single atomic operation. If any step fails, nothing is saved (all-or-nothing).
  - **Input:** `activity *models.Activity` (the activity to create), `tags []string` (tag names like `["running", "morning"]`)
  - **Returns:** `error` - nil on success, error if any database operation fails. On success, the `activity` struct is updated with the generated ID and timestamps.
  - **Outcome:** Activity exists in database with ID populated, and all tags are linked in the activity_tags junction table.
  - **Logic:**
    1. Start transaction with `db.BeginTx(ctx, nil)` and defer `tx.Rollback()` (safe to call after commit)
    2. INSERT activity into activities table using the transaction, get generated ID with RETURNING clause
    3. For each tag in tags slice:
       - Use INSERT ... ON CONFLICT to get or create tag (returns tag ID)
       - INSERT into activity_tags junction table to link activity and tag
    4. If any step fails, return error (deferred Rollback will execute)
    5. If all succeed, call `tx.Commit()` to save changes
    6. Return nil on success

  - **API Endpoint (Create Activity with Tags):**
    - **HTTP Method:** `POST`
    - **URL:** `/api/v1/activities`
    - **Headers:**
      ```
      Content-Type: application/json
      Authorization: Bearer <your-jwt-token>
      ```
    - **Request Body:**
      ```json
      {
        "activity_type": "running",
        "duration_minutes": 45,
        "distance_km": 7.5,
        "notes": "Morning run in the park",
        "activity_date": "2024-01-15T06:30:00Z",
        "tags": ["morning", "outdoor", "cardio"]
      }
      ```
    - **Success Response (201 Created):**
      ```json
      {
        "id": 123,
        "user_id": 1,
        "activity_type": "running",
        "duration_minutes": 45,
        "distance_km": 7.5,
        "notes": "Morning run in the park",
        "activity_date": "2024-01-15T06:30:00Z",
        "tags": ["morning", "outdoor", "cardio"],
        "created_at": "2024-01-15T06:35:22Z",
        "updated_at": "2024-01-15T06:35:22Z"
      }
      ```
    - **Error Response (400 Bad Request):**
      ```json
      {
        "error": "invalid input",
        "message": "duration_minutes must be positive"
      }
      ```
    - **Error Response (401 Unauthorized):**
      ```json
      {
        "error": "unauthorized",
        "message": "missing or invalid token"
      }
      ```

- [X] Start transaction with `db.BeginTx(ctx, nil)`
- [X] Insert activity and get ID back (RETURNING clause)
- [X] Loop through tags: get/create tag, then link to activity
- [X] Implement proper error handling with tx.Rollback()
- [X] Commit transaction if all succeeds
- [X] Test rollback behavior (simulate failure after activity insert)

**Task 5: Fix N+1 Query Problem** (90 min)
- [X] Create `GetActivitiesWithTags(ctx, userID) ([]*Activity, error)` method
  - **Purpose:** Efficiently fetch all activities for a user WITH their associated tags in a single database query. Solves the N+1 query problem.
  - **Returns:** `([]*Activity, error)` - Slice of Activity structs where each Activity has its `Tags []string` field populated. Example:
    ```go
    []*Activity {
        {ID: 1, Type: "running", Tags: ["morning", "outdoor"]},
        {ID: 2, Type: "yoga", Tags: ["evening"]},
        {ID: 3, Type: "swimming", Tags: []},  // no tags
    }
    ```
  - **Why this matters:** Without this, you'd do 1 query for activities + N queries (one per activity) to get tags = N+1 queries. This does it in 1.
  - **Logic:**
    1. Execute single query: SELECT activities.*, tags.id, tags.name FROM activities LEFT JOIN activity_tags LEFT JOIN tags WHERE user_id = $1
    2. Loop through rows (one row per activity-tag combination, or one row for activities with no tags)
    3. Use map[int]*Activity to deduplicate - if activity ID already in map, append tag to existing activity; if not, add new activity to map
    4. Handle NULL tag values using sql.NullInt64 and sql.NullString (when activity has no tags)
    5. Convert map values to slice and return
    - **Why:** Instead of 1 query for activities + N queries for tags (N+1 problem), this uses 1 query total

  - **API Endpoint (Get Activities with Tags):**
    - **HTTP Method:** `GET`
    - **URL:** `/api/v1/activities`
    - **Headers:**
      ```
      Authorization: Bearer <your-jwt-token>
      ```
    - **Query Parameters (optional):**
      ```
      ?limit=10&offset=0&sort=date_desc
      ```
    - **Success Response (200 OK):**
      ```json
      {
        "activities": [
          {
            "id": 123,
            "user_id": 1,
            "activity_type": "running",
            "duration_minutes": 45,
            "distance_km": 7.5,
            "notes": "Morning run in the park",
            "activity_date": "2024-01-15T06:30:00Z",
            "tags": ["morning", "outdoor", "cardio"],
            "created_at": "2024-01-15T06:35:22Z",
            "updated_at": "2024-01-15T06:35:22Z"
          },
          {
            "id": 122,
            "user_id": 1,
            "activity_type": "yoga",
            "duration_minutes": 30,
            "distance_km": 0,
            "notes": "Evening yoga session",
            "activity_date": "2024-01-14T18:00:00Z",
            "tags": ["evening", "flexibility"],
            "created_at": "2024-01-14T18:05:10Z",
            "updated_at": "2024-01-14T18:05:10Z"
          },
          {
            "id": 121,
            "user_id": 1,
            "activity_type": "swimming",
            "duration_minutes": 60,
            "distance_km": 2.0,
            "notes": "Pool laps",
            "activity_date": "2024-01-13T07:00:00Z",
            "tags": [],
            "created_at": "2024-01-13T07:15:33Z",
            "updated_at": "2024-01-13T07:15:33Z"
          }
        ],
        "total": 3,
        "limit": 10,
        "offset": 0
      }
      ```
    - **Error Response (401 Unauthorized):**
      ```json
      {
        "error": "unauthorized",
        "message": "missing or invalid token"
      }
      ```
    - **Error Response (500 Internal Server Error):**
      ```json
      {
        "error": "database error",
        "message": "failed to fetch activities"
      }
      ```

- [X] Write JOIN query (activities LEFT JOIN activity_tags LEFT JOIN tags)
- [X] Handle NULL values for activities without tags (sql.NullInt64, sql.NullString)
- [X] Build activityMap to deduplicate rows
- [X] Append tags to each activity
- [X] Compare query count: old approach vs new (should be 1 query vs N+1)
- [X] Add database query logging to verify

**Task 6: Write Transaction Tests** (45 min)
- [ ] Test `CreateWithTags` with multiple tags
- [ ] Test transaction rollback on tag insertion failure
- [ ] Test duplicate tag handling (should reuse existing tags)
- [ ] Verify all data committed or none (atomic behavior)

**Task 7: Verify N+1 Fix** (20 min)
- [ ] Enable PostgreSQL query logging (edit postgresql.conf if needed)
- [ ] Create 10 activities with tags
- [ ] Call old method and count queries (should see N+1)
- [ ] Call new method and count queries (should see 1-2)
- [ ] Document the performance improvement

### 📦 Files You'll Create/Modify

```
migrations/
├── 003_create_tags.up.sql         [CREATE]
└── 003_create_tags.down.sql       [CREATE]

internal/
├── models/
│   └── activity.go                [MODIFY - add Tags field]
├── repository/
│   ├── tag_repository.go          [CREATE]
│   ├── activity_repository.go     [MODIFY - add CreateWithTags, GetActivitiesWithTags]
│   └── activity_repository_test.go [MODIFY - add transaction tests]
```

### 🔄 Implementation Order

1. **Database first**: Migration → Run migration
2. **Models**: Update Activity model with Tags field
3. **Repository layer**: Tag repository → ActivityRepository methods
4. **Testing**: Transaction tests → N+1 verification
5. **Optimization**: Measure before/after query counts

### ⚠️ Blockers to Watch For

- **Transaction scope**: Don't forget `defer tx.Rollback()` - won't auto-rollback on error
- **NULL handling**: Use `sql.NullInt64` and `sql.NullString` for optional JOIN columns
- **Map deduplication**: Activities appear multiple times in JOIN results (one row per tag)
- **ON CONFLICT**: Requires unique constraint on tag name - check migration
- **Context cancellation**: Transactions respect context timeout - test this

### ✅ Definition of Done

- [ ] Can create activity with tags in single transaction
- [ ] Tags are reused if they already exist (no duplicates)
- [ ] GetActivitiesWithTags uses 1 query instead of N+1
- [ ] Transaction rolls back if any step fails
- [ ] All tests passing with transaction scenarios
- [ ] Query performance verified (logs or EXPLAIN ANALYZE)

---

## Week 10: Complex Queries + Joins + Graceful Shutdown

### 📋 Implementation Tasks

**Task 1: Implement Analytics Queries** (60 min)
- [ ] Create `internal/repository/stats_repository.go`
- [ ] Implement `GetWeeklyStats(ctx, userID) (*WeeklyStats, error)`
  - **Purpose:** Calculate aggregate statistics for a user's activities over the past 7 days. Used for weekly summary emails and dashboard.
  - **Returns:** `(*WeeklyStats, error)` - Pointer to struct with aggregated data:
    ```go
    type WeeklyStats struct {
        TotalActivities int     `json:"total_activities"`
        TotalDuration   int     `json:"total_duration_minutes"`
        TotalDistance   float64 `json:"total_distance_km"`
        AvgDuration     float64 `json:"avg_duration_minutes"`
    }
    // Example: &WeeklyStats{TotalActivities: 12, TotalDuration: 360, TotalDistance: 45.5, AvgDuration: 30.0}
    ```
  - **Logic:**
    1. Query activities WHERE user_id = $1 AND activity_date >= NOW() - INTERVAL '7 days'
    2. Use aggregate functions: COUNT(*) as total_activities, SUM(duration_minutes) as total_duration, SUM(distance_km) as total_distance, AVG(duration_minutes) as avg_duration
    3. Optionally GROUP BY activity_type if you want per-type breakdown
    4. Scan results into WeeklyStats struct with fields like TotalActivities, TotalDuration, TotalDistance, AvgDuration
  - Use SUM, COUNT, AVG aggregate functions
  - Filter by date range (past 7 days)
  - GROUP BY activity_type
- [ ] Implement `GetMonthlyStats(ctx, userID) (*MonthlyStats, error)`
  - **Purpose:** Calculate aggregate statistics for a user's activities over the past 30 days. Used for monthly reports.
  - **Returns:** `(*MonthlyStats, error)` - Same structure as WeeklyStats but covers 30-day period.
  - **Logic:** Same as GetWeeklyStats but use '30 days' or '1 month' interval. Return MonthlyStats struct with same aggregate fields.
- [ ] Implement `GetActivityCountByType(ctx, userID) (map[string]int, error)`
  - **Purpose:** Get a breakdown of how many activities of each type a user has logged. Shows distribution across activity types.
  - **Returns:** `(map[string]int, error)` - Map where key is activity type, value is count. Example:
    ```go
    map[string]int{
        "running":    25,
        "cycling":    15,
        "swimming":   8,
        "basketball": 12,
    }
    ```
  - **Logic:** SELECT activity_type, COUNT(*) FROM activities WHERE user_id = $1 GROUP BY activity_type. Loop through rows and build map[string]int where key is activity type and value is count.
- [ ] Test with real data to verify correctness

**Task 2: Create Complex JOIN Queries** (45 min)
- [ ] Implement `GetUserActivitySummary(ctx, userID) (*UserActivitySummary, error)`
  - **Purpose:** Get a complete overview of a user's activity profile - total activities and unique tags they've used.
  - **Returns:** `(*UserActivitySummary, error)` - Struct with user summary:
    ```go
    type UserActivitySummary struct {
        Username     string `json:"username"`
        ActivityCount int   `json:"activity_count"`
        UniqueTagCount int  `json:"unique_tag_count"`
    }
    // Example: &UserActivitySummary{Username: "john_doe", ActivityCount: 150, UniqueTagCount: 12}
    ```
  - **Logic:** SELECT users.username, COUNT(DISTINCT activities.id) as activity_count, COUNT(DISTINCT tags.id) as unique_tags FROM users LEFT JOIN activities LEFT JOIN activity_tags LEFT JOIN tags WHERE users.id = $1 GROUP BY users.id. Returns summary with user info + aggregate stats.

  - **API Endpoint (Get User Activity Summary):**
    - **HTTP Method:** `GET`
    - **URL:** `/api/v1/users/me/summary`
    - **Headers:**
      ```
      Authorization: Bearer <your-jwt-token>
      ```
    - **Success Response (200 OK):**
      ```json
      {
        "username": "john_doe",
        "activity_count": 150,
        "unique_tag_count": 12,
        "member_since": "2023-06-15T10:00:00Z",
        "last_activity": "2024-01-15T06:30:00Z"
      }
      ```
    - **Error Response (401 Unauthorized):**
      ```json
      {
        "error": "unauthorized",
        "message": "missing or invalid token"
      }
      ```

- [ ] Implement `GetTopTagsByUser(ctx, userID, limit int) ([]TagUsage, error)`
  - **Purpose:** Find which tags a user uses most frequently. Useful for showing "top categories" in analytics.
  - **Returns:** `([]TagUsage, error)` - Slice of structs ordered by usage count:
    ```go
    type TagUsage struct {
        TagName string `json:"tag_name"`
        Count   int    `json:"count"`
    }
    // Example: []TagUsage{{"outdoor", 45}, {"morning", 32}, {"cardio", 28}}
    ```
  - **Logic:** SELECT tags.name, COUNT(*) as usage_count FROM tags JOIN activity_tags JOIN activities WHERE activities.user_id = $1 GROUP BY tags.id, tags.name ORDER BY usage_count DESC LIMIT $2. Returns slice of tag names sorted by most used.

  - **API Endpoint (Get Top Tags):**
    - **HTTP Method:** `GET`
    - **URL:** `/api/v1/users/me/tags/top`
    - **Headers:**
      ```
      Authorization: Bearer <your-jwt-token>
      ```
    - **Query Parameters (optional):**
      ```
      ?limit=10
      ```
    - **Success Response (200 OK):**
      ```json
      {
        "tags": [
          {
            "tag_name": "outdoor",
            "count": 45
          },
          {
            "tag_name": "morning",
            "count": 32
          },
          {
            "tag_name": "cardio",
            "count": 28
          },
          {
            "tag_name": "evening",
            "count": 20
          },
          {
            "tag_name": "strength",
            "count": 18
          }
        ],
        "total_unique_tags": 12
      }
      ```
    - **Error Response (401 Unauthorized):**
      ```json
      {
        "error": "unauthorized",
        "message": "missing or invalid token"
      }
      ```

- [ ] Use LEFT JOIN vs INNER JOIN appropriately
- [ ] Handle NULL values in results
- [ ] Add LIMIT and ORDER BY for performance

**Task 3: Implement Graceful Shutdown** (90 min)
- [ ] Open `cmd/api/main.go`
  - **Logic:**
    1. Create buffered signal channel and register for SIGINT/SIGTERM
    2. Start HTTP server in a goroutine (non-blocking)
    3. Main goroutine blocks on signal channel with `<-quit`
    4. When signal received, create context with 30s timeout
    5. Call `srv.Shutdown(ctx)` - this stops accepting new connections and waits for active requests to finish (up to 30s)
    6. Close database and other resources after shutdown completes
    7. Log each step for observability
    - **Why:** Prevents abrupt termination that could corrupt in-flight requests or leave database connections open
- [ ] Import `os`, `os/signal`, `syscall`, `context`
- [ ] Create signal channel: `quit := make(chan os.Signal, 1)`
- [ ] Register signals: `signal.Notify(quit, os.Interrupt, syscall.SIGTERM)`
- [ ] Start server in goroutine
- [ ] Wait for signal with `<-quit`
- [ ] Create shutdown context with 30s timeout
- [ ] Call `srv.Shutdown(ctx)` to drain connections
- [ ] Close database connections after shutdown
- [ ] Log shutdown steps for debugging

**Task 4: Test Graceful Shutdown** (30 min)
- [ ] Start server: `go run cmd/api/main.go`
- [ ] Send test request that takes 10s to complete (add sleep in handler)
- [ ] Send SIGTERM while request is processing: `kill -TERM <pid>`
- [ ] Verify request completes before server exits
- [ ] Verify new requests are rejected during shutdown
- [ ] Check logs show "Server exited" message

**Task 5: Add Query Timeouts** (20 min)
- [ ] Wrap all repository queries with context timeout
- [ ] Use `context.WithTimeout(ctx, 5*time.Second)`
- [ ] Test timeout behavior with slow query: `SELECT pg_sleep(10)`
- [ ] Verify context.DeadlineExceeded error returned

**Task 6: Create Analytics Endpoint** (45 min)
- [ ] Create `internal/handlers/stats_handler.go`
- [ ] Implement `GetWeeklyStats(w, r)` handler
- [ ] Implement `GetMonthlyStats(w, r)` handler
- [ ] Add routes to router: `/api/v1/users/me/stats/weekly`, `/monthly`
- [ ] Protect with auth middleware
- [ ] Test with curl/Postman

  - **API Endpoint (Get Weekly Stats):**
    - **HTTP Method:** `GET`
    - **URL:** `/api/v1/users/me/stats/weekly`
    - **Headers:**
      ```
      Authorization: Bearer <your-jwt-token>
      ```
    - **Success Response (200 OK):**
      ```json
      {
        "total_activities": 12,
        "total_duration_minutes": 360,
        "total_distance_km": 45.5,
        "avg_duration_minutes": 30.0,
        "period": "last_7_days",
        "start_date": "2024-01-08T00:00:00Z",
        "end_date": "2024-01-15T23:59:59Z"
      }
      ```
    - **Error Response (401 Unauthorized):**
      ```json
      {
        "error": "unauthorized",
        "message": "missing or invalid token"
      }
      ```

  - **API Endpoint (Get Monthly Stats):**
    - **HTTP Method:** `GET`
    - **URL:** `/api/v1/users/me/stats/monthly`
    - **Headers:**
      ```
      Authorization: Bearer <your-jwt-token>
      ```
    - **Success Response (200 OK):**
      ```json
      {
        "total_activities": 52,
        "total_duration_minutes": 1560,
        "total_distance_km": 195.8,
        "avg_duration_minutes": 30.0,
        "period": "last_30_days",
        "start_date": "2023-12-16T00:00:00Z",
        "end_date": "2024-01-15T23:59:59Z"
      }
      ```
    - **Error Response (401 Unauthorized):**
      ```json
      {
        "error": "unauthorized",
        "message": "missing or invalid token"
      }
      ```

  - **API Endpoint (Get Activity Count by Type):**
    - **HTTP Method:** `GET`
    - **URL:** `/api/v1/users/me/stats/by-type`
    - **Headers:**
      ```
      Authorization: Bearer <your-jwt-token>
      ```
    - **Success Response (200 OK):**
      ```json
      {
        "activity_breakdown": {
          "running": 25,
          "cycling": 15,
          "swimming": 8,
          "basketball": 12,
          "yoga": 10,
          "gym": 20
        },
        "total_activities": 90
      }
      ```
    - **Error Response (401 Unauthorized):**
      ```json
      {
        "error": "unauthorized",
        "message": "missing or invalid token"
      }
      ```

### 📦 Files You'll Create/Modify

```
internal/
├── repository/
│   ├── stats_repository.go        [CREATE]
│   └── stats_repository_test.go   [CREATE]
├── handlers/
│   ├── stats_handler.go           [CREATE]
│   └── stats_handler_test.go      [CREATE]
└── models/
    └── stats.go                   [CREATE - WeeklyStats, MonthlyStats structs]

cmd/api/
└── main.go                        [MODIFY - add graceful shutdown]
```

### 🔄 Implementation Order

1. **Stats models**: Define struct types for statistics
2. **Repository**: Implement aggregate queries
3. **Testing**: Test queries with sample data
4. **Handlers**: Wire up HTTP endpoints
5. **Graceful shutdown**: Modify main.go last (affects server lifecycle)

### ⚠️ Blockers to Watch For

- **NULL in aggregates**: COUNT(*) includes NULLs, COUNT(column) doesn't
- **GROUP BY**: All non-aggregated SELECT columns must be in GROUP BY
- **Date ranges**: Use `>= AND <` for date ranges, not BETWEEN (timezone issues)
- **Shutdown timeout**: 30s might be too short for long-running requests - adjust as needed
- **Signal handling**: Only works on Unix systems - Windows uses different signals

### ✅ Definition of Done

- [ ] Can get weekly/monthly activity statistics
- [ ] All aggregate queries return correct results
- [ ] Server shuts down gracefully on SIGTERM/SIGINT
- [ ] In-flight requests complete during shutdown (tested)
- [ ] Database connections close cleanly
- [ ] Analytics endpoints working and protected by auth

---

## Week 11: Table-Driven Tests + Mocking + Mock Generation

### 📋 Implementation Tasks

**Task 1: Install Testing Dependencies** (10 min)
- [ ] Install testify: `go get github.com/stretchr/testify`
- [ ] Install gomock: `go get github.com/golang/mock/mockgen`
- [ ] Install mockgen tool: `go install github.com/golang/mock/mockgen@latest`
- [ ] Verify installation: `mockgen -version`

**Task 2: Define Repository Interfaces** (30 min)
- [ ] Create `internal/repository/interfaces.go`
- [ ] Define `ActivityRepository` interface with all methods
- [ ] Define `UserRepository` interface
- [ ] Define `StatsRepository` interface
- [ ] Update existing repositories to implement interfaces explicitly
- [ ] Add `//go:generate mockgen` directives above each interface

**Task 3: Generate Mocks** (15 min)
- [ ] Add to each interface:
  ```go
  //go:generate mockgen -destination=mocks/mock_activity_repository.go -package=mocks . ActivityRepository
  ```
- [ ] Run `go generate ./...` from project root
- [ ] Verify mocks created in `internal/repository/mocks/`
- [ ] Check mocks compile: `go build ./internal/repository/mocks`

**Task 4: Convert Tests to Table-Driven Pattern** (90 min)
- [ ] Refactor `activity_repository_test.go` to table-driven tests
- [ ] Refactor `user_repository_test.go` to table-driven tests
- [ ] Each test should have:
  - `name` field for test case description
  - Input fields
  - Expected output fields
  - `wantErr bool` field
- [ ] Use `t.Run()` to execute subtests
- [ ] Use `testify/assert` for cleaner assertions

**Task 5: Write Mock-Based Handler Tests** (120 min)
- [ ] Create `internal/handlers/activity_handler_test.go`
- [ ] Test `CreateActivity` handler:
  - Mock repository returning success
  - Mock repository returning error
  - Invalid JSON payload
  - Missing required fields
- [ ] Test `GetActivities` handler with mock
- [ ] Test `UpdateActivity` handler with mock
- [ ] Use `gomock.NewController(t)` and `EXPECT()` chains
- [ ] Verify mock expectations with `ctrl.Finish()`

**Task 6: Test Error Paths** (45 min)
- [ ] Test database connection errors
- [ ] Test context cancellation
- [ ] Test invalid input validation
- [ ] Test concurrent access (use goroutines)
- [ ] Verify proper error messages returned

**Task 7: Measure Code Coverage** (20 min)
- [ ] Run `go test -cover ./...` to see overall coverage
- [ ] Run `go test -coverprofile=coverage.out ./...`
- [ ] View HTML report: `go tool cover -html=coverage.out`
- [ ] Identify untested code paths
- [ ] Add tests to reach 70%+ coverage

### 📦 Files You'll Create/Modify

```
internal/
├── repository/
│   ├── interfaces.go              [CREATE]
│   ├── mocks/                     [CREATE DIR]
│   │   ├── mock_activity_repository.go  [GENERATED]
│   │   ├── mock_user_repository.go      [GENERATED]
│   │   └── mock_stats_repository.go     [GENERATED]
│   ├── activity_repository_test.go [MODIFY - table-driven]
│   └── user_repository_test.go    [MODIFY - table-driven]
└── handlers/
    ├── activity_handler_test.go   [CREATE - mock tests]
    ├── user_handler_test.go       [MODIFY - add mock tests]
    └── stats_handler_test.go      [CREATE - mock tests]

coverage.out                       [GENERATED]
```

### 🔄 Implementation Order

1. **Setup**: Install dependencies
2. **Interfaces**: Define repository interfaces
3. **Generation**: Generate mocks with go generate
4. **Repository tests**: Convert to table-driven
5. **Handler tests**: Write mock-based tests
6. **Coverage**: Measure and improve

### ⚠️ Blockers to Watch For

- **Interface location**: Must be in same package as implementation for `go generate`
- **Mock regeneration**: Re-run `go generate` after interface changes
- **gomock.Any()**: Use for parameters you don't want to verify
- **EXPECT() order**: Calls must happen in declared order unless using `.AnyTimes()`
- **Controller.Finish()**: Must call or use `defer ctrl.Finish()` to verify mocks
- **Table test isolation**: Each test case should be independent (clean state)

### ✅ Definition of Done

- [ ] All repository interfaces defined
- [ ] Mocks auto-generated with `go generate`
- [ ] All repository tests use table-driven pattern
- [ ] All handler tests use mocks (no real database)
- [ ] Code coverage >= 70% (run `go test -cover ./...`)
- [ ] Error paths tested (not just happy path)
- [ ] Tests run fast (mocks = no database overhead)

---

## Week 12: Benchmarking + Optimization + Testcontainers

### 📋 Implementation Tasks

**Task 1: Install Testcontainers** (15 min)
- [ ] Install package: `go get github.com/testcontainers/testcontainers-go`
- [ ] Install postgres module: `go get github.com/testcontainers/testcontainers-go/modules/postgres`
- [ ] Ensure Docker is running: `docker ps`
- [ ] Pull postgres image: `docker pull postgres:15`

**Task 2: Create Testcontainer Setup Helper** (45 min)
- [ ] Create `internal/repository/testhelpers/container.go`
- [ ] Implement `SetupTestDB(t *testing.T) (*sql.DB, func())`
  - **Purpose:** Create a real PostgreSQL database running in a Docker container for integration tests. Each test gets a fresh, isolated database.
  - **Returns:** `(*sql.DB, func())` - Two values:
    1. `*sql.DB` - Connection to the test database (fully migrated and ready to use)
    2. `func()` - Cleanup function that must be called with `defer cleanup()` to stop the container and close the connection
  - **Usage Example:**
    ```go
    func TestActivityRepository(t *testing.T) {
        db, cleanup := SetupTestDB(t)
        defer cleanup()  // Always call cleanup!

        repo := NewActivityRepository(db)
        // ... test repository methods ...
    }
    ```
  - **Logic:**
    1. Create testcontainers.ContainerRequest with postgres:15 image and test credentials
    2. Start container with testcontainers.GenericContainer
    3. Wait for "database system is ready" log message using wait.ForLog
    4. Get mapped host and port from container
    5. Build connection string and open sql.DB connection
    6. Run all migrations from migrations/ folder on the test database
    7. Return (*sql.DB, cleanup func) where cleanup stops container and closes DB
    - **Why:** Each test gets a fresh, isolated database in a Docker container - no shared state between tests
- [ ] Start postgres container with testcontainers
- [ ] Wait for database to be ready
- [ ] Run migrations on test container
- [ ] Return cleanup function to stop container
- [ ] Test helper works

**Task 3: Write Integration Tests with Testcontainers** (60 min)
- [ ] Create `internal/repository/integration_test.go`
- [ ] Test full transaction flow (create activity with tags)
- [ ] Test concurrent insertions (multiple goroutines)
- [ ] Test foreign key constraints
- [ ] Test unique constraint violations
- [ ] Verify actual database state after operations

**Task 4: Write Benchmark Tests** (90 min)
- [ ] Create `internal/repository/activity_repository_bench_test.go`
- [ ] Benchmark `Create()`: `BenchmarkActivityRepository_Create`
- [ ] Benchmark `GetByID()`: `BenchmarkActivityRepository_GetByID`
- [ ] Benchmark `GetActivitiesWithTags()` (with N+1 comparison)
- [ ] Benchmark `CreateWithTags()` with varying tag counts (1, 5, 10 tags)
- [ ] Use `b.ResetTimer()` to exclude setup time
- [ ] Use `b.ReportAllocs()` to track memory allocations

**Task 5: Profile CPU and Memory** (45 min)
- [ ] Run benchmarks with CPU profile: `go test -bench=. -cpuprofile=cpu.out`
- [ ] Analyze CPU profile: `go tool pprof cpu.out`
- [ ] Run with memory profile: `go test -bench=. -memprofile=mem.out`
- [ ] Identify top memory allocators
- [ ] Look for optimization opportunities

**Task 6: Optimize Slow Queries** (60 min)
- [ ] Use `EXPLAIN ANALYZE` on GetActivitiesWithTags query
- [ ] Verify indexes are being used (check EXPLAIN output)
- [ ] Add missing indexes if needed
- [ ] Optimize query by reducing columns selected
- [ ] Re-run benchmark to measure improvement
- [ ] Document before/after performance

**Task 7: Add Query Performance Logging** (30 min)
- [ ] Create middleware to log slow queries (>100ms)
- [ ] Wrap `db.QueryContext()` calls with timing
- [ ] Log query, duration, and params for slow queries
- [ ] Test with intentionally slow query
- [ ] Integrate into repository layer

### 📦 Files You'll Create/Modify

```
internal/
├── repository/
│   ├── testhelpers/
│   │   └── container.go           [CREATE]
│   ├── integration_test.go        [CREATE]
│   ├── activity_repository_bench_test.go [CREATE]
│   └── query_logger.go            [CREATE]

*.out                              [GENERATED - profiles]
coverage.html                      [GENERATED]
```

### 🔄 Implementation Order

1. **Setup**: Install testcontainers and verify Docker
2. **Test infrastructure**: Create container helper
3. **Integration tests**: Write tests using real database
4. **Benchmarks**: Write and run benchmark tests
5. **Profiling**: Analyze CPU and memory
6. **Optimization**: Fix slow queries, add indexes
7. **Monitoring**: Add query performance logging

### ⚠️ Blockers to Watch For

- **Docker required**: Testcontainers needs Docker daemon running
- **Port conflicts**: Container might conflict with local postgres on 5432
- **Slow tests**: Integration tests are slower - don't run in CI on every commit
- **Benchmark stability**: Run multiple times, results vary with system load
- **b.N value**: Don't use b.N directly - let testing package control it
- **Profile cleanup**: Delete .out files before re-running to avoid confusion
- **Container cleanup**: Always call cleanup function to stop containers

### ✅ Definition of Done

- [ ] Testcontainers working (can start/stop postgres in tests)
- [ ] Integration tests running against real database in container
- [ ] Benchmarks for all critical repository methods
- [ ] CPU and memory profiles analyzed
- [ ] Slow queries identified and optimized
- [ ] Query performance improved (benchmark comparison)
- [ ] Slow query logging added (>100ms threshold)
- [ ] All tests still passing (unit + integration)

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
