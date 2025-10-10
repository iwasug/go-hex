package middleware

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"

	"golang-chi-hexagonal-base/internal/domain/repository"
)

// InjectAdminStatus is a middleware that checks if the user has admin role
// and injects is_admin flag into context for downstream handlers
func InjectAdminStatus(userRepo repository.UserRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract user_id from context (should be set by Auth middleware)
			userID := getUserIDFromContext(r.Context())
			if userID == "" {
				// If no user ID, continue without admin status
				next.ServeHTTP(w, r)
				return
			}

			// Get user roles from database
			roles, err := userRepo.GetUserRoles(r.Context(), userID)
			if err != nil {
				// Log error but continue without admin status
				next.ServeHTTP(w, r)
				return
			}

			// Check if user has admin role
			isAdmin := false
			for _, role := range roles {
				if role.Name == "admin" {
					isAdmin = true
					break
				}
			}

			// Inject is_admin into context
			ctx := context.WithValue(r.Context(), "is_admin", isAdmin)

			// Continue with the request
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ValidateSessionOwnership is a middleware that validates if the user owns the WhatsApp session
// or is an admin. This should be used on endpoints that operate on specific sessions.
func ValidateSessionOwnership(sessionRepo repository.WhatsAppSessionRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract session ID from URL parameter
			sessionID := chi.URLParam(r, "id")
			if sessionID == "" {
				// No session ID in URL, skip validation (might be a list endpoint)
				next.ServeHTTP(w, r)
				return
			}

			// Extract user_id from context
			userID := getUserIDFromContext(r.Context())
			if userID == "" {
				respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
				return
			}

			// Check if user is admin
			isAdmin := false
			if adminVal := r.Context().Value("is_admin"); adminVal != nil {
				if admin, ok := adminVal.(bool); ok {
					isAdmin = admin
				}
			}

			// If admin, bypass ownership check
			if isAdmin {
				next.ServeHTTP(w, r)
				return
			}

			// Get session from database
			session, err := sessionRepo.FindByID(r.Context(), sessionID)
			if err != nil {
				respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Session not found")
				return
			}

			// Check if user owns the session
			if session.UserID != userID {
				respondWithError(w, http.StatusForbidden, "FORBIDDEN", "You don't have access to this session")
				return
			}

			// User owns the session, continue
			next.ServeHTTP(w, r)
		})
	}
}
