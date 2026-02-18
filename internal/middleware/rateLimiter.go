package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	cacheTypes "github.com/valentinesamuel/activelog/internal/cache/types"
	"github.com/valentinesamuel/activelog/internal/config"
	queueTypes "github.com/valentinesamuel/activelog/internal/queue/types"
	requestcontext "github.com/valentinesamuel/activelog/internal/requestContext"
)

// getClientIP extracts the client IP address from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (for proxies/load balancers)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP in the list
		if idx := strings.Index(xff, ","); idx != -1 {
			return strings.TrimSpace(xff[:idx])
		}
		return strings.TrimSpace(xff)
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	// Fall back to RemoteAddr (strip port)
	ip := r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	return ip
}

// cachedRateLimitConfig is the schema stored in Redis.
type cachedRateLimitConfig struct {
	CachedAt time.Time              `json:"cached_at"`
	Config   config.RateLimitConfig `json:"config"`
}

var (
	rateLimitConfigOpts = cacheTypes.CacheOptions{
		DB:           cacheTypes.CacheDBRateLimits,
		PartitionKey: cacheTypes.CachePartitionRateLimitConfig,
	}
	rateLimitLockOpts = cacheTypes.CacheOptions{
		DB:           cacheTypes.CacheDBRateLimits,
		PartitionKey: cacheTypes.CachePartitionRateLimitConfig,
	}
	rateLimitCounterOpts = cacheTypes.CacheOptions{
		DB:           cacheTypes.CacheDBRateLimits,
		PartitionKey: cacheTypes.CachePartitionRateLimitCounters,
	}
)

const (
	rateLimitConfigTTL  = 48 * time.Hour
	refreshLockTTL      = 5 * time.Minute
	staleThreshold      = 3 * time.Minute
)

type RateLimiter struct {
	rlCache     cacheTypes.RateLimitCacheProvider
	configCache cacheTypes.CacheAdapter
	queue       queueTypes.QueueProvider
	fallback    *config.RateLimitConfig
}

func NewRateLimiter(
	rlCache cacheTypes.RateLimitCacheProvider,
	configCache cacheTypes.CacheAdapter,
	queue queueTypes.QueueProvider,
	cfg *config.RateLimitConfig,
) *RateLimiter {
	return &RateLimiter{
		rlCache:     rlCache,
		configCache: configCache,
		queue:       queue,
		fallback:    cfg,
	}
}

// resolveConfig returns the rate limit config, preferring the cached value
// from Redis. Falls back to the in-memory config on any error.
// If the cached value is near expiry, it triggers a background refresh.
func (rl *RateLimiter) resolveConfig(ctx context.Context) *config.RateLimitConfig {
	raw, err := rl.configCache.Get(ctx, "config", rateLimitConfigOpts)
	if err != nil || raw == "" {
		return rl.fallback
	}

	var cached cachedRateLimitConfig
	if err := json.Unmarshal([]byte(raw), &cached); err != nil {
		return rl.fallback
	}

	// Recompile patterns (lost during JSON round-trip due to unexported field)
	cached.Config.CompilePatterns()

	// Check if near expiry — trigger stale-while-revalidate in background
	age := time.Since(cached.CachedAt)
	if age > rateLimitConfigTTL-staleThreshold {
		go rl.tryEnqueueRefresh()
	}

	return &cached.Config
}

// tryEnqueueRefresh tries to acquire a refresh lock and enqueues the refresh
// job if successful. Runs in a goroutine so it never blocks requests.
func (rl *RateLimiter) tryEnqueueRefresh() {
	ctx := context.Background()
	acquired, err := rl.rlCache.SetNX(ctx, "refresh_lock", "1", refreshLockTTL, rateLimitLockOpts)
	if err != nil || !acquired {
		return
	}

	payload := queueTypes.JobPayload{Event: queueTypes.EventRefreshRateLimitConfig}
	if _, err := rl.queue.Enqueue(ctx, queueTypes.InboxQueue, payload); err != nil {
		log.Printf("Warning: failed to enqueue rate limit config refresh: %v", err)
	}
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Resolve config (cached or fallback)
		cfg := rl.resolveConfig(ctx)

		// Look up limit for this method + path
		limit, window := cfg.FindRule(r.Method, r.URL.Path)

		// Build key with method for separate counters
		var key string
		if requestUser, ok := requestcontext.FromContext(ctx); ok && requestUser != nil && requestUser.Id != 0 {
			key = fmt.Sprintf("ratelimit:user:%d:%s:%s", requestUser.Id, r.Method, r.URL.Path)
		} else {
			key = fmt.Sprintf("ratelimit:ip:%s:%s:%s", getClientIP(r), r.Method, r.URL.Path)
		}

		// Increment counter
		count, err := rl.rlCache.Increment(ctx, key, rateLimitCounterOpts)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		if count == 1 {
			rl.rlCache.Expire(ctx, key, window, rateLimitCounterOpts)
		}

		if count > int64(limit) {
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(limit))
			w.Header().Set("X-RateLimit-Remaining", "0")
			w.Header().Set("X-Retry-After", strconv.Itoa(int(window.Seconds())))

			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		w.Header().Set("X-RateLimit-Limit", strconv.Itoa(limit))
		w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(limit-int(count)))
		next.ServeHTTP(w, r)
	})
}
