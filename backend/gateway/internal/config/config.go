package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all configuration for the gateway
type Config struct {
	Port        int
	Environment string
	JWT         JWTConfig
	CORS        CORSConfig
	RateLimit   RateLimitConfig
	Redis       RedisConfig
	Services    ServicesConfig
	LogLevel    string
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret           string
	ExpiresIn        time.Duration
	RefreshExpiresIn time.Duration
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Requests int
	Window   time.Duration
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// Addr returns the Redis address
func (c RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// ServicesConfig holds backend service URLs
type ServicesConfig struct {
	AuthURL    string
	CaptchaURL string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		Port:        getEnvInt("GATEWAY_PORT", 8080),
		Environment: getEnv("GATEWAY_ENV", "development"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),

		JWT: JWTConfig{
			Secret:           getEnv("JWT_SECRET", "your-secret-key-min-32-characters-long"),
			ExpiresIn:        getEnvDuration("JWT_EXPIRES_IN", 24*time.Hour),
			RefreshExpiresIn: getEnvDuration("JWT_REFRESH_EXPIRES_IN", 7*24*time.Hour),
		},

		CORS: CORSConfig{
			AllowedOrigins: getEnvSlice("CORS_ORIGINS", []string{"http://localhost:3000"}),
			AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
			AllowedHeaders: []string{
				"Accept",
				"Authorization",
				"Content-Type",
				"X-API-Key",
				"X-Request-ID",
				"X-Requested-With",
			},
		},

		RateLimit: RateLimitConfig{
			Requests: getEnvInt("RATE_LIMIT_REQUESTS", 100),
			Window:   time.Duration(getEnvInt("RATE_LIMIT_WINDOW_SECONDS", 60)) * time.Second,
		},

		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnvInt("REDIS_PORT", 6379),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvInt("REDIS_DB", 0),
		},

		Services: ServicesConfig{
			AuthURL:    getEnv("AUTH_SERVICE_URL", "http://localhost:8081"),
			CaptchaURL: getEnv("CAPTCHA_SERVICE_URL", "http://localhost:8082"),
		},
	}

	// Validate required configuration
	if cfg.JWT.Secret == "" || len(cfg.JWT.Secret) < 32 {
		return nil, fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}

	return cfg, nil
}

// Helper functions

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}

func getEnvSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}