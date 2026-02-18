# Cache Adapter with Multi-DB + Partition Support (incl. Rate Limiter)

## Context

The current cache setup uses a single `CacheProvider` connected to Redis DB 0 for everything. Use cases and the rate limiter call it with raw string keys — no DB isolation, no namespacing, no config caching.

Goals:
1. Introduce a `CacheAdapter` layer that routes to the correct Redis DB + namespaces every key — both are **mandatory**, no defaults.
2. Add a `RateLimitCacheProvider` interface (separate, for Increment/Expire) used by the rate limiter.
3. The rate limit config (from `ratelimit.yaml`) is cached to Redis at startup with a 48-hour TTL, read back to verify, and stale-while-revalidate refreshed atomically when < 3 minutes remain before expiry.
4. Migrate existing use cases (`ListActivities`, `UpdateActivity`) to the new `CacheAdapter`.

Inbox/outbox queues (DB 0 via asynq) are untouched.

---

## Architecture Overview

```
Use Case / RateLimiter
        │
        ▼
CacheAdapter (Get/Set/Del)          ← types.CacheAdapter interface
RateLimitCacheProvider (Incr/Expire) ← types.RateLimitCacheProvider interface
        │
        └── both implemented by internal/cache/adapter/redis.Adapter
                │  holds map[dbNumber]*redis.Client (lazy, mutex-protected)
                │  builds key: "<partition>:<caller-key>"
                ▼
        Redis DB 1 (activity data)
        Redis DB 2 (stats)
        Redis DB 3 (rate limits — config + counters)
```

The existing `CacheProvider` + its DI registration remain **untouched** (no regressions).

---

## Implementation Phases

### Phase 1 — Config: named Redis DB numbers
**File:** `internal/config/redis.go`

```go
type CacheDBNumbers struct {
    ActivityData int  // REDIS_DB_ACTIVITY_DATA (default 1)
    Stats        int  // REDIS_DB_STATS         (default 2)
    RateLimits   int  // REDIS_DB_RATE_LIMITS   (default 3)
}

type CacheConfigType struct {
    Provider string
    Redis    RedisConfigType
    DBs      CacheDBNumbers   // ← new
}
```

`loadCache()` addition:
```go
DBs: CacheDBNumbers{
    ActivityData: GetEnvInt("REDIS_DB_ACTIVITY_DATA", 1),
    Stats:        GetEnvInt("REDIS_DB_STATS", 2),
    RateLimits:   GetEnvInt("REDIS_DB_RATE_LIMITS", 3),
},
```

---

### Phase 2 — Cache types: enums + adapter interfaces
**File:** `internal/cache/types/types.go`

Append below the existing `CacheProvider` interface (do not modify it):

```go
// --- Multi-DB Adapter types ---

type CacheDBName string

const (
    CacheDBActivityData CacheDBName = "ACTIVITY_DATA"
    CacheDBStats        CacheDBName = "STATS"
    CacheDBRateLimits   CacheDBName = "RATE_LIMITS"
)

type CachePartition string

const (
    CachePartitionActivities        CachePartition = "activities"
    CachePartitionStats             CachePartition = "stats"
    CachePartitionRateLimitConfig   CachePartition = "ratelimit:config"
    CachePartitionRateLimitCounters CachePartition = "ratelimit:counters"
)

// CacheOptions is required on every CacheAdapter call.
type CacheOptions struct {
    DB           CacheDBName
    PartitionKey CachePartition
}

// CacheAdapter is the high-level interface for general caching.
// DB and PartitionKey are always required — no defaults.
type CacheAdapter interface {
    Get(ctx context.Context, key string, opts CacheOptions) (string, error)
    Set(ctx context.Context, key string, value string, ttl time.Duration, opts CacheOptions) error
    Del(ctx context.Context, key string, opts CacheOptions) error
}

// RateLimitCacheProvider is the dedicated interface for rate limiter counter operations.
type RateLimitCacheProvider interface {
    Increment(ctx context.Context, key string, opts CacheOptions) (int64, error)
    Expire(ctx context.Context, key string, ttl time.Duration, opts CacheOptions) (bool, error)
    SetNX(ctx context.Context, key string, value string, ttl time.Duration, opts CacheOptions) (bool, error)
}
```

> `SetNX` is needed for the atomic refresh lock.

---

