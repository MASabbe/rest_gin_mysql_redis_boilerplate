package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/auth/domain/entity"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/auth/domain/repository"
	appErrors "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/errors"
	"github.com/go-sql-driver/mysql"
)

type MySQLUserRepository struct {
	db *sql.DB
}

// NewMySQLUserRepository creates a new MySQL user repository.
func NewMySQLUserRepository(db *sql.DB) repository.UserRepository {
	return &MySQLUserRepository{db: db}
}

// Create inserts a new user record.
func (r *MySQLUserRepository) Create(ctx context.Context, user *entity.User) error {
	query := `
		INSERT INTO users (id, email, password_hash, status, last_login_at, last_activity_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.Email,
		user.PasswordHash,
		string(user.Status),
		user.LastLoginAt,
		user.LastActivityAt,
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return appErrors.NewConflictError("email already registered", err)
		}
		if strings.Contains(err.Error(), "Duplicate entry") {
			return appErrors.NewConflictError("email already registered", err)
		}
		return appErrors.NewInternalError("failed to create user", err)
	}

	return nil
}

// FindByEmail finds a user record by email.
func (r *MySQLUserRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	query := `
		SELECT id, email, password_hash, status, last_login_at, last_activity_at, created_at, updated_at
		FROM users
		WHERE email = ?
		LIMIT 1
	`
	var user entity.User
	var statusStr string
	var lastLoginAt sql.NullTime
	var lastActivityAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, strings.ToLower(strings.TrimSpace(email))).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&statusStr,
		&lastLoginAt,
		&lastActivityAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, appErrors.NewNotFoundError("user not found", err)
		}
		return nil, appErrors.NewInternalError("failed to find user by email", err)
	}

	user.Status = entity.UserStatus(statusStr)
	if lastLoginAt.Valid {
		t := lastLoginAt.Time.UTC()
		user.LastLoginAt = &t
	}
	if lastActivityAt.Valid {
		t := lastActivityAt.Time.UTC()
		user.LastActivityAt = &t
	}

	return &user, nil
}

// FindByID finds a user record by user ID.
func (r *MySQLUserRepository) FindByID(ctx context.Context, id string) (*entity.User, error) {
	query := `
		SELECT id, email, password_hash, status, last_login_at, last_activity_at, created_at, updated_at
		FROM users
		WHERE id = ?
		LIMIT 1
	`
	var user entity.User
	var statusStr string
	var lastLoginAt sql.NullTime
	var lastActivityAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&statusStr,
		&lastLoginAt,
		&lastActivityAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, appErrors.NewNotFoundError("user not found", err)
		}
		return nil, appErrors.NewInternalError("failed to find user by id", err)
	}

	user.Status = entity.UserStatus(statusStr)
	if lastLoginAt.Valid {
		t := lastLoginAt.Time.UTC()
		user.LastLoginAt = &t
	}
	if lastActivityAt.Valid {
		t := lastActivityAt.Time.UTC()
		user.LastActivityAt = &t
	}

	return &user, nil
}

// Update updates an existing user record.
func (r *MySQLUserRepository) Update(ctx context.Context, user *entity.User) error {
	query := `
		UPDATE users
		SET password_hash = ?, status = ?, last_login_at = ?, last_activity_at = ?, updated_at = ?
		WHERE id = ?
	`
	result, err := r.db.ExecContext(ctx, query,
		user.PasswordHash,
		string(user.Status),
		user.LastLoginAt,
		user.LastActivityAt,
		user.UpdatedAt,
		user.ID,
	)

	if err != nil {
		return appErrors.NewInternalError("failed to update user", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return appErrors.NewInternalError("failed to check update result", err)
	}

	if rows == 0 {
		return appErrors.NewNotFoundError("user not found for update", fmt.Errorf("id: %s", user.ID))
	}

	return nil
}

// UpdateLastLoginAt explicitly updates the last_login_at and updated_at timestamps in UTC.
func (r *MySQLUserRepository) UpdateLastLoginAt(ctx context.Context, userID string, at time.Time) error {
	utcTime := at.UTC()
	query := `
		UPDATE users
		SET last_login_at = ?, updated_at = ?
		WHERE id = ?
	`
	result, err := r.db.ExecContext(ctx, query, utcTime, utcTime, userID)
	if err != nil {
		return appErrors.NewInternalError("failed to update last_login_at", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return appErrors.NewInternalError("failed to check update result", err)
	}

	if rows == 0 {
		return appErrors.NewNotFoundError("user not found for last_login_at update", fmt.Errorf("id: %s", userID))
	}

	return nil
}

// UpdateLastActivityAt explicitly updates the last_activity_at and updated_at timestamps in UTC.
func (r *MySQLUserRepository) UpdateLastActivityAt(ctx context.Context, userID string, at time.Time) error {
	utcTime := at.UTC()
	query := `
		UPDATE users
		SET last_activity_at = ?, updated_at = ?
		WHERE id = ?
	`
	result, err := r.db.ExecContext(ctx, query, utcTime, utcTime, userID)
	if err != nil {
		return appErrors.NewInternalError("failed to update last_activity_at", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return appErrors.NewInternalError("failed to check update result", err)
	}

	if rows == 0 {
		return appErrors.NewNotFoundError("user not found for last_activity_at update", fmt.Errorf("id: %s", userID))
	}

	return nil
}
