package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	if !rw.written {
		rw.statusCode = statusCode
		rw.written = true
		rw.ResponseWriter.WriteHeader(statusCode)
	}
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.written {
		rw.WriteHeader(http.StatusOK)
	}
	return rw.ResponseWriter.Write(b)
}

// Logger is a middleware that logs HTTP requests
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Generate request ID for tracing
		requestID := uuid.New().String()
		
		// Add request ID to response header
		w.Header().Set("X-Request-ID", requestID)

		// Extract user_id from context if available
		userID := ""
		if uid := r.Context().Value("user_id"); uid != nil {
			userID = uid.(string)
		}

		// Record start time
		start := time.Now()

		// Wrap response writer to capture status code
		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
			written:        false,
		}

		// Log incoming request
		fmt.Printf("[INFO] [%s] Request started - Method: %s, Path: %s, UserID: %s\n",
			requestID, r.Method, r.URL.Path, userID)

		// Process request
		next.ServeHTTP(wrapped, r)

		// Calculate duration
		duration := time.Since(start)

		// Log response
		fmt.Printf("[INFO] [%s] Request completed - Method: %s, Path: %s, Status: %d, Duration: %dms, UserID: %s\n",
			requestID, r.Method, r.URL.Path, wrapped.statusCode, duration.Milliseconds(), userID)
	})
}
