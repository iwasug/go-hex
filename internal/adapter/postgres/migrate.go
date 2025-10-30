package postgres

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// MigrateConfig holds migration configuration
type MigrateConfig struct {
	MigrationsPath string
	Enabled        bool
}

// RunMigrations runs database migrations
func RunMigrations(db *sql.DB, config MigrateConfig) error {
	if !config.Enabled {
		return nil
	}

	if config.MigrationsPath == "" {
		return fmt.Errorf("migrations path is required when auto-migration is enabled")
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(config.MigrationsPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Verify the directory exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("migrations directory does not exist: %s", absPath)
	}

	// Convert to proper file URL format for Windows
	// Replace backslashes with forward slashes and ensure proper file:// prefix
	sourceURL := "file://" + strings.ReplaceAll(absPath, "\\", "/")

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		sourceURL,
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migration instance (path: %s): %w", sourceURL, err)
	}

	// Run migrations
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// GetMigrationVersion returns the current migration version
func GetMigrationVersion(db *sql.DB, config MigrateConfig) (uint, bool, error) {
	if config.MigrationsPath == "" {
		return 0, false, fmt.Errorf("migrations path is required")
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(config.MigrationsPath)
	if err != nil {
		return 0, false, fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Convert to proper file URL format for Windows
	sourceURL := "file://" + strings.ReplaceAll(absPath, "\\", "/")

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return 0, false, fmt.Errorf("failed to create migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		sourceURL,
		"postgres",
		driver,
	)
	if err != nil {
		return 0, false, fmt.Errorf("failed to create migration instance: %w", err)
	}

	version, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return 0, false, fmt.Errorf("failed to get migration version: %w", err)
	}

	return version, dirty, nil
}
