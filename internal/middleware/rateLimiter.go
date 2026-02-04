package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/valentinesamuel/activelog/internal/cache/types"
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
	limit  int
	window time.Duration
}

func NewRateLimiter(cache types.CacheProvider, limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		cache:  cache,
		limit:  limit,
		window: window,
	}
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Determine rate limit key: use user ID if authenticated, otherwise use IP
		var key string
		if requestUser, ok := requestcontext.FromContext(ctx); ok && requestUser != nil && requestUser.Id != 0 {
			key = fmt.Sprintf("ratelimit:user:%d", requestUser.Id)
		} else {
			key = fmt.Sprintf("ratelimit:ip:%s", getClientIP(r))
		}

		// increment counter
		count, err := rl.cache.Increment(key)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		if count == 1 {
			rl.cache.Expire(key, rl.window)
		}

		if count > int64(rl.limit) {
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(rl.limit))
			w.Header().Set("X-RateLimit-Remaining", "0")
			w.Header().Set("X-Retry-After", strconv.Itoa(int(rl.window.Seconds())))

			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		w.Header().Set("X-RateLimit-Limit", strconv.Itoa(rl.limit))
		w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(rl.limit-int(count)))
		next.ServeHTTP(w, r)
	})

}
