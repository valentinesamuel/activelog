# Plan: Standardized Response Interceptor

## Context

Currently, handlers write responses using three different patterns:
- `response.SendJSON(w, statusCode, data)` for success
- `response.Error(w, statusCode, message)` for errors
- Inline `json.NewEncoder(w).Encode(...)` for validation errors

These produce inconsistent shapes with no `statusCode`, `success`, `path`, or `duration` fields. The goal is to standardize every response into one of three envelopes:

**Success:** `{ statusCode, success: true, message: "Request successful", result: {...}, path, duration }`
**Error:** `{ statusCode, success: false, message, errors: [], path, duration }`
**Validation:** `{ statusCode: 400, success: false, message: "Bad Request", errors: [{field, errors:[]}], path, duration }`

> **Note on `"error"` vs `"errors"`:** The todo.md uses `"errors"` (plural, empty `[]`) for simple errors but `"error"` (singular) for validation errors. This plan uses `"errors"` (plural) consistently for API consistency.

---

## Implementation Steps

### Step 1 — Update `pkg/response/json.go`

Add to the file:
1. **Private context key type** to avoid collisions:
   ```go
   type contextKey int
   var RequestStartKey contextKey = 0
   ```
2. **`WithStartTime` helper** (used by timing middleware):
   ```go
   func WithStartTime(ctx context.Context, t time.Time) context.Context {
       return context.WithValue(ctx, RequestStartKey, t)
   }
   ```
3. **`ValidationErrorItem` struct**:
   ```go
   type ValidationErrorItem struct {
       Field  string   `json:"field"`
       Errors []string `json:"errors"`
   }
   ```
4. **Three new response helpers** (all read `r.URL.RequestURI()` for path, compute duration from context start time in milliseconds):
   - `Success(w, r, statusCode, result)` → success envelope
   - `Fail(w, r, statusCode, message)` → error envelope with `errors: []`
   - `ValidationFail(w, r, []ValidationErrorItem)` → validation envelope

Add imports: `"context"` and `"time"`.

Keep `SendJSON`, `Error`, `AppError` in place until all call sites are migrated (removed at the end of Step 6).

**File:** `pkg/response/json.go`

---

### Step 2 — Update `internal/validator/validator.go`

Change `FormatValidationErrors` return type from `map[string]string` to `[]response.ValidationErrorItem`.

New logic: iterate `validator.ValidationErrors`, group messages by field into a `map[string][]string` accumulator preserving insertion order, then convert to `[]response.ValidationErrorItem`.

Updated tag messages:
- `required` → `"<field> should not be empty"`
- `min` → `"<field> must be at least <param>"`
- `max` → `"<field> must be at most <param> characters"`
- `email` → `"<field> must be a valid email"`
- default → `"<field> is invalid"`

Add import: `"github.com/valentinesamuel/activelog/pkg/response"`

**File:** `internal/validator/validator.go`

---

### Step 3 — Create `internal/middleware/timing.go` (new file)

