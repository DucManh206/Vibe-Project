package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/captcha-platform/auth/internal/config"
	"github.com/captcha-platform/auth/internal/models"
	"github.com/captcha-platform/auth/internal/repository"
	"github.com/captcha-platform/auth/pkg/jwt"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserNotActive      = errors.New("user account is not active")
	ErrInvalidToken       = errors.New("invalid or expired token")
	ErrMaxAPIKeysReached  = errors.New("maximum number of API keys reached")
)

const (
	MaxAPIKeysPerUser = 10
)

// AuthService handles authentication business logic
type AuthService struct {
	userRepo   *repository.UserRepository
	apiKeyRepo *repository.APIKeyRepository
	jwtConfig  config.JWTConfig
	bcryptCost int
}

// NewAuthService creates a new AuthService
func NewAuthService(
	userRepo *repository.UserRepository,
	apiKeyRepo *repository.APIKeyRepository,
	jwtConfig config.JWTConfig,
	bcryptConfig config.BCryptConfig,
) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		apiKeyRepo: apiKeyRepo,
		jwtConfig:  jwtConfig,
		bcryptCost: bcryptConfig.Cost,
	}
}

// Register registers a new user
func (s *AuthService) Register(ctx context.Context, req *models.RegisterRequest) (*models.User, error) {
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), s.bcryptCost)
	if err != nil {
		return nil, err
	}

	// Create user
	user, err := s.userRepo.Create(ctx, req.Email, string(hashedPassword))
	if err != nil {
		return nil, err
	}

	return user, nil
}

// Login authenticates a user and returns tokens
func (s *AuthService) Login(ctx context.Context, req *models.LoginRequest) (*models.LoginResponse, error) {
	// Find user by email
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	// Check if user is active
	if !user.IsActive {
		return nil, ErrUserNotActive
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Generate tokens
	accessToken, err := jwt.GenerateAccessToken(user.ID, user.Email, user.Role, s.jwtConfig.Secret, s.jwtConfig.ExpiresIn)
	if err != nil {
		return nil, err
	}

	refreshToken, err := jwt.GenerateRefreshToken(user.ID, s.jwtConfig.Secret, s.jwtConfig.RefreshExpiresIn)
	if err != nil {
		return nil, err
	}

	// Update last login time
	_ = s.userRepo.UpdateLastLogin(ctx, user.ID)

	return &models.LoginResponse{
		User:         user.ToResponse(),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.jwtConfig.ExpiresIn.Seconds()),
	}, nil
}

// RefreshToken refreshes an access token using a refresh token
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*models.LoginResponse, error) {
	// Validate refresh token
	claims, err := jwt.ValidateRefreshToken(refreshToken, s.jwtConfig.Secret)
	if err != nil {
		return nil, ErrInvalidToken
	}

	// Get user
	user, err := s.userRepo.FindByID(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}

	// Check if user is active
	if !user.IsActive {
		return nil, ErrUserNotActive
	}

	// Generate new access token
	accessToken, err := jwt.GenerateAccessToken(user.ID, user.Email, user.Role, s.jwtConfig.Secret, s.jwtConfig.ExpiresIn)
	if err != nil {
		return nil, err
	}

	// Generate new refresh token
	newRefreshToken, err := jwt.GenerateRefreshToken(user.ID, s.jwtConfig.Secret, s.jwtConfig.RefreshExpiresIn)
	if err != nil {
		return nil, err
	}

	return &models.LoginResponse{
		User:         user.ToResponse(),
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    int64(s.jwtConfig.ExpiresIn.Seconds()),
	}, nil
}

// GetUserByID gets a user by ID
func (s *AuthService) GetUserByID(ctx context.Context, userID uint64) (*models.User, error) {
	return s.userRepo.FindByID(ctx, userID)
}

// UpdateUser updates a user's information
func (s *AuthService) UpdateUser(ctx context.Context, userID uint64, req *models.UpdateUserRequest) (*models.User, error) {
	if req.Email != "" {
		if err := s.userRepo.UpdateEmail(ctx, userID, req.Email); err != nil {
			return nil, err
		}
	}

	return s.userRepo.FindByID(ctx, userID)
}

