package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the auth service
type Config struct {
	Port        int
	Environment string
	Database    DatabaseConfig
	JWT         JWTConfig
	BCrypt      BCryptConfig
	CORS        CORSConfig
	Log         LogConfig
}

// DatabaseConfig holds database connection settings
type DatabaseConfig struct {
	Host     string
	Port     int
	Name     string
	User     string
	Password string
	MaxConns int
	Timeout  time.Duration
}

// JWTConfig holds JWT settings
type JWTConfig struct {
	Secret           string
	ExpiresIn        time.Duration
	RefreshExpiresIn time.Duration
	Issuer           string
}

// BCryptConfig holds password hashing settings
type BCryptConfig struct {
	Cost int
}

// CORSConfig holds CORS settings
type CORSConfig struct {
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
}

// LogConfig holds logging settings
type LogConfig struct {
	Level  string
	Format string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if exists (for development)
	_ = godotenv.Load()

	cfg := &Config{
		Port:        getEnvInt("AUTH_SERVICE_PORT", 8081),
		Environment: getEnvString("GATEWAY_ENV", "development"),
		Database: DatabaseConfig{
			Host:     getEnvString("DB_HOST", "localhost"),
			Port:     getEnvInt("DB_PORT", 3306),
			Name:     getEnvString("DB_NAME", "captcha_platform"),
			User:     getEnvString("DB_USER", "captcha_user"),
			Password: getEnvString("DB_PASSWORD", ""),
			MaxConns: getEnvInt("DB_MAX_CONNS", 25),
			Timeout:  time.Duration(getEnvInt("DB_TIMEOUT_SECONDS", 30)) * time.Second,
		},
		JWT: JWTConfig{
			Secret:           getEnvString("JWT_SECRET", ""),
			ExpiresIn:        parseDuration(getEnvString("JWT_EXPIRES_IN", "24h")),
			RefreshExpiresIn: parseDuration(getEnvString("JWT_REFRESH_EXPIRES_IN", "7d")),
			Issuer:           getEnvString("JWT_ISSUER", "captcha-platform"),
		},
		BCrypt: BCryptConfig{
			Cost: getEnvInt("BCRYPT_COST", 12),
		},
		CORS: CORSConfig{
			AllowedOrigins: strings.Split(getEnvString("CORS_ORIGINS", "http://localhost:3000"), ","),
			AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID"},
		},
		Log: LogConfig{
			Level:  getEnvString("LOG_LEVEL", "debug"),
			Format: getEnvString("LOG_FORMAT", "json"),
		},
	}

	// Validate required fields
	if cfg.JWT.Secret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	if len(cfg.JWT.Secret) < 32 {
		return nil, fmt.Errorf("JWT_SECRET must be at least 32 characters long")
	}

	if cfg.Database.Password == "" {
		return nil, fmt.Errorf("DB_PASSWORD is required")
	}

	return cfg, nil
}

// DSN returns the database connection string
func (d *DatabaseConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=UTC&timeout=%s",
		d.User,
		d.Password,
		d.Host,
		d.Port,
		d.Name,
		d.Timeout.String(),
	)
}

// Helper functions

func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func parseDuration(value string) time.Duration {
	// Handle special cases like "7d" for 7 days
	if strings.HasSuffix(value, "d") {
		days, err := strconv.Atoi(strings.TrimSuffix(value, "d"))
		if err == nil {
			return time.Duration(days) * 24 * time.Hour
		}
	}

	duration, err := time.ParseDuration(value)
	if err != nil {
		return 24 * time.Hour // Default to 24 hours
	}
	return duration
}