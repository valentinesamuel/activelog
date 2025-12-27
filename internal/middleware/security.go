package middleware

import "net/http"

func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// prevent MEME type sniffing
			w.Header().Set("x-Content-Type-Options", "nosniff")

			// prevent click jacking attachks
			w.Header().Set("x-Frame-Options", "DENY")

			// Enable XSS protection(older browsers)
			w.Header().Set("X-XSS-Protection", "1; mode=block")

			// Force HTTPS for 1 year (only enable when using HTTPS!)
			// w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

			// Content Security Policy (basic - customize as needed)
			w.Header().Set("Content-Security-Policy", "default-src 'self'")

			next.ServeHTTP(w, r)

		},
	)
}
