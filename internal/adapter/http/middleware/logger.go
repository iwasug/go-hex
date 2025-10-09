package middleware

import (
	"fmt"
	"net/http"
	"time"
)

// loggerResponseWriter wraps http.ResponseWriter to capture status code
type loggerResponseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func (rw *loggerResponseWriter) WriteHeader(statusCode int) {
	if !rw.written {
		rw.statusCode = statusCode
		rw.written = true
		rw.ResponseWriter.WriteHeader(statusCode)
	}
}

func (rw *loggerResponseWriter) Write(b []byte) (int, error) {
	if !rw.written {
		rw.WriteHeader(http.StatusOK)
	}
	return rw.ResponseWriter.Write(b)
}

// Logger is a middleware that logs HTTP requests
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract request ID from context (set by RequestID middleware)
		requestID := GetRequestID(r.Context())
		if requestID == "" {
			requestID = "unknown"
		}

		// Extract user_id from context if available
		userID := ""
		if uid := r.Context().Value("user_id"); uid != nil {
			userID = fmt.Sprintf("%v", uid)
		}

		// Record start time
		start := time.Now()

		// Wrap response writer to capture status code
		wrapped := &loggerResponseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
			written:        false,
		}

		// Log incoming request
		fmt.Printf("[INFO] [%s] Request started - Method: %s, Path: %s, IP: %s, UserID: %s\n",
			requestID, r.Method, r.URL.Path, getIPAddress(r), userID)

		// Process request
		next.ServeHTTP(wrapped, r)

		// Calculate duration
		duration := time.Since(start)

		// Log response
		fmt.Printf("[INFO] [%s] Request completed - Method: %s, Path: %s, Status: %d, Duration: %dms, UserID: %s\n",
			requestID, r.Method, r.URL.Path, wrapped.statusCode, duration.Milliseconds(), userID)
	})
}
