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

func TestWhatsAppConfig_Defaults(t *testing.T) {
	// Set required environment variables
	os.Setenv("JWT_SECRET", "this-is-a-test-secret-key-with-at-least-32-characters")
	defer os.Unsetenv("JWT_SECRET")

	config, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Verify WhatsApp default values
	if !config.WhatsApp.Enabled {
		t.Error("Expected WhatsApp to be enabled by default")
	}

	if config.WhatsApp.QRTimeout != 60*time.Second {
		t.Errorf("Expected default QR timeout 60s, got %v", config.WhatsApp.QRTimeout)
	}

	if config.WhatsApp.ReconnectInterval != 30*time.Second {
		t.Errorf("Expected default reconnect interval 30s, got %v", config.WhatsApp.ReconnectInterval)
	}

	if config.WhatsApp.MaxReconnectAttempts != 5 {
		t.Errorf("Expected default max reconnect attempts 5, got %d", config.WhatsApp.MaxReconnectAttempts)
	}

	if config.WhatsApp.MessageRetentionDays != 90 {
		t.Errorf("Expected default message retention 90 days, got %d", config.WhatsApp.MessageRetentionDays)
	}

	if config.WhatsApp.MediaMaxSizeMB != 16 {
		t.Errorf("Expected default media max size 16MB, got %d", config.WhatsApp.MediaMaxSizeMB)
	}

	if config.WhatsApp.MediaStoragePath != "./data/whatsapp/media" {
		t.Errorf("Expected default media storage path ./data/whatsapp/media, got %s", config.WhatsApp.MediaStoragePath)
	}

	if config.WhatsApp.WebhookTimeout != 10*time.Second {
		t.Errorf("Expected default webhook timeout 10s, got %v", config.WhatsApp.WebhookTimeout)
	}

	if config.WhatsApp.WebhookMaxRetries != 3 {
		t.Errorf("Expected default webhook max retries 3, got %d", config.WhatsApp.WebhookMaxRetries)
	}

	if config.WhatsApp.WebhookRetryDelay != 5*time.Second {
		t.Errorf("Expected default webhook retry delay 5s, got %v", config.WhatsApp.WebhookRetryDelay)
	}
}

func TestWhatsAppConfig_CustomValues(t *testing.T) {
	// Set required environment variables
	os.Setenv("JWT_SECRET", "this-is-a-test-secret-key-with-at-least-32-characters")
	os.Setenv("WHATSAPP_ENABLED", "false")
	os.Setenv("WHATSAPP_QR_TIMEOUT", "120s")
	os.Setenv("WHATSAPP_RECONNECT_INTERVAL", "60s")
	os.Setenv("WHATSAPP_MAX_RECONNECT_ATTEMPTS", "10")
	os.Setenv("WHATSAPP_MESSAGE_RETENTION_DAYS", "30")
	os.Setenv("WHATSAPP_MEDIA_MAX_SIZE_MB", "32")
	os.Setenv("WHATSAPP_MEDIA_STORAGE_PATH", "/custom/path")
	os.Setenv("WHATSAPP_WEBHOOK_TIMEOUT", "20s")
	os.Setenv("WHATSAPP_WEBHOOK_MAX_RETRIES", "5")
	os.Setenv("WHATSAPP_WEBHOOK_RETRY_DELAY", "10s")

	defer func() {
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("WHATSAPP_ENABLED")
		os.Unsetenv("WHATSAPP_QR_TIMEOUT")
		os.Unsetenv("WHATSAPP_RECONNECT_INTERVAL")
		os.Unsetenv("WHATSAPP_MAX_RECONNECT_ATTEMPTS")
		os.Unsetenv("WHATSAPP_MESSAGE_RETENTION_DAYS")
		os.Unsetenv("WHATSAPP_MEDIA_MAX_SIZE_MB")
		os.Unsetenv("WHATSAPP_MEDIA_STORAGE_PATH")
		os.Unsetenv("WHATSAPP_WEBHOOK_TIMEOUT")
		os.Unsetenv("WHATSAPP_WEBHOOK_MAX_RETRIES")
		os.Unsetenv("WHATSAPP_WEBHOOK_RETRY_DELAY")
	}()

	config, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Verify custom WhatsApp values
	if config.WhatsApp.Enabled {
		t.Error("Expected WhatsApp to be disabled")
	}

	if config.WhatsApp.QRTimeout != 120*time.Second {
		t.Errorf("Expected QR timeout 120s, got %v", config.WhatsApp.QRTimeout)
	}

	if config.WhatsApp.ReconnectInterval != 60*time.Second {
		t.Errorf("Expected reconnect interval 60s, got %v", config.WhatsApp.ReconnectInterval)
	}

	if config.WhatsApp.MaxReconnectAttempts != 10 {
		t.Errorf("Expected max reconnect attempts 10, got %d", config.WhatsApp.MaxReconnectAttempts)
	}

	if config.WhatsApp.MessageRetentionDays != 30 {
		t.Errorf("Expected message retention 30 days, got %d", config.WhatsApp.MessageRetentionDays)
	}

	if config.WhatsApp.MediaMaxSizeMB != 32 {
		t.Errorf("Expected media max size 32MB, got %d", config.WhatsApp.MediaMaxSizeMB)
	}

	if config.WhatsApp.MediaStoragePath != "/custom/path" {
		t.Errorf("Expected media storage path /custom/path, got %s", config.WhatsApp.MediaStoragePath)
	}

	if config.WhatsApp.WebhookTimeout != 20*time.Second {
		t.Errorf("Expected webhook timeout 20s, got %v", config.WhatsApp.WebhookTimeout)
	}

	if config.WhatsApp.WebhookMaxRetries != 5 {
		t.Errorf("Expected webhook max retries 5, got %d", config.WhatsApp.WebhookMaxRetries)
	}

	if config.WhatsApp.WebhookRetryDelay != 10*time.Second {
		t.Errorf("Expected webhook retry delay 10s, got %v", config.WhatsApp.WebhookRetryDelay)
	}
}