```go
package middleware

import (
    "net/http"
    "time"
    "github.com/valentinesamuel/activelog/pkg/response"
)

func TimingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ctx := response.WithStartTime(r.Context(), time.Now())
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

**File:** `internal/middleware/timing.go`

---

### Step 4 — Register `TimingMiddleware` in `cmd/api/main.go`

In `setupRoutes()`, add as the very first `router.Use` call (before `MetricsMiddleware`):

```go
router.Use(middleware.TimingMiddleware)   // FIRST
router.Use(middleware.MetricsMiddleware)
// ... rest unchanged
```

**File:** `cmd/api/main.go`

---

### Step 5 — Update middleware error responses

**`internal/middleware/jwt.go`** — two occurrences (lines 20, 34):
```go
response.Error(w, http.StatusUnauthorized, "Unauthorized request")
→ response.Fail(w, r, http.StatusUnauthorized, "Unauthorized request")
```

**`internal/middleware/rateLimiter.go`** — one occurrence:
```go
http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
→ response.Fail(w, r, http.StatusTooManyRequests, "Rate limit exceeded")
```
(add `"github.com/valentinesamuel/activelog/pkg/response"` import)

---

### Step 6 — Migrate all handler call sites

Apply these substitutions across all handler files. After all call sites are migrated, remove the old `SendJSON`, `Error`, and `AppError` helpers from `pkg/response/json.go`.

#### `internal/handlers/activity.go`

| Line    | Before                                                                                    | After                                                                  |
| ------- | ----------------------------------------------------------------------------------------- | ---------------------------------------------------------------------- |
| 83      | `response.Error(w, http.StatusBadRequest, "Invalid request body")`                        | `response.Fail(w, r, ...)`                                             |
| 90–97   | inline validation block                                                                   | `response.ValidationFail(w, r, validator.FormatValidationErrors(err))` |
| 113     | `response.Error(w, http.StatusInternalServerError, "Failed to create activity")`          | `response.Fail(w, r, ...)`                                             |
| 118     | `response.SendJSON(w, http.StatusCreated, result.Activity)`                               | `response.Success(w, r, ...)`                                          |
| 140     | `response.Error(w, http.StatusBadRequest, "Invalid activity ID")`                         | `response.Fail(w, r, ...)`                                             |
| 156     | `response.Error(w, http.StatusNotFound, "Activity not found")`                            | `response.Fail(w, r, ...)`                                             |
| 161     | `response.Error(w, http.StatusInternalServerError, "Failed to fetch activity")`           | `response.Fail(w, r, ...)`                                             |
| 165     | `response.SendJSON(w, http.StatusOK, result.Activity)`                                    | `response.Success(w, r, ...)`                                          |
| 193     | `response.Error(w, http.StatusInternalServerError, "Failed to fetch activities")`         | `response.Fail(w, r, ...)`                                             |
| 200     | `response.Error(w, http.StatusBadRequest, "Invalid query parameters")`                    | `response.Fail(w, r, ...)`                                             |
| 264     | `response.Error(w, http.StatusBadRequest, err.Error())`                                   | `response.Fail(w, r, ...)`                                             |
| 272     | `response.Error(w, http.StatusBadRequest, err.Error())`                                   | `response.Fail(w, r, ...)`                                             |
| 289     | `response.Error(w, http.StatusInternalServerError, "Failed to fetch activities")`         | `response.Fail(w, r, ...)`                                             |
| 301–304 | `response.SendJSON(w, http.StatusOK, map[string]interface{}{...})`                        | `response.Success(w, r, ...)`                                          |
| 328     | `response.Error(w, http.StatusBadRequest, "Invalid activity ID")`                         | `response.Fail(w, r, ...)`                                             |
| 334     | `response.Error(w, http.StatusBadRequest, "Invalid JSON")`                                | `response.Fail(w, r, ...)`                                             |
| 340–347 | inline validation block                                                                   | `response.ValidationFail(w, r, validator.FormatValidationErrors(err))` |
| 366     | `response.Error(w, http.StatusInternalServerError, "Failed to update activity")`          | `response.Fail(w, r, ...)`                                             |
| 368     | `response.SendJSON(w, http.StatusOK, result.Activity)`                                    | `response.Success(w, r, ...)`                                          |
| 390     | `response.Error(w, http.StatusBadRequest, "Invalid activity ID")`                         | `response.Fail(w, r, ...)`                                             |
| 407     | `response.Error(w, http.StatusNotFound, "Activity not found")`                            | `response.Fail(w, r, ...)`                                             |
| 411     | `w.WriteHeader(http.StatusNoContent)`                                                     | **leave as-is** (HTTP 204 has no body)                                 |
| 449     | `response.Error(w, http.StatusBadRequest, "Invalid request body")`                        | `response.Fail(w, r, ...)`                                             |
| 453     | `response.Error(w, http.StatusBadRequest, "activities must have between 1 and 50 items")` | `response.Fail(w, r, ...)`                                             |
| 483     | `response.SendJSON(w, http.StatusMultiStatus, results)`                                   | `response.Success(w, r, ...)`                                          |
| 506     | `response.Error(w, http.StatusBadRequest, "Invalid request body")`                        | `response.Fail(w, r, ...)`                                             |
| 509     | `response.Error(w, http.StatusBadRequest, "ids must have between 1 and 50 items")`        | `response.Fail(w, r, ...)`                                             |
| 538     | `response.SendJSON(w, http.StatusMultiStatus, results)`                                   | `response.Success(w, r, ...)`                                          |
| 584     | `response.Error(w, http.StatusInternalServerError, "Failed to get statistics")`           | `response.Fail(w, r, ...)`                                             |
| 588     | `response.SendJSON(w, http.StatusOK, result.Stats)`                                       | `response.Success(w, r, ...)`                                          |

#### `internal/handlers/user.go`

| Line    | Before                                                                         | After                                                                  |
| ------- | ------------------------------------------------------------------------------ | ---------------------------------------------------------------------- |
| 33      | `response.Error(w, http.StatusBadRequest, "Invalid request body")`             | `response.Fail(w, r, ...)`                                             |
| 38–46   | inline validation block                                                        | `response.ValidationFail(w, r, validator.FormatValidationErrors(err))` |
| 58      | `response.Error(w, http.StatusBadRequest, "User already exists")`              | `response.Fail(w, r, ...)`                                             |
| 68      | `response.Error(w, http.StatusInternalServerError, "Invalid password")`        | `response.Fail(w, r, ...)`                                             |
| 79      | `response.Error(w, http.StatusInternalServerError, "❌ Failed to create user")` | `response.Fail(w, r, ...)`                                             |
| 84–89   | `response.SendJSON(w, http.StatusCreated, ...)`                                | `response.Success(w, r, ...)`                                          |
| 98      | `response.Error(w, http.StatusBadRequest, "Invalid request body")`             | `response.Fail(w, r, ...)`                                             |
| 103–111 | inline validation block                                                        | `response.ValidationFail(w, r, validator.FormatValidationErrors(err))` |
| 119     | `response.Error(w, http.StatusNotFound, "User not found")`                     | `response.Fail(w, r, ...)`                                             |
| 123     | `response.Error(w, http.StatusInternalServerError, "Invalid Credentials")`     | `response.Fail(w, r, ...)`                                             |
| 130–132 | `response.Error(w, http.StatusInternalServerError, "Invalid Credentials")`     | `response.Fail(w, r, ...)`                                             |
| 136–138 | `response.Error(w, http.StatusInternalServerError, "Invalid credentials")`     | `response.Fail(w, r, ...)`                                             |
| 143–145 | `response.Error(w, http.StatusInternalServerError, "Server error")`            | `response.Fail(w, r, ...)`                                             |
| 148–151 | `response.SendJSON(w, http.StatusOK, ...)`                                     | `response.Success(w, r, ...)`                                          |

#### `internal/handlers/stats.go`

| Line | Before                                                   | After                         |
| ---- | -------------------------------------------------------- | ----------------------------- |
| 29   | `response.Error(w, http.StatusInternalServerError, ...)` | `response.Fail(w, r, ...)`    |
| 33   | `response.SendJSON(w, http.StatusOK, weeklyStats)`       | `response.Success(w, r, ...)` |
| 43   | `response.Error(w, http.StatusInternalServerError, ...)` | `response.Fail(w, r, ...)`    |
| 46   | `response.SendJSON(w, http.StatusOK, monthlyStats)`      | `response.Success(w, r, ...)` |
| 56   | `response.Error(w, http.StatusInternalServerError, ...)` | `response.Fail(w, r, ...)`    |
| 60   | `response.SendJSON(w, http.StatusOK, activitySummary)`   | `response.Success(w, r, ...)` |
| 80   | `response.Error(w, http.StatusInternalServerError, ...)` | `response.Fail(w, r, ...)`    |
| 90   | `response.SendJSON(w, http.StatusOK, responseData)`      | `response.Success(w, r, ...)` |
| 100  | `response.Error(w, http.StatusInternalServerError, ...)` | `response.Fail(w, r, ...)`    |
| 116  | `response.SendJSON(w, http.StatusOK, responseData)`      | `response.Success(w, r, ...)` |

#### `internal/handlers/webhook_handler.go`

| Line | Before                                                                       | After                         |
| ---- | ---------------------------------------------------------------------------- | ----------------------------- |
| 38   | `response.Error(w, http.StatusBadRequest, "Invalid request body")`           | `response.Fail(w, r, ...)`    |
| 42   | `response.Error(w, http.StatusBadRequest, "URL is required")`                | `response.Fail(w, r, ...)`    |
| 46   | `response.Error(w, http.StatusBadRequest, "At least one event is required")` | `response.Fail(w, r, ...)`    |
| 52   | `response.Error(w, http.StatusInternalServerError, ...)`                     | `response.Fail(w, r, ...)`    |
| 64   | `response.Error(w, http.StatusInternalServerError, ...)`                     | `response.Fail(w, r, ...)`    |
| 73   | `response.SendJSON(w, http.StatusCreated, webhookResponse{...})`             | `response.Success(w, r, ...)` |
| 83   | `response.Error(w, http.StatusInternalServerError, ...)`                     | `response.Fail(w, r, ...)`    |
| 89   | `response.SendJSON(w, http.StatusOK, webhooks)`                              | `response.Success(w, r, ...)` |
| 99   | `response.Error(w, http.StatusNotFound, "Webhook not found")`                | `response.Fail(w, r, ...)`    |
| 102  | `w.WriteHeader(http.StatusNoContent)`                                        | **leave as-is**               |

#### `internal/handlers/photo_handler.go`

| Line | Before                                                            | After                         |
| ---- | ----------------------------------------------------------------- | ----------------------------- |
| 49   | `response.Error(w, http.StatusBadRequest, "Invalid activity ID")` | `response.Fail(w, r, ...)`    |
| 59   | `response.Error(w, http.StatusBadRequest, err.Error())`           | `response.Fail(w, r, ...)`    |
| 65   | `response.Error(w, http.StatusBadRequest, "Too many files")`      | `response.Fail(w, r, ...)`    |
| 117  | `response.Error(w, http.StatusBadRequest, v.err.Error())`         | `response.Fail(w, r, ...)`    |
| 136  | `response.Error(w, http.StatusInternalServerError, ...)`          | `response.Fail(w, r, ...)`    |
| 141  | `response.SendJSON(w, http.StatusCreated, result.ActivityPhotos)` | `response.Success(w, r, ...)` |
| 152  | `response.Error(w, http.StatusBadRequest, "Invalid activity ID")` | `response.Fail(w, r, ...)`    |
| 167  | `response.Error(w, http.StatusInternalServerError, ...)`          | `response.Fail(w, r, ...)`    |
| 172  | `response.SendJSON(w, http.StatusOK, result.Photos)`              | `response.Success(w, r, ...)` |

#### `internal/handlers/export_handler.go`

| Line    | Before                                                   | After                                                                           |
| ------- | -------------------------------------------------------- | ------------------------------------------------------------------------------- |
| 51      | `response.Error(w, http.StatusInternalServerError, ...)` | `response.Fail(w, r, ...)` (before CSV headers set — safe)                      |
| 59      | `response.Error(w, http.StatusInternalServerError, ...)` | **leave as-is** (after `text/csv` headers written — cannot switch Content-Type) |
| 76      | `response.Error(w, http.StatusInternalServerError, ...)` | `response.Fail(w, r, ...)`                                                      |
| 87      | `response.Error(w, http.StatusInternalServerError, ...)` | `response.Fail(w, r, ...)`                                                      |
| 97      | `response.Error(w, http.StatusInternalServerError, ...)` | `response.Fail(w, r, ...)`                                                      |
| 101–103 | `response.SendJSON(w, http.StatusAccepted, ...)`         | `response.Success(w, r, ...)`                                                   |
| 113     | `response.Error(w, http.StatusNotFound, ...)`            | `response.Fail(w, r, ...)`                                                      |
| 118     | `response.SendJSON(w, http.StatusOK, record)`            | `response.Success(w, r, ...)`                                                   |
| 129     | `response.Error(w, http.StatusNotFound, ...)`            | `response.Fail(w, r, ...)`                                                      |
| 134     | `response.Error(w, http.StatusConflict, ...)`            | `response.Fail(w, r, ...)`                                                      |
| 139     | `response.Error(w, http.StatusInternalServerError, ...)` | `response.Fail(w, r, ...)`                                                      |
| 150     | `response.Error(w, http.StatusInternalServerError, ...)` | `response.Fail(w, r, ...)`                                                      |
| 153–155 | `response.SendJSON(w, http.StatusOK, ...)`               | `response.Success(w, r, ...)`                                                   |

#### `internal/handlers/features_handler.go`

| Line | Before                                          | After                         |
| ---- | ----------------------------------------------- | ----------------------------- |
| 29   | `response.SendJSON(w, http.StatusOK, features)` | `response.Success(w, r, ...)` |

#### `internal/handlers/health.go`

| Line | Before                                              | After                         |
| ---- | --------------------------------------------------- | ----------------------------- |
| 32   | `response.SendJSON(w, http.StatusOK, responseData)` | `response.Success(w, r, ...)` |

---

## Critical Files Summary

| File                                    | Change Type                                                        |
| --------------------------------------- | ------------------------------------------------------------------ |
| `pkg/response/json.go`                  | Add types + 3 new helpers + context key; remove old helpers at end |
| `internal/validator/validator.go`       | Change return type of `FormatValidationErrors`                     |
| `internal/middleware/timing.go`         | **New file**                                                       |
| `cmd/api/main.go`                       | Register `TimingMiddleware` first                                  |
| `internal/middleware/jwt.go`            | `response.Error` → `response.Fail` (2 calls)                       |
| `internal/middleware/rateLimiter.go`    | `http.Error` → `response.Fail` (1 call)                            |
| `internal/handlers/activity.go`         | 26 call sites migrated                                             |
| `internal/handlers/user.go`             | 14 call sites migrated                                             |
| `internal/handlers/stats.go`            | 10 call sites migrated                                             |
| `internal/handlers/webhook_handler.go`  | 10 call sites migrated                                             |
| `internal/handlers/photo_handler.go`    | 9 call sites migrated                                              |
| `internal/handlers/export_handler.go`   | 12 call sites migrated (1 left as-is)                              |
| `internal/handlers/features_handler.go` | 1 call site migrated                                               |
| `internal/handlers/health.go`           | 1 call site migrated                                               |

---

## Import Safety (No Circular Deps)

```
pkg/response     (adds: "context", "time" — no internal/ imports)
    ↑
internal/validator  (adds: pkg/response import)
    ↑
internal/middleware (already imports pkg/response — no change needed)
    ↑
internal/handlers   (already imports pkg/response and internal/validator)
```

---

## Verification

1. `go build ./...` — must compile with zero errors
2. Run the server: `go run ./cmd/api/`
3. **Success response test:**
   ```
   POST /api/v1/auth/register with valid body
   → { "statusCode": 201, "success": true, "message": "Request successful", "result": {...}, "path": "/api/v1/auth/register", "duration": <ms> }
   ```
4. **Validation error test:**
   ```
   POST /api/v1/auth/register with empty body
   → { "statusCode": 400, "success": false, "message": "Bad Request", "errors": [{"field": "username", "errors": ["username should not be empty"]}, ...], "path": "...", "duration": <ms> }
   ```
5. **Generic error test:**
   ```
   GET /api/v1/activities/999999
   → { "statusCode": 404, "success": false, "message": "Activity not found", "errors": [], "path": "...", "duration": <ms> }
   ```
