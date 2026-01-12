package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/captcha-platform/gateway/internal/config"
	"github.com/captcha-platform/gateway/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// Logger returns a middleware that logs requests
func Logger(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Log details
		latency := time.Since(start)
		status := c.Writer.Status()

		log.Info("HTTP Request",
			"method", c.Request.Method,
			"path", path,
			"query", query,
			"status", status,
			"latency", latency.String(),
			"client_ip", c.ClientIP(),
			"request_id", c.GetString("request_id"),
		)
	}
}

// RequestID adds a unique request ID to each request
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)

		c.Next()
	}
}

// CORS returns a middleware that handles CORS
func CORS(cfg config.CORSConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		// Check if origin is allowed
		allowed := false
		for _, allowedOrigin := range cfg.AllowedOrigins {
			if allowedOrigin == "*" || allowedOrigin == origin {
				allowed = true
				break
			}
		}

		if allowed && origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
		}

		c.Header("Access-Control-Allow-Methods", strings.Join(cfg.AllowedMethods, ", "))
		c.Header("Access-Control-Allow-Headers", strings.Join(cfg.AllowedHeaders, ", "))
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400")

		// Handle preflight requests
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// SecurityHeaders adds security headers to responses
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Only add HSTS in production
		if gin.Mode() == gin.ReleaseMode {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		c.Next()
	}
}

// AuthRequired returns a middleware that requires JWT authentication
func AuthRequired(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Authorization header is required",
			})
			return
		}

		// Check Bearer prefix
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Invalid authorization header format",
			})
			return
		}

		token := parts[1]

		// Validate token using internal JWT validation
		claims, err := validateJWT(token, secret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Invalid or expired token",
			})
			return
		}

		// Set user info in context
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Role)

		// Forward token to downstream services
		c.Request.Header.Set("X-User-ID", uintToString(claims.UserID))
		c.Request.Header.Set("X-User-Email", claims.Email)
		c.Request.Header.Set("X-User-Role", claims.Role)

		c.Next()
	}
}

// APIKeyOrJWTAuth allows authentication via API key or JWT
func APIKeyOrJWTAuth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check for API key first
		apiKey := c.GetHeader("X-API-Key")
		if apiKey != "" {
			// Hash the API key to validate
			keyHash := sha256.Sum256([]byte(apiKey))
			c.Set("api_key_hash", hex.EncodeToString(keyHash[:]))
			c.Set("auth_type", "api_key")

			// Forward the API key hash to downstream service for validation
			c.Request.Header.Set("X-API-Key-Hash", hex.EncodeToString(keyHash[:]))

			c.Next()
			return
		}

		// Check for JWT token
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
				token := parts[1]
				claims, err := validateJWT(token, jwtSecret)
				if err == nil {
					c.Set("user_id", claims.UserID)
					c.Set("user_email", claims.Email)
					c.Set("user_role", claims.Role)
					c.Set("auth_type", "jwt")

					c.Request.Header.Set("X-User-ID", uintToString(claims.UserID))
					c.Request.Header.Set("X-User-Email", claims.Email)
					c.Request.Header.Set("X-User-Role", claims.Role)

					c.Next()
					return
				}
			}
		}

		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "API key or valid JWT token required",
		})
	}
}

// NewRedisClient creates a new Redis client
func NewRedisClient(cfg config.RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr(),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return client, nil
}

// RateLimiterRedis returns a Redis-based rate limiter middleware
func RateLimiterRedis(client *redis.Client, maxRequests int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		key := "rate_limit:" + c.ClientIP()

		// Increment and get current count
		pipe := client.Pipeline()
		incr := pipe.Incr(ctx, key)
		pipe.Expire(ctx, key, window)
		_, err := pipe.Exec(ctx)

		if err != nil {
			// If Redis fails, allow the request
			c.Next()
			return
		}

		count := incr.Val()
		remaining := maxRequests - int(count)
		if remaining < 0 {
			remaining = 0
		}

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", intToString(maxRequests))
		c.Header("X-RateLimit-Remaining", intToString(remaining))
		c.Header("X-RateLimit-Reset", intToString(int(window.Seconds())))

		if count > int64(maxRequests) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":   "rate_limit_exceeded",
				"message": "Too many requests, please try again later",
			})
			return
		}

		c.Next()
	}
}

// RateLimiterMemory returns an in-memory rate limiter middleware
func RateLimiterMemory(maxRequests int) gin.HandlerFunc {
	type client struct {
		count    int
		lastSeen time.Time
	}

	var mu sync.Mutex
	clients := make(map[string]*client)
	window := time.Minute

	// Cleanup goroutine
	go func() {
		for {
			time.Sleep(time.Minute)
			mu.Lock()
			for ip, cl := range clients {
				if time.Since(cl.lastSeen) > window*2 {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return func(c *gin.Context) {
		ip := c.ClientIP()
		now := time.Now()

		mu.Lock()
		if cl, exists := clients[ip]; exists {
			if now.Sub(cl.lastSeen) > window {
				cl.count = 1
				cl.lastSeen = now
			} else {
				cl.count++
			}

			remaining := maxRequests - cl.count
			if remaining < 0 {
				remaining = 0
			}

			c.Header("X-RateLimit-Limit", intToString(maxRequests))
			c.Header("X-RateLimit-Remaining", intToString(remaining))

			if cl.count > maxRequests {
				mu.Unlock()
				c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
					"error":   "rate_limit_exceeded",
					"message": "Too many requests, please try again later",
				})
				return
			}
		} else {
			clients[ip] = &client{count: 1, lastSeen: now}
			c.Header("X-RateLimit-Limit", intToString(maxRequests))
			c.Header("X-RateLimit-Remaining", intToString(maxRequests-1))
		}
		mu.Unlock()

		c.Next()
	}
}

// JWT Claims structure
type JWTClaims struct {
	UserID uint64
	Email  string
	Role   string
}

// Simple JWT validation (in production, use proper JWT library)
func validateJWT(tokenString, secret string) (*JWTClaims, error) {
	// This is a simplified version - the actual implementation would use
	// the golang-jwt library. For now, we'll forward the token to auth service
	// which will do the actual validation.

	// Parse token manually (simplified for gateway)
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidToken
	}

	// In a real implementation, decode and verify the signature here
	// For now, just extract basic claims and forward to auth service

	// This is a placeholder - the actual validation happens in auth service
	// Gateway just checks format and forwards
	return &JWTClaims{
		UserID: 0,
		Email:  "",
		Role:   "",
	}, nil
}

// Error types
type TokenError struct {
	Message string
}

func (e *TokenError) Error() string {
	return e.Message
}

var ErrInvalidToken = &TokenError{"invalid token"}

// Helper functions
func uintToString(n uint64) string {
	return strconv.FormatUint(n, 10)
}

func intToString(n int) string {
	return strconv.Itoa(n)
}