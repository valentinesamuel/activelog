package middleware

import (
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/valentinesamuel/activelog/internal/config"
	requestcontext "github.com/valentinesamuel/activelog/internal/requestContext"
	"github.com/valentinesamuel/activelog/pkg/auth"
	"github.com/valentinesamuel/activelog/pkg/response"
)

func AuthMiddleware(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			response.Error(w, http.StatusUnauthorized, "Unauthorized request")
			return
		}

		// Parse "Bearer <token>"
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Validate token
		claims := &auth.CustomClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
			return []byte(config.Common.Auth.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			response.Error(w, http.StatusUnauthorized, "Unauthorized request")
			return
		}

		requestUser := &requestcontext.User{
			Id:    claims.UserID,
			Email: claims.Email,
		}
		ctx := requestcontext.NewContext(r.Context(), requestUser)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
