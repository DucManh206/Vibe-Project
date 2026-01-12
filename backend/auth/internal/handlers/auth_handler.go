package handlers

import (
	"net/http"
	"strconv"

	"github.com/captcha-platform/auth/internal/models"
	"github.com/captcha-platform/auth/internal/repository"
	"github.com/captcha-platform/auth/internal/services"
	"github.com/captcha-platform/auth/pkg/logger"

	"github.com/gin-gonic/gin"
)

// AuthHandler handles authentication HTTP requests
type AuthHandler struct {
	authService *services.AuthService
	logger      *logger.Logger
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(authService *services.AuthService, logger *logger.Logger) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		logger:      logger,
	}
}

// Register handles user registration
// @Summary Register a new user
// @Description Register a new user with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.RegisterRequest true "Registration request"
// @Success 201 {object} models.UserResponse
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Debug("Invalid registration request", "error", err)
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "validation_error",
			Message: "Invalid request body",
			Details: err.Error(),
		})
		return
	}

	user, err := h.authService.Register(c.Request.Context(), &req)
	if err != nil {
		if err == repository.ErrUserAlreadyExists {
			c.JSON(http.StatusConflict, ErrorResponse{
				Error:   "user_exists",
				Message: "A user with this email already exists",
			})
			return
		}
		h.logger.Error("Failed to register user", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to create user",
		})
		return
	}

	h.logger.Info("User registered", "user_id", user.ID, "email", user.Email)
	c.JSON(http.StatusCreated, user.ToResponse())
}

// Login handles user login
// @Summary Login user
// @Description Authenticate user and return JWT tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.LoginRequest true "Login request"
// @Success 200 {object} models.LoginResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "validation_error",
			Message: "Invalid request body",
			Details: err.Error(),
		})
		return
	}

	response, err := h.authService.Login(c.Request.Context(), &req)
	if err != nil {
		switch err {
		case services.ErrInvalidCredentials:
			c.JSON(http.StatusUnauthorized, ErrorResponse{
				Error:   "invalid_credentials",
				Message: "Invalid email or password",
			})
		case services.ErrUserNotActive:
			c.JSON(http.StatusForbidden, ErrorResponse{
				Error:   "user_inactive",
				Message: "User account is not active",
			})
		default:
			h.logger.Error("Failed to login user", "error", err)
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error:   "internal_error",
				Message: "Failed to authenticate",
			})
		}
		return
	}

	h.logger.Info("User logged in", "user_id", response.User.ID)
	c.JSON(http.StatusOK, response)
}

// RefreshToken handles token refresh
// @Summary Refresh access token
// @Description Get a new access token using a refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.RefreshTokenRequest true "Refresh token request"
// @Success 200 {object} models.LoginResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req models.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "validation_error",
			Message: "Invalid request body",
			Details: err.Error(),
		})
		return
	}

	response, err := h.authService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "invalid_token",
			Message: "Invalid or expired refresh token",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// Logout handles user logout
// @Summary Logout user
// @Description Invalidate user tokens (client-side token removal)
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} SuccessResponse
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// In a stateless JWT setup, logout is handled client-side
	// For a more secure implementation, you could:
	// 1. Add the token to a blacklist in Redis
	// 2. Use short-lived access tokens with refresh token rotation

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Successfully logged out",
	})
}

// GetCurrentUser returns the current authenticated user
// @Summary Get current user
// @Description Get the currently authenticated user's information
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.UserResponse
// @Failure 401 {object} ErrorResponse
// @Router /auth/me [get]
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
		return
	}

	user, err := h.authService.GetUserByID(c.Request.Context(), userID.(uint64))
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "not_found",
			Message: "User not found",
		})
		return
	}

	c.JSON(http.StatusOK, user.ToResponse())
}

