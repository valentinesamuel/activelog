# MONTH 5: Caching & Performance

**Weeks:** 17-20
**Phase:** Optimization & Scalability
**Theme:** Make your application fast and efficient

---

## Overview

This month focuses on performance optimization through caching, rate limiting, and monitoring. You'll learn how to use Redis for caching and session management, implement rate limiting to protect your API, set up performance monitoring, and implement soft deletes for data integrity. By the end, your API will be production-ready for high traffic.

---

## Learning Path

### Week 17: Redis Setup + Basic Caching
- Install and configure Redis
- Connect Go application to Redis
- Implement cache-aside pattern
- Cache activity listings

### Week 18: Cache Invalidation Strategies + Soft Deletes Pattern (45 min)
- When to invalidate cache
- TTL (Time-To-Live) strategies
- Write-through vs write-behind caching
- **NEW:** Soft deletes with `deleted_at` timestamp

### Week 19: Rate Limiting
- Implement rate limiting middleware
- Token bucket algorithm
- Per-user and per-IP limits
- Custom rate limit responses

### Week 20: Performance Monitoring
- Prometheus metrics integration
- Grafana dashboards
- Request duration tracking
- Error rate monitoring

---

# WEEKLY TASK BREAKDOWNS

## Week 17: Redis Setup + Basic Caching

### 📋 Implementation Tasks

**Task 1: Install and Configure Redis** (30 min)
- [ ] Install Redis: `brew install redis` (Mac) or `apt-get install redis` (Linux)
- [ ] Start Redis server: `redis-server` or `brew services start redis`
- [ ] Test connection: `redis-cli ping` (should return PONG)
- [ ] Install Go Redis client: `go get github.com/redis/go-redis/v9`
- [ ] Configure Redis connection in application config

**Task 2: Create Redis Client Wrapper** (45 min)
- [ ] Create `pkg/cache/redis_client.go`
- [ ] Implement `NewRedisClient(addr, password string) (*RedisClient, error)`
  - **Logic:** Create `redis.Options` with addr and password. Call `redis.NewClient(opts)` to get client. Ping server with `client.Ping(ctx)` to verify connection. If ping fails, return error. If success, return RedisClient struct wrapping the client.
- [ ] Add connection pooling configuration
- [ ] Implement `Ping()` to test connection
  - **Logic:** Call `client.Ping(ctx).Result()`. Returns "PONG" if Redis is up, error otherwise. Used for health checks.
- [ ] Add graceful disconnect on shutdown
- [ ] Test connection and basic operations