// ChangePassword changes a user's password
func (s *AuthService) ChangePassword(ctx context.Context, userID uint64, req *models.ChangePasswordRequest) error {
	// Get user
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.CurrentPassword)); err != nil {
		return ErrInvalidCredentials
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), s.bcryptCost)
	if err != nil {
		return err
	}

	// Update password
	return s.userRepo.UpdatePassword(ctx, userID, string(hashedPassword))
}

// CreateAPIKey creates a new API key for a user
func (s *AuthService) CreateAPIKey(ctx context.Context, userID uint64, req *models.CreateAPIKeyRequest) (*models.APIKeyWithSecret, error) {
	// Check if user has reached max API keys
	count, err := s.apiKeyRepo.CountByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if count >= MaxAPIKeysPerUser {
		return nil, ErrMaxAPIKeysReached
	}

	// Generate random API key
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		return nil, err
	}

	fullKey := "cp_" + hex.EncodeToString(keyBytes) // cp_ prefix for captcha-platform
	keyPrefix := fullKey[:11]                        // First 11 chars including prefix

	// Hash the key for storage
	keyHash := sha256.Sum256([]byte(fullKey))
	keyHashHex := hex.EncodeToString(keyHash[:])

	// Set rate limit
	rateLimit := req.RateLimit
	if rateLimit <= 0 {
		rateLimit = 100 // Default rate limit
	}

	// Set expiration
	var expiresAt *time.Time
	if req.ExpiresIn > 0 {
		exp := time.Now().AddDate(0, 0, req.ExpiresIn)
		expiresAt = &exp
	}

	// Convert scopes to JSON string
	scopesJSON := "[]"
	if len(req.Scopes) > 0 {
		// Simple JSON array construction
		scopesJSON = "["
		for i, scope := range req.Scopes {
			if i > 0 {
				scopesJSON += ","
			}
			scopesJSON += `"` + scope + `"`
		}
		scopesJSON += "]"
	}

	// Create API key
	apiKey, err := s.apiKeyRepo.Create(ctx, userID, req.Name, keyPrefix, keyHashHex, scopesJSON, rateLimit, expiresAt)
	if err != nil {
		return nil, err
	}

	// Return with full key (only time it's shown)
	return &models.APIKeyWithSecret{
		APIKeyResponse: *apiKey.ToResponse(),
		Key:            fullKey,
	}, nil
}

// ListAPIKeys lists all API keys for a user
func (s *AuthService) ListAPIKeys(ctx context.Context, userID uint64) ([]*models.APIKeyResponse, error) {
	apiKeys, err := s.apiKeyRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	responses := make([]*models.APIKeyResponse, len(apiKeys))
	for i, apiKey := range apiKeys {
		responses[i] = apiKey.ToResponse()
	}

	return responses, nil
}

// DeleteAPIKey deletes an API key
func (s *AuthService) DeleteAPIKey(ctx context.Context, userID, keyID uint64) error {
	return s.apiKeyRepo.Delete(ctx, keyID, userID)
}

// ValidateAPIKey validates an API key and returns the associated user
func (s *AuthService) ValidateAPIKey(ctx context.Context, key string) (*models.User, *models.APIKey, error) {
	// Hash the provided key
	keyHash := sha256.Sum256([]byte(key))
	keyHashHex := hex.EncodeToString(keyHash[:])

	// Find API key by hash
	apiKey, err := s.apiKeyRepo.FindByKeyHash(ctx, keyHashHex)
	if err != nil {
		return nil, nil, err
	}

	// Check if expired
	if apiKey.ExpiresAt.Valid && apiKey.ExpiresAt.Time.Before(time.Now()) {
		return nil, nil, ErrInvalidToken
	}

	// Get user
	user, err := s.userRepo.FindByID(ctx, apiKey.UserID)
	if err != nil {
		return nil, nil, err
	}

	// Check if user is active
	if !user.IsActive {
		return nil, nil, ErrUserNotActive
	}

	// Increment usage
	_ = s.apiKeyRepo.IncrementUsage(ctx, apiKey.ID)

	return user, apiKey, nil
}