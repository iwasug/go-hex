package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"golang-chi-hexagonal-base/internal/domain/repository"
)

// RequirePermission is a middleware that checks if the user has the required permission
// Permission format: "resource:action" (e.g., "users:read", "roles:write")
func RequirePermission(userRepo repository.UserRepository, requiredPermission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract user_id from context (should be set by Auth middleware)
			userID := getUserIDFromContext(r.Context())
			if userID == "" {
				respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
				return
			}

			// Get user permissions from database
			permissions, err := userRepo.GetUserPermissions(r.Context(), userID)
			if err != nil {
				fmt.Printf("[ERROR] Failed to get user permissions: %v\n", err)
				respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to check permissions")
				return
			}

			// Check if user has the required permission
			hasPermission := false
			for _, perm := range permissions {
				// Permission name format: "resource:action"
				permName := fmt.Sprintf("%s:%s", perm.Resource, perm.Action)
				if permName == requiredPermission {
					hasPermission = true
					break
				}
			}

			if !hasPermission {
				respondWithError(w, http.StatusForbidden, "FORBIDDEN", "Insufficient permissions")
				return
			}

			// User has permission, continue
			next.ServeHTTP(w, r)
		})
	}
}

// RequireAnyPermission checks if the user has at least one of the required permissions
func RequireAnyPermission(userRepo repository.UserRepository, requiredPermissions ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract user_id from context
			userID := getUserIDFromContext(r.Context())
			if userID == "" {
				respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
				return
			}

			// Get user permissions from database
			permissions, err := userRepo.GetUserPermissions(r.Context(), userID)
			if err != nil {
				fmt.Printf("[ERROR] Failed to get user permissions: %v\n", err)
				respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to check permissions")
				return
			}

			// Build permission map for quick lookup
			userPermissions := make(map[string]bool)
			for _, perm := range permissions {
				permName := fmt.Sprintf("%s:%s", perm.Resource, perm.Action)
				userPermissions[permName] = true
			}

			// Check if user has any of the required permissions
			hasPermission := false
			for _, requiredPerm := range requiredPermissions {
				if userPermissions[requiredPerm] {
					hasPermission = true
					break
				}
			}

			if !hasPermission {
				respondWithError(w, http.StatusForbidden, "FORBIDDEN", "Insufficient permissions")
				return
			}

			// User has at least one required permission, continue
			next.ServeHTTP(w, r)
		})
	}
}

// RequireAllPermissions checks if the user has all of the required permissions
func RequireAllPermissions(userRepo repository.UserRepository, requiredPermissions ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract user_id from context
			userID := getUserIDFromContext(r.Context())
			if userID == "" {
				respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
				return
			}

			// Get user permissions from database
			permissions, err := userRepo.GetUserPermissions(r.Context(), userID)
			if err != nil {
				fmt.Printf("[ERROR] Failed to get user permissions: %v\n", err)
				respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to check permissions")
				return
			}

			// Build permission map for quick lookup
			userPermissions := make(map[string]bool)
			for _, perm := range permissions {
				permName := fmt.Sprintf("%s:%s", perm.Resource, perm.Action)
				userPermissions[permName] = true
			}

			// Check if user has all required permissions
			for _, requiredPerm := range requiredPermissions {
				if !userPermissions[requiredPerm] {
					respondWithError(w, http.StatusForbidden, "FORBIDDEN", "Insufficient permissions")
					return
				}
			}

			// User has all required permissions, continue
			next.ServeHTTP(w, r)
		})
	}
}

// RequireRole checks if the user has a specific role
func RequireRole(roleName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract user roles from context (set by Auth middleware)
			roles := getUserRolesFromContext(r.Context())
			if len(roles) == 0 {
				respondWithError(w, http.StatusForbidden, "FORBIDDEN", "No roles assigned")
				return
			}

			// Check if user has the required role
			hasRole := false
			for _, role := range roles {
				if role == roleName {
					hasRole = true
					break
				}
			}

			if !hasRole {
				respondWithError(w, http.StatusForbidden, "FORBIDDEN", "Insufficient role")
				return
			}

			// User has the required role, continue
			next.ServeHTTP(w, r)
		})
	}
}

// getUserIDFromContext extracts user_id from context
func getUserIDFromContext(ctx context.Context) string {
	if userID := ctx.Value("user_id"); userID != nil {
		if id, ok := userID.(string); ok {
			return id
		}
	}
	return ""
}

// getUserRolesFromContext extracts user roles from context
func getUserRolesFromContext(ctx context.Context) []string {
	if roles := ctx.Value("user_roles"); roles != nil {
		if r, ok := roles.([]string); ok {
			return r
		}
	}
	return []string{}
}

// parsePermission parses a permission string into resource and action
func parsePermission(permission string) (resource, action string) {
	parts := strings.Split(permission, ":")
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return "", ""
}
