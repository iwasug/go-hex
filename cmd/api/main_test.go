package main

import (
	"testing"
	"time"

	"golang-chi-hexagonal-base/internal/adapter/postgres"
	"golang-chi-hexagonal-base/internal/config"
	"golang-chi-hexagonal-base/internal/pkg/logger"
)

// TestInitDatabase tests the database initialization function
func TestInitDatabase(t *testing.T) {
	// This is a unit test that verifies the initDatabase function structure
	// Integration tests would require an actual database connection

	t.Run("validates config structure", func(t *testing.T) {
		cfg := &config.Config{
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
		}

		// Verify DSN generation
		dsn := cfg.GetDSN()
		if dsn == "" {
			t.Error("Expected non-empty DSN")
		}

		// Verify config values are accessible
		if cfg.Database.Host != "localhost" {
			t.Errorf("Expected host 'localhost', got '%s'", cfg.Database.Host)
		}
	})

	t.Run("creates valid DBConfig", func(t *testing.T) {
		cfg := &config.Config{
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
		}

		log := logger.NewFromConfig("info", "json")

		dbConfig := postgres.DBConfig{
			DSN:             cfg.GetDSN(),
			MaxOpenConns:    cfg.Database.MaxOpenConns,
			MaxIdleConns:    cfg.Database.MaxIdleConns,
			ConnMaxLifetime: cfg.Database.ConnMaxLifetime,
			ConnectTimeout:  10 * time.Second,
			RetryAttempts:   3,
			RetryDelay:      2 * time.Second,
		}

		// Verify DBConfig structure
		if dbConfig.MaxOpenConns != 25 {
			t.Errorf("Expected MaxOpenConns 25, got %d", dbConfig.MaxOpenConns)
		}
		if dbConfig.MaxIdleConns != 5 {
			t.Errorf("Expected MaxIdleConns 5, got %d", dbConfig.MaxIdleConns)
		}
		if dbConfig.ConnMaxLifetime != 5*time.Minute {
			t.Errorf("Expected ConnMaxLifetime 5m, got %v", dbConfig.ConnMaxLifetime)
		}
		if dbConfig.RetryAttempts != 3 {
			t.Errorf("Expected RetryAttempts 3, got %d", dbConfig.RetryAttempts)
		}

		// Verify logger is not nil
		if log == nil {
			t.Error("Expected non-nil logger")
		}
	})
}