func TestValidate_WhatsAppConfig_InvalidQRTimeout(t *testing.T) {
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
			Cost: 10,
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "json",
		},
		WhatsApp: WhatsAppConfig{
			Enabled:              true,
			QRTimeout:            0, // Invalid
			ReconnectInterval:    30 * time.Second,
			MaxReconnectAttempts: 5,
			MessageRetentionDays: 90,
			MediaMaxSizeMB:       16,
			MediaStoragePath:     "./data/whatsapp/media",
			WebhookTimeout:       10 * time.Second,
			WebhookMaxRetries:    3,
			WebhookRetryDelay:    5 * time.Second,
		},
	}

	err := config.Validate()
	if err == nil {
		t.Error("Expected validation error for invalid QR timeout")
	}
}

func TestValidate_WhatsAppConfig_InvalidMediaMaxSize(t *testing.T) {
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
			Cost: 10,
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "json",
		},
		WhatsApp: WhatsAppConfig{
			Enabled:              true,
			QRTimeout:            60 * time.Second,
			ReconnectInterval:    30 * time.Second,
			MaxReconnectAttempts: 5,
			MessageRetentionDays: 90,
			MediaMaxSizeMB:       0, // Invalid
			MediaStoragePath:     "./data/whatsapp/media",
			WebhookTimeout:       10 * time.Second,
			WebhookMaxRetries:    3,
			WebhookRetryDelay:    5 * time.Second,
		},
	}

	err := config.Validate()
	if err == nil {
		t.Error("Expected validation error for invalid media max size")
	}
}

func TestValidate_WhatsAppConfig_MissingMediaStoragePath(t *testing.T) {
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
			Cost: 10,
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "json",
		},
		WhatsApp: WhatsAppConfig{
			Enabled:              true,
			QRTimeout:            60 * time.Second,
			ReconnectInterval:    30 * time.Second,
			MaxReconnectAttempts: 5,
			MessageRetentionDays: 90,
			MediaMaxSizeMB:       16,
			MediaStoragePath:     "", // Missing
			WebhookTimeout:       10 * time.Second,
			WebhookMaxRetries:    3,
			WebhookRetryDelay:    5 * time.Second,
		},
	}

	err := config.Validate()
	if err == nil {
		t.Error("Expected validation error for missing media storage path")
	}
}

func TestValidate_WhatsAppConfig_DisabledSkipsValidation(t *testing.T) {
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
			Cost: 10,
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "json",
		},
		WhatsApp: WhatsAppConfig{
			Enabled:              false, // Disabled
			QRTimeout:            0,      // Invalid but should be ignored
			ReconnectInterval:    0,      // Invalid but should be ignored
			MaxReconnectAttempts: -1,     // Invalid but should be ignored
			MessageRetentionDays: -1,     // Invalid but should be ignored
			MediaMaxSizeMB:       0,      // Invalid but should be ignored
			MediaStoragePath:     "",     // Missing but should be ignored
			WebhookTimeout:       0,      // Invalid but should be ignored
			WebhookMaxRetries:    -1,     // Invalid but should be ignored
			WebhookRetryDelay:    0,      // Invalid but should be ignored
		},
	}

	err := config.Validate()
	if err != nil {
		t.Errorf("Expected no validation error when WhatsApp is disabled, got: %v", err)
	}
}
