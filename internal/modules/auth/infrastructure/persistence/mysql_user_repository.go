package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

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
		INSERT INTO users (id, email, password_hash, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.Email,
		user.PasswordHash,
		string(user.Status),
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
		SELECT id, email, password_hash, status, created_at, updated_at
		FROM users
		WHERE email = ?
		LIMIT 1
	`
	var user entity.User
	var statusStr string

	err := r.db.QueryRowContext(ctx, query, strings.ToLower(strings.TrimSpace(email))).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&statusStr,
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
	return &user, nil
}

// FindByID finds a user record by user ID.
func (r *MySQLUserRepository) FindByID(ctx context.Context, id string) (*entity.User, error) {
	query := `
		SELECT id, email, password_hash, status, created_at, updated_at
		FROM users
		WHERE id = ?
		LIMIT 1
	`
	var user entity.User
	var statusStr string

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&statusStr,
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
	return &user, nil
}

// Update updates an existing user record.
func (r *MySQLUserRepository) Update(ctx context.Context, user *entity.User) error {
	query := `
		UPDATE users
		SET password_hash = ?, status = ?, updated_at = ?
		WHERE id = ?
	`
	result, err := r.db.ExecContext(ctx, query,
		user.PasswordHash,
		string(user.Status),
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
