package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/captcha-platform/auth/internal/models"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user with this email already exists")
)

// UserRepository handles database operations for users
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new UserRepository
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new user
func (r *UserRepository) Create(ctx context.Context, email, passwordHash string) (*models.User, error) {
	query := `
		INSERT INTO users (email, password_hash, role, is_active, created_at, updated_at)
		VALUES (?, ?, 'user', TRUE, NOW(), NOW())
	`

	result, err := r.db.ExecContext(ctx, query, email, passwordHash)
	if err != nil {
		// Check for duplicate entry error
		if isDuplicateKeyError(err) {
			return nil, ErrUserAlreadyExists
		}
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return r.FindByID(ctx, uint64(id))
}

// FindByID finds a user by ID
func (r *UserRepository) FindByID(ctx context.Context, id uint64) (*models.User, error) {
	query := `
		SELECT id, email, password_hash, role, is_active, email_verified_at, last_login_at, created_at, updated_at
		FROM users
		WHERE id = ?
	`

	user := &models.User{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.IsActive,
		&user.EmailVerifiedAt,
		&user.LastLoginAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return user, nil
}

// FindByEmail finds a user by email
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, email, password_hash, role, is_active, email_verified_at, last_login_at, created_at, updated_at
		FROM users
		WHERE email = ?
	`

	user := &models.User{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.IsActive,
		&user.EmailVerifiedAt,
		&user.LastLoginAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return user, nil
}

// UpdateEmail updates a user's email
func (r *UserRepository) UpdateEmail(ctx context.Context, id uint64, email string) error {
	query := `UPDATE users SET email = ?, updated_at = NOW() WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, email, id)
	if err != nil {
		if isDuplicateKeyError(err) {
			return ErrUserAlreadyExists
		}
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

// UpdatePassword updates a user's password
func (r *UserRepository) UpdatePassword(ctx context.Context, id uint64, passwordHash string) error {
	query := `UPDATE users SET password_hash = ?, updated_at = NOW() WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, passwordHash, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

// UpdateLastLogin updates a user's last login time
func (r *UserRepository) UpdateLastLogin(ctx context.Context, id uint64) error {
	query := `UPDATE users SET last_login_at = NOW(), updated_at = NOW() WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// Delete soft-deletes a user by deactivating them
func (r *UserRepository) Delete(ctx context.Context, id uint64) error {
	query := `UPDATE users SET is_active = FALSE, updated_at = NOW() WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

// ExistsByEmail checks if a user exists with the given email
func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = ?)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, email).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

// GetActiveUsersCount returns the count of active users
func (r *UserRepository) GetActiveUsersCount(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM users WHERE is_active = TRUE`

	var count int64
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// List returns a paginated list of users
func (r *UserRepository) List(ctx context.Context, offset, limit int) ([]*models.User, int64, error) {
	// Get total count
	countQuery := `SELECT COUNT(*) FROM users`
	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Get users
	query := `
		SELECT id, email, password_hash, role, is_active, email_verified_at, last_login_at, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.PasswordHash,
			&user.Role,
			&user.IsActive,
			&user.EmailVerifiedAt,
			&user.LastLoginAt,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// Helper function to check for duplicate key error
func isDuplicateKeyError(err error) bool {
	return err != nil && (contains(err.Error(), "Duplicate entry") || contains(err.Error(), "1062"))
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Transaction support
type TxFn func(*sql.Tx) error

// WithTransaction executes a function within a database transaction
func (r *UserRepository) WithTransaction(ctx context.Context, fn TxFn) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// Ensure the struct implements necessary time operations
var _ = time.Now