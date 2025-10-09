package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"

	"golang-chi-hexagonal-base/internal/adapter/http/response"
)

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

				errorDetail := response.ErrorDetail{
					Code:    response.ErrCodeInternal,
					Message: "Internal server error",
				}

				json.NewEncoder(w).Encode(response.NewErrorResponseWrapper(errorDetail))
			}
		}()

		next.ServeHTTP(w, r)
	})
}
