package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/captcha-platform/gateway/internal/config"
	"github.com/captcha-platform/gateway/internal/handlers"
	"github.com/captcha-platform/gateway/internal/middleware"
	"github.com/captcha-platform/gateway/internal/proxy"
	"github.com/captcha-platform/gateway/pkg/logger"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize logger
	log := logger.NewLogger()
	defer log.Sync()

	log.Info("Starting API Gateway...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration", "error", err)
	}

	// Set Gin mode
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize Redis client for rate limiting
	redisClient, err := middleware.NewRedisClient(cfg.Redis)
	if err != nil {
		log.Warn("Failed to connect to Redis, using in-memory rate limiting", "error", err)
	}

	// Initialize service proxies
	authProxy := proxy.NewServiceProxy(cfg.Services.AuthURL, log)
	captchaProxy := proxy.NewServiceProxy(cfg.Services.CaptchaURL, log)

	// Initialize handlers
	proxyHandler := handlers.NewProxyHandler(authProxy, captchaProxy, log)

	// Setup Gin router
	router := gin.New()

	// Add global middleware
	router.Use(gin.Recovery())
	router.Use(middleware.Logger(log))
	router.Use(middleware.RequestID())
	router.Use(middleware.CORS(cfg.CORS))
	router.Use(middleware.SecurityHeaders())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "gateway",
			"time":    time.Now().UTC().Format(time.RFC3339),
		})
	})

	// API routes
	api := router.Group("/api")
	{
		v1 := api.Group("/v1")
		{
			// Rate limiting for all v1 routes
			if redisClient != nil {
				v1.Use(middleware.RateLimiterRedis(redisClient, cfg.RateLimit.Requests, cfg.RateLimit.Window))
			} else {
				v1.Use(middleware.RateLimiterMemory(cfg.RateLimit.Requests))
			}

			// Auth routes - proxy to auth service
			auth := v1.Group("/auth")
			{
				auth.POST("/register", proxyHandler.ProxyToAuth)
				auth.POST("/login", proxyHandler.ProxyToAuth)
				auth.POST("/refresh", proxyHandler.ProxyToAuth)
				auth.POST("/logout", proxyHandler.ProxyToAuth)

				// Protected auth routes
				authProtected := auth.Group("")
				authProtected.Use(middleware.AuthRequired(cfg.JWT.Secret))
				{
					authProtected.GET("/me", proxyHandler.ProxyToAuth)
					authProtected.PUT("/me", proxyHandler.ProxyToAuth)
					authProtected.PUT("/me/password", proxyHandler.ProxyToAuth)
				}
			}

			// API Keys routes - proxy to auth service (protected)
			apiKeys := v1.Group("/api-keys")
			apiKeys.Use(middleware.AuthRequired(cfg.JWT.Secret))
			{
				apiKeys.GET("", proxyHandler.ProxyToAuth)
				apiKeys.POST("", proxyHandler.ProxyToAuth)
				apiKeys.DELETE("/:id", proxyHandler.ProxyToAuth)
			}

			// Captcha routes - proxy to captcha service
			captcha := v1.Group("/captcha")
			{
				// Public route for solving (with API key auth)
				captcha.POST("/solve", middleware.APIKeyOrJWTAuth(cfg.JWT.Secret), proxyHandler.ProxyToCaptcha)
				captcha.POST("/solve/batch", middleware.APIKeyOrJWTAuth(cfg.JWT.Secret), proxyHandler.ProxyToCaptcha)

				// Protected routes
				captchaProtected := captcha.Group("")
				captchaProtected.Use(middleware.AuthRequired(cfg.JWT.Secret))
				{
					captchaProtected.GET("/models", proxyHandler.ProxyToCaptcha)
					captchaProtected.POST("/models/upload", proxyHandler.ProxyToCaptcha)
					captchaProtected.POST("/train", proxyHandler.ProxyToCaptcha)
					captchaProtected.GET("/train/:job_id", proxyHandler.ProxyToCaptcha)
					captchaProtected.GET("/logs", proxyHandler.ProxyToCaptcha)
					captchaProtected.GET("/stats", proxyHandler.ProxyToCaptcha)
				}
			}
		}
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Info("API Gateway started", "port", cfg.Port, "environment", cfg.Environment)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server", "error", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down API Gateway...")

	// Create a deadline to wait for
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Close Redis connection
	if redisClient != nil {
		redisClient.Close()
	}

	// Attempt graceful shutdown
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown", "error", err)
	}

	log.Info("API Gateway stopped")
}