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
