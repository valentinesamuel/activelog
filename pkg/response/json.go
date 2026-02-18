package response

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
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
		"result":     normalizeResult(result),
		"path":       r.URL.RequestURI(),
		"duration":   duration,
	})
}

// normalizeResult ensures nil slices become [] and nil maps/pointers become {}
// so the frontend always receives a consistent shape instead of null.
func normalizeResult(result interface{}) interface{} {
	if result == nil {
		return map[string]interface{}{}
	}
	return normalizeValue(reflect.ValueOf(result))
}

func normalizeValue(v reflect.Value) interface{} {
	// Unwrap pointers and interfaces
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		if v.IsNil() {
			return map[string]interface{}{}
		}
		v = v.Elem()
	}
	switch v.Kind() {
	case reflect.Slice:
		if v.IsNil() {
			return []interface{}{}
		}
		return v.Interface()
	case reflect.Map:
		if v.IsNil() {
			return map[string]interface{}{}
		}
		// Walk string-keyed maps to normalize nested nil values
		if v.Type().Key().Kind() == reflect.String {
			out := make(map[string]interface{}, v.Len())
			for _, key := range v.MapKeys() {
				out[key.String()] = normalizeValue(v.MapIndex(key))
			}
			return out
		}
		return v.Interface()
	default:
		return v.Interface()
	}
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
