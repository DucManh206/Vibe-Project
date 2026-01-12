package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/captcha-platform/auth/internal/models"
)

var (
	ErrAPIKeyNotFound = errors.New("API key not found")
)

// APIKeyRepository handles database operations for API keys
type APIKeyRepository struct {
	db *sql.DB
}

// NewAPIKeyRepository creates a new APIKeyRepository
func NewAPIKeyRepository(db *sql.DB) *APIKeyRepository {
	return &APIKeyRepository{db: db}
}

// Create creates a new API key
func (r *APIKeyRepository) Create(
	ctx context.Context,
	userID uint64,
	name string,
	keyPrefix string,
	keyHash string,
	scopesJSON string,
	rateLimit int,
	expiresAt *time.Time,
) (*models.APIKey, error) {
	query := `
		INSERT INTO api_keys (user_id, name, key_prefix, key_hash, scopes, rate_limit, is_active, expires_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, TRUE, ?, NOW(), NOW())
	`

	result, err := r.db.ExecContext(ctx, query, userID, name, keyPrefix, keyHash, scopesJSON, rateLimit, expiresAt)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return r.FindByID(ctx, uint64(id))
}

// FindByID finds an API key by ID
func (r *APIKeyRepository) FindByID(ctx context.Context, id uint64) (*models.APIKey, error) {
	query := `
		SELECT id, user_id, name, key_prefix, key_hash, scopes, rate_limit, 
		       total_requests, last_used_at, is_active, expires_at, created_at, updated_at
		FROM api_keys
		WHERE id = ?
	`

	apiKey := &models.APIKey{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&apiKey.ID,
		&apiKey.UserID,
		&apiKey.Name,
		&apiKey.KeyPrefix,
		&apiKey.KeyHash,
		&apiKey.Scopes,
		&apiKey.RateLimit,
		&apiKey.TotalRequests,
		&apiKey.LastUsedAt,
		&apiKey.IsActive,
		&apiKey.ExpiresAt,
		&apiKey.CreatedAt,
		&apiKey.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrAPIKeyNotFound
		}
		return nil, err
	}

	return apiKey, nil
}

// FindByUserID finds all API keys for a user
func (r *APIKeyRepository) FindByUserID(ctx context.Context, userID uint64) ([]*models.APIKey, error) {
	query := `
		SELECT id, user_id, name, key_prefix, key_hash, scopes, rate_limit, 
		       total_requests, last_used_at, is_active, expires_at, created_at, updated_at
		FROM api_keys
		WHERE user_id = ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var apiKeys []*models.APIKey
	for rows.Next() {
		apiKey := &models.APIKey{}
		err := rows.Scan(
			&apiKey.ID,
			&apiKey.UserID,
			&apiKey.Name,
			&apiKey.KeyPrefix,
			&apiKey.KeyHash,
			&apiKey.Scopes,
			&apiKey.RateLimit,
			&apiKey.TotalRequests,
			&apiKey.LastUsedAt,
			&apiKey.IsActive,
			&apiKey.ExpiresAt,
			&apiKey.CreatedAt,
			&apiKey.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		apiKeys = append(apiKeys, apiKey)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return apiKeys, nil
}

// FindByKeyHash finds an API key by its hash
func (r *APIKeyRepository) FindByKeyHash(ctx context.Context, keyHash string) (*models.APIKey, error) {
	query := `
		SELECT id, user_id, name, key_prefix, key_hash, scopes, rate_limit, 
		       total_requests, last_used_at, is_active, expires_at, created_at, updated_at
		FROM api_keys
		WHERE key_hash = ? AND is_active = TRUE
	`

	apiKey := &models.APIKey{}
	err := r.db.QueryRowContext(ctx, query, keyHash).Scan(
		&apiKey.ID,
		&apiKey.UserID,
		&apiKey.Name,
		&apiKey.KeyPrefix,
		&apiKey.KeyHash,
		&apiKey.Scopes,
		&apiKey.RateLimit,
		&apiKey.TotalRequests,
		&apiKey.LastUsedAt,
		&apiKey.IsActive,
		&apiKey.ExpiresAt,
		&apiKey.CreatedAt,
		&apiKey.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrAPIKeyNotFound
		}
		return nil, err
	}

	return apiKey, nil
}

// Delete deletes an API key (sets is_active to false)
func (r *APIKeyRepository) Delete(ctx context.Context, id uint64, userID uint64) error {
	query := `UPDATE api_keys SET is_active = FALSE, updated_at = NOW() WHERE id = ? AND user_id = ?`

	result, err := r.db.ExecContext(ctx, query, id, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrAPIKeyNotFound
	}

	return nil
}

// CountByUserID counts API keys for a user
func (r *APIKeyRepository) CountByUserID(ctx context.Context, userID uint64) (int, error) {
	query := `SELECT COUNT(*) FROM api_keys WHERE user_id = ? AND is_active = TRUE`

	var count int
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// IncrementUsage increments the usage counter and updates last used time
func (r *APIKeyRepository) IncrementUsage(ctx context.Context, id uint64) error {
	query := `
		UPDATE api_keys 
		SET total_requests = total_requests + 1, last_used_at = NOW(), updated_at = NOW() 
		WHERE id = ?
	`

	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// UpdateRateLimit updates the rate limit for an API key
func (r *APIKeyRepository) UpdateRateLimit(ctx context.Context, id uint64, rateLimit int) error {
	query := `UPDATE api_keys SET rate_limit = ?, updated_at = NOW() WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query, rateLimit, id)
	return err
}