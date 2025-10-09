package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang-chi-hexagonal-base/internal/adapter/postgres"
	httpAdapter "golang-chi-hexagonal-base/internal/adapter/http"
	"golang-chi-hexagonal-base/internal/config"
	"golang-chi-hexagonal-base/internal/pkg/logger"
)

func main() {
	// Initialize logger
	log := logger.NewFromConfig("info", "json")

	log.Info("Starting application...")

	// Load configuration from environment
	cfg, err := config.Load()
	if err != nil {
		log.Error("Failed to load configuration", logger.Fields{
			"error": err.Error(),
		})
		os.Exit(1)
	}

	// Update logger with config settings
	log = logger.NewFromConfig(cfg.Logging.Level, cfg.Logging.Format)

	log.Info("Configuration loaded successfully", logger.Fields{
		"env":     cfg.Server.Env,
		"port":    cfg.Server.Port,
		"db_host": cfg.Database.Host,
	})

	// Initialize database connection
	db, err := initDatabase(cfg, log)
	if err != nil {
		log.Error("Failed to initialize database", logger.Fields{
			"error": err.Error(),
		})
		os.Exit(1)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Error("Error closing database connection", logger.Fields{
				"error": err.Error(),
			})
		}
	}()

	log.Info("Database connection established successfully")

	// Setup HTTP router with all dependencies
	router := httpAdapter.NewRouter(cfg, db)
	handler := router.Setup()

	// Setup HTTP server with graceful shutdown
	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	serverErrors := make(chan error, 1)
	go func() {
		log.Info("Starting HTTP server", logger.Fields{
			"port": cfg.Server.Port,
			"env":  cfg.Server.Env,
		})
		serverErrors <- server.ListenAndServe()
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	// Block until we receive a signal or server error
	select {
	case err := <-serverErrors:
		if err != nil && err != http.ErrServerClosed {
			log.Error("Server error", logger.Fields{
				"error": err.Error(),
			})
			os.Exit(1)
		}
	case sig := <-shutdown:
		log.Info("Shutdown signal received", logger.Fields{
			"signal": sig.String(),
		})

		// Create context with timeout for graceful shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Attempt graceful shutdown
		if err := server.Shutdown(ctx); err != nil {
			log.Error("Error during graceful shutdown", logger.Fields{
				"error": err.Error(),
			})
			// Force close if graceful shutdown fails
			if err := server.Close(); err != nil {
				log.Error("Error forcing server close", logger.Fields{
					"error": err.Error(),
				})
			}
			os.Exit(1)
		}

		log.Info("Server stopped gracefully")
	}
}

// initDatabase initializes the database connection with retry logic
func initDatabase(cfg *config.Config, log *logger.Logger) (*postgres.DB, error) {
	dbConfig := postgres.DBConfig{
		DSN:             cfg.GetDSN(),
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		ConnMaxLifetime: cfg.Database.ConnMaxLifetime,
		ConnectTimeout:  10 * time.Second,
		RetryAttempts:   3,
		RetryDelay:      2 * time.Second,
	}

	log.Info("Connecting to database...", logger.Fields{
		"host":     cfg.Database.Host,
		"port":     cfg.Database.Port,
		"database": cfg.Database.Name,
	})

	db, err := postgres.NewDB(dbConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Verify connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Log connection pool stats
	stats := db.Stats()
	log.Info("Database connection pool configured", logger.Fields{
		"max_open_conns":   cfg.Database.MaxOpenConns,
		"max_idle_conns":   cfg.Database.MaxIdleConns,
		"open_connections": stats.OpenConnections,
	})

	return db, nil
}
