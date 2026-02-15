# Transaction Usage Examples

Quick reference for using optional transactions in ActiveLog repositories.

## Basic Usage

### 1. Simple Create (No Transaction)

```go
// Handler or service layer
activity := &models.Activity{
    UserID:          userID,
    ActivityType:    "running",
    Title:           "Morning Run",
    DurationMinutes: 30,
}

// Pass nil for tx - no transaction needed
err := activityRepo.Create(ctx, nil, activity)
```

### 2. Using WithTransaction Helper

```go
// Complex operation requiring multiple steps
err := repository.WithTransaction(ctx, db, func(tx repository.TxConn) error {
    // Create activity within transaction
    if err := activityRepo.Create(ctx, tx, activity); err != nil {
        return err
    }

    // Link tags within same transaction
    for _, tagID := range tagIDs {
        if err := tagRepo.LinkActivityTag(ctx, tx, activity.ID, tagID); err != nil {
            return err // Automatic rollback
        }
    }

    // Update user stats
    if err := userRepo.UpdateStats(ctx, tx, userID); err != nil {
        return err // Automatic rollback
    }

    return nil // Automatic commit
})
```

### 3. Built-in Multi-Step Method

```go
// CreateWithTags already uses WithTransaction internally
activity := &models.Activity{
    UserID:          userID,
    ActivityType:    "running",
    Title:           "Morning Run",
    DurationMinutes: 30,
}

tags := []*models.Tag{
    {Name: "outdoor"},
    {Name: "cardio"},
}

// Single call - transaction handled automatically
err := activityRepo.CreateWithTags(ctx, activity, tags)
```

## When to Use Transactions

### ✅ Use Transactions

- **Multiple related writes** that must succeed/fail together
- **Data consistency** requirements (e.g., transfer operation)
- **Foreign key relationships** (e.g., create parent + children)
- **Aggregation updates** (e.g., update counters after insert)

### ❌ Don't Use Transactions

- **Single read** operations
- **Single write** operations with no dependencies
- **Independent operations** that can fail separately
- **Long-running** operations (holds locks)

## Real-World Examples

### Example 1: Simple Update (No Transaction)

```go
// Handler or service layer - simple update
activity := &models.Activity{
    UserID:          userID,
    ActivityType:    "running",
    Title:           "Updated Morning Run",
    DurationMinutes: 45,
}

// Pass nil for tx - no transaction needed
err := activityRepo.Update(ctx, nil, activityID, activity)
```

### Example 2: Simple Delete (No Transaction)

```go
// Handler or service layer - simple delete
err := activityRepo.Delete(ctx, nil, activityID, userID)
```

### Example 3: Delete Activity with Cleanup (Transaction)

```go
func (s *ActivityService) DeleteActivity(ctx context.Context, activityID int64, userID int) error {
    return repository.WithTransaction(ctx, s.db, func(tx repository.TxConn) error {
        // 1. Delete activity tags
        if err := s.tagRepo.UnlinkAllTags(ctx, tx, activityID); err != nil {
            return err
        }

        // 2. Delete activity photos
        if err := s.photoRepo.DeleteByActivity(ctx, tx, activityID); err != nil {
            return err
        }

        // 3. Delete activity
        if err := s.activityRepo.Delete(ctx, tx, int(activityID), userID); err != nil {
            return err
        }

        // 4. Update user statistics
        if err := s.userRepo.RecalculateStats(ctx, tx, userID); err != nil {
            return err
        }

        return nil // All deletions succeed together or fail together
    })
}
```

### Example 4: Bulk Import

```go
func (s *ActivityService) BulkImport(ctx context.Context, userID int, activities []*models.Activity) error {
    return repository.WithTransaction(ctx, s.db, func(tx repository.TxConn) error {
        for _, activity := range activities {
            activity.UserID = userID

            // Create each activity within the transaction
            if err := s.activityRepo.Create(ctx, tx, activity); err != nil {
                return fmt.Errorf("failed to import activity %q: %w", activity.Title, err)
            }
        }

        // Update totals after all imports
        if err := s.userRepo.UpdateActivityCount(ctx, tx, userID); err != nil {
            return err
        }

        return nil // Either all activities import or none do
    })
}
```

### Example 5: Transfer Ownership

