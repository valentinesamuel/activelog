package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/valentinesamuel/activelog/internal/cache/types"
	"github.com/valentinesamuel/activelog/internal/config"
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

type RateLimiter struct {
	cache  types.CacheProvider
	config *config.RateLimitConfig
}

func NewRateLimiter(cache types.CacheProvider, cfg *config.RateLimitConfig) *RateLimiter {
	return &RateLimiter{
		cache:  cache,
		config: cfg,
	}
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Look up limit for this method + path
		limit, window := rl.config.FindRule(r.Method, r.URL.Path)

		// Build key with method for separate counters
		var key string
		if requestUser, ok := requestcontext.FromContext(ctx); ok && requestUser != nil && requestUser.Id != 0 {
			key = fmt.Sprintf("ratelimit:user:%d:%s:%s", requestUser.Id, r.Method, r.URL.Path)
		} else {
			key = fmt.Sprintf("ratelimit:ip:%s:%s:%s", getClientIP(r), r.Method, r.URL.Path)
		}

		// Increment counter
		count, err := rl.cache.Increment(key)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		if count == 1 {
			rl.cache.Expire(key, window)
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
