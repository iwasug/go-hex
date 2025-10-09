package http

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"golang-chi-hexagonal-base/internal/adapter/postgres"
	"golang-chi-hexagonal-base/internal/config"
)

// TestHealthCheckEndpoint tests the health check endpoint
func TestHealthCheckEndpoint(t *testing.T) {
	// Create a minimal config for testing
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port: "8080",
			Env:  "test",
		},
		Database: config.DatabaseConfig{
			Host:            "localhost",
			Port:            "5432",
			User:            "postgres",
			Password:        "postgres",
			Name:            "testdb",
			SSLMode:         "disable",
			MaxOpenConns:    25,
			MaxIdleConns:    5,
			ConnMaxLifetime: 5 * time.Minute,
		},
		JWT: config.JWTConfig{
			Secret:             "test-secret-key-at-least-32-characters-long",
			AccessTokenExpiry:  15 * time.Minute,
			RefreshTokenExpiry: 168 * time.Hour,
		},
		Bcrypt: config.BcryptConfig{
			Cost: 10,
		},
		Logging: config.LoggingConfig{
			Level:  "info",
			Format: "json",
		},
	}

	// Create a mock DB (we don't need actual connection for health check test)
	// In a real scenario, you'd use a test database or mock
	db := &postgres.DB{}

	// Create router
	router := NewRouter(cfg, db)
	handler := router.Setup()

	// Create test request for liveness endpoint (doesn't need real DB)
	req := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	w := httptest.NewRecorder()

	// Execute request
	handler.ServeHTTP(w, req)

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	responseBody := w.Body.String()
	if !strings.Contains(responseBody, `"status":"alive"`) {
		t.Errorf("Expected body to contain status:alive, got %s", responseBody)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
	}
}

// TestRouterSetup tests that the router is properly configured
func TestRouterSetup(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port: "8080",
			Env:  "test",
		},
		Database: config.DatabaseConfig{
			Host:            "localhost",
			Port:            "5432",
			User:            "postgres",
			Password:        "postgres",
			Name:            "testdb",
			SSLMode:         "disable",
			MaxOpenConns:    25,
			MaxIdleConns:    5,
			ConnMaxLifetime: 5 * time.Minute,
		},
		JWT: config.JWTConfig{
			Secret:             "test-secret-key-at-least-32-characters-long",
			AccessTokenExpiry:  15 * time.Minute,
			RefreshTokenExpiry: 168 * time.Hour,
		},
		Bcrypt: config.BcryptConfig{
			Cost: 10,
		},
		Logging: config.LoggingConfig{
			Level:  "info",
			Format: "json",
		},
	}

	db := &postgres.DB{}

	router := NewRouter(cfg, db)
	handler := router.Setup()

	if handler == nil {
		t.Error("Expected non-nil handler")
	}
}
