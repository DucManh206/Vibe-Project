package models

import (
	"database/sql"
	"time"
)

// APIKey represents an API key in the system
type APIKey struct {
	ID            uint64         `json:"id"`
	UserID        uint64         `json:"user_id"`
	Name          string         `json:"name"`
	KeyPrefix     string         `json:"key_prefix"`
	KeyHash       string         `json:"-"` // Never expose key hash
	Scopes        sql.NullString `json:"scopes,omitempty"`
	RateLimit     int            `json:"rate_limit"`
	TotalRequests uint64         `json:"total_requests"`
	LastUsedAt    sql.NullTime   `json:"last_used_at,omitempty"`
	IsActive      bool           `json:"is_active"`
	ExpiresAt     sql.NullTime   `json:"expires_at,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

// APIKeyResponse is the public representation of an API key
type APIKeyResponse struct {
	ID            uint64   `json:"id"`
	Name          string   `json:"name"`
	KeyPrefix     string   `json:"key_prefix"`
	Scopes        []string `json:"scopes,omitempty"`
	RateLimit     int      `json:"rate_limit"`
	TotalRequests uint64   `json:"total_requests"`
	LastUsedAt    *string  `json:"last_used_at,omitempty"`
	IsActive      bool     `json:"is_active"`
	ExpiresAt     *string  `json:"expires_at,omitempty"`
	CreatedAt     string   `json:"created_at"`
}

// APIKeyWithSecret is returned only when creating a new API key
type APIKeyWithSecret struct {
	APIKeyResponse
	Key string `json:"key"` // Full key, only shown once
}

// CreateAPIKeyRequest represents a request to create an API key
type CreateAPIKeyRequest struct {
	Name      string   `json:"name" binding:"required,min=1,max=100"`
	Scopes    []string `json:"scopes,omitempty"`
	RateLimit int      `json:"rate_limit,omitempty"`
	ExpiresIn int      `json:"expires_in,omitempty"` // Duration in days
}

// ToResponse converts an APIKey to APIKeyResponse
func (a *APIKey) ToResponse() *APIKeyResponse {
	response := &APIKeyResponse{
		ID:            a.ID,
		Name:          a.Name,
		KeyPrefix:     a.KeyPrefix,
		RateLimit:     a.RateLimit,
		TotalRequests: a.TotalRequests,
		IsActive:      a.IsActive,
		CreatedAt:     a.CreatedAt.Format(time.RFC3339),
	}

	if a.LastUsedAt.Valid {
		formatted := a.LastUsedAt.Time.Format(time.RFC3339)
		response.LastUsedAt = &formatted
	}

	if a.ExpiresAt.Valid {
		formatted := a.ExpiresAt.Time.Format(time.RFC3339)
		response.ExpiresAt = &formatted
	}

	// Parse scopes from JSON string
	if a.Scopes.Valid && a.Scopes.String != "" {
		// Simple parsing - in production, use proper JSON unmarshal
		response.Scopes = []string{}
	}

	return response
}