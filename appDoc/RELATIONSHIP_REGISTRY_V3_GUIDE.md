# RelationshipRegistry v3.0 - Implementation Guide

**Version:** 3.0.0
**Date:** 2026-01-08
**Status:** Production Ready

---

## Overview

Relationship Registry v3.0 extends the auto-JOIN system to support advanced relationship patterns:

- ✅ **Deep Nesting** - Unlimited levels (`user.company.department.country`)
- ✅ **Self-Referential** - Tables referencing themselves (`parent.author`)
- ✅ **Polymorphic** - Multi-type relationships (`commentable → Post|Activity`)
- ✅ **Complex Conditions** - Additional WHERE clauses in JOINs
- ✅ **Cross-Registry Paths** - Automatic path resolution across tables
- ✅ **Cycle Detection** - Automatic duplicate JOIN prevention

**Backward Compatible:** All v2.0 code continues to work without changes.

---

## Table of Contents

1. [Quick Start](#quick-start)
2. [Feature 1: Deep Nesting](#feature-1-deep-nesting)
3. [Feature 2: Self-Referential Relationships](#feature-2-self-referential-relationships)
4. [Feature 3: Polymorphic Relationships](#feature-3-polymorphic-relationships)
5. [Feature 4: Complex JOIN Conditions](#feature-4-complex-join-conditions)
6. [Feature 5: Cross-Registry Resolution](#feature-5-cross-registry-resolution)
7. [API Reference](#api-reference)
8. [Migration from v2.0](#migration-from-v20)
9. [Best Practices](#best-practices)

---

## Quick Start

### What's New in v3.0?

**v2.0 (Before):**
```go
// Only 1-level relationships
GET /activities?filter[tags.name]=cardio  // ✅ Works
GET /activities?filter[user.company.name]=Acme  // ❌ Fails
```

**v3.0 (Now):**
```go
// Unlimited nesting
GET /activities?filter[tags.name]=cardio  // ✅ Works (backward compatible)
GET /activities?filter[user.company.name]=Acme  // ✅ Works (new!)
GET /posts?filter[user.company.department.country]=USA  // ✅ Works (new!)
GET /comments?filter[parent.author]=john  // ✅ Works (new!)
```

---

## Feature 1: Deep Nesting

### Problem Solved

Query relationships 3+ levels deep without manual JOIN logic.

### Code Example

```go
// 1. Setup RegistryManager (one time, in main/init)
manager := query.NewRegistryManager()

// 2. Register each table's relationships
activitiesRegistry := query.NewRelationshipRegistry("activities")
activitiesRegistry.Register(query.ManyToOneRelationship("user", "users", "user_id"))
manager.RegisterTable("activities", activitiesRegistry)

usersRegistry := query.NewRelationshipRegistry("users")
usersRegistry.Register(query.ManyToOneRelationship("company", "companies", "company_id"))
manager.RegisterTable("users", usersRegistry)

companiesRegistry := query.NewRelationshipRegistry("companies")
companiesRegistry.Register(query.ManyToOneRelationship("department", "departments", "department_id"))
manager.RegisterTable("companies", companiesRegistry)

departmentsRegistry := query.NewRelationshipRegistry("departments")
departmentsRegistry.Register(query.ManyToOneRelationship("country", "countries", "country_id"))
manager.RegisterTable("departments", departmentsRegistry)

// 3. Use in repository (exactly like v2.0)
func (ar *ActivityRepository) ListActivitiesWithQuery(ctx context.Context, opts *query.QueryOptions) (*query.PaginatedResult, error) {
    joins := ar.registry.GenerateJoins(opts)  // Auto-generates 4 JOINs!
    return FindAndPaginate[models.Activity](ctx, ar.db, "activities", opts, ar.scanActivity, joins...)
}
```

### API Usage

```bash
# Filter by nested relationship (4 levels deep)
GET /api/v1/activities?filter[user.company.department.country]=USA

# Order by nested field
GET /api/v1/activities?order[user.company.name]=ASC

# Search in nested relationship
GET /api/v1/activities?search[user.company.department.name]=engineering
```

### Generated SQL

```sql
SELECT activities.*
FROM activities
LEFT JOIN users ON users.id = activities.user_id
LEFT JOIN companies ON companies.id = users.company_id
LEFT JOIN departments ON departments.id = companies.department_id
LEFT JOIN countries ON countries.id = departments.country_id
WHERE countries.name = $1
```

### Key Benefits

- **No Code Changes** - Just register relationships in each registry
- **Automatic Path Resolution** - System traverses registries to build full path
- **Deduplication** - Multiple references to same path generate JOIN only once

---

## Feature 2: Self-Referential Relationships

### Problem Solved

Query tables that reference themselves (comments with parent comments, employees with managers).

### Code Example

```go
// Setup: Comments table with parent_id referencing other comments
commentsRegistry := query.NewRelationshipRegistry("comments")

commentsRegistry.Register(query.SelfReferentialRelationship(
    "parent",           // Relationship name
    "comments",         // Same table
    "parent_id",        // FK to self
    3,                  // Max depth (prevents infinite loops)
))

// Use in handler whitelist
allowedFilters := []string{
    "content",
    "author",
    "parent.content",   // Filter by parent comment content
    "parent.author",    // Filter by parent comment author
}
```

### API Usage

```bash
# Filter by parent comment's author
GET /api/v1/comments?filter[parent.author]=john

# Order by parent comment date
GET /api/v1/comments?order[parent.created_at]=DESC

# Search in parent comments
GET /api/v1/comments?search[parent.content]=great
```

### Generated SQL

```sql
SELECT comments.*
FROM comments
LEFT JOIN comments AS parent_comments ON parent_comments.id = comments.parent_id
WHERE parent_comments.author = $1
ORDER BY parent_comments.created_at DESC
```

### Key Benefits

- **Automatic Aliasing** - System generates unique alias (`parent_comments`)
- **Cycle Prevention** - MaxDepth prevents infinite recursion
- **Natural Names** - Use `parent.author` instead of `pc.author`

---

## Feature 3: Polymorphic Relationships

### Problem Solved

Query relationships where a record can belong to multiple table types (comments on Posts OR Activities).

### Code Example

```go
// Setup: Comments can belong to Posts or Activities
commentsRegistry := query.NewRelationshipRegistry("comments")

commentsRegistry.Register(query.PolymorphicRelationship(
    "commentable",              // Relationship name
    "commentable_type",         // Type discriminator column
    "commentable_id",           // ID column
    map[string]string{
        "Post":     "posts",     // Type "Post" → JOIN posts table
        "Activity": "activities", // Type "Activity" → JOIN activities table
    },
))

// Use in handler whitelist
allowedFilters := []string{
    "content",
    "commentable_type",      // REQUIRED for polymorphic filtering
    "commentable.title",     // Filter by target title
    "commentable.content",   // Filter by target content
}
```

### API Usage

```bash
# Filter comments on Posts with specific title
GET /api/v1/comments?filter[commentable_type]=Post&filter[commentable.title]=Hello

# Filter comments on Activities
GET /api/v1/comments?filter[commentable_type]=Activity&filter[commentable.content]=workout

# Search across polymorphic targets
GET /api/v1/comments?filter[commentable_type]=Post&search[commentable.title]=news
```

### Generated SQL

```sql
-- When commentable_type = 'Post'
SELECT comments.*
FROM comments
LEFT JOIN posts ON posts.id = comments.commentable_id
WHERE comments.commentable_type = 'Post'
  AND posts.title = $1

-- When commentable_type = 'Activity'
SELECT comments.*
FROM comments
LEFT JOIN activities ON activities.id = comments.commentable_id
WHERE comments.commentable_type = 'Activity'
  AND activities.content = $1
```

### Key Benefits

- **Type-Aware JOINs** - System chooses correct table based on type filter
- **Secure** - Requires type filter to prevent ambiguous queries
- **Flexible** - Easy to add new polymorphic types

### Important Notes

⚠️ **Polymorphic relationships REQUIRE a type filter** - Without `commentable_type=Post`, the system cannot determine which table to JOIN.

---

## Feature 4: Complex JOIN Conditions

### Problem Solved

Add additional WHERE clauses to JOIN conditions (soft deletes, active records, tenant filtering).

### Code Example

```go
// Setup: Only JOIN active tags, exclude soft-deleted
postsRegistry := query.NewRelationshipRegistry("posts")

rel := query.ManyToManyRelationship("tags", "tags", "post_tags", "post_id", "tag_id").
    WithConditions(
        query.AdditionalCondition{
            Column:   "tags.is_active",
            Operator: "eq",
            Value:    true,
        },
        query.AdditionalCondition{
            Column:   "tags.deleted_at",
            Operator: "eq",
            Value:    nil,
        },
    )

postsRegistry.Register(rel)
```

### API Usage

```bash
# Standard filtering - only active tags are joined
GET /api/v1/posts?filter[tags.name]=tech

# System automatically filters out:
# - Inactive tags (is_active = false)
# - Deleted tags (deleted_at IS NOT NULL)
```

### Generated SQL

```sql
SELECT posts.*
FROM posts
LEFT JOIN post_tags ON post_tags.post_id = posts.id
LEFT JOIN tags ON tags.id = post_tags.tag_id
    AND tags.is_active = true        -- ← Additional condition
    AND tags.deleted_at IS NULL      -- ← Additional condition
WHERE tags.name = $1
```

### Key Benefits

- **Declarative** - Conditions defined once, applied everywhere
- **Automatic** - No need to remember to filter soft deletes
- **Secure** - Users can't bypass soft delete filters

### Supported Operators

- `eq` - Equals
- `ne` - Not equals
- `gt` - Greater than
- `gte` - Greater than or equal
- `lt` - Less than
- `lte` - Less than or equal

---

## Feature 5: Cross-Registry Resolution

### Problem Solved

Automatically resolve relationship paths across multiple registries.

### Code Example

```go
// main.go or dependency injection setup
func SetupRegistries() *query.RegistryManager {
    manager := query.NewRegistryManager()

    // Register ALL table registries with manager
    activitiesRegistry := query.NewRelationshipRegistry("activities")
    activitiesRegistry.Register(query.ManyToOneRelationship("user", "users", "user_id"))
    manager.RegisterTable("activities", activitiesRegistry)

    usersRegistry := query.NewRelationshipRegistry("users")
    usersRegistry.Register(query.ManyToOneRelationship("company", "companies", "company_id"))
    manager.RegisterTable("users", usersRegistry)

    companiesRegistry := query.NewRelationshipRegistry("companies")
    companiesRegistry.Register(query.ManyToOneRelationship("department", "departments", "department_id"))
    manager.RegisterTable("companies", companiesRegistry)

    return manager
}

// Now paths like "user.company.department" are auto-resolved across registries!
```

### How It Works

1. **Path Parsing:** `user.company.department` → `["user", "company", "department"]`
2. **Registry Traversal:**
   - Start at `activities` registry → find `user` → points to `users` table
   - Load `users` registry → find `company` → points to `companies` table
   - Load `companies` registry → find `department` → points to `departments` table
3. **JOIN Generation:** Create 3 chained JOINs

### Key Benefits

- **Modular** - Each table's relationships defined independently
- **Scalable** - Add new registries without modifying existing ones
- **Maintainable** - Relationship definitions colocated with table logic

---

## API Reference

### New Relationship Types

```go
// SelfReferentialRelationship(name, table, foreignKey, maxDepth)
query.SelfReferentialRelationship("parent", "comments", "parent_id", 3)

// PolymorphicRelationship(name, typeColumn, idColumn, typeMap)
query.PolymorphicRelationship(
    "commentable",
    "commentable_type",
    "commentable_id",
    map[string]string{
        "Post": "posts",
        "Activity": "activities",
    },
)

// WithConditions() - Chainable method
rel := query.ManyToManyRelationship(...).WithConditions(
    query.AdditionalCondition{
        Column:   "tags.is_active",
        Operator: "eq",
        Value:    true,
    },
)
```

### RegistryManager

```go
// Create manager
manager := query.NewRegistryManager()

// Register table registries
manager.RegisterTable("activities", activitiesRegistry)
manager.RegisterTable("users", usersRegistry)

// Get registry (for debugging/testing)
registry, found := manager.GetRegistry("users")
```

### AdditionalCondition

```go
type AdditionalCondition struct {
    Column   string      // Column name (e.g., "tags.is_active")
    Operator string      // Operator: "eq", "ne", "gt", "gte", "lt", "lte"
    Value    interface{} // Value to compare
}
```

---

## Migration from v2.0

### Good News: Backward Compatible

**All v2.0 code works in v3.0 without changes!**

```go
// v2.0 code - still works
registry := query.NewRelationshipRegistry("activities")
registry.Register(query.ManyToManyRelationship("tags", "tags", "activity_tags", "activity_id", "tag_id"))

// Use exactly as before
joins := registry.GenerateJoins(opts)
```

### Opt-In New Features

To use v3.0 features:

**1. Deep Nesting:** Create RegistryManager and register all tables

```go
// Before (v2.0)
registry := query.NewRelationshipRegistry("activities")

// After (v3.0)
manager := query.NewRegistryManager()
activitiesRegistry := query.NewRelationshipRegistry("activities")
manager.RegisterTable("activities", activitiesRegistry)
```

**2. Self-Referential:** Use new relationship type

```go
// New in v3.0
registry.Register(query.SelfReferentialRelationship("parent", "comments", "parent_id", 3))
```

**3. Polymorphic:** Use new relationship type

```go
// New in v3.0
registry.Register(query.PolymorphicRelationship(
    "commentable", "commentable_type", "commentable_id",
    map[string]string{"Post": "posts", "Activity": "activities"},
))
```

**4. Complex Conditions:** Chain `.WithConditions()`

```go
// New in v3.0
rel := query.ManyToManyRelationship(...).WithConditions(
    query.AdditionalCondition{Column: "tags.is_active", Operator: "eq", Value: true},
)
registry.Register(rel)
```

---

## Best Practices

### 1. Registry Setup

**✅ DO:**
```go
// Centralize registry setup in main/init
func SetupRegistries() *query.RegistryManager {
    manager := query.NewRegistryManager()

    // Register all tables
    manager.RegisterTable("activities", setupActivitiesRegistry())
    manager.RegisterTable("users", setupUsersRegistry())
    // ...

    return manager
}
```

**❌ DON'T:**
```go
// Create manager per request - expensive!
func (h *Handler) ListActivities(w, r) {
    manager := query.NewRegistryManager() // ❌ Wrong!
    // ...
}
```

### 2. Whitelist Natural Names

**✅ DO:**
```go
allowedFilters := []string{
    "title",
    "user.username",                  // Natural
    "user.company.name",              // Natural
    "parent.author",                  // Natural
    "commentable.title",              // Natural
}
```

**❌ DON'T:**
```go
allowedFilters := []string{
    "u.username",                     // ❌ Manual alias
    "c.name",                         // ❌ Manual alias
}
```

### 3. Polymorphic Type Filters

**✅ DO:**
```go
// Require type filter for polymorphic relationships
operatorWhitelists := query.OperatorWhitelist{
    "commentable_type": query.StrictEqualityOnly(), // eq only
}
```

**❌ DON'T:**
```go
// Allow polymorphic filtering without type
filter[commentable.title]=Test  // ❌ Ambiguous!
```

### 4. Self-Referential Max Depth

**✅ DO:**
```go
// Set reasonable max depth
query.SelfReferentialRelationship("parent", "comments", "parent_id", 3)  // ✅ Max 3 levels
```

**❌ DON'T:**
```go
// Unlimited depth - risk of infinite loops
query.SelfReferentialRelationship("parent", "comments", "parent_id", 0)  // ❌ No limit!
```

### 5. Additional Conditions

**✅ DO:**
```go
// Use for global constraints (soft deletes, tenancy)
rel.WithConditions(
    query.AdditionalCondition{Column: "tags.deleted_at", Operator: "eq", Value: nil},
)
```

**❌ DON'T:**
```go
// Use for user-specific filters - those belong in query params
rel.WithConditions(
    query.AdditionalCondition{Column: "tags.user_id", Operator: "eq", Value: currentUser.ID},  // ❌ Wrong!
)
```

---

## Performance Considerations

### 1. JOIN Deduplication

v3.0 automatically prevents duplicate JOINs:

```go
// Query references tags 3 times
opts := &query.QueryOptions{
    Filter: map[string]interface{}{"tags.name": "tech"},
    Search: map[string]interface{}{"tags.description": "coding"},
    Order: map[string]string{"tags.name": "ASC"},
}

// Only 2 JOINs generated (not 6!)
joins := registry.GenerateJoins(opts)  // len(joins) == 2
```

### 2. Registry Caching

RegistryManager caches registries in memory - cheap lookups.

### 3. Deep Nesting Cost

Each level adds one JOIN. Limit nesting depth in production:

```go
// ✅ Good: 2-3 levels
filter[user.company.name]=Acme

// ⚠️ Acceptable: 4-5 levels
filter[user.company.department.country.continent]=Europe

// ❌ Avoid: 6+ levels (performance impact)
filter[a.b.c.d.e.f.g]=value
```

---

## Troubleshooting

### Issue: "Unknown relationship 'company' in column 'user.company.name'"

**Cause:** Registry not registered with manager

**Solution:**
```go
manager.RegisterTable("users", usersRegistry)  // Don't forget this!
```

### Issue: Polymorphic JOIN not generated

**Cause:** Missing type filter

**Solution:**
```bash
# Add type filter
GET /comments?filter[commentable_type]=Post&filter[commentable.title]=Hello
```

### Issue: Infinite loop with self-referential

**Cause:** MaxDepth set to 0

**Solution:**
```go
query.SelfReferentialRelationship("parent", "comments", "parent_id", 3)  // Set maxDepth > 0
```

---

## What's Next?

✅ **v3.0 is production-ready** - All features implemented and tested

**Future Enhancements (v4.0?):**
- Recursive CTEs for true deep self-referential queries
- Automatic index suggestions
- Query plan visualization
- Performance benchmarking tools

---

## Summary

**v3.0 brings powerful features while maintaining backward compatibility:**

| Feature | v2.0 | v3.0 |
|---------|------|------|
| 1-level relationships | ✅ | ✅ |
| Deep nesting (3+ levels) | ❌ | ✅ |
| Self-referential | ❌ | ✅ |
| Polymorphic | ❌ | ✅ |
| Complex JOIN conditions | ❌ | ✅ |
| Cross-registry resolution | ❌ | ✅ |
| Automatic deduplication | ✅ | ✅ |

**Ready to use in production today!** 🚀

---

**Version:** 3.0.0
**Last Updated:** 2026-01-08
**Status:** Production Ready
