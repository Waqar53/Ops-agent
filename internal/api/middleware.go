package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/opsagent/opsagent/internal/config"
)

// Claims represents JWT claims
type Claims struct {
	UserID         string `json:"user_id"`
	OrganizationID string `json:"org_id"`
	Email          string `json:"email"`
	jwt.RegisteredClaims
}

// ContextKey for request context
type ContextKey string

const (
	ContextUserID  ContextKey = "user_id"
	ContextOrgID   ContextKey = "org_id"
	ContextEmail   ContextKey = "email"
)

// AuthMiddleware validates JWT tokens
func AuthMiddleware(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get token from header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				writeError(w, http.StatusUnauthorized, "missing authorization header")
				return
			}

			// Extract bearer token
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				writeError(w, http.StatusUnauthorized, "invalid authorization format")
				return
			}

			tokenString := parts[1]

			// Parse and validate token
			claims := &Claims{}
			token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
				return []byte(cfg.Auth.JWTSecret), nil
			})

			if err != nil || !token.Valid {
				writeError(w, http.StatusUnauthorized, "invalid token")
				return
			}

			// Add claims to context
			ctx := context.WithValue(r.Context(), ContextUserID, claims.UserID)
			ctx = context.WithValue(ctx, ContextOrgID, claims.OrganizationID)
			ctx = context.WithValue(ctx, ContextEmail, claims.Email)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Helper functions

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func getUserID(r *http.Request) string {
	if id, ok := r.Context().Value(ContextUserID).(string); ok {
		return id
	}
	return ""
}

func getOrgID(r *http.Request) string {
	if id, ok := r.Context().Value(ContextOrgID).(string); ok {
		return id
	}
	return ""
}