// UpdateCurrentUser updates the current user's information
// @Summary Update current user
// @Description Update the currently authenticated user's information
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.UpdateUserRequest true "Update request"
// @Success 200 {object} models.UserResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /auth/me [put]
func (h *AuthHandler) UpdateCurrentUser(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
		return
	}

	var req models.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "validation_error",
			Message: "Invalid request body",
			Details: err.Error(),
		})
		return
	}

	user, err := h.authService.UpdateUser(c.Request.Context(), userID.(uint64), &req)
	if err != nil {
		if err == repository.ErrUserAlreadyExists {
			c.JSON(http.StatusConflict, ErrorResponse{
				Error:   "email_exists",
				Message: "A user with this email already exists",
			})
			return
		}
		h.logger.Error("Failed to update user", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to update user",
		})
		return
	}

	c.JSON(http.StatusOK, user.ToResponse())
}

// ChangePassword changes the current user's password
// @Summary Change password
// @Description Change the currently authenticated user's password
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.ChangePasswordRequest true "Password change request"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /auth/me/password [put]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
		return
	}

	var req models.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "validation_error",
			Message: "Invalid request body",
			Details: err.Error(),
		})
		return
	}

	err := h.authService.ChangePassword(c.Request.Context(), userID.(uint64), &req)
	if err != nil {
		if err == services.ErrInvalidCredentials {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "invalid_password",
				Message: "Current password is incorrect",
			})
			return
		}
		h.logger.Error("Failed to change password", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to change password",
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Password changed successfully",
	})
}

// ListAPIKeys lists all API keys for the current user
// @Summary List API keys
// @Description Get all API keys for the authenticated user
// @Tags api-keys
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.APIKeyResponse
// @Failure 401 {object} ErrorResponse
// @Router /api-keys [get]
func (h *AuthHandler) ListAPIKeys(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
		return
	}

	apiKeys, err := h.authService.ListAPIKeys(c.Request.Context(), userID.(uint64))
	if err != nil {
		h.logger.Error("Failed to list API keys", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to list API keys",
		})
		return
	}

	c.JSON(http.StatusOK, apiKeys)
}

// CreateAPIKey creates a new API key for the current user
// @Summary Create API key
// @Description Create a new API key for the authenticated user
// @Tags api-keys
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.CreateAPIKeyRequest true "API key creation request"
// @Success 201 {object} models.APIKeyWithSecret
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api-keys [post]
func (h *AuthHandler) CreateAPIKey(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
		return
	}

	var req models.CreateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "validation_error",
			Message: "Invalid request body",
			Details: err.Error(),
		})
		return
	}

	apiKey, err := h.authService.CreateAPIKey(c.Request.Context(), userID.(uint64), &req)
	if err != nil {
		if err == services.ErrMaxAPIKeysReached {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "limit_exceeded",
				Message: "Maximum number of API keys reached",
			})
			return
		}
		h.logger.Error("Failed to create API key", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to create API key",
		})
		return
	}

	h.logger.Info("API key created", "user_id", userID, "key_id", apiKey.ID)
	c.JSON(http.StatusCreated, apiKey)
}

// DeleteAPIKey deletes an API key
// @Summary Delete API key
// @Description Delete an API key by ID
// @Tags api-keys
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "API key ID"
// @Success 200 {object} SuccessResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api-keys/{id} [delete]
func (h *AuthHandler) DeleteAPIKey(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
		return
	}

	keyIDStr := c.Param("id")
	keyID, err := strconv.ParseUint(keyIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid API key ID",
		})
		return
	}

	err = h.authService.DeleteAPIKey(c.Request.Context(), userID.(uint64), keyID)
	if err != nil {
		if err == repository.ErrAPIKeyNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "not_found",
				Message: "API key not found",
			})
			return
		}
		h.logger.Error("Failed to delete API key", "error", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to delete API key",
		})
		return
	}

	h.logger.Info("API key deleted", "user_id", userID, "key_id", keyID)
	c.JSON(http.StatusOK, SuccessResponse{
		Message: "API key deleted successfully",
	})
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// SuccessResponse represents a success response
type SuccessResponse struct {
	Message string `json:"message"`
}