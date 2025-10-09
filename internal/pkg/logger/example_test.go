package logger_test

import (
	"golang-chi-hexagonal-base/internal/pkg/logger"
)

func ExampleLogger_Info() {
	// Create a logger with INFO level and JSON format
	log := logger.NewFromConfig("info", "json")

	// Log a simple message
	log.Info("Application started")

	// Log with additional context fields
	log.Info("User logged in", logger.Fields{
		"user_id": "123",
		"email":   "user@example.com",
	})
}

func ExampleLogger_WithFields() {
	// Create a logger
	log := logger.NewFromConfig("info", "json")

	// Create a context logger with default fields
	requestLogger := log.WithFields(logger.Fields{
		"request_id": "abc-123",
		"method":     "GET",
		"path":       "/api/v1/users",
	})

	// All logs from this logger will include the default fields
	requestLogger.Info("Request started")
	requestLogger.Info("Request completed", logger.Fields{
		"status":      200,
		"duration_ms": 45,
	})
}

func ExampleLogger_levels() {
	// Create a logger with DEBUG level
	log := logger.NewFromConfig("debug", "json")

	// Different log levels
	log.Debug("Detailed debugging information")
	log.Info("General information")
	log.Warn("Warning message")
	log.Error("Error occurred", logger.Fields{
		"error": "connection timeout",
	})
}

func ExampleNewFromConfig() {
	// Create logger from configuration strings
	// Level: debug, info, warn, error
	// Format: json, text
	log := logger.NewFromConfig("info", "json")

	log.Info("Logger initialized")
}

func ExampleNew() {
	// Create logger with explicit level and format
	log := logger.New(logger.INFO, logger.JSONFormat, nil)

	log.Info("Custom logger created")
}
