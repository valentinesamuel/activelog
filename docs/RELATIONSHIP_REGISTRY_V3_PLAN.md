# RelationshipRegistry v3.0 - Advanced Features Plan

**Created:** 2026-01-08
**Status:** Planning
**Goal:** Extend RelationshipRegistry to handle advanced relationship patterns (the 10%)

---

## Table of Contents

1. [Overview](#overview)
2. [Current Limitations](#current-limitations)
3. [Proposed Enhancements](#proposed-enhancements)
4. [API Design](#api-design)
5. [Implementation Plan](#implementation-plan)
6. [Developer Experience](#developer-experience)
7. [Backward Compatibility](#backward-compatibility)

---

## Overview

**Current State (v2.0):**
- ✅ Handles 90% of use cases (standard FK relationships)
- ✅ Auto-JOIN detection via dot notation
- ✅ 1-level deep relationships work perfectly

**Target State (v3.0):**
- ✅ All of v2.0 features
- ✅ Deep nesting (unlimited levels)
- ✅ Self-referential relationships
- ✅ Polymorphic relationships
- ✅ Complex join conditions
- ✅ Relationship path resolution
- ✅ Cycle detection

---

## Current Limitations

### 1. Deep Nesting (3+ Levels)

**What doesn't work:**
```bash
# User -> Company -> Department -> Country
GET /posts?filter[user.company.department.country]=USA
```

**Why it fails:**
- RelationshipRegistry only resolves 1 level deep
- No recursive relationship traversal

**Use cases:**
- Organizational hierarchies
- Multi-level foreign keys
- Nested data structures

---

### 2. Self-Referential Relationships

**What doesn't work:**
```go
// Comments table:
// id, content, parent_id (references comments.id)

registry.Register(query.SelfReferentialRelationship(
    "parent",
    "comments",
    "parent_id",
))
```

**Why it fails:**
- Table name collision (can't JOIN `comments` twice)
- No alias generation for self-references

**Use cases:**
- Comment threads (parent/child)
- Category trees
- Employee hierarchies (manager_id)
- File/folder structures

---

### 3. Polymorphic Relationships

**What doesn't work:**
```go
// Comments can belong to Posts OR Activities
// Table: comments (id, content, commentable_id, commentable_type)

registry.Register(query.PolymorphicRelationship(
    "commentable",
    "commentable_type", // Type discriminator column
    "commentable_id",   // ID column
    map[string]string{
        "Post":     "posts",
        "Activity": "activities",
    },
))
```

**Why it fails:**
- No polymorphic relationship type exists
- No type discrimination logic

**Use cases:**
- Comments on multiple models
- Attachments on various entities
- Activity feeds (polymorphic subjects)
- Taggable interface

---

### 4. Complex Join Conditions

**What doesn't work:**
```go
registry.Register(query.ManyToManyRelationship(
    "tags",
    "tags",
    "post_tags",
    "post_id",
    "tag_id",
).WithCondition("tags.is_active = true")) // Additional WHERE clause
```

**Why it fails:**
- JOINs only support simple FK equality
- No support for additional conditions

**Use cases:**
- Soft deletes (WHERE deleted_at IS NULL)
- Active/inactive records
- Tenant filtering in JOINs
- Date range conditions

---

## Proposed Enhancements

### Feature 1: Deep Nesting Support

**API Design:**
```go
// Register nested relationships
func NewUserRepository(db DBConn) *UserRepository {
    registry := query.NewRelationshipRegistry("users")

    // Register company relationship
    registry.Register(query.ManyToOneRelationship("company", "companies", "company_id"))

    return &UserRepository{db: db, registry: registry}
}

func NewCompanyRepository(db DBConn) *CompanyRepository {
    registry := query.NewRelationshipRegistry("companies")

    // Register department relationship
    registry.Register(query.ManyToOneRelationship("department", "departments", "department_id"))

    return &CompanyRepository{db: db, registry: registry}
}

// CROSS-REGISTRY RESOLUTION (new feature)
// When parsing "user.company.department.name", the system:
// 1. Starts at UserRegistry, finds "company" -> companies table
// 2. Loads CompanyRegistry, finds "department" -> departments table
// 3. Continues until full path resolved
```

**Developer Experience:**
```bash
# Just works - no extra code needed!
GET /posts?filter[user.company.department.name]=Engineering
```

**Implementation Approach:**
```go
// Registry Manager (new)
type RegistryManager struct {
    registries map[string]*RelationshipRegistry
}

func (rm *RegistryManager) Register(tableName string, registry *RelationshipRegistry)
func (rm *RegistryManager) ResolvePath(startTable string, path string) ([]JoinConfig, error)

// Example path: "user.company.department.name"
// Returns: [
//   {Table: "companies", Condition: "companies.id = users.company_id"},
//   {Table: "departments", Condition: "departments.id = companies.department_id"},
// ]
```

**Complexity:** Medium
**Impact:** High - unlocks complex data models

---

### Feature 2: Self-Referential Relationships

**API Design:**
```go
func NewCommentRepository(db DBConn) *CommentRepository {
    registry := query.NewRelationshipRegistry("comments")

    // Self-referential relationship with automatic aliasing
    registry.Register(query.SelfReferentialRelationship(
        "parent",           // Relationship name
        "comments",         // Same table
        "parent_id",        // FK to self
        query.SelfRefConfig{
            Alias: "parent_comment", // Auto-generated if not provided
            MaxDepth: 5,             // Prevent infinite recursion
        },
    ))

    return &CommentRepository{db: db, registry: registry}
}
```

**Developer Experience:**
```bash
# Filter by parent comment's author
GET /comments?filter[parent.author]=john

# Order by parent comment date
GET /comments?order[parent.created_at]=DESC
```

**Generated SQL:**
```sql
SELECT comments.*
FROM comments
LEFT JOIN comments AS parent_comment ON parent_comment.id = comments.parent_id
WHERE parent_comment.author = $1
ORDER BY parent_comment.created_at DESC
```

**Implementation Approach:**
```go
type SelfReferentialRelationship struct {
    Name      string
    Table     string
    ForeignKey string
    Alias     string // Auto-generated: "parent_comments", "child_comments"
    MaxDepth  int    // Default: 3
}

func (r *RelationshipRegistry) GenerateJoins(opts *QueryOptions) []JoinConfig {
    // Detect self-references
    if rel.Table == r.baseTable {
        return []JoinConfig{
            {
                Table:     rel.Table + " AS " + rel.Alias,
                Condition: rel.Alias + ".id = " + r.baseTable + "." + rel.ForeignKey,
            },
        }
    }
}
```

**Complexity:** Low-Medium
**Impact:** High - common pattern in many apps

---

### Feature 3: Polymorphic Relationships

**API Design:**
```go
func NewCommentRepository(db DBConn) *CommentRepository {
    registry := query.NewRelationshipRegistry("comments")

    // Polymorphic relationship
    registry.Register(query.PolymorphicRelationship(
        "commentable",              // Relationship name
        "commentable_type",         // Type discriminator column
        "commentable_id",           // ID column
        query.PolymorphicConfig{
            Types: map[string]string{
                "Post":     "posts",     // commentable_type = 'Post' -> JOIN posts
                "Activity": "activities", // commentable_type = 'Activity' -> JOIN activities
            },
            TypeColumn: "commentable_type", // Optional: already specified above
        },
    ))

    return &CommentRepository{db: db, registry: registry}
}
```

**Developer Experience:**
```bash
# Filter comments on posts by post title
GET /comments?filter[commentable.title]=Hello&filter[commentable_type]=Post

# Search across polymorphic targets
GET /comments?search[commentable.content]=workout
```

**Generated SQL:**
```sql
-- When commentable_type = 'Post'
SELECT comments.*
FROM comments
LEFT JOIN posts ON posts.id = comments.commentable_id
WHERE comments.commentable_type = 'Post'
  AND posts.title = $1

-- System intelligently chooses JOIN based on type filter
```

**Implementation Approach:**
```go
type PolymorphicRelationship struct {
    Name        string
    TypeColumn  string
    IDColumn    string
    Types       map[string]string // "Post" -> "posts"
}

func (r *RelationshipRegistry) GenerateJoins(opts *QueryOptions) []JoinConfig {
    // Detect polymorphic relationship
    if polyRel, ok := rel.(*PolymorphicRelationship); ok {
        // Check if type filter exists
        typeValue := opts.Filter[polyRel.TypeColumn] // e.g., "Post"

        if typeValue != "" {
            targetTable := polyRel.Types[typeValue] // "posts"
            return []JoinConfig{
                {
                    Table: targetTable,
                    Condition: targetTable + ".id = " + r.baseTable + "." + polyRel.IDColumn,
                    Where: r.baseTable + "." + polyRel.TypeColumn + " = '" + typeValue + "'",
                },
            }
        }

        // No type specified: can't JOIN (or UNION all types - advanced)
        return nil
    }
}
```

**Complexity:** Medium-High
**Impact:** Medium - specific use cases but very powerful

---

### Feature 4: Complex Join Conditions

**API Design:**
```go
func NewPostRepository(db DBConn) *PostRepository {
    registry := query.NewRelationshipRegistry("posts")

    // Standard relationship with additional conditions
    registry.Register(query.ManyToManyRelationship(
        "tags",
        "tags",
        "post_tags",
        "post_id",
        "tag_id",
    ).WithConditions(
        query.JoinCondition{
            Column:   "tags.is_active",
            Operator: "eq",
            Value:    true,
        },
        query.JoinCondition{
            Column:   "tags.deleted_at",
            Operator: "is_null",
        },
    ))

    return &PostRepository{db: db, registry: registry}
}
```

**Developer Experience:**
```bash
# Only joins active, non-deleted tags automatically
GET /posts?filter[tags.name]=golang
```

**Generated SQL:**
```sql
SELECT posts.*
FROM posts
LEFT JOIN post_tags ON post_tags.post_id = posts.id
LEFT JOIN tags ON tags.id = post_tags.tag_id
    AND tags.is_active = true
    AND tags.deleted_at IS NULL
WHERE tags.name = $1
```

**Implementation Approach:**
```go
type JoinCondition struct {
    Column   string      // "tags.is_active"
    Operator string      // "eq", "ne", "is_null", "is_not_null"
    Value    interface{} // true, false, nil
}

type Relationship interface {
    WithConditions(...JoinCondition) Relationship
}

func (r *ManyToManyRelationship) WithConditions(conditions ...JoinCondition) *ManyToManyRelationship {
    r.JoinConditions = conditions
    return r
}

func (r *RelationshipRegistry) GenerateJoins(opts *QueryOptions) []JoinConfig {
    // Append conditions to JOIN clause
    condition := baseCondition
    for _, jc := range rel.JoinConditions {
        condition += " AND " + jc.Column + " " + operatorToSQL(jc.Operator) + " " + valueToSQL(jc.Value)
    }
}
```

**Complexity:** Low-Medium
**Impact:** High - very common need (soft deletes, active records)

---

### Feature 5: Relationship Path Resolution

**API Design:**
```go
// Global registry manager (new singleton)
var GlobalRegistryManager = query.NewRegistryManager()

// Auto-register when creating repositories
func NewUserRepository(db DBConn) *UserRepository {
    registry := query.NewRelationshipRegistry("users")
    registry.Register(query.ManyToOneRelationship("company", "companies", "company_id"))

    // Auto-register with global manager
    GlobalRegistryManager.Register("users", registry)

    return &UserRepository{db: db, registry: registry}
}

// Or explicit registration in main.go
func main() {
    // Register all table registries
    query.GlobalRegistryManager.RegisterAll(map[string]*query.RelationshipRegistry{
        "users":      userRepo.GetRegistry(),
        "companies":  companyRepo.GetRegistry(),
        "departments": deptRepo.GetRegistry(),
    })
}
```

**Developer Experience:**
```bash
# Deep paths just work - system resolves across registries
GET /posts?filter[author.company.department.country]=USA
GET /comments?filter[post.author.company.name]=Acme Corp
```

**Implementation Approach:**
```go
type RegistryManager struct {
    registries map[string]*RelationshipRegistry
    mu         sync.RWMutex
}

func (rm *RegistryManager) ResolvePath(startTable string, path string) ([]JoinConfig, error) {
    parts := strings.Split(path, ".")

    joins := []JoinConfig{}
    currentTable := startTable

    for i := 0; i < len(parts)-1; i++ { // Last part is the column
        relationshipName := parts[i]

        // Get registry for current table
        registry, exists := rm.registries[currentTable]
        if !exists {
            return nil, fmt.Errorf("no registry for table: %s", currentTable)
        }

        // Find relationship
        rel, exists := registry.relationships[relationshipName]
        if !exists {
            return nil, fmt.Errorf("relationship %s not found in %s", relationshipName, currentTable)
        }

        // Generate JOIN for this level
        join := generateJoinForRelationship(currentTable, rel)
        joins = append(joins, join)

        // Move to next table
        currentTable = rel.GetTargetTable()
    }

    return joins, nil
}
```

**Complexity:** Medium
**Impact:** High - enables unlimited nesting

---

### Feature 6: Cycle Detection

**API Design:**
```go
// Automatically prevents infinite loops
registry := query.NewRelationshipRegistry("users")

registry.Register(query.ManyToOneRelationship("manager", "users", "manager_id"))
registry.Register(query.ManyToOneRelationship("department", "departments", "dept_id"))

// System detects cycle: users -> manager (users) -> manager (users) ...
// Throws error or limits depth automatically
```

**Implementation Approach:**
```go
func (rm *RegistryManager) ResolvePath(startTable string, path string) ([]JoinConfig, error) {
    visited := make(map[string]bool)
    depth := 0
    maxDepth := 10 // Configurable

    for _, relationshipName := range pathParts {
        if visited[currentTable+"."+relationshipName] {
            return nil, fmt.Errorf("cycle detected: %s.%s already visited", currentTable, relationshipName)
        }

        if depth > maxDepth {
            return nil, fmt.Errorf("max depth exceeded: %d", maxDepth)
        }

        visited[currentTable+"."+relationshipName] = true
        depth++

        // ... continue traversal
    }
}
```

**Complexity:** Low
**Impact:** Medium - safety feature

---

## API Design Summary

### Developer-Friendly Features

**1. Backward Compatible:**
```go
// v2.0 code still works exactly the same
registry.Register(query.ManyToOneRelationship("user", "users", "user_id"))
```

**2. Fluent API:**
```go
// Chain methods for readability
registry.Register(query.ManyToManyRelationship(
    "tags", "tags", "post_tags", "post_id", "tag_id",
).WithConditions(
    query.Active(),           // Built-in helpers
    query.NotDeleted(),
    query.Custom("tags.verified = true"),
))
```

**3. Auto-Configuration:**
```go
// System auto-detects common patterns
registry.Register(query.AutoRelationship("company")) // Assumes company_id FK
```

**4. Built-in Helpers:**
```go
// Common condition helpers
query.Active()                    // is_active = true
query.NotDeleted()               // deleted_at IS NULL
query.InTenant(tenantID)         // tenant_id = ?
query.DateRange(start, end)      // created_at BETWEEN ? AND ?
```

**5. Clear Error Messages:**
```go
// Instead of: "relationship not found"
// Show: "Relationship 'company' not found in 'users' registry. Did you mean 'companies'? Available: [department, manager]"
```

---

## Implementation Plan

### Phase 1: Foundation (Week 1)

**Goals:**
- Create RegistryManager
- Implement path resolution algorithm
- Add cycle detection

**Deliverables:**
- `pkg/query/registry_manager.go`
- `pkg/query/path_resolver.go`
- Unit tests for path resolution

**Acceptance Criteria:**
- [ ] Can resolve 2-level paths
- [ ] Detects cycles correctly
- [ ] Clear error messages

---

### Phase 2: Self-Referential (Week 2)

**Goals:**
- Implement SelfReferentialRelationship type
- Auto-alias generation
- Max depth limiting

**Deliverables:**
- `pkg/query/self_referential.go`
- Integration tests with comment threads
- Documentation

**Acceptance Criteria:**
- [ ] Can filter by parent.field
- [ ] Can order by parent.field
- [ ] Auto-generates unique aliases
- [ ] Respects max depth

---

### Phase 3: Complex Conditions (Week 2)

**Goals:**
- Add WithConditions() method
- Implement JoinCondition type
- Built-in helper functions

**Deliverables:**
- `pkg/query/join_conditions.go`
- Helper functions (Active, NotDeleted, etc.)
- Tests

**Acceptance Criteria:**
- [ ] Can add custom WHERE clauses to JOINs
- [ ] Helpers work correctly
- [ ] Conditions apply only to JOIN, not main query

---

### Phase 4: Deep Nesting (Week 3)

**Goals:**
- Cross-registry path resolution
- Unlimited depth support
- Performance optimization

**Deliverables:**
- Full RegistryManager implementation
- Cross-registry tests
- Benchmarks

**Acceptance Criteria:**
- [ ] Can resolve 5+ level paths
- [ ] Performance acceptable (<1ms for 5 levels)
- [ ] Clear errors for missing registries

---

### Phase 5: Polymorphic (Week 4)

**Goals:**
- Implement PolymorphicRelationship type
- Type discrimination logic
- Multi-JOIN support

**Deliverables:**
- `pkg/query/polymorphic.go`
- Tests with comments on multiple models
- Documentation

**Acceptance Criteria:**
- [ ] Can JOIN based on type discriminator
- [ ] Works with filter[commentable_type]=Post
- [ ] Generates correct SQL

---

### Phase 6: Documentation & Examples (Week 5)

**Goals:**
- Update QUERY_GUIDE.md
- Create migration guide
- Real-world examples

**Deliverables:**
- Updated documentation
- Example repositories
- Video/blog post

**Acceptance Criteria:**
- [ ] Developers can implement all features
- [ ] Migration path clear
- [ ] Examples cover common use cases

---

## Developer Experience

### Before (v2.0):
```bash
# Only 1 level works
✅ GET /posts?filter[user.username]=john
❌ GET /posts?filter[user.company.name]=Acme
```

### After (v3.0):
```bash
# Unlimited levels
✅ GET /posts?filter[user.company.department.country]=USA

# Self-referential
✅ GET /comments?filter[parent.author]=john

# Polymorphic
✅ GET /comments?filter[commentable.title]=Hello&filter[commentable_type]=Post

# Complex conditions (automatic)
✅ GET /posts?filter[tags.name]=active # Only active, non-deleted tags
```

### Code Comparison:

**v2.0 (limited):**
```go
// Can't do deep nesting
registry.Register(query.ManyToOneRelationship("company", "companies", "company_id"))
// ❌ Can't go deeper
```

**v3.0 (unlimited):**
```go
// Just register each level - system connects them!
// users repository
registry.Register(query.ManyToOneRelationship("company", "companies", "company_id"))

// companies repository
registry.Register(query.ManyToOneRelationship("department", "departments", "dept_id"))

// departments repository
registry.Register(query.ManyToOneRelationship("country", "countries", "country_id"))

// ✅ user.company.department.country just works!
```

---

## Backward Compatibility

### Guaranteed:
✅ All v2.0 code continues to work
✅ No breaking changes to existing API
✅ Optional features only

### Migration:
```go
// v2.0 code (still works)
registry := query.NewRelationshipRegistry("users")
registry.Register(query.ManyToOneRelationship("company", "companies", "company_id"))

// v3.0 enhancements (optional)
query.GlobalRegistryManager.Register("users", registry) // Enable cross-registry

// Use new features
registry.Register(query.ManyToOneRelationship("company", "companies", "company_id").
    WithConditions(query.Active()))
```

---

## Success Metrics

### Coverage:
- **Current:** 90% of use cases
- **Target:** 99% of use cases

### Developer Time:
- **Current:** 10 minutes per table
- **Target:** 10 minutes per table (same, but more powerful)

### Performance:
- **Current:** <1ms for 1-level joins
- **Target:** <5ms for 5-level joins

### Adoption:
- **Target:** 80% of projects can use without custom SQL

---

## Open Questions

1. **Global Registry Management:**
   - Auto-register on repository creation?
   - Or explicit registration in main.go?
   - **Decision:** Both options supported

2. **Polymorphic UNION:**
   - Support querying across all polymorphic types?
   - **Decision:** Phase 2 feature (v3.1)

3. **Performance:**
   - Cache compiled paths?
   - **Decision:** Yes, with TTL

4. **Validation:**
   - Validate registries on startup?
   - **Decision:** Optional, enabled in dev mode

---

## Next Steps

1. **Review this plan** - Get feedback from team
2. **Create POC** - Implement Phase 1 (RegistryManager)
3. **Benchmark** - Ensure performance acceptable
4. **Iterate** - Adjust based on learnings

---

## Summary

**What we're building:**
A RelationshipRegistry v3.0 that handles 99% of relationship patterns with the same easy API as v2.0.

**Why it matters:**
Developers can focus on building features, not writing JOIN logic.

**Timeline:**
5 weeks to production-ready v3.0

**Effort:**
Medium complexity, high impact

**Risk:**
Low - backward compatible, well-planned

---

**Questions?** Let's discuss the plan and prioritize features!
