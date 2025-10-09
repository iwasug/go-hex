package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

// DB wraps sql.DB with additional functionality
type DB struct {
	*sql.DB
}

// DBConfig holds database connection configuration
type DBConfig struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnectTimeout  time.Duration
	RetryAttempts   int
	RetryDelay      time.Duration
}

// NewDB creates a new database connection with connection pooling
func NewDB(config DBConfig) (*DB, error) {
	// Set default values
	if config.ConnectTimeout == 0 {
		config.ConnectTimeout = 10 * time.Second
	}
	if config.RetryAttempts == 0 {
		config.RetryAttempts = 3
	}
	if config.RetryDelay == 0 {
		config.RetryDelay = 2 * time.Second
	}

	var db *sql.DB
	var err error

	// Retry logic for connection
	for attempt := 1; attempt <= config.RetryAttempts; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), config.ConnectTimeout)
		defer cancel()

		db, err = sql.Open("postgres", config.DSN)
		if err != nil {
			if attempt < config.RetryAttempts {
				time.Sleep(config.RetryDelay)
				continue
			}
			return nil, fmt.Errorf("failed to open database connection after %d attempts: %w", config.RetryAttempts, err)
		}

		// Test the connection
		err = db.PingContext(ctx)
		if err != nil {
			db.Close()
			if attempt < config.RetryAttempts {
				time.Sleep(config.RetryDelay)
				continue
			}
			return nil, fmt.Errorf("failed to ping database after %d attempts: %w", config.RetryAttempts, err)
		}

		// Connection successful
		break
	}

	// Configure connection pool
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)

	return &DB{DB: db}, nil
}

// Ping checks if the database connection is alive
func (db *DB) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return db.PingContext(ctx)
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.DB.Close()
}

// Stats returns database statistics
func (db *DB) Stats() sql.DBStats {
	return db.DB.Stats()
}