**Task 3: Implement Cache-Aside Pattern** (90 min)
- [ ] Create `internal/services/activity_service.go`
- [ ] Implement `GetActivity(ctx, id)` with cache-aside pattern:
  - **Logic:**
    1. Build cache key: `activity:{id}`
    2. Try `redis.Get(ctx, key)` to check cache
    3. If cache HIT: unmarshal JSON string to Activity struct, return
    4. If cache MISS (redis.Nil error): fetch from database using repository
    5. If DB returns activity: marshal to JSON, store in Redis with `redis.Set(ctx, key, json, 5*time.Minute)`
    6. Return activity from DB
    7. If Redis errors (connection failed): log error, skip cache, fetch from DB (fail open - don't break app)
    - **Why:** Cache-aside = application manages cache. Check cache first (fast), fall back to DB (slow), then populate cache for next request.
  - Check cache first
  - On cache miss, fetch from database
  - Store in cache with TTL
- [ ] Use `json.Marshal/Unmarshal` for cache serialization
- [ ] Set TTL to 5 minutes
- [ ] Handle cache errors gracefully (fail open)

**Task 4: Cache Activity Listings** (60 min)
- [ ] Implement `GetActivitiesByUser(ctx, userID)` with caching
  - **Logic:** Same cache-aside pattern but key includes pagination: `activities:user:{userID}:page:{page}:limit:{limit}`. Marshal entire activity slice to JSON. Shorter TTL (2 min) since lists change often when user adds activities. Invalidate on create/update/delete.
- [ ] Generate cache key: `activities:user:{userID}`
- [ ] Cache paginated results separately
- [ ] Set TTL to 2 minutes (changes frequently)
- [ ] Test cache hit/miss scenarios

**Task 5: Add Cache Metrics** (30 min)
- [ ] Track cache hits and misses
- [ ] Add counters to Prometheus metrics
- [ ] Log cache performance
- [ ] Create helper function `recordCacheHit()` and `recordCacheMiss()`

**Task 6: Write Tests** (45 min)
- [ ] Test cache hit scenario
- [ ] Test cache miss scenario
- [ ] Test cache error handling (Redis down)
- [ ] Test TTL expiration
- [ ] Mock Redis client for unit tests

### 📦 Files You'll Create/Modify

```
pkg/
└── cache/
    ├── redis_client.go            [CREATE]
    └── redis_client_test.go       [CREATE]

internal/
├── services/
│   ├── activity_service.go        [CREATE]
│   └── activity_service_test.go   [CREATE]
└── config/
    └── config.go                  [MODIFY - add Redis config]
```

### 🔄 Implementation Order

1. **Setup**: Install Redis → Test manually → Install Go client
2. **Client**: Redis wrapper → Connection pooling
3. **Service**: Activity service with cache-aside pattern
4. **Testing**: Unit tests with mocked Redis
5. **Metrics**: Add cache performance tracking

### ⚠️ Blockers to Watch For

- **Redis not running**: Ensure `redis-server` is running before tests
- **Serialization**: JSON marshaling can fail - handle errors
- **TTL**: Too long = stale data, too short = cache thrashing
- **Memory**: Monitor Redis memory usage (use `redis-cli info memory`)
- **Fail open**: If Redis fails, still serve from database (don't break app)

### ✅ Definition of Done

- [ ] Redis installed and running locally
- [ ] Can connect to Redis from Go application
- [ ] Cache-aside pattern working for single activities
- [ ] Activity listings cached with pagination
- [ ] Cache metrics tracked (hits/misses)
- [ ] All tests passing (cache hit, miss, error scenarios)

---

## Week 18: Cache Invalidation + Soft Deletes

### 📋 Implementation Tasks

**Task 1: Create Migration for Soft Deletes** (20 min)
- [ ] Create migration `migrations/006_add_soft_deletes.up.sql`
- [ ] Add `deleted_at TIMESTAMP NULL` to activities table
- [ ] Add `deleted_at TIMESTAMP NULL` to users table
- [ ] Create partial index: `WHERE deleted_at IS NULL` for performance
- [ ] Run migration

**Task 2: Implement Soft Delete in Repository** (60 min)
- [ ] Update `ActivityRepository.Delete()` to soft delete
- [ ] Set `deleted_at = NOW()` instead of actual DELETE
- [ ] Check rows affected (return ErrNotFound if 0)
- [ ] Update all query methods to exclude soft-deleted records
- [ ] Add `WHERE deleted_at IS NULL` to all SELECT queries

**Task 3: Implement Restore Functionality** (30 min)
- [ ] Add `Restore(ctx, id) error` method to repository
- [ ] Set `deleted_at = NULL` to restore
- [ ] Add endpoint: `POST /api/v1/activities/:id/restore`
- [ ] Protect with auth (only owner can restore)
- [ ] Test restore flow

**Task 4: Implement Cache Invalidation on Updates** (90 min)
- [ ] Update `ActivityService.Update()` to invalidate cache
- [ ] Delete cache key: `activity:{id}`
- [ ] Delete user list cache: `activities:user:{userID}`
- [ ] Test invalidation on update
- [ ] Test invalidation on soft delete
- [ ] Ensure database updated before cache deleted

**Task 5: Implement Write-Through Caching** (60 min)
- [ ] Update `CreateActivity` to write to cache immediately
- [ ] Update database first, then cache
- [ ] Handle cache write failures gracefully
- [ ] Set same TTL as read operations
- [ ] Compare write-through vs cache-aside performance

**Task 6: Add Permanent Delete (Admin Only)** (45 min)
- [ ] Add `PermanentDelete(ctx, id) error` method
- [ ] Actually DELETE from database
- [ ] Add admin authorization check
- [ ] Cascade delete related records (photos, tags)
- [ ] Log permanent deletes for audit

### 📦 Files You'll Create/Modify

```
migrations/
├── 006_add_soft_deletes.up.sql    [CREATE]
└── 006_add_soft_deletes.down.sql  [CREATE]

internal/
├── repository/
│   └── activity_repository.go     [MODIFY - soft deletes]
├── services/
│   └── activity_service.go        [MODIFY - cache invalidation]
└── handlers/
    └── activity_handler.go        [MODIFY - restore endpoint]
```

### 🔄 Implementation Order

1. **Database**: Migration → Run migration
2. **Repository**: Soft delete logic → Update queries
3. **Service**: Cache invalidation on updates
4. **Restore**: Restore functionality
5. **Admin**: Permanent delete for admins

### ⚠️ Blockers to Watch For

- **Query updates**: ALL queries must exclude `deleted_at IS NOT NULL`
- **Foreign keys**: Cascade deletes might interfere - test carefully
- **Cache timing**: Invalidate cache AFTER database update succeeds
- **Partial index**: Improves query performance on non-deleted records
- **Restore race**: User could restore while admin permanently deletes

### ✅ Definition of Done

- [ ] Activities soft-deleted instead of hard-deleted
- [ ] All queries exclude soft-deleted records
- [ ] Can restore soft-deleted activities
- [ ] Cache invalidated on update/delete
- [ ] Write-through caching implemented
- [ ] Permanent delete available for admins
- [ ] All tests passing

---

## Week 19: Rate Limiting

### 📋 Implementation Tasks

**Task 1: Design Rate Limit Strategy** (20 min)
- [ ] Decide on limits: 100 requests/minute per user (adjust as needed)
- [ ] Anonymous users: 20 requests/minute per IP
- [ ] Premium users: 500 requests/minute
- [ ] Document rate limit strategy

**Task 2: Implement Rate Limiter with Redis** (90 min)
- [ ] Create `internal/middleware/rate_limiter.go`
- [ ] Implement token bucket algorithm using Redis INCR
- [ ] Use key format: `ratelimit:{userID}` or `ratelimit:ip:{IP}`
- [ ] Set expiration with `EXPIRE` on first request
- [ ] Return 429 Too Many Requests when limit exceeded
- [ ] Add rate limit headers: X-RateLimit-Limit, X-RateLimit-Remaining, Retry-After

**Task 3: Create Rate Limit Middleware** (60 min)
- [ ] Implement `RateLimiter.Middleware(next) http.Handler`
- [ ] Extract user ID from context or IP from request
- [ ] Check/increment counter in Redis
- [ ] Add rate limit headers to all responses
- [ ] Fail open if Redis unavailable (allow request)
- [ ] Apply middleware to router

**Task 4: Implement Per-Endpoint Rate Limits** (45 min)
- [ ] Different limits for different endpoints:
  - POST /activities: 10/minute
  - GET /activities: 100/minute
  - POST /auth/login: 5/minute
- [ ] Use endpoint-specific keys: `ratelimit:{userID}:create_activity`
- [ ] Test each endpoint's limit

**Task 5: Add Rate Limit Bypass for Premium Users** (30 min)
- [ ] Check user tier from database
- [ ] Apply higher limits for premium users
- [ ] Cache user tier in Redis (TTL: 1 hour)
- [ ] Test different user tiers

**Task 6: Monitor Rate Limit Violations** (30 min)
- [ ] Log rate limit violations
- [ ] Add Prometheus counter for rate limit hits
- [ ] Alert on excessive violations (potential attack)
- [ ] Create dashboard visualization

### 📦 Files You'll Create/Modify

```
internal/
├── middleware/
│   ├── rate_limiter.go            [CREATE]
│   └── rate_limiter_test.go       [CREATE]
└── models/
    └── user.go                    [MODIFY - add tier field]

cmd/api/
└── main.go                        [MODIFY - add rate limit middleware]
```

### 🔄 Implementation Order

1. **Design**: Rate limit strategy and tiers
2. **Implementation**: Rate limiter with Redis
3. **Middleware**: HTTP middleware wrapper
4. **Per-endpoint**: Different limits for different endpoints
5. **Premium**: Tier-based limits
6. **Monitoring**: Metrics and alerts

### ⚠️ Blockers to Watch For

- **Clock skew**: Redis EXPIRE is time-based - ensure clocks synced
- **Distributed systems**: Multiple servers need shared Redis
- **Fail open**: If Redis down, allow requests (or fail closed for security)
- **Header typos**: Rate limit headers must match RFC standards
- **IP spoofing**: Use X-Forwarded-For carefully (can be spoofed)

### ✅ Definition of Done

- [ ] Rate limiting working (100 req/min default)
- [ ] Different limits for different endpoints
- [ ] Premium users have higher limits
- [ ] 429 status returned when limit exceeded
- [ ] Rate limit headers in all responses
- [ ] Metrics tracking violations
- [ ] All tests passing

---

## Week 20: Performance Monitoring

### 📋 Implementation Tasks

**Task 1: Install Prometheus and Grafana** (30 min)
- [ ] Create `docker-compose.yml` for Prometheus + Grafana
- [ ] Create `prometheus.yml` config to scrape app metrics
- [ ] Start containers: `docker-compose up -d`
- [ ] Access Grafana: http://localhost:3000 (admin/admin)
- [ ] Add Prometheus datasource in Grafana

**Task 2: Implement Prometheus Metrics** (90 min)
- [ ] Install client: `go get github.com/prometheus/client_golang/prometheus`
- [ ] Create `internal/middleware/metrics.go`
- [ ] Add HTTP request counter (method, endpoint, status code)
- [ ] Add HTTP request duration histogram
- [ ] Add cache hit/miss counters
- [ ] Add database query duration histogram
- [ ] Expose `/metrics` endpoint

**Task 3: Create Metrics Middleware** (45 min)
- [ ] Implement `MetricsMiddleware(next) http.Handler`
- [ ] Wrap ResponseWriter to capture status code
- [ ] Time request duration
- [ ] Record metrics after request completes
- [ ] Apply to all routes

**Task 4: Add Custom Business Metrics** (60 min)
- [ ] Activities created counter
- [ ] Users registered counter
- [ ] Photos uploaded counter (gauge)
- [ ] Active WebSocket connections (gauge)
- [ ] Background jobs processed counter

**Task 5: Create Grafana Dashboards** (90 min)
- [ ] Create dashboard for HTTP metrics:
  - Request rate (req/sec)
  - Response time (p50, p95, p99)
  - Error rate (4xx, 5xx)
- [ ] Create dashboard for caching:
  - Cache hit ratio
  - Cache size
- [ ] Create dashboard for business metrics:
  - Activities per hour
  - Users registered per day
- [ ] Export dashboards as JSON

**Task 6: Set Up Alerts** (45 min)
- [ ] Configure Prometheus alert rules
- [ ] Alert on high error rate (>1% 5xx)
- [ ] Alert on slow response time (p99 > 2s)
- [ ] Alert on cache miss ratio (> 50%)
- [ ] Test alerts trigger correctly

### 📦 Files You'll Create/Modify

```
docker-compose.yml                 [CREATE]
prometheus.yml                     [CREATE]
grafana/
└── dashboards/
    ├── http_metrics.json          [CREATE]
    ├── cache_metrics.json         [CREATE]
    └── business_metrics.json      [CREATE]

internal/
├── middleware/
│   ├── metrics.go                 [CREATE]
│   └── metrics_test.go            [CREATE]
└── monitoring/
    └── metrics.go                 [CREATE - custom metrics]

cmd/api/
└── main.go                        [MODIFY - add /metrics endpoint]
```

### 🔄 Implementation Order

1. **Setup**: Docker compose → Prometheus → Grafana
2. **Metrics**: HTTP metrics → Custom metrics
3. **Middleware**: Metrics middleware → Apply to routes
4. **Dashboards**: Create visualizations in Grafana
5. **Alerts**: Configure alert rules

### ⚠️ Blockers to Watch For

- **High cardinality**: Don't use user IDs in labels (too many unique values)
- **Label limits**: Prometheus has label count limits
- **Metric naming**: Follow Prometheus naming conventions (snake_case, _total suffix)
- **Histogram buckets**: Configure appropriate buckets for your use case
- **Dashboard overload**: Too many metrics = slow dashboards

### ✅ Definition of Done

- [ ] Prometheus scraping application metrics
- [ ] Grafana dashboards showing HTTP metrics
- [ ] Cache hit ratio visible in dashboard
- [ ] Business metrics tracked
- [ ] Alerts configured and tested
- [ ] Can identify slow endpoints via metrics
- [ ] All services running in Docker Compose

---

## Redis Use Cases

- **Cache activity listings**
  - Reduce database queries
  - Fast response times
  - Pagination support

- **Cache user statistics**
  - Daily/weekly/monthly stats
  - Expensive aggregate queries
  - Refresh on new activity

- **Session storage**
  - Store refresh tokens
  - User sessions
  - Temporary data

- **Rate limit counters**
  - Track requests per user
  - Sliding window algorithm
  - Automatic expiration

---

## Database Patterns

- 🔴 **Soft deletes with `deleted_at` timestamp**
  - Preserve data for auditing
  - Enable "undo" functionality
  - Comply with data retention policies
  - Support data recovery

- **Cache invalidation on data changes**
  - Delete cache on UPDATE/DELETE
  - Use cache tags for bulk invalidation
  - Event-driven invalidation

- **Query result caching strategies**
  - Cache expensive joins
  - Cache paginated results
  - Cache user-specific data

---

## Performance Improvements

### Cache-Aside Pattern
```go
import "github.com/redis/go-redis/v9"

type ActivityService struct {
    repo  *repository.ActivityRepository
    cache *redis.Client
}

func (s *ActivityService) GetActivity(ctx context.Context, id int) (*models.Activity, error) {
    // 1. Try cache first
    cacheKey := fmt.Sprintf("activity:%d", id)
    cached, err := s.cache.Get(ctx, cacheKey).Result()

    if err == nil {
        // Cache hit - unmarshal and return
        var activity models.Activity
        if err := json.Unmarshal([]byte(cached), &activity); err == nil {
            return &activity, nil
        }
    }

    // 2. Cache miss - fetch from database
    activity, err := s.repo.GetByID(ctx, id)
    if err != nil {
        return nil, err
    }

    // 3. Store in cache for future requests
    data, _ := json.Marshal(activity)
    s.cache.Set(ctx, cacheKey, data, 5*time.Minute) // 5 min TTL

    return activity, nil
}
```

### 🔴 Soft Deletes Implementation
```sql
-- Add deleted_at column to tables
ALTER TABLE activities ADD COLUMN deleted_at TIMESTAMP NULL;
ALTER TABLE users ADD COLUMN deleted_at TIMESTAMP NULL;

-- Index for faster queries
CREATE INDEX idx_activities_deleted_at ON activities(deleted_at) WHERE deleted_at IS NULL;
```

```go
// Soft delete instead of hard delete
func (r *ActivityRepository) SoftDelete(ctx context.Context, id int) error {
    query := `
        UPDATE activities
        SET deleted_at = NOW()
        WHERE id = $1 AND deleted_at IS NULL
    `
    result, err := r.db.ExecContext(ctx, query, id)
    if err != nil {
        return err
    }

    rows, _ := result.RowsAffected()
    if rows == 0 {
        return ErrNotFound
    }

    // Invalidate cache
    cacheKey := fmt.Sprintf("activity:%d", id)
    r.cache.Del(ctx, cacheKey)

    return nil
}

// Exclude soft-deleted records in queries
func (r *ActivityRepository) GetByUserID(ctx context.Context, userID int) ([]*models.Activity, error) {
    query := `
        SELECT id, user_id, activity_type, duration_minutes, distance_km, notes, activity_date
        FROM activities
        WHERE user_id = $1 AND deleted_at IS NULL
        ORDER BY activity_date DESC
    `
    // ... rest of implementation
}

// Restore soft-deleted record
func (r *ActivityRepository) Restore(ctx context.Context, id int) error {
    query := `
        UPDATE activities
        SET deleted_at = NULL
        WHERE id = $1 AND deleted_at IS NOT NULL
    `
    _, err := r.db.ExecContext(ctx, query, id)
    return err
}

// Permanent delete (admin only)
func (r *ActivityRepository) PermanentDelete(ctx context.Context, id int) error {
    query := `DELETE FROM activities WHERE id = $1`
    _, err := r.db.ExecContext(ctx, query, id)
    return err
}
```

### Cache Invalidation
```go
func (s *ActivityService) UpdateActivity(ctx context.Context, id int, updates *models.Activity) error {
    // Update database
    if err := s.repo.Update(ctx, id, updates); err != nil {
        return err
    }

    // Invalidate cache
    cacheKey := fmt.Sprintf("activity:%d", id)
    s.cache.Del(ctx, cacheKey)

    // Also invalidate list cache for this user
    userListKey := fmt.Sprintf("activities:user:%d", updates.UserID)
    s.cache.Del(ctx, userListKey)

    return nil
}
```

### Rate Limiting Middleware
```go
import "github.com/redis/go-redis/v9"

type RateLimiter struct {
    redis  *redis.Client
    limit  int           // requests allowed
    window time.Duration // time window
}

func NewRateLimiter(redis *redis.Client, limit int, window time.Duration) *RateLimiter {
    return &RateLimiter{
        redis:  redis,
        limit:  limit,
        window: window,
    }
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Get user ID or IP address
        identifier := getUserID(r.Context())
        if identifier == 0 {
            identifier = getIPAddress(r)
        }

        key := fmt.Sprintf("ratelimit:%v", identifier)

        // Increment counter
        count, err := rl.redis.Incr(r.Context(), key).Result()
        if err != nil {
            // On Redis error, allow request (fail open)
            next.ServeHTTP(w, r)
            return
        }

        // Set expiration on first request
        if count == 1 {
            rl.redis.Expire(r.Context(), key, rl.window)
        }

        // Check if limit exceeded
        if count > int64(rl.limit) {
            w.Header().Set("X-RateLimit-Limit", strconv.Itoa(rl.limit))
            w.Header().Set("X-RateLimit-Remaining", "0")
            w.Header().Set("Retry-After", strconv.Itoa(int(rl.window.Seconds())))

            http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
            return
        }

        // Add rate limit headers
        w.Header().Set("X-RateLimit-Limit", strconv.Itoa(rl.limit))
        w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(rl.limit-int(count)))

        next.ServeHTTP(w, r)
    })
}

// Usage in router
rateLimiter := NewRateLimiter(redisClient, 100, time.Minute) // 100 req/min
router.Use(rateLimiter.Middleware)
```

---

## Monitoring

### Prometheus Metrics
```go
import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
    httpRequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "endpoint", "status"},
    )

    httpRequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Help:    "HTTP request duration in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "endpoint"},
    )

    cacheHits = promauto.NewCounter(
        prometheus.CounterOpts{
            Name: "cache_hits_total",
            Help: "Total number of cache hits",
        },
    )

    cacheMisses = promauto.NewCounter(
        prometheus.CounterOpts{
            Name: "cache_misses_total",
            Help: "Total number of cache misses",
        },
    )
)

// Metrics middleware
func MetricsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()

        // Wrap response writer to capture status code
        wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

        next.ServeHTTP(wrapped, r)

        duration := time.Since(start).Seconds()

        httpRequestsTotal.WithLabelValues(
            r.Method,
            r.URL.Path,
            strconv.Itoa(wrapped.statusCode),
        ).Inc()

        httpRequestDuration.WithLabelValues(
            r.Method,
            r.URL.Path,
        ).Observe(duration)
    })
}

// Expose metrics endpoint
router.Handle("/metrics", promhttp.Handler())
```

### Grafana Dashboard
```yaml
# docker-compose.yml
services:
  prometheus:
    image: prom/prometheus:latest
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
    ports:
      - "9090:9090"

  grafana:
    image: grafana/grafana:latest
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
    volumes:
      - grafana-storage:/var/lib/grafana

volumes:
  grafana-storage:
```

---

## Performance Tips

1. **Cache hot data only**
   - Don't cache everything
   - Focus on frequently accessed data
   - Monitor cache hit ratio

2. **Set appropriate TTLs**
   - Short TTL for frequently changing data (1-5 min)
   - Long TTL for static data (1 hour+)
   - No TTL for rarely changing data

3. **Use connection pooling**
   - Redis client pools connections
   - Configure max connections
   - Monitor connection usage

4. **Batch operations when possible**
   - Use MGET for multiple keys
   - Pipeline commands
   - Reduce network round trips

5. **Monitor cache memory**
   - Set maxmemory in Redis
   - Use LRU eviction policy
   - Monitor memory usage

---

## Common Pitfalls

1. **Caching without invalidation**
   - ❌ Stale data forever
   - ✅ Invalidate on updates

2. **Over-caching**
   - ❌ Cache everything, waste memory
   - ✅ Cache strategically

3. **Hard deletes**
   - ❌ Lose data permanently
   - ✅ Use soft deletes for user data

4. **No rate limiting**
   - ❌ API abuse, DDoS vulnerable
   - ✅ Implement rate limiting

5. **Ignoring monitoring**
   - ❌ No visibility into performance
   - ✅ Set up metrics and dashboards

---

## Resources

- [Redis Documentation](https://redis.io/docs/)
- [go-redis Client](https://redis.uptrace.dev/)
- [Prometheus Documentation](https://prometheus.io/docs/)
- [Grafana Tutorials](https://grafana.com/tutorials/)
- [Caching Strategies](https://aws.amazon.com/caching/best-practices/)

---

## Next Steps

After completing Month 5, you'll move to **Month 6: Background Jobs & Email**, where you'll learn:
- Job queue systems
- Email integration
- Scheduled tasks (cron)
- Export features (PDF/CSV)

**Your API is now fast and scalable!** ⚡
