package middleware

import (
	"net/http"
	"strings"
)

func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// prevent MIME type sniffing
			w.Header().Set("X-Content-Type-Options", "nosniff")

			// prevent click jacking attacks
			w.Header().Set("X-Frame-Options", "DENY")

			// Enable XSS protection (older browsers)
			w.Header().Set("X-XSS-Protection", "1; mode=block")

			// Force HTTPS for 1 year (only enable when using HTTPS!)
			// w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

			// Content Security Policy
			// Swagger UI requires inline scripts and styles to function
			if strings.HasPrefix(r.URL.Path, "/swagger/") {
				w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data:")
			} else {
				w.Header().Set("Content-Security-Policy", "default-src 'self'")
			}

			next.ServeHTTP(w, r)
		},
	)
}
