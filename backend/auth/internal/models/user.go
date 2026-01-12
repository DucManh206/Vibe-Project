package models

import (
	"database/sql"
	"time"
)

// User represents a user in the system
type User struct {
	ID              uint64       `json:"id"`
	Email           string       `json:"email"`
	PasswordHash    string       `json:"-"` // Never expose password hash
	Role            string       `json:"role"`
	IsActive        bool         `json:"is_active"`
	EmailVerifiedAt sql.NullTime `json:"email_verified_at,omitempty"`
	LastLoginAt     sql.NullTime `json:"last_login_at,omitempty"`
	CreatedAt       time.Time    `json:"created_at"`
	UpdatedAt       time.Time    `json:"updated_at"`
}

// UserResponse is the public representation of a user
type UserResponse struct {
	ID              uint64  `json:"id"`
	Email           string  `json:"email"`
	Role            string  `json:"role"`
	IsActive        bool    `json:"is_active"`
	EmailVerifiedAt *string `json:"email_verified_at,omitempty"`
	LastLoginAt     *string `json:"last_login_at,omitempty"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
}

// ToResponse converts a User to UserResponse
func (u *User) ToResponse() *UserResponse {
	response := &UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		Role:      u.Role,
		IsActive:  u.IsActive,
		CreatedAt: u.CreatedAt.Format(time.RFC3339),
		UpdatedAt: u.UpdatedAt.Format(time.RFC3339),
	}

	if u.EmailVerifiedAt.Valid {
		formatted := u.EmailVerifiedAt.Time.Format(time.RFC3339)
		response.EmailVerifiedAt = &formatted
	}

	if u.LastLoginAt.Valid {
		formatted := u.LastLoginAt.Time.Format(time.RFC3339)
		response.LastLoginAt = &formatted
	}

	return response
}

// RegisterRequest represents a registration request
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email,max=255"`
	Password string `json:"password" binding:"required,min=8,max=72"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	User         *UserResponse `json:"user"`
	AccessToken  string        `json:"access_token"`
	RefreshToken string        `json:"refresh_token"`
	ExpiresIn    int64         `json:"expires_in"`
}

// RefreshTokenRequest represents a refresh token request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// UpdateUserRequest represents an update user request
type UpdateUserRequest struct {
	Email string `json:"email" binding:"omitempty,email,max=255"`
}

// ChangePasswordRequest represents a change password request
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8,max=72"`
}