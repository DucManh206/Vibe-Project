package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/captcha-platform/auth/internal/config"
	"github.com/captcha-platform/auth/pkg/jwt"
	"github.com/captcha-platform/auth/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

// AuthRequired returns a middleware that requires JWT authentication
func AuthRequired(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// First check for X-User-ID header (forwarded from gateway)
		userIDHeader := c.GetHeader("X-User-ID")
		userEmailHeader := c.GetHeader("X-User-Email")
		userRoleHeader := c.GetHeader("X-User-Role")

		if userIDHeader != "" && userEmailHeader != "" {
			// Request already authenticated by gateway
			var userID uint64
			if _, err := parseUint64(userIDHeader, &userID); err == nil {
				c.Set("user_id", userID)
				c.Set("user_email", userEmailHeader)
				c.Set("user_role", userRoleHeader)
				c.Next()
				return
			}
		}

		// Fallback to JWT validation
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

		// Validate token
		claims, err := jwt.ValidateAccessToken(token, secret)
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

		c.Next()
	}
}

// AdminRequired returns a middleware that requires admin role
func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("user_role")
		if !exists || role != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":   "forbidden",
				"message": "Admin access required",
			})
			return
		}
		c.Next()
	}
}

// parseUint64 parses a string to uint64
func parseUint64(s string, result *uint64) (bool, error) {
	var n uint64
	for _, c := range s {
		if c < '0' || c > '9' {
			return false, nil
		}
		n = n*10 + uint64(c-'0')
	}
	*result = n
	return true, nil
}