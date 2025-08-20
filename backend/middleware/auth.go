package middleware

import (
	"context"
	"net/http"
	"strings"

	"laplogger/handlers"
)

type contextKey string

const UserContextKey contextKey = "user"

// JWTMiddleware validates JWT tokens
func JWTMiddleware(authHandler *handlers.AuthHandler) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Authorization header required", http.StatusUnauthorized)
				return
			}

			// Check for Bearer token format
			tokenParts := strings.Split(authHeader, " ")
			if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
				http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
				return
			}

			// Validate token
			claims, err := authHandler.ValidateToken(tokenParts[1])
			if err != nil {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			// Add user info to context
			ctx := context.WithValue(r.Context(), UserContextKey, *claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserFromContext extracts user information from request context
func GetUserFromContext(r *http.Request) (map[string]interface{}, bool) {
	user, ok := r.Context().Value(UserContextKey).(map[string]interface{})
	return user, ok
}
