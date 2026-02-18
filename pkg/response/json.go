package response

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

type contextKey int

var RequestStartKey contextKey = 0

func WithStartTime(ctx context.Context, t time.Time) context.Context {
	return context.WithValue(ctx, RequestStartKey, t)
}

type ValidationErrorItem struct {
	Field  string   `json:"field"`
	Errors []string `json:"errors"`
}

func Success(w http.ResponseWriter, r *http.Request, statusCode int, result interface{}) {
	duration := computeDuration(r.Context())
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"statusCode": statusCode,
		"success":    true,
		"message":    "Request successful",
		"result":     result,
		"path":       r.URL.RequestURI(),
		"duration":   duration,
	})
}

func Fail(w http.ResponseWriter, r *http.Request, statusCode int, message string) {
	duration := computeDuration(r.Context())
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"statusCode": statusCode,
		"success":    false,
		"message":    message,
		"errors":     []interface{}{},
		"path":       r.URL.RequestURI(),
		"duration":   duration,
	})
}

func ValidationFail(w http.ResponseWriter, r *http.Request, errs []ValidationErrorItem) {
	duration := computeDuration(r.Context())
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"statusCode": http.StatusBadRequest,
		"success":    false,
		"message":    "Bad Request",
		"errors":     errs,
		"path":       r.URL.RequestURI(),
		"duration":   duration,
	})
}

func computeDuration(ctx context.Context) float64 {
	if start, ok := ctx.Value(RequestStartKey).(time.Time); ok {
		return float64(time.Since(start).Milliseconds())
	}
	return 0
}
