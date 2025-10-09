package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"golang-chi-hexagonal-base/internal/adapter/crypto"
)

// Auth is a middleware that validates JWT tokens and injects user information into context
func Auth(tokenManager crypto.TokenManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing authorization header")
				return
			}

			// Check if it's a Bearer token
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid authorization header format")
				return
			}

			tokenString := parts[1]

			// Validate token
			claims, err := tokenManager.ValidateToken(tokenString)
			if err != nil {
				if err == crypto.ErrExpiredToken {
					respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Token has expired")
					return
				}
				respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid token")
				return
			}

			// Check if it's an access token
			if claims.Type != crypto.AccessToken {
				respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid token type")
				return
			}

			// Inject user information into context
			ctx := r.Context()
			ctx = context.WithValue(ctx, "user_id", claims.UserID)
			ctx = context.WithValue(ctx, "user_email", claims.Email)
			ctx = context.WithValue(ctx, "user_roles", claims.Roles)

			// Continue with the request
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// respondWithError sends an error response
func respondWithError(w http.ResponseWriter, statusCode int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	errorResponse := ErrorResponse{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
		},
	}

	json.NewEncoder(w).Encode(errorResponse)
}