```go
func (s *ActivityService) TransferOwnership(ctx context.Context, activityID int64, fromUserID, toUserID int) error {
    return repository.WithTransaction(ctx, s.db, func(tx repository.TxConn) error {
        // 1. Verify current ownership
        activity, err := s.activityRepo.GetByIDWithTx(ctx, tx, activityID)
        if err != nil {
            return err
        }
        if activity.UserID != fromUserID {
            return errors.New("not the owner")
        }

        // 2. Update activity owner
        activity.UserID = toUserID
        if err := s.activityRepo.Update(ctx, tx, int(activityID), activity); err != nil {
            return err
        }

        // 3. Decrement old user's count
        if err := s.userRepo.DecrementActivityCount(ctx, tx, fromUserID); err != nil {
            return err
        }

        // 4. Increment new user's count
        if err := s.userRepo.IncrementActivityCount(ctx, tx, toUserID); err != nil {
            return err
        }

        // 5. Log transfer
        if err := s.auditRepo.LogTransfer(ctx, tx, activityID, fromUserID, toUserID); err != nil {
            return err
        }

        return nil
    })
}
```

## Testing with Transactions

### Test with No Transaction

```go
func TestActivityRepo_Create(t *testing.T) {
    db, loggingDB := setupTestDB(t)
    defer cleanupTestDB(t, db)

    repo := NewActivityRepository(loggingDB, nil)
    activity := &models.Activity{/*...*/}

    // Simple create without transaction
    err := repo.Create(context.Background(), nil, activity)
    assert.NoError(t, err)
    assert.NotZero(t, activity.ID)
}
```

### Test with Transaction

```go
func TestActivityRepo_CreateWithTags(t *testing.T) {
    db, loggingDB := setupTestDB(t)
    defer cleanupTestDB(t, db)

    repo := NewActivityRepository(loggingDB, tagRepo)
    activity := &models.Activity{/*...*/}
    tags := []*models.Tag{{Name: "test"}}

    // CreateWithTags uses WithTransaction internally
    err := repo.CreateWithTags(context.Background(), activity, tags)
    assert.NoError(t, err)
    assert.NotZero(t, activity.ID)

    // Verify tags were created
    activityTags, _ := tagRepo.GetTagsForActivity(context.Background(), activity.ID)
    assert.Len(t, activityTags, 1)
}
```

## Quick Reference

| Operation | Transaction? | Code |
|-----------|-------------|------|
| Single read | ❌ No | `repo.GetByID(ctx, id)` |
| Single write | ❌ No | `repo.Create(ctx, nil, entity)` |
| Multi-step write | ✅ Yes | `WithTransaction(ctx, db, func(tx) {...})` |
| Built-in multi-step | ✅ Yes (internal) | `repo.CreateWithTags(ctx, entity, tags)` |
| Conditional write | ⚠️ Depends | Use tx if part of larger operation |

## Common Patterns

### Pattern: Create + Link

```go
// ✅ Good - uses transaction
WithTransaction(ctx, db, func(tx TxConn) error {
    repo.Create(ctx, tx, entity)
    repo.Link(ctx, tx, entityID, relatedID)
})

// ❌ Bad - no transaction, inconsistent state possible
repo.Create(ctx, nil, entity)
repo.Link(ctx, nil, entityID, relatedID) // Could fail leaving orphan
```

### Pattern: Update + Recalculate

```go
// ✅ Good - atomic update
WithTransaction(ctx, db, func(tx TxConn) error {
    repo.Update(ctx, tx, entity)
    statsRepo.Recalculate(ctx, tx, userID)
})
```

### Pattern: Read + Write (Conditional)

```go
// ✅ Good - consistent read and write
WithTransaction(ctx, db, func(tx TxConn) error {
    entity, _ := repo.GetByID(ctx, tx, id) // Read with tx for consistency
    entity.Field = newValue
    repo.Update(ctx, tx, id, entity)
})
```

## Summary

- **Default**: Pass `nil` for `tx` parameter (no transaction)
- **Multi-step**: Use `WithTransaction` helper
- **Built-in**: Use methods like `CreateWithTags` that handle transactions internally
- **Testing**: Works the same way - pass `nil` or use `WithTransaction`

Happy coding! 🎯
