# Plan: Remove Redundant Cache Provider Layer

## Context

When the cache layer was originally built, a simple single-DB `CacheProvider` was implemented in `internal/cache/redis/provider.go`. This was later superseded by the multi-DB `Adapter` in `internal/cache/adapter/redis/adapter.go`, which added context support, lazy per-DB client initialization, key namespacing, and `SetNX`. The old provider was never removed — it's still registered in the DI container, its connection is eagerly resolved at startup, but the resolved object is never actually used for any real operation. Additionally, a dead utility function (`mustMarshal`) was found in `internal/jobs/types.go`.

## Changes

### 1. Delete `internal/cache/redis/` directory
- Delete `internal/cache/redis/provider.go` (entire file — the old single-DB provider)
- The directory becomes empty after this and should also be deleted

### 2. `internal/cache/types/types.go`
- Remove the `CacheProvider` interface (lines 8–14): the context-free `Get/Set/Del/Increment/Expire` signatures
- Keep everything else: `CacheDBName`, `CachePartition`, `CacheOptions`, `CacheAdapter`, `RateLimitCacheProvider`
- The `"time"` import remains (still needed by `CacheAdapter` and `RateLimitCacheProvider`)

### 3. `internal/cache/di/register.go`
- Remove `RegisterCache()` function
- Remove `createProvider()` helper (only called by `RegisterCache`)
- Remove the old redis import: `"github.com/valentinesamuel/activelog/internal/cache/redis"`
- Remove the import of `types` if it becomes unused (check — it's still used for the log line in `RegisterCacheAdapter` return type? No — `types` is only referenced in `createProvider`'s return type, so remove that import too)
- Keep `RegisterCacheAdapter()` and the `redisadapter` import

> Note: After removing `createProvider()`, the `"github.com/valentinesamuel/activelog/internal/cache/types"` import will also become unused — remove it.

### 4. `internal/cache/di/keys.go`
- Remove `CacheProviderKey = "cacheProvider"` constant
- Keep `CacheAdapterKey`

### 5. `cmd/api/container.go`
- Remove the call `cacheRegister.RegisterCache(c)` (registers the old provider)
- Remove the call `c.MustResolve(cacheRegister.CacheProviderKey)` (eagerly resolves the old provider; result was always discarded)
- Keep `cacheRegister.RegisterCacheAdapter(c)` and `c.MustResolve(cacheRegister.CacheAdapterKey)`

### 6. `internal/jobs/types.go`
- Remove the `mustMarshal()` function (unexported, never called anywhere in the codebase; comment itself says "used only in tests/examples" but no tests call it)
- Remove the `encoding/json` import (only used by `mustMarshal`)

## Files NOT Changed

| File                                                                              | Reason                                                                                                                 |
| --------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------- |
| `internal/cache/adapter/redis/adapter.go`                                         | Current live implementation — keep                                                                                     |
| `internal/config/redis.go` (incl. `RedisConfigType.DB`)                           | `RedisConfigType.DB` is still consumed by `internal/webhook/redis/provider.go` for its Redis Streams connection (DB 0) |
| `.env.example` (`REDIS_DB=0`)                                                     | Still valid for the webhook stream provider                                                                            |
| `internal/webhook/redis/provider.go`                                              | Different purpose (Redis Streams for webhooks) — not redundant                                                         |
| `internal/middleware/rateLimiter.go`                                              | `rateLimitLockOpts` partition key issue is a bug, not redundant code — out of scope                                    |
| `internal/queue/memory/` + `internal/queue/asynq/`                                | Intentional strategy pattern — chosen at runtime by `QUEUE_PROVIDER` env var; neither is dead                          |
| `internal/webhook/memory/` + `internal/webhook/redis/` + `internal/webhook/nats/` | Intentional three-way strategy — chosen by `WEBHOOK_PROVIDER` env var; all three are live                              |

## Scope Note

A full codebase audit confirmed the cache layer is the **only** area with this true "old superseded by new but never removed" redundancy pattern. All other multi-implementation areas (queue, webhook) use deliberate environment-driven strategy switches where every provider is an active code path.

## Critical Files

| File                               | Action                                                          |
| ---------------------------------- | --------------------------------------------------------------- |
| `internal/cache/redis/provider.go` | Delete                                                          |
| `internal/cache/types/types.go`    | Edit — remove `CacheProvider` interface + unused import cleanup |
| `internal/cache/di/register.go`    | Edit — remove `RegisterCache`, `createProvider`, old imports    |
| `internal/cache/di/keys.go`        | Edit — remove `CacheProviderKey`                                |
| `cmd/api/container.go`             | Edit — remove 2 lines                                           |
| `internal/jobs/types.go`           | Edit — remove `mustMarshal` + `encoding/json` import            |

## Verification

1. `go build ./...` — must compile with zero errors (no unused imports, no undefined references)
2. `go vet ./...` — must pass clean
3. Confirm no remaining references to `CacheProviderKey`, `RegisterCache`, or `createProvider` via grep
4. Confirm `internal/cache/redis/` directory no longer exists
5. Confirm `mustMarshal` is gone from `internal/jobs/types.go`
6. Start the API server (`make run` or equivalent) — startup logs should show only "Cache adapter initialized: Redis multi-DB", not the old provider log line
