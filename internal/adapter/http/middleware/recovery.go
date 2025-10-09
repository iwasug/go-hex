package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"
)

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail contains error information
type ErrorDetail struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// Recovery is a middleware that recovers from panics
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Log panic with stack trace
				stack := debug.Stack()
				fmt.Printf("[PANIC] %v\n%s\n", err, stack)

				// Return 500 error response
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)

				errorResponse := ErrorResponse{
					Error: ErrorDetail{
						Code:    "INTERNAL_ERROR",
						Message: "Internal server error",
					},
				}

				json.NewEncoder(w).Encode(errorResponse)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