### Phase 3 — Redis adapter (implements both interfaces)
**New file:** `internal/cache/adapter/redis/adapter.go`

```go
package redis

// Adapter implements both CacheAdapter and RateLimitCacheProvider.
// It lazily creates one *redis.Client per DB number, protected by a mutex.
type Adapter struct {
    addr     string
    password string
    dbMap    map[types.CacheDBName]int   // name → DB number
    clients  map[int]*redis.Client       // DB number → client (lazy)
    mu       sync.Mutex
}

func New() *Adapter {
    return &Adapter{
        addr:     config.Cache.Redis.Address,
        password: config.Cache.Redis.Password,
        dbMap: map[types.CacheDBName]int{
            types.CacheDBActivityData: config.Cache.DBs.ActivityData,
            types.CacheDBStats:        config.Cache.DBs.Stats,
            types.CacheDBRateLimits:   config.Cache.DBs.RateLimits,
        },
        clients: make(map[int]*redis.Client),
    }
}

// client lazily initialises and returns the redis.Client for the given DB name.
func (a *Adapter) client(db types.CacheDBName) (*redis.Client, error) { ... }

// buildKey → "<partition>:<key>"
func buildKey(opts types.CacheOptions, key string) string {
    return fmt.Sprintf("%s:%s", opts.PartitionKey, key)
}

// CacheAdapter methods: Get, Set, Del
// RateLimitCacheProvider methods: Increment, Expire, SetNX
```

---

### Phase 4 — DI wiring
**File:** `internal/cache/di/keys.go`

```go
const CacheAdapterKey = "cacheAdapter"
```

**File:** `internal/cache/di/register.go`

Add `RegisterCacheAdapter`:
```go
func RegisterCacheAdapter(c *container.Container) {
    c.Register(CacheAdapterKey, func(c *container.Container) (interface{}, error) {
        switch config.Cache.Provider {
        case "redis":
            adapter := redisadapter.New()
            log.Printf("Cache adapter initialized: Redis multi-DB")
            return adapter, nil
        default:
            return nil, fmt.Errorf("unsupported provider: %s", config.Cache.Provider)
        }
    })
}
```

Call `RegisterCacheAdapter` in the app bootstrap alongside `RegisterCache`.

---

### Phase 5 — Rate limit config caching (startup + stale-while-revalidate)

#### Startup sequence (wherever `NewRateLimiter` is constructed, e.g., `main.go`)

```
1. config.MustLoad()                             // loads ratelimit.yaml → config.RateLimit (in-memory)
2. RegisterCacheAdapter(container)
3. adapter = container.MustResolve(CacheAdapterKey)
4. rateOpts = CacheOptions{DB: CacheDBRateLimits, PartitionKey: CachePartitionRateLimitConfig}
5. Write config.RateLimit (as CachedRateLimitConfig JSON) to Redis with 48h TTL
6. Read back from Redis → verify presence; if miss, log warning (in-memory still available)
7. app.RateLimiter = middleware.NewRateLimiter(rlCacheProvider, cacheAdapter, config.RateLimit)
```

#### Cached value schema (stored in Redis as JSON)

```go
type CachedRateLimitConfig struct {
    CachedAt time.Time             `json:"cached_at"`
    Config   config.RateLimitConfig `json:"config"`
}
```

#### Runtime request path (inside `RateLimiter.Middleware`)

```
1. Try cacheAdapter.Get(ctx, "config", rateConfigOpts)
2a. On SUCCESS:
    - Parse CachedRateLimitConfig
    - If time.Since(cachedAt) > (48h - 3min):
        → Try to acquire refresh lock via rlCacheProvider.SetNX(ctx, "refresh_lock", "1", 5min, lockOpts)
        → If lock acquired: enqueue EventRefreshRateLimitConfig job on InboxQueue
    - Use parsed config for this request
2b. On FAILURE (Redis error or key not found):
    - Use in-memory config.RateLimit (fallback)
3. Proceed with rate limit check using resolved config
```

#### Background refresh job

**Event type** (add to `internal/queue/types/types.go`):
```go
EventRefreshRateLimitConfig EventType = "refresh_rate_limit_config"
```

**Handler** (new file following existing inbox handler pattern):
- Read `ratelimit.yaml` from disk
- Parse into `config.RateLimitConfig`
- Write new `CachedRateLimitConfig` with fresh `CachedAt` to Redis with 48h TTL (overwrite atomically with SET)
- Log success/failure

