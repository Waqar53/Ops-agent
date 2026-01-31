package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/opsagent/opsagent/internal/auth"
)

type contextKey string

const UserContextKey contextKey = "user"

// AuthMiddleware validates JWT tokens or API keys
func AuthMiddleware(authService *auth.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Extract token
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 {
				http.Error(w, "Invalid authorization header", http.StatusUnauthorized)
				return
			}

			var claims *auth.Claims
			var err error

			// Check if it's a Bearer token (JWT) or API key
			if parts[0] == "Bearer" {
				claims, err = authService.VerifyToken(parts[1])
			} else if parts[0] == "ApiKey" {
				claims, err = authService.VerifyAPIKey(parts[1])
			} else {
				http.Error(w, "Invalid authorization type", http.StatusUnauthorized)
				return
			}

			if err != nil {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			// Add user to context
			ctx := context.WithValue(r.Context(), UserContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OptionalAuthMiddleware validates tokens but doesn't require them
func OptionalAuthMiddleware(authService *auth.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				parts := strings.Split(authHeader, " ")
				if len(parts) == 2 {
					var claims *auth.Claims
					var err error

					if parts[0] == "Bearer" {
						claims, err = authService.VerifyToken(parts[1])
					} else if parts[0] == "ApiKey" {
						claims, err = authService.VerifyAPIKey(parts[1])
					}

					if err == nil {
						ctx := context.WithValue(r.Context(), UserContextKey, claims)
						next.ServeHTTP(w, r.WithContext(ctx))
						return
					}
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetUser retrieves user claims from context
func GetUser(r *http.Request) *auth.Claims {
	user, _ := r.Context().Value(UserContextKey).(*auth.Claims)
	return user
}

// RequireRole checks if user has required role
func RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := GetUser(r)
			if user == nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// TODO: Check user role from database
			// For now, allow all authenticated users
			next.ServeHTTP(w, r)
		})
	}
}
