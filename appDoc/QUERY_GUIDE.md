# ActiveLog Query Guide

**Last Updated:** 2026-01-08
**Version:** 3.0 (Advanced Relationships + Deep Nesting)
**Status:** Production Ready

**New in v3.0:** Deep nesting, self-referential relationships, polymorphic relationships, and complex JOIN conditions. See [v3.0 Guide](./RELATIONSHIP_REGISTRY_V3_GUIDE.md) for details.

---

## Table of Contents

1. [Overview](#overview)
2. [Quick Start](#quick-start)
3. [Query Parameters](#query-parameters)
4. [Relationship Filtering (Auto-JOINs)](#relationship-filtering-auto-joins)
5. [Operator Filtering](#operator-filtering)
6. [Security & Whitelisting](#security--whitelisting)
7. [For Developers: Implementation Guide](#for-developers-implementation-guide)
8. [API Examples](#api-examples)
9. [Performance](#performance)
10. [Troubleshooting](#troubleshooting)

---

## Overview

ActiveLog implements a **TypeORM-style dynamic filtering system** with automatic JOIN detection. This allows you to build powerful queries through URL parameters without writing custom endpoint code.

### What Can You Do?

```bash
# Filter by column
GET /activities?filter[activity_type]=running

# Filter by relationship (auto-JOIN)
GET /activities?filter[tags.name]=cardio

# Date range filtering
GET /activities?filter[activity_date][gte]=2024-01-01&filter[activity_date][lte]=2024-12-31

# Search (case-insensitive)
GET /activities?search[title]=morning

# Sort results
GET /activities?order[distance_km]=DESC

# Paginate
GET /activities?page=2&limit=20

# Complex queries
GET /activities?filter[tags.name]=cardio&filter[user.username]=john&search[title]=run&order[activity_date]=DESC
```

### Key Features

- ✅ **Natural column names** - Use `tags.name` instead of SQL aliases
- ✅ **Auto-JOIN detection** - Relationships detected from dot notation
- ✅ **Operator filtering** - Support for `eq`, `ne`, `gt`, `gte`, `lt`, `lte`
- ✅ **Secure** - Multi-layered security with column whitelisting
- ✅ **Performant** - Parameterized queries, automatic indexing
- ✅ **Type-safe** - Go generics for compile-time safety
- ✅ **Reusable** - Generic pattern works across all entities

---

## Quick Start

### Basic Filtering

**Filter by activity type:**
```bash
GET /activities?filter[activity_type]=running
```

**Response:**
```json
{
  "data": [
    {
      "id": 1,
      "activityType": "running",
      "title": "Morning Run",
      "durationMinutes": 30,
      "distanceKm": 5.2
    }
  ],
  "meta": {
    "page": 1,
    "limit": 10,
    "count": 1,
    "totalRecords": 1,
    "pageCount": 1,
    "previousPage": false,
    "nextPage": false
  }
}
```

### Relationship Filtering (Auto-JOIN)

**Filter by tag (automatically generates JOINs):**
```bash
GET /activities?filter[tags.name]=cardio
```

**Generated SQL:**
```sql
SELECT activities.*
FROM activities
LEFT JOIN activity_tags ON activity_tags.activity_id = activities.id
LEFT JOIN tags ON tags.id = activity_tags.tag_id
WHERE activities.user_id = $1 AND tags.name = $2
ORDER BY activities.created_at DESC
LIMIT 10
```

### Date Range Filtering

**Activities from January 2024:**
```bash
GET /activities?filter[activity_date][gte]=2024-01-01&filter[activity_date][lte]=2024-01-31
```

### Search & Sort

**Search for "morning" and sort by distance:**
```bash
GET /activities?search[title]=morning&order[distance_km]=DESC
```

---

## Query Parameters

### Pagination

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `page` | integer | 1 | Page number (1-indexed) |
| `limit` | integer | 10 | Items per page (max 100) |

**Example:**
```bash
GET /activities?page=2&limit=20
```

**Response includes metadata:**
```json
{
  "meta": {
    "page": 2,
    "limit": 20,
    "count": 20,
    "previousPage": 1,
    "nextPage": 3,
    "pageCount": 5,
    "totalRecords": 87
  }
}
```

### Filtering

#### Basic Filters

**Syntax:** `filter[column_name]=value`

**Single value (exact match):**
```bash
GET /activities?filter[activity_type]=running
# WHERE activity_type = $1
```

**Multiple values (IN clause):**
```bash
GET /activities?filter[activity_type]=[running,cycling]
# WHERE activity_type IN ($1, $2)
```

**Multiple filters (AND logic):**
```bash
GET /activities?filter[activity_type]=running&filter[duration_minutes]=30
# WHERE activity_type = $1 AND duration_minutes = $2
```

**Type conversion:**
- Strings: `filter[title]=Morning Run`
- Numbers: `filter[duration_minutes]=30`
- Booleans: `filter[is_public]=true`
- Arrays: `filter[activity_type]=[running,cycling]`

#### Operator-Based Filters (v1.1.0+)

**Syntax:** `filter[column_name][operator]=value`

| Operator | Description | Example | SQL |
|----------|-------------|---------|-----|
| `eq` | Equal (default) | `filter[distance_km][eq]=10` | `WHERE distance_km = $1` |
| `ne` | Not equal | `filter[activity_type][ne]=running` | `WHERE activity_type != $1` |
| `gt` | Greater than | `filter[distance_km][gt]=5.0` | `WHERE distance_km > $1` |
| `gte` | Greater than or equal | `filter[activity_date][gte]=2024-01-01` | `WHERE activity_date >= $1` |
| `lt` | Less than | `filter[duration_minutes][lt]=60` | `WHERE duration_minutes < $1` |
| `lte` | Less than or equal | `filter[activity_date][lte]=2024-12-31` | `WHERE activity_date <= $1` |

**Date range example:**
```bash
GET /activities?filter[activity_date][gte]=2024-01-01&filter[activity_date][lte]=2024-12-31
# WHERE activity_date >= $1 AND activity_date <= $2
```

**Numeric range example:**
```bash
GET /activities?filter[distance_km][gt]=5.0&filter[distance_km][lt]=15.0
# WHERE distance_km > $1 AND distance_km < $2
```

**Not equal example:**
```bash
GET /activities?filter[activity_type][ne]=running
# WHERE activity_type != $1
```

**Backward compatibility:**
```bash
# Old syntax (v1.0) - Still works!
filter[activity_type]=running

# New syntax (v1.1) - Explicit equality
filter[activity_type][eq]=running

# Both generate: WHERE activity_type = $1
```

### Search

**Syntax:** `search[column_name]=pattern`

**Case-insensitive pattern matching:**
```bash
GET /activities?search[title]=run
# WHERE title ILIKE '%run%'
```

**Multiple search terms (OR logic):**
```bash
GET /activities?search[title]=morning&search[description]=workout
# WHERE (title ILIKE '%morning%' OR description ILIKE '%workout%')
```

### Ordering

**Syntax:** `order[column_name]=direction`

**Valid directions:** `ASC`, `DESC` (case-insensitive)

**Single column:**
```bash
GET /activities?order[created_at]=DESC
# ORDER BY created_at DESC
```

**Multiple columns:**
```bash
GET /activities?order[activity_date]=DESC&order[distance_km]=ASC
# ORDER BY activity_date DESC, distance_km ASC
```

**Default:** If no order specified, defaults to `created_at DESC`

---

## Relationship Filtering (Auto-JOINs)

ActiveLog automatically generates SQL JOINs when you use **dot notation** in column names.

### How It Works

**You write:**
```bash
GET /activities?filter[tags.name]=cardio
```

**System automatically:**
1. Detects "tags" relationship from `tags.name`
2. Looks up relationship configuration in `RelationshipRegistry`
3. Generates appropriate JOINs based on relationship type
4. Executes query with natural column names

**Generated SQL:**
```sql
SELECT activities.*
FROM activities
LEFT JOIN activity_tags ON activity_tags.activity_id = activities.id
LEFT JOIN tags ON tags.id = activity_tags.tag_id
WHERE activities.user_id = $1 AND tags.name = $2
ORDER BY activities.created_at DESC
```

### Supported Relationship Types

#### 1. Many-to-Many (Activities ↔ Tags)

**Definition:**
```go
registry.Register(query.ManyToManyRelationship(
    "tags",            // Name (users write: tags.name)
    "tags",            // Target table
    "activity_tags",   // Junction table
    "activity_id",     // FK to activities
    "tag_id",          // FK to tags
))
```

**Usage:**
```bash
# Filter by tag name
GET /activities?filter[tags.name]=cardio

# Search tag names
GET /activities?search[tags.name]=run

# Order by tag name
GET /activities?order[tags.name]=ASC

# Filter by tag ID
GET /activities?filter[tags.id]=123
```

#### 2. Many-to-One (Activities → Users)

**Definition:**
```go
registry.Register(query.ManyToOneRelationship(
    "user",      // Name (users write: user.username)
    "users",     // Target table
    "user_id",   // FK in activities table
))
```

**Usage:**
```bash
# Filter by username
GET /activities?filter[user.username]=john

# Search by user email
GET /activities?search[user.email]=@example.com

# Order by user creation date
GET /activities?order[user.created_at]=DESC
```

#### 3. One-to-Many (Users → Activities)

**Definition (in UserRepository):**
```go
registry.Register(query.OneToManyRelationship(
    "activities",       // Name
    "activities",       // Target table
    "user_id",          // FK in activities table
))
```

**Usage:**
```bash
# Find users who do running
GET /users?filter[activities.activity_type]=running

# Search users by activity titles
GET /users?search[activities.title]=marathon

# Order users by latest activity
GET /users?order[activities.created_at]=DESC
```

### Complex Relationship Queries

**Multiple relationships:**
```bash
GET /activities?filter[tags.name]=cardio&filter[user.username]=john&search[tags.name]=run&order[tags.name]=ASC
```

**Generated SQL:**
```sql
SELECT activities.*
FROM activities
LEFT JOIN activity_tags ON activity_tags.activity_id = activities.id
LEFT JOIN tags ON tags.id = activity_tags.tag_id
LEFT JOIN users ON users.id = activities.user_id
WHERE activities.user_id = $1
  AND tags.name = $2
  AND users.username = $3
  AND tags.name ILIKE '%run%'
ORDER BY tags.name ASC
```

### Why Natural Column Names?

**Before (Manual aliases):**
```bash
# User had to know internal aliases
GET /activities?filter[tags]=cardio
# Behind the scenes: translates to t.name
```

**After (Natural names):**
```bash
# User writes what they mean
GET /activities?filter[tags.name]=cardio
# System handles everything automatically
```

**Benefits:**
- ✅ Intuitive API
- ✅ Self-documenting
- ✅ No manual translation needed
- ✅ Works for any relationship depth
- ✅ 90% less code per entity

---

## Operator Filtering

### Available Operators

**Equality Operators:**
- `eq` - Equal to (default behavior)
- `ne` - Not equal to

**Comparison Operators:**
- `gt` - Greater than (exclusive)
- `gte` - Greater than or equal (inclusive)
- `lt` - Less than (exclusive)
- `lte` - Less than or equal (inclusive)

### Use Cases

**Date Ranges (inclusive):**
```bash
# January 2024 (includes Jan 1 and Jan 31)
GET /activities?filter[activity_date][gte]=2024-01-01&filter[activity_date][lte]=2024-01-31
```

**Date Ranges (exclusive):**
```bash
# Excludes boundaries
GET /activities?filter[activity_date][gt]=2024-01-01&filter[activity_date][lt]=2024-01-31
```

**Numeric Ranges:**
```bash
# Distances between 5km and 15km
GET /activities?filter[distance_km][gte]=5.0&filter[distance_km][lte]=15.0

# Duration less than 60 minutes
GET /activities?filter[duration_minutes][lt]=60
```

**Exclusion:**
```bash
# All activities except running
GET /activities?filter[activity_type][ne]=running

# Exclude specific tag
GET /activities?filter[tags.name][ne]=yoga
```

**Mixed Syntax:**
```bash
# Combine old and new syntax
GET /activities?filter[activity_type]=running&filter[activity_date][gte]=2024-01-01&filter[distance_km][gte]=5.0
```

### Operator Whitelisting

For security, each column can restrict which operators are allowed:

```go
operatorWhitelists := query.OperatorWhitelist{
    // Date/timestamp columns: Allow all comparison operators
    "activity_date": query.ComparisonOperators(), // eq, ne, gt, gte, lt, lte

    // Numeric columns: Allow all comparison operators
    "distance_km":      query.ComparisonOperators(),
    "duration_minutes": query.ComparisonOperators(),

    // String columns: Equality only
    "activity_type": query.EqualityOperators(), // eq, ne

    // ID columns: Strict equality only
    "user_id": query.StrictEqualityOnly(), // eq only

    // Relationship columns
    "tags.name":      query.EqualityOperators(), // eq, ne
    "user.username":  query.EqualityOperators(), // eq, ne
}
```

**Helper Functions:**
- `query.AllOperators()` - All 6 operators
- `query.ComparisonOperators()` - eq, ne, gt, gte, lt, lte
- `query.EqualityOperators()` - eq, ne
- `query.StrictEqualityOnly()` - eq only

---

## Security & Whitelisting

### Why Security Matters

**Without whitelisting:**
```bash
# Attacker tries to access password hash
GET /users?filter[password_hash]=...

# Without whitelist, this could execute:
SELECT * FROM users WHERE password_hash = '...'
```

**With whitelisting:**
```bash
# Same attack attempt
GET /users?filter[password_hash]=...

# Returns 400 Bad Request:
{
  "error": "Invalid query fields",
  "details": "column 'password_hash' is not allowed for filtering"
}
```

### Multi-Layered Security

ActiveLog implements **5 security layers**:

1. **Column Whitelisting** - Only approved columns can be queried
2. **Operator Whitelisting** - Per-column operator restrictions
3. **Parameterized Queries** - Prevents SQL injection via Squirrel
4. **Multi-tenancy Filtering** - Automatic user_id filtering
5. **Type Safety** - Go generics ensure compile-time checking

### Whitelist Configuration

**Define whitelists per endpoint:**

```go
func (h *ActivityHandler) ListActivities(w http.ResponseWriter, r *http.Request) {
    queryOpts, err := query.ParseQueryParams(r.URL.Query())
    if err != nil {
        response.Error(w, http.StatusBadRequest, "Invalid query parameters", err)
        return
    }

    // 1. Column whitelisting
    allowedFilters := []string{
        "activity_type", "duration_minutes", "distance_km", "activity_date",
        "tags.name", "tags.id", "user.username",  // Natural column names
    }
    allowedSearch := []string{"title", "description", "tags.name"}
    allowedOrder := []string{"created_at", "activity_date", "distance_km", "tags.name"}

    // 2. Operator whitelisting (v1.1.0+)
    operatorWhitelists := query.OperatorWhitelist{
        "activity_date":    query.ComparisonOperators(),
        "distance_km":      query.ComparisonOperators(),
        "duration_minutes": query.ComparisonOperators(),
        "activity_type":    query.EqualityOperators(),
        "tags.name":        query.EqualityOperators(),
    }

    // 3. Validate columns
    if err := query.ValidateQueryOptions(queryOpts, allowedFilters, allowedSearch, allowedOrder); err != nil {
        response.Error(w, http.StatusBadRequest, "Invalid query fields", err)
        return
    }

    // 4. Validate operators
    if err := query.ValidateFilterConditions(queryOpts, allowedFilters, operatorWhitelists); err != nil {
        response.Error(w, http.StatusBadRequest, "Invalid operator", err)
        return
    }

    // 5. Multi-tenancy: Add user_id filter in use case
    result, err := h.broker.RunUseCases(...)
}
```

### Security Best Practices

**1. Never expose sensitive columns:**
```go
// ❌ BAD
allowedFilters := []string{"username", "password_hash"}

// ✅ GOOD
allowedFilters := []string{"username", "email", "is_active"}
```

**2. Always validate before executing:**
```go
// Always validate
if err := query.ValidateQueryOptions(opts, allowedFilters, allowedSearch, allowedOrder); err != nil {
    return response.Error(w, http.StatusBadRequest, "Invalid query fields", err)
}
```

**3. Add multi-tenancy filtering:**
```go
// In use case: Always filter by authenticated user
queryOpts.Filter["user_id"] = getUserIDFromContext(ctx)
```

**4. Set maximum page size:**
```go
if opts.Limit > 100 {
    opts.Limit = 100 // Cap at 100 items
}
```

**5. Use operator whitelists:**
```go
// Prevent inappropriate operators on ID columns
operatorWhitelists := query.OperatorWhitelist{
    "user_id": query.StrictEqualityOnly(), // Only eq allowed
}
```

### What Gets Rejected

```bash
# ❌ REJECTED: user_id cannot use gt operator
GET /activities?filter[user_id][gt]=100
# Error: "operator 'gt' is not allowed for column 'user_id'"

# ❌ REJECTED: password_hash not in whitelist
GET /users?filter[password_hash]=...
# Error: "column 'password_hash' is not allowed for filtering"

# ✅ ALLOWED: activity_date can use gte
GET /activities?filter[activity_date][gte]=2024-01-01

# ✅ ALLOWED: tags.name is whitelisted
GET /activities?filter[tags.name]=cardio
```

---

## For Developers: Implementation Guide

This section shows how to add dynamic filtering with auto-JOINs to a new entity.

### Step 1: Register Relationships in Repository

```go
// File: internal/repository/activity_repository.go

import "github.com/valentinesamuel/activelog/pkg/query"

type ActivityRepository struct {
    db       DBConn
    tagRepo  *TagRepository
    registry *query.RelationshipRegistry  // Add this
}

func NewActivityRepository(db DBConn, tagRepo *TagRepository) *ActivityRepository {
    // Create registry for activities table
    registry := query.NewRelationshipRegistry("activities")

    // Register many-to-many: activities <-> tags
    registry.Register(query.ManyToManyRelationship(
        "tags",            // Relationship name (users write: tags.name)
        "tags",            // Target table
        "activity_tags",   // Junction table
        "activity_id",     // Junction FK to activities
        "tag_id",          // Junction FK to tags
    ))

    // Register many-to-one: activities -> users
    registry.Register(query.ManyToOneRelationship(
        "user",           // Relationship name (users write: user.username)
        "users",          // Target table
        "user_id",        // Foreign key in activities
    ))

    return &ActivityRepository{
        db:       db,
        tagRepo:  tagRepo,
        registry: registry,
    }
}
```

### Step 2: Use Auto-Joins in Repository Method

```go
func (ar *ActivityRepository) ListActivitiesWithQuery(
    ctx context.Context,
    opts *query.QueryOptions,
) (*query.PaginatedResult, error) {
    // Auto-generate JOINs from column names - that's it!
    joins := ar.registry.GenerateJoins(opts)

    return FindAndPaginate[models.Activity](
        ctx,
        ar.db,
        "activities",
        opts,
        ar.scanActivity,
        joins...,
    )
}
```

### Step 3: Update Handler Whitelists

```go
// File: internal/handlers/activity_v2.go

func (h *ActivityHandlerV2) ListActivities(w http.ResponseWriter, r *http.Request) {
    queryOpts, err := query.ParseQueryParams(r.URL.Query())
    if err != nil {
        response.Error(w, http.StatusBadRequest, "Invalid query parameters", err)
        return
    }

    // Define whitelists with natural column names
    allowedFilters := []string{
        // Direct columns
        "activity_type",
        "distance_km",
        "duration_minutes",
        "activity_date",

        // Relationship columns (natural names!)
        "tags.name",        // ✅ Natural: users write tags.name
        "tags.id",          // ✅ Can filter by tag ID too
        "user.username",    // ✅ Natural: join to users table
        "user.email",
    }

    allowedSearch := []string{
        "title",
        "description",
        "tags.name",        // ✅ Search tag names
        "user.username",    // ✅ Search username
    }

    allowedOrder := []string{
        "created_at",
        "activity_date",
        "distance_km",
        "tags.name",        // ✅ Order by tag name
        "user.username",    // ✅ Order by username
    }

    // Operator whitelisting (v1.1.0+)
    operatorWhitelists := query.OperatorWhitelist{
        "activity_date":    query.ComparisonOperators(),
        "distance_km":      query.ComparisonOperators(),
        "duration_minutes": query.ComparisonOperators(),
        "tags.name":        query.EqualityOperators(),  // eq, ne
        "user.username":    query.EqualityOperators(),
    }

    // Validate
    if err := query.ValidateQueryOptions(queryOpts, allowedFilters, allowedSearch, allowedOrder); err != nil {
        response.Error(w, http.StatusBadRequest, "Invalid query fields", err)
        return
    }

    if err := query.ValidateFilterConditions(queryOpts, allowedFilters, operatorWhitelists); err != nil {
        response.Error(w, http.StatusBadRequest, "Invalid operator", err)
        return
    }

    // Execute use case
    result, err := h.broker.RunUseCases(
        r.Context(),
        []broker.UseCase{h.listActivitiesUC},
        map[string]interface{}{
            "user_id":       getUserIDFromContext(r.Context()),
            "query_options": queryOpts,
        },
    )

    response.JSON(w, http.StatusOK, result["activities"])
}
```

### Complete Example: Adding to Users Entity

**1. Repository:**
```go
func NewUserRepository(db DBConn) *UserRepository {
    registry := query.NewRelationshipRegistry("users")

    // One user has many activities
    registry.Register(query.OneToManyRelationship(
        "activities",
        "activities",
        "user_id",
    ))

    return &UserRepository{db: db, registry: registry}
}

func (ur *UserRepository) ListUsersWithQuery(
    ctx context.Context,
    opts *query.QueryOptions,
) (*query.PaginatedResult, error) {
    joins := ur.registry.GenerateJoins(opts)
    return FindAndPaginate[models.User](ctx, ur.db, "users", opts, ur.scanUser, joins...)
}
```

**2. Handler:**
```go
allowedFilters := []string{
    "username",
    "email",
    "is_active",
    "activities.activity_type",  // ✅ Filter users by their activity types
}

allowedSearch := []string{
    "username",
    "email",
    "activities.title",          // ✅ Search users by activity titles
}

allowedOrder := []string{
    "created_at",
    "username",
    "activities.created_at",     // ✅ Order users by latest activity
}
```

**3. API Usage:**
```bash
# Find users who do running
GET /users?filter[activities.activity_type]=running

# Find users with "marathon" in activity titles
GET /users?search[activities.title]=marathon

# Order users by their latest activity
GET /users?order[activities.created_at]=DESC
```

---

## API Examples

### Activities Endpoint

**Available Filters:**
- Direct: `activity_type`, `duration_minutes`, `distance_km`, `activity_date`
- Relationships: `tags.name`, `tags.id`, `user.username`

**Available Search:**
- `title`, `description`, `tags.name`

**Available Order:**
- `created_at`, `activity_date`, `distance_km`, `tags.name`

#### Example 1: Simple Filter

```bash
GET /activities?filter[activity_type]=running
```

```json
{
  "data": [
    {
      "id": 1,
      "activityType": "running",
      "title": "Morning Run",
      "distanceKm": 10.5
    }
  ],
  "meta": {
    "page": 1,
    "limit": 10,
    "count": 1,
    "totalRecords": 1,
    "pageCount": 1
  }
}
```

#### Example 2: Relationship Filter

```bash
GET /activities?filter[tags.name]=cardio
```

**Auto-generated SQL:**
```sql
SELECT activities.*
FROM activities
LEFT JOIN activity_tags ON activity_tags.activity_id = activities.id
LEFT JOIN tags ON tags.id = activity_tags.tag_id
WHERE activities.user_id = $1 AND tags.name = $2
ORDER BY activities.created_at DESC
LIMIT 10
```

#### Example 3: Date Range

```bash
GET /activities?filter[activity_date][gte]=2024-01-01&filter[activity_date][lte]=2024-12-31
```

**Result:** All activities from 2024

#### Example 4: Complex Query

```bash
GET /activities?filter[activity_type]=running&filter[tags.name]=cardio&filter[distance_km][gte]=5.0&search[title]=morning&order[activity_date]=DESC&page=1&limit=20
```

**Result:** Running activities tagged "cardio" with distance >= 5km and "morning" in title, sorted by date, first 20 results

#### Example 5: Multiple Relationships

```bash
GET /activities?filter[tags.name]=cardio&filter[user.username]=john&order[tags.name]=ASC
```

**Auto-generated SQL:**
```sql
SELECT activities.*
FROM activities
LEFT JOIN activity_tags ON activity_tags.activity_id = activities.id
LEFT JOIN tags ON tags.id = activity_tags.tag_id
LEFT JOIN users ON users.id = activities.user_id
WHERE activities.user_id = $1
  AND tags.name = $2
  AND users.username = $3
ORDER BY tags.name ASC
```

### Tags Endpoint

```bash
# List all tags
GET /tags?limit=50

# Search tags
GET /tags?search[name]=cardio

# Sort alphabetically
GET /tags?order[name]=ASC
```

---

## Performance

### Database Indexes

**Ensure indexes exist for filtered/sorted columns:**

```sql
-- Activities
CREATE INDEX idx_activities_activity_type ON activities(activity_type);
CREATE INDEX idx_activities_created_at ON activities(created_at DESC);
CREATE INDEX idx_activities_activity_date ON activities(activity_date);
CREATE INDEX idx_activities_distance_km ON activities(distance_km);

-- Composite index for common filters
CREATE INDEX idx_activities_user_type ON activities(user_id, activity_type);

-- For JOINs
CREATE INDEX idx_activity_tags_activity ON activity_tags(activity_id);
CREATE INDEX idx_activity_tags_tag ON activity_tags(tag_id);
CREATE INDEX idx_tags_name ON tags(name);

-- For search (GIN index for better ILIKE performance)
CREATE INDEX idx_activities_title_gin ON activities USING gin(title gin_trgm_ops);
```

### Benchmark Results

| Operation | Time | Allocations |
|-----------|------|-------------|
| Simple Filter | 268 ns/op | 3 allocs/op |
| Multi-Filter | 445 ns/op | 5 allocs/op |
| Search | 587 ns/op | 6 allocs/op |
| Pagination | 215 ns/op | 2 allocs/op |
| Auto-JOIN | ~9% overhead | Minimal |

**Verdict:** Minimal overhead (~8-12%) for significantly improved flexibility

### Optimization Tips

**1. Use specific filters over search:**
```bash
# Faster: Exact filter
GET /activities?filter[activity_type]=running

# Slower: Full-text search
GET /activities?search[title]=run
```

**2. Limit page size:**
```bash
# Faster: Fetch 20 items
GET /activities?limit=20

# Slower: Fetch 1000 items
GET /activities?limit=1000
```

**3. Use indexed columns for sorting:**
```bash
# Faster: Indexed column
GET /activities?order[created_at]=DESC

# Slower: Non-indexed column
GET /activities?order[notes]=ASC
```

**4. Check query plans:**
```sql
EXPLAIN ANALYZE
SELECT activities.*
FROM activities
WHERE user_id = 1 AND activity_type = 'running'
ORDER BY activity_date DESC
LIMIT 10;
```

Look for `Index Scan` instead of `Seq Scan`.

---

## Troubleshooting

### Common Errors

#### "column 'X' is not allowed for filtering"

**Cause:** Column not in whitelist

**Fix:** Add column to appropriate whitelist
```go
allowedFilters := []string{"activity_type", "your_column"}
```

#### "operator 'X' is not allowed for column 'Y'"

**Cause:** Operator not whitelisted for that column

**Fix:** Add operator to whitelist
```go
operatorWhitelists := query.OperatorWhitelist{
    "your_column": query.ComparisonOperators(), // Allow all operators
}
```

#### "sql: expected N destination arguments in Scan, not M"

**Cause:** Scan function doesn't match table schema

**Fix:** Ensure scan function matches actual table columns exactly
```go
func (r *Repository) scanEntity(rows *sql.Rows) (*models.Entity, error) {
    var entity models.Entity
    err := rows.Scan(
        &entity.ID,
        &entity.Name,
        &entity.CreatedAt,
        // Must match table schema EXACTLY
    )
    return &entity, err
}
```

### Debugging Tips

**1. Enable SQL logging:**
```go
// In database/logger.go
logger.SetLevel(logrus.DebugLevel)
```

**2. Check generated SQL:**
```go
// Add temporary debug logging
sql, args, err := qb.baseQuery.PlaceholderFormat(sq.Dollar).ToSql()
log.Printf("Generated SQL: %s", sql)
log.Printf("With args: %v", args)
```

**3. Test query in PostgreSQL:**
```sql
-- Copy generated SQL from logs
-- Replace $1, $2 with actual values
SELECT activities.*
FROM activities
WHERE user_id = 1 AND activity_type = 'running'
ORDER BY activity_date DESC
LIMIT 10;
```

**4. Verify indexes:**
```sql
-- Check if indexes exist
\d activities

-- Check index usage
EXPLAIN ANALYZE SELECT * FROM activities WHERE activity_type = 'running';
```

### Performance Issues

**Slow queries?**

1. **Check EXPLAIN ANALYZE**
   ```sql
   EXPLAIN ANALYZE
   SELECT * FROM activities WHERE activity_type = 'running';
   ```

2. **Look for Seq Scan (bad) vs Index Scan (good)**

3. **Add missing indexes**
   ```sql
   CREATE INDEX idx_activities_activity_type ON activities(activity_type);
   ```

4. **Consider composite indexes for common filters**
   ```sql
   CREATE INDEX idx_activities_user_type ON activities(user_id, activity_type);
   ```

---

## Architecture Overview

### How It Works

```
┌─────────────────────────────────────────────────────────────────┐
│                     HTTP REQUEST                                 │
│  /activities?filter[tags.name]=cardio&order[date]=DESC          │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         ↓
┌─────────────────────────────────────────────────────────────────┐
│                 HANDLER (Security Layer)                         │
│  • Parse query params → QueryOptions                            │
│  • Validate against whitelist                                   │
│  • Validate operators                                           │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         ↓
┌─────────────────────────────────────────────────────────────────┐
│                  USE CASE (Business Logic)                       │
│  • Add system filters (user_id for multi-tenancy)               │
│  • Pass QueryOptions to repository                              │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         ↓
┌─────────────────────────────────────────────────────────────────┐
│              REPOSITORY (Auto-JOIN Detection)                    │
│  • registry.GenerateJoins(opts)                                 │
│    - Scans for dot notation (tags.name)                         │
│    - Looks up relationship config                               │
│    - Generates appropriate JOINs                                │
│  • FindAndPaginate[T]() - Generic query execution              │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         ↓
┌─────────────────────────────────────────────────────────────────┐
│                  QUERY BUILDER                                   │
│  • Builds SQL from QueryOptions                                 │
│  • Adds JOINs, WHERE, ORDER BY, LIMIT                           │
│  • Generates parameterized queries                              │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         ↓
┌─────────────────────────────────────────────────────────────────┐
│                  DATABASE (PostgreSQL)                           │
│  • Executes safe parameterized query                            │
│  • Returns results                                               │
└─────────────────────────────────────────────────────────────────┘
```

### Core Components

**1. QueryOptions (`pkg/query/types.go`)**
```go
type QueryOptions struct {
    Page             int
    Limit            int
    Filter           map[string]interface{}
    FilterConditions []FilterCondition  // v1.1.0+ operators
    FilterOr         map[string]interface{}
    Search           map[string]interface{}
    Order            map[string]string
}
```

**2. RelationshipRegistry (`pkg/query/relationships.go`)**
```go
type RelationshipRegistry struct {
    baseTable     string
    relationships map[string]*Relationship
}

func (r *RelationshipRegistry) GenerateJoins(opts *QueryOptions) []JoinConfig
```

**3. QueryBuilder (`pkg/query/builder.go`)**
- Builds SQL using Squirrel library
- Automatic column qualification for JOINs
- Parameterized queries (SQL injection safe)

**4. Generic Repository (`internal/repository/base_repository.go`)**
```go
func FindAndPaginate[T any](
    ctx context.Context,
    db DBConn,
    tableName string,
    opts *QueryOptions,
    scanFunc func(*sql.Rows) (*T, error),
    joins ...JoinConfig,
) (*PaginatedResult, error)
```

---

## Summary

ActiveLog's query system provides:

✅ **Intuitive API** - Natural column names match database schema
✅ **Auto-JOINs** - Zero-configuration relationship filtering
✅ **Powerful Operators** - Full range filtering with 6 operators
✅ **Secure** - Multi-layered security with whitelisting
✅ **Performant** - Minimal overhead with automatic indexing
✅ **Type-Safe** - Go generics for compile-time safety
✅ **Reusable** - Generic pattern works across all entities
✅ **Testable** - Comprehensive test coverage

**For more details, see:**
- [Architecture Documentation](./ARCHITECTURE.md)
- [Legacy Code Reference](./LEGACY_CODE.md)
- [Project README](../README.md)

---

**Last Updated:** 2026-01-08
**Version:** 2.0.0
