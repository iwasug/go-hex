package config

import (
	"os"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	// Set required environment variables
	os.Setenv("JWT_SECRET", "this-is-a-test-secret-key-with-at-least-32-characters")
	defer os.Unsetenv("JWT_SECRET")

	config, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Verify default values
	if config.Server.Port != "8080" {
		t.Errorf("Expected default port 8080, got %s", config.Server.Port)
	}

	if config.Database.Host != "localhost" {
		t.Errorf("Expected default DB host localhost, got %s", config.Database.Host)
	}

	if config.Bcrypt.Cost != 10 {
		t.Errorf("Expected default bcrypt cost 10, got %d", config.Bcrypt.Cost)
	}
}

func TestValidate_MissingJWTSecret(t *testing.T) {
	config := &Config{
		Server: ServerConfig{
			Port: "8080",
			Env:  "development",
		},
		Database: DatabaseConfig{
			Host: "localhost",
			Port: "5432",
			User: "postgres",
			Name: "myapp",
		},
		JWT: JWTConfig{
			Secret:             "", // Missing
			AccessTokenExpiry:  15 * time.Minute,
			RefreshTokenExpiry: 168 * time.Hour,
		},
		Bcrypt: BcryptConfig{
			Cost: 10,
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "json",
		},
	}

	err := config.Validate()
	if err == nil {
		t.Error("Expected validation error for missing JWT_SECRET")
	}
}

func TestValidate_ShortJWTSecret(t *testing.T) {
	config := &Config{
		Server: ServerConfig{
			Port: "8080",
			Env:  "development",
		},
		Database: DatabaseConfig{
			Host: "localhost",
			Port: "5432",
			User: "postgres",
			Name: "myapp",
		},
		JWT: JWTConfig{
			Secret:             "short", // Too short
			AccessTokenExpiry:  15 * time.Minute,
			RefreshTokenExpiry: 168 * time.Hour,
		},
		Bcrypt: BcryptConfig{
			Cost: 10,
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "json",
		},
	}

	err := config.Validate()
	if err == nil {
		t.Error("Expected validation error for short JWT_SECRET")
	}
}

func TestValidate_InvalidBcryptCost(t *testing.T) {
	config := &Config{
		Server: ServerConfig{
			Port: "8080",
			Env:  "development",
		},
		Database: DatabaseConfig{
			Host: "localhost",
			Port: "5432",
			User: "postgres",
			Name: "myapp",
		},
		JWT: JWTConfig{
			Secret:             "this-is-a-test-secret-key-with-at-least-32-characters",
			AccessTokenExpiry:  15 * time.Minute,
			RefreshTokenExpiry: 168 * time.Hour,
		},
		Bcrypt: BcryptConfig{
			Cost: 50, // Invalid (too high)
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "json",
		},
	}

	err := config.Validate()
	if err == nil {
		t.Error("Expected validation error for invalid bcrypt cost")
	}
}

func TestGetDSN(t *testing.T) {
	config := &Config{
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     "5432",
			User:     "testuser",
			Password: "testpass",
			Name:     "testdb",
			SSLMode:  "disable",
		},
	}

	expected := "host=localhost port=5432 user=testuser password=testpass dbname=testdb sslmode=disable"
	dsn := config.GetDSN()

	if dsn != expected {
		t.Errorf("Expected DSN %s, got %s", expected, dsn)
	}
}

func TestIsDevelopment(t *testing.T) {
	config := &Config{
		Server: ServerConfig{
			Env: "development",
		},
	}

	if !config.IsDevelopment() {
		t.Error("Expected IsDevelopment() to return true")
	}

	if config.IsProduction() {
		t.Error("Expected IsProduction() to return false")
	}
}

func TestIsProduction(t *testing.T) {
	config := &Config{
		Server: ServerConfig{
			Env: "production",
		},
	}

	if !config.IsProduction() {
		t.Error("Expected IsProduction() to return true")
	}

	if config.IsDevelopment() {
		t.Error("Expected IsDevelopment() to return false")
	}
}
