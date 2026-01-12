package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/captcha-platform/auth/internal/config"
	"github.com/captcha-platform/auth/internal/database"
	"github.com/captcha-platform/auth/internal/handlers"
	"github.com/captcha-platform/auth/internal/middleware"
	"github.com/captcha-platform/auth/internal/repository"
	"github.com/captcha-platform/auth/internal/services"
	"github.com/captcha-platform/auth/pkg/logger"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize logger
	log := logger.NewLogger()
	defer log.Sync()

	log.Info("Starting Auth Service...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration", "error", err)
	}

	// Set Gin mode
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize database connection
	db, err := database.NewConnection(cfg.Database)
	if err != nil {
		log.Fatal("Failed to connect to database", "error", err)
	}
	defer db.Close()

	log.Info("Connected to database successfully")

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	apiKeyRepo := repository.NewAPIKeyRepository(db)

	// Initialize services
	authService := services.NewAuthService(userRepo, apiKeyRepo, cfg.JWT, cfg.BCrypt)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService, log)

	// Setup Gin router
	router := gin.New()

	// Add middleware
	router.Use(gin.Recovery())
	router.Use(middleware.Logger(log))
	router.Use(middleware.CORS(cfg.CORS))
	router.Use(middleware.RequestID())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "auth",
			"time":    time.Now().UTC().Format(time.RFC3339),
		})
	})

	// API routes
	v1 := router.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
			auth.POST("/logout", authHandler.Logout)

			// Protected routes
			protected := auth.Group("")
			protected.Use(middleware.AuthRequired(cfg.JWT.Secret))
			{
				protected.GET("/me", authHandler.GetCurrentUser)
				protected.PUT("/me", authHandler.UpdateCurrentUser)
				protected.PUT("/me/password", authHandler.ChangePassword)
			}
		}

		// API Keys management (protected)
		apiKeys := v1.Group("/api-keys")
		apiKeys.Use(middleware.AuthRequired(cfg.JWT.Secret))
		{
			apiKeys.GET("", authHandler.ListAPIKeys)
			apiKeys.POST("", authHandler.CreateAPIKey)
			apiKeys.DELETE("/:id", authHandler.DeleteAPIKey)
		}
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Info("Auth Service started", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server", "error", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down Auth Service...")

	// Create a deadline to wait for
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown", "error", err)
	}

	log.Info("Auth Service stopped")
}