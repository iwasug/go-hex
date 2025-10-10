package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Bcrypt   BcryptConfig
	Logging  LoggingConfig
	Webhook  WebhookConfig
	WhatsApp WhatsAppConfig
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port string
	Env  string
}

// DatabaseConfig holds database-related configuration
type DatabaseConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	Name            string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// JWTConfig holds JWT-related configuration
type JWTConfig struct {
	Secret             string
	AccessTokenExpiry  time.Duration
	RefreshTokenExpiry time.Duration
}

// BcryptConfig holds bcrypt-related configuration
type BcryptConfig struct {
	Cost int
}

// LoggingConfig holds logging-related configuration
type LoggingConfig struct {
	Level  string
	Format string
}

// WebhookConfig holds webhook-related configuration
type WebhookConfig struct {
	Timeout       time.Duration
	MaxRetries    int
	RetryDelay    time.Duration
	MaxRetryDelay time.Duration
}

// WhatsAppConfig holds WhatsApp-related configuration
type WhatsAppConfig struct {
	Enabled                bool
	QRTimeout              time.Duration
	ReconnectInterval      time.Duration
	MaxReconnectAttempts   int
	MessageRetentionDays   int
	MediaMaxSizeMB         int
	MediaStoragePath       string
	WebhookTimeout         time.Duration
	WebhookMaxRetries      int
	WebhookRetryDelay      time.Duration
}

// Load loads configuration from environment variables
// It attempts to load .env file in development mode
func Load() (*Config, error) {
	// Try to load .env file (ignore error if file doesn't exist)
	_ = godotenv.Load()

	config := &Config{
		Server: ServerConfig{
			Port: getEnv("PORT", "8080"),
			Env:  getEnv("ENV", "development"),
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnv("DB_PORT", "5432"),
			User:            getEnv("DB_USER", "postgres"),
			Password:        getEnv("DB_PASSWORD", ""),
			Name:            getEnv("DB_NAME", "myapp"),
			SSLMode:         getEnv("DB_SSLMODE", "disable"),
			MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getEnvAsDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
		},
		JWT: JWTConfig{
			Secret:             getEnv("JWT_SECRET", ""),
			AccessTokenExpiry:  getEnvAsDuration("JWT_ACCESS_TOKEN_EXPIRY", 15*time.Minute),
			RefreshTokenExpiry: getEnvAsDuration("JWT_REFRESH_TOKEN_EXPIRY", 168*time.Hour),
		},
		Bcrypt: BcryptConfig{
			Cost: getEnvAsInt("BCRYPT_COST", 10),
		},
		Logging: LoggingConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
		Webhook: WebhookConfig{
			Timeout:       getEnvAsDuration("WEBHOOK_TIMEOUT", 10*time.Second),
			MaxRetries:    getEnvAsInt("WEBHOOK_MAX_RETRIES", 3),
			RetryDelay:    getEnvAsDuration("WEBHOOK_RETRY_DELAY", 5*time.Second),
			MaxRetryDelay: getEnvAsDuration("WEBHOOK_MAX_RETRY_DELAY", 60*time.Second),
		},
		WhatsApp: WhatsAppConfig{
			Enabled:                getEnvAsBool("WHATSAPP_ENABLED", true),
			QRTimeout:              getEnvAsDuration("WHATSAPP_QR_TIMEOUT", 60*time.Second),
			ReconnectInterval:      getEnvAsDuration("WHATSAPP_RECONNECT_INTERVAL", 30*time.Second),
			MaxReconnectAttempts:   getEnvAsInt("WHATSAPP_MAX_RECONNECT_ATTEMPTS", 5),
			MessageRetentionDays:   getEnvAsInt("WHATSAPP_MESSAGE_RETENTION_DAYS", 90),
			MediaMaxSizeMB:         getEnvAsInt("WHATSAPP_MEDIA_MAX_SIZE_MB", 16),
			MediaStoragePath:       getEnv("WHATSAPP_MEDIA_STORAGE_PATH", "./data/whatsapp/media"),
			WebhookTimeout:         getEnvAsDuration("WHATSAPP_WEBHOOK_TIMEOUT", 10*time.Second),
			WebhookMaxRetries:      getEnvAsInt("WHATSAPP_WEBHOOK_MAX_RETRIES", 3),
			WebhookRetryDelay:      getEnvAsDuration("WHATSAPP_WEBHOOK_RETRY_DELAY", 5*time.Second),
		},
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return config, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate server config
	if c.Server.Port == "" {
		return fmt.Errorf("PORT is required")
	}

	// Validate database config
	if c.Database.Host == "" {
		return fmt.Errorf("DB_HOST is required")
	}
	if c.Database.Port == "" {
		return fmt.Errorf("DB_PORT is required")
	}
	if c.Database.User == "" {
		return fmt.Errorf("DB_USER is required")
	}
	if c.Database.Name == "" {
		return fmt.Errorf("DB_NAME is required")
	}

	// Validate JWT config
	if c.JWT.Secret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}
	if len(c.JWT.Secret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters long")
	}
	if c.JWT.AccessTokenExpiry <= 0 {
		return fmt.Errorf("JWT_ACCESS_TOKEN_EXPIRY must be positive")
	}
	if c.JWT.RefreshTokenExpiry <= 0 {
		return fmt.Errorf("JWT_REFRESH_TOKEN_EXPIRY must be positive")
	}

	// Validate bcrypt config
	if c.Bcrypt.Cost < 4 || c.Bcrypt.Cost > 31 {
		return fmt.Errorf("BCRYPT_COST must be between 4 and 31")
	}

	// Validate logging config
	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLogLevels[c.Logging.Level] {
		return fmt.Errorf("LOG_LEVEL must be one of: debug, info, warn, error")
	}

	validLogFormats := map[string]bool{
		"json": true,
		"text": true,
	}
	if !validLogFormats[c.Logging.Format] {
		return fmt.Errorf("LOG_FORMAT must be one of: json, text")
	}

	// Validate WhatsApp config
	if c.WhatsApp.Enabled {
		if c.WhatsApp.QRTimeout <= 0 {
			return fmt.Errorf("WHATSAPP_QR_TIMEOUT must be positive")
		}
		if c.WhatsApp.ReconnectInterval <= 0 {
			return fmt.Errorf("WHATSAPP_RECONNECT_INTERVAL must be positive")
		}
		if c.WhatsApp.MaxReconnectAttempts < 0 {
			return fmt.Errorf("WHATSAPP_MAX_RECONNECT_ATTEMPTS must be non-negative")
		}
		if c.WhatsApp.MessageRetentionDays < 0 {
			return fmt.Errorf("WHATSAPP_MESSAGE_RETENTION_DAYS must be non-negative")
		}
		if c.WhatsApp.MediaMaxSizeMB <= 0 {
			return fmt.Errorf("WHATSAPP_MEDIA_MAX_SIZE_MB must be positive")
		}
		if c.WhatsApp.MediaStoragePath == "" {
			return fmt.Errorf("WHATSAPP_MEDIA_STORAGE_PATH is required when WhatsApp is enabled")
		}
		if c.WhatsApp.WebhookTimeout <= 0 {
			return fmt.Errorf("WHATSAPP_WEBHOOK_TIMEOUT must be positive")
		}
		if c.WhatsApp.WebhookMaxRetries < 0 {
			return fmt.Errorf("WHATSAPP_WEBHOOK_MAX_RETRIES must be non-negative")
		}
		if c.WhatsApp.WebhookRetryDelay <= 0 {
			return fmt.Errorf("WHATSAPP_WEBHOOK_RETRY_DELAY must be positive")
		}
	}

	return nil
}

// GetDSN returns the database connection string
func (c *Config) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.Name,
		c.Database.SSLMode,
	)
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.Server.Env == "development"
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return c.Server.Env == "production"
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getEnvAsInt gets an environment variable as int or returns a default value
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}

// getEnvAsDuration gets an environment variable as duration or returns a default value
func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := time.ParseDuration(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}

// getEnvAsBool gets an environment variable as bool or returns a default value
func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}
