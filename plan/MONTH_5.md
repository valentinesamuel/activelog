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
