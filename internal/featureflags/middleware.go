package featureflags

import (
	"net/http"

	"github.com/valentinesamuel/activelog/pkg/response"
)

// Middleware wraps feature flag checks for HTTP routes
type Middleware struct {
	flags *FeatureFlags
}

// NewMiddleware creates a Middleware for the given flags
func NewMiddleware(flags *FeatureFlags) *Middleware {
	return &Middleware{flags: flags}
}

// Check returns an HTTP middleware that responds with 403 when the feature is disabled
func (m *Middleware) Check(feature string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !m.flags.IsEnabled(feature) {
				response.Error(w, http.StatusForbidden, "feature_not_available")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
