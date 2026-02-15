# Transaction Patterns in ActiveLog

This guide explains how to use transactions in repository methods, allowing each method to decide whether to use a transaction or not.

## Table of Contents
- [Basic Concept](#basic-concept)
- [Pattern 1: Optional Transaction Parameter](#pattern-1-optional-transaction-parameter)
- [Pattern 2: Transaction Helper](#pattern-2-transaction-helper)
- [Examples](#examples)
- [Best Practices](#best-practices)

---

## Basic Concept

In ActiveLog, repository methods can work **both with and without transactions**:

- **READ operations** (Get, List) → Usually don't need transactions
- **WRITE operations** (Create, Update, Delete) → May need transactions
- **MULTI-STEP operations** (CreateWithTags) → Always need transactions

---

## Pattern 1: Optional Transaction Parameter

### Method Signature

```go
// Method accepts optional transaction
// If tx is nil, uses direct DB connection
func (r *Repository) Create(ctx context.Context, tx TxConn, entity *Model) error
```

### Implementation Example

```go
func (ar *ActivityRepository) Create(ctx context.Context, tx TxConn, activity *models.Activity) error {
    query := `
        INSERT INTO activities
        (user_id, activity_type, title, description, duration_minutes, distance_km, calories_burned, notes, activity_date)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
        RETURNING id, created_at, updated_at
    `

    // Use transaction if provided, otherwise use direct DB
    var row *sql.Row
    if tx != nil {
        row = tx.QueryRowContext(ctx, query, activity.UserID, activity.ActivityType, /*...*/)
    } else {
        row = ar.db.QueryRowContext(ctx, query, activity.UserID, activity.ActivityType, /*...*/)
    }

    return row.Scan(&activity.ID, &activity.CreatedAt, &activity.UpdatedAt)
}
```

### Usage

```go
// Without transaction (simple create)
err := activityRepo.Create(ctx, nil, activity)

// With transaction (part of larger operation)
tx, _ := db.BeginTx(ctx, nil)
err := activityRepo.Create(ctx, tx, activity)
err = tagRepo.LinkActivityTag(ctx, tx, activityID, tagID)
tx.Commit()
```

---

## Pattern 2: Transaction Helper

For cleaner code, use the `WithTransaction` helper and helper functions.

### Using Transaction Helpers

```go
func (ar *ActivityRepository) Create(ctx context.Context, tx TxConn, activity *models.Activity) error {
    query := `
        INSERT INTO activities
        (user_id, activity_type, title, description, duration_minutes, distance_km, calories_burned, notes, activity_date)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
        RETURNING id, created_at, updated_at
    `

    // Helper automatically chooses tx or db
    row := QueryRowInTx(ctx, tx, ar.db, query,
        activity.UserID, activity.ActivityType, activity.Title, activity.Description,
        activity.DurationMinutes, activity.DistanceKm, activity.CaloriesBurned,
        activity.Notes, activity.ActivityDate)

    return row.Scan(&activity.ID, &activity.CreatedAt, &activity.UpdatedAt)
}
```

**Benefits:**
- No `if tx != nil` checks
- Cleaner code
- Same signature for all methods

---

## Examples

### Example 1: Simple Read (No Transaction Needed)

```go
func (ar *ActivityRepository) GetByID(ctx context.Context, id int64) (*models.Activity, error) {
    query := `
        SELECT id, user_id, activity_type, title, description, duration_minutes, distance_km, calories_burned, notes, activity_date, created_at, updated_at
        FROM activities
        WHERE id = $1
    `

    activity := &models.Activity{}

    // No transaction needed for simple read
    err := ar.db.QueryRowContext(ctx, query, id).Scan(
        &activity.ID,
        &activity.UserID,
        &activity.ActivityType,
        // ... other fields
    )

    return activity, err
}
```

### Example 2: Simple Write (Optional Transaction)

```go
func (ar *ActivityRepository) Update(ctx context.Context, tx TxConn, id int, activity *models.Activity) error {
    query := `
        UPDATE activities
        SET activity_type = $1, title = $2, description = $3,
            duration_minutes = $4, distance_km = $5, calories_burned = $6,
            notes = $7, activity_date = $8, updated_at = CURRENT_TIMESTAMP
        WHERE id = $9 AND user_id = $10
        RETURNING updated_at
    `

    // Use helper - works with or without transaction
    row := QueryRowInTx(ctx, tx, ar.db, query,
        activity.ActivityType,
        activity.Title,
        activity.Description,
        activity.DurationMinutes,
        activity.DistanceKm,
        activity.CaloriesBurned,
        activity.Notes,
        activity.ActivityDate,
        id,
        activity.UserID,
    )

    return row.Scan(&activity.UpdatedAt)
}
```

**Usage:**

```go
// Without transaction
err := repo.Update(ctx, nil, activityID, activity)

// With transaction
err := WithTransaction(ctx, db, func(tx TxConn) error {
    if err := repo.Update(ctx, tx, activityID, activity); err != nil {
        return err
    }
    // Other operations...
    return nil
})
```

### Example 3: Multi-Step Operation (Transaction Required)

```go
func (ar *ActivityRepository) CreateWithTags(ctx context.Context, activity *models.Activity, tags []*models.Tag) error {
    // Use WithTransaction helper for multi-step operations
    return WithTransaction(ctx, ar.db, func(tx TxConn) error {
        // 1. Create activity (within transaction)
        activityQuery := `
            INSERT INTO activities
            (user_id, activity_type, title, description, duration_minutes, distance_km, calories_burned, notes, activity_date)
            VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
            RETURNING id, created_at, updated_at
        `
        row := QueryRowInTx(ctx, tx, ar.db, activityQuery,
            activity.UserID, activity.ActivityType, activity.Title, activity.Description,
            activity.DurationMinutes, activity.DistanceKm, activity.CaloriesBurned,
            activity.Notes, activity.ActivityDate)

        if err := row.Scan(&activity.ID, &activity.CreatedAt, &activity.UpdatedAt); err != nil {
            return fmt.Errorf("failed to insert activity: %w", err)
        }

        // 2. Create tags and link them (within same transaction)
        for _, tag := range tags {
            tagQuery := `
                INSERT INTO tags (name)
                VALUES ($1)
                ON CONFLICT (name) DO UPDATE
                SET name = EXCLUDED.name
                RETURNING id
            `
            var tagID int
            row := QueryRowInTx(ctx, tx, ar.db, tagQuery, tag.Name)
            if err := row.Scan(&tagID); err != nil {
                return fmt.Errorf("failed to create tag: %w", err)
            }

            linkQuery := `
                INSERT INTO activity_tags (tag_id, activity_id)
                VALUES ($1, $2)
                ON CONFLICT (tag_id, activity_id) DO NOTHING
            `
            _, err := ExecInTx(ctx, tx, ar.db, linkQuery, tagID, activity.ID)
            if err != nil {
                return fmt.Errorf("failed to link tag: %w", err)
            }
        }

        return nil // Commit happens automatically
    })
}
```

### Example 4: Complex Business Logic

```go
// Service layer orchestrating multiple repository operations
func (s *ActivityService) TransferActivityOwnership(ctx context.Context, activityID int64, newUserID int) error {
    return WithTransaction(ctx, s.db, func(tx TxConn) error {
        // 1. Get activity (read can use transaction for consistency)
        activity, err := s.activityRepo.GetByIDWithTx(ctx, tx, activityID)
        if err != nil {
            return err
        }

        // 2. Update activity owner
        activity.UserID = newUserID
        if err := s.activityRepo.Update(ctx, tx, int(activityID), activity); err != nil {
            return err
        }

        // 3. Log the transfer
        if err := s.auditRepo.LogTransfer(ctx, tx, activityID, newUserID); err != nil {
            return err
        }

        // 4. Update user statistics
        if err := s.userRepo.RecalculateStats(ctx, tx, newUserID); err != nil {
            return err
        }

        return nil // All or nothing
    })
}
```

---

## Best Practices

### 1. **Use Transactions for Multi-Step Operations**

✅ **DO:**
```go
func (r *Repo) ComplexOperation(ctx context.Context, data *Data) error {
    return WithTransaction(ctx, r.db, func(tx TxConn) error {
        // Multiple related operations
        if err := r.Step1(ctx, tx, data); err != nil {
            return err
        }
        if err := r.Step2(ctx, tx, data); err != nil {
            return err
        }
        return nil
    })
}
```

❌ **DON'T:**
```go
func (r *Repo) ComplexOperation(ctx context.Context, data *Data) error {
    // Without transaction - inconsistent if Step2 fails
    r.Step1(ctx, nil, data)
    r.Step2(ctx, nil, data) // If this fails, Step1 already committed!
}
```

### 2. **Don't Nest Transactions**

❌ **DON'T:**
```go
WithTransaction(ctx, db, func(tx TxConn) error {
    WithTransaction(ctx, db, func(tx2 TxConn) error { // Nested!
        // ...
    })
})
```

✅ **DO:**
```go
WithTransaction(ctx, db, func(tx TxConn) error {
    // Pass the same tx to all operations
    repo.Method1(ctx, tx, data)
    repo.Method2(ctx, tx, data)
})
```

### 3. **Keep Transactions Short**

✅ **DO:**
```go
// Quick transaction
WithTransaction(ctx, db, func(tx TxConn) error {
    repo.Create(ctx, tx, entity)
    repo.Link(ctx, tx, entityID, tagID)
    return nil
})
```

❌ **DON'T:**
```go
// Transaction held too long
WithTransaction(ctx, db, func(tx TxConn) error {
    repo.Create(ctx, tx, entity)
    time.Sleep(5 * time.Second)  // ❌ Holding locks!
    externalAPI.Call()           // ❌ Network call in transaction!
    repo.Update(ctx, tx, entity)
    return nil
})
```

### 4. **Make Transaction Parameter Optional**

✅ **DO:**
```go
// Flexible - works with or without transaction
func (r *Repo) Create(ctx context.Context, tx TxConn, entity *Entity) error {
    row := QueryRowInTx(ctx, tx, r.db, query, args...)
    return row.Scan(&entity.ID)
}
```

❌ **DON'T:**
```go
// Inflexible - always requires transaction
func (r *Repo) Create(ctx context.Context, tx TxConn, entity *Entity) error {
    if tx == nil {
        return errors.New("transaction required") // ❌ Too restrictive
    }
    // ...
}
```

### 5. **Return Errors, Let Caller Handle Rollback**

✅ **DO:**
```go
WithTransaction(ctx, db, func(tx TxConn) error {
    if err := repo.Create(ctx, tx, entity); err != nil {
        return err // Automatic rollback
    }
    if err := repo.Link(ctx, tx, id, tagID); err != nil {
        return err // Automatic rollback
    }
    return nil // Automatic commit
})
```

❌ **DON'T:**
```go
tx, _ := db.BeginTx(ctx, nil)
if err := repo.Create(ctx, tx, entity); err != nil {
    tx.Rollback()
    return err
}
if err := repo.Link(ctx, tx, id, tagID); err != nil {
    tx.Rollback() // Manual rollback - error-prone
    return err
}
tx.Commit()
```

---

## Decision Tree

```
Need Transaction?
│
├─ Single database operation?
│  └─ NO → Use tx = nil
│
├─ Multiple operations that must succeed/fail together?
│  └─ YES → Use WithTransaction
│
├─ Operation is part of larger workflow?
│  └─ YES → Accept tx parameter, let caller decide
│
└─ Simple read query?
   └─ NO → Use direct DB (no tx parameter needed)
```

---

## Migration Guide

### Refactoring Existing Methods

**Before:**
```go
func (r *Repo) Create(ctx context.Context, entity *Entity) error {
    row := r.db.QueryRowContext(ctx, query, args...)
    return row.Scan(&entity.ID)
}
```

**After:**
```go
func (r *Repo) Create(ctx context.Context, tx TxConn, entity *Entity) error {
    row := QueryRowInTx(ctx, tx, r.db, query, args...)
    return row.Scan(&entity.ID)
}
```

**Updating Callers:**
```go
// Old code still works with tx = nil
repo.Create(ctx, nil, entity)

// New code can use transactions
WithTransaction(ctx, db, func(tx TxConn) error {
    return repo.Create(ctx, tx, entity)
})
```

---

## Summary

✅ **Use `TxConn` parameter** for methods that might need transactions
✅ **Use helper functions** (`QueryRowInTx`, `ExecInTx`, `QueryInTx`) for cleaner code
✅ **Use `WithTransaction`** helper for multi-step operations
✅ **Pass `nil`** when transaction isn't needed
✅ **Keep transactions short** and focused
✅ **Return errors** immediately, let `WithTransaction` handle rollback/commit

This pattern gives you **maximum flexibility** while keeping the code clean and maintainable! 🎯
