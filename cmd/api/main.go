package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	httpAdapter "golang-chi-hexagonal-base/internal/adapter/http"
	"golang-chi-hexagonal-base/internal/adapter/postgres"
	"golang-chi-hexagonal-base/internal/config"
	"golang-chi-hexagonal-base/internal/pkg/logger"

	_ "golang-chi-hexagonal-base/docs"
)

// @title Go Hexagonal API
// @version 1.0
// @description Production-ready Go API using Chi router with Hexagonal Architecture (Ports and Adapters pattern). Features authentication, user management, and role-based access control (RBAC).
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.example.com/support
// @contact.email support@example.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

// @tag.name Auth
// @tag.description Authentication endpoints for user registration, login, and token management

// @tag.name Users
// @tag.description User management endpoints (requires authentication)

// @tag.name Roles
// @tag.description Role management endpoints (requires authentication and admin privileges)

// @tag.name Health
// @tag.description Health check and monitoring endpoints

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

	// Run database migrations if enabled
	if cfg.Database.AutoMigrate {
		log.Info("Running database migrations...", logger.Fields{
			"migrations_path": cfg.Database.MigrationsPath,
		})

		migrateConfig := postgres.MigrateConfig{
			MigrationsPath: cfg.Database.MigrationsPath,
			Enabled:        cfg.Database.AutoMigrate,
		}

		if err := postgres.RunMigrations(db.DB, migrateConfig); err != nil {
			log.Error("Failed to run migrations", logger.Fields{
				"error": err.Error(),
			})
			os.Exit(1)
		}

		// Get and log current migration version
		version, dirty, err := postgres.GetMigrationVersion(db.DB, migrateConfig)
		if err != nil {
			log.Warn("Failed to get migration version", logger.Fields{
				"error": err.Error(),
			})
		} else {
			log.Info("Database migrations completed", logger.Fields{
				"version": version,
				"dirty":   dirty,
			})
		}
	} else {
		log.Info("Auto-migration is disabled")
	}

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
