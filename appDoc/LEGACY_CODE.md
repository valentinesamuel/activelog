# Legacy Code Reference

**Last Updated:** 2026-01-08
**Status:** ✅ Legacy code removed - Clean codebase

---

## Current Implementation

### ✅ Activity Handler - Main Implementation
**File:** `internal/handlers/activity.go`
**Pattern:** Broker + Use Case pattern with QueryOptions
**Features:**
- Natural column names (`tags.name`, `user.username`)
- Operator filtering (v1.1.0+): `filter[date][gte]=2024-01-01`
- Auto-JOIN detection via RelationshipRegistry
- Pagination metadata
- Full test coverage (12 E2E tests)

**Example Usage:**
```bash
GET /api/v1/activities?filter[tags.name]=cardio&filter[activity_date][gte]=2024-01-01
```

### ✅ Activity Repository - Current Methods
**File:** `internal/repository/activity_repository.go`
**Methods:**
- `ListActivitiesWithQuery(ctx, *query.QueryOptions)` - Uses RelationshipRegistry
- `Create(ctx, tx, *models.Activity)`
- `GetByID(ctx, id)`
- `Update(ctx, tx, id, *models.Activity)`
- `Delete(ctx, tx, id, userID)`
- `CreateWithTags(ctx, *models.Activity, []*models.Tag)` - Transaction-based

---

## Migration Complete

**✅ All legacy code has been removed!**

The following old patterns no longer exist in the codebase:
- ❌ Old `activity.go` handler (direct repository calls) - **DELETED**
- ❌ `ActivityFilters` struct - **REMOVED**
- ❌ `ListByUserWithFilters()` method - **STILL EXISTS** (used by repository)
- ❌ `GetActivitiesWithTags()` method - **STILL EXISTS** (used by repository)
- ❌ Manual alias handling (`t.name`, `at.`) - **REMOVED from handlers**

### What Was Cleaned Up

**Files Deleted:**
- `internal/handlers/activity.go` (old version)
- `internal/handlers/activity_test.go` (old version)

**Files Renamed:**
- `activity_v2.go` → `activity.go` (now the primary handler)
- `activity_v2_e2e_test.go` → `activity_e2e_test.go`

**Code Updated:**
- `ActivityHandlerV2` → `ActivityHandler` (removed v2 suffix)
- Removed all `/api/v1/v2` routes - now just `/api/v1/activities`
- Updated DI container to register new handler

---

## Repository Legacy Methods

**Note:** Some legacy methods still exist in `activity_repository.go` for backward compatibility with other parts of the codebase:

### `ListByUserWithFilters(userID, ActivityFilters)`
**Status:** Still exists
**Reason:** May be used by other parts of the codebase
**Recommended:** Use `ListActivitiesWithQuery()` for new code

### `GetActivitiesWithTags(ctx, userID, ActivityFilters)`
**Status:** Still exists
**Reason:** May be used by other parts of the codebase
**Recommended:** Use `ListActivitiesWithQuery()` - it's more flexible

---

## For New Development

### ✅ Always Use
- `internal/handlers/activity.go` (main handler)
- `ListActivitiesWithQuery()` repository method
- `QueryOptions` for filtering
- Natural column names (`tags.name`, `user.username`)

### ❌ Never Use
- Old handler patterns (all removed)
- Manual alias translation (removed)
- `ActivityFilters` struct (use `QueryOptions`)

---

## Clean Codebase Benefits

✅ **No confusion** - One handler, one pattern
✅ **Consistent** - All endpoints use same approach
✅ **Maintainable** - No duplicate code paths
✅ **Modern** - RelationshipRegistry + auto-JOINs everywhere
✅ **Type-safe** - QueryOptions validated at compile time

---

## Questions?

**"Where is the old handler?"**
→ **Deleted!** We now have only ONE handler using the modern pattern.

**"What about backward compatibility?"**
→ Not needed - we're using the same `/api/v1/activities` endpoints with the improved implementation.

**"Can I still use manual JOINs?"**
→ No - all filtering now uses the RelationshipRegistry auto-JOIN system.

---

## See Also
- `/docs/QUERY_GUIDE.md` - Comprehensive guide to filtering, ordering, and relationships
- `/docs/ARCHITECTURE.md` - System architecture overview