---

### Phase 6 — Rate limiter middleware update
**File:** `internal/middleware/rateLimiter.go`

```go
type RateLimiter struct {
    rlCache     types.RateLimitCacheProvider  // for counters (Increment/Expire)
    configCache types.CacheAdapter            // for config caching (Get/Set)
    fallback    *config.RateLimitConfig       // in-memory fallback
}
```

- Counter key format unchanged: `ratelimit:user:<id>:<method>:<path>`
  → stored via `Increment(ctx, key, CacheOptions{DB: CacheDBRateLimits, PartitionKey: CachePartitionRateLimitCounters})`
- Config key: `"config"` with `CacheOptions{DB: CacheDBRateLimits, PartitionKey: CachePartitionRateLimitConfig}`
- Lock key: `"refresh_lock"` with same DB, `CachePartitionRateLimitConfig` partition

---

### Phase 7 — Migrate activity use cases to CacheAdapter

#### `list_activities.go`
- Field: `cache types.CacheAdapter`
- `cache.Get(ctx, cacheKey, CacheOptions{DB: CacheDBActivityData, PartitionKey: CachePartitionActivities})`
- `cache.Set(ctx, cacheKey, data, cacheTTL, CacheOptions{DB: CacheDBActivityData, PartitionKey: CachePartitionActivities})`

#### `update_activity.go`
- Field: `cache types.CacheAdapter`
- `cache.Del(ctx, key, CacheOptions{DB: CacheDBActivityData, PartitionKey: CachePartitionActivities})`

#### `di/register.go` (activity)
- Resolve `cacheDI.CacheAdapterKey`, cast to `types.CacheAdapter` for both `ListActivitiesUCKey` and `UpdateActivityUCKey`

---

### Phase 8 — Env files

**`.env` and `.env.example`:**
```
REDIS_DB_ACTIVITY_DATA=1
REDIS_DB_STATS=2
REDIS_DB_RATE_LIMITS=3
```

---

## Files Modified / Created

| File                                                        | Action                                                                 |
| ----------------------------------------------------------- | ---------------------------------------------------------------------- |
| `internal/config/redis.go`                                  | Modify — add `CacheDBNumbers`                                          |
| `internal/cache/types/types.go`                             | Modify — add enums, CacheOptions, CacheAdapter, RateLimitCacheProvider |
| `internal/cache/adapter/redis/adapter.go`                   | **Create** — Redis impl of both interfaces                             |
| `internal/cache/di/keys.go`                                 | Modify — add `CacheAdapterKey`                                         |
| `internal/cache/di/register.go`                             | Modify — add `RegisterCacheAdapter`                                    |
| `internal/middleware/rateLimiter.go`                        | Modify — use CacheAdapter + stale-while-revalidate logic               |
| `internal/queue/types/types.go`                             | Modify — add `EventRefreshRateLimitConfig`                             |
| `internal/application/broker/` (handler file)               | Modify — register refresh handler                                      |
| `internal/application/activity/usecases/list_activities.go` | Modify — use CacheAdapter                                              |
| `internal/application/activity/usecases/update_activity.go` | Modify — use CacheAdapter                                              |
| `internal/application/activity/usecases/di/register.go`     | Modify — inject CacheAdapter                                           |
| `.env`                                                      | Modify — add 3 DB number vars                                          |
| `.env.example`                                              | Modify — document new vars                                             |

`CacheProvider` + its DI registration → **untouched**.

---

## Verification

1. `go build ./...` — zero errors
2. Start app — expect:
   - `"Cache adapter initialized: Redis multi-DB"` log line
   - `"Rate limit config cached to Redis (DB 3)"` log line
   - `"Rate limit config verified from Redis"` log line
3. `redis-cli -p 6377 -n 1 keys '*'` — confirm activity keys prefixed with `activities:`
4. `redis-cli -p 6377 -n 3 keys '*'` — confirm `ratelimit:config:config` key present with TTL ~48h
5. `GET /activities` twice — first miss, second hit (check `X-Cache` headers)
6. `PUT /activities/:id` — subsequent GET returns fresh data (cache invalidated)
7. Simulate near-expiry (short TTL in test) — verify refresh job is enqueued exactly once (lock prevents double-enqueue)
