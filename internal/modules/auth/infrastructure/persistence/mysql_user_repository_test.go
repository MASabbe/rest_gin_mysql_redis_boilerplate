package persistence_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/auth/domain/entity"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/auth/infrastructure/persistence"
	appErrors "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMySQLUserRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := persistence.NewMySQLUserRepository(db)
	ctx := context.Background()

	user, _ := entity.NewUser("test@example.com", "hashedpassword")

	// 1. Success
	mock.ExpectExec("INSERT INTO users").
		WithArgs(user.ID, user.Email, user.PasswordHash, string(user.Status), nil, nil, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Create(ctx, user)
	assert.NoError(t, err)

	// 2. Database Error
	mock.ExpectExec("INSERT INTO users").
		WithArgs(user.ID, user.Email, user.PasswordHash, string(user.Status), nil, nil, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(errors.New("db connection lost"))

	err = repo.Create(ctx, user)
	assert.Error(t, err)
}

func TestMySQLUserRepository_FindByEmail(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := persistence.NewMySQLUserRepository(db)
	ctx := context.Background()

	now := time.Now().UTC()

	// 1. Found with activity timestamps
	rows := sqlmock.NewRows([]string{"id", "email", "password_hash", "status", "last_login_at", "last_activity_at", "created_at", "updated_at"}).
		AddRow("user-1", "test@example.com", "hash", "active", now, now, now, now)

	mock.ExpectQuery("SELECT (.+) FROM users WHERE email = ?").
		WithArgs("test@example.com").
		WillReturnRows(rows)

	found, err := repo.FindByEmail(ctx, "test@example.com")
	assert.NoError(t, err)
	assert.Equal(t, "user-1", found.ID)
	assert.Equal(t, "test@example.com", found.Email)
	assert.NotNil(t, found.LastLoginAt)
	assert.NotNil(t, found.LastActivityAt)

	// 2. Found with NULL timestamps (existing user)
	rowsNull := sqlmock.NewRows([]string{"id", "email", "password_hash", "status", "last_login_at", "last_activity_at", "created_at", "updated_at"}).
		AddRow("user-2", "null@example.com", "hash", "active", nil, nil, now, now)

	mock.ExpectQuery("SELECT (.+) FROM users WHERE email = ?").
		WithArgs("null@example.com").
		WillReturnRows(rowsNull)

	foundNull, err := repo.FindByEmail(ctx, "null@example.com")
	assert.NoError(t, err)
	assert.Nil(t, foundNull.LastLoginAt)
	assert.Nil(t, foundNull.LastActivityAt)

	// 3. Not Found
	mock.ExpectQuery("SELECT (.+) FROM users WHERE email = ?").
		WithArgs("notfound@example.com").
		WillReturnError(sql.ErrNoRows)

	_, err = repo.FindByEmail(ctx, "notfound@example.com")
	assert.Error(t, err)
	appErr := appErrors.AsAppError(err)
	assert.Equal(t, appErrors.TypeNotFound, appErr.Type)
}

func TestMySQLUserRepository_FindByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := persistence.NewMySQLUserRepository(db)
	ctx := context.Background()

	now := time.Now().UTC()

	// 1. Found
	rows := sqlmock.NewRows([]string{"id", "email", "password_hash", "status", "last_login_at", "last_activity_at", "created_at", "updated_at"}).
		AddRow("user-123", "idtest@example.com", "hash", "active", now, now, now, now)

	mock.ExpectQuery("SELECT (.+) FROM users WHERE id = ?").
		WithArgs("user-123").
		WillReturnRows(rows)

	found, err := repo.FindByID(ctx, "user-123")
	assert.NoError(t, err)
	assert.Equal(t, "user-123", found.ID)

	// 2. Not Found
	mock.ExpectQuery("SELECT (.+) FROM users WHERE id = ?").
		WithArgs("unknown-id").
		WillReturnError(sql.ErrNoRows)

	_, err = repo.FindByID(ctx, "unknown-id")
	assert.Error(t, err)
	appErr := appErrors.AsAppError(err)
	assert.Equal(t, appErrors.TypeNotFound, appErr.Type)
}

func TestMySQLUserRepository_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := persistence.NewMySQLUserRepository(db)
	ctx := context.Background()

	user, _ := entity.NewUser("update@example.com", "newhash")

	// 1. Success
	mock.ExpectExec("UPDATE users SET").
		WithArgs(user.PasswordHash, string(user.Status), nil, nil, sqlmock.AnyArg(), user.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Update(ctx, user)
	assert.NoError(t, err)

	// 2. Not Found (0 rows affected)
	mock.ExpectExec("UPDATE users SET").
		WithArgs(user.PasswordHash, string(user.Status), nil, nil, sqlmock.AnyArg(), user.ID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = repo.Update(ctx, user)
	assert.Error(t, err)
	appErr := appErrors.AsAppError(err)
	assert.Equal(t, appErrors.TypeNotFound, appErr.Type)
}

func TestMySQLUserRepository_UpdateLastLoginAt(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := persistence.NewMySQLUserRepository(db)
	ctx := context.Background()
	now := time.Now().UTC()
	userID := "user-login-update-1"

	// 1. Success
	mock.ExpectExec("UPDATE users SET last_login_at = \\?, updated_at = \\? WHERE id = \\?").
		WithArgs(now, now, userID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.UpdateLastLoginAt(ctx, userID, now)
	assert.NoError(t, err)

	// 2. Not Found (0 rows affected)
	mock.ExpectExec("UPDATE users SET last_login_at = \\?, updated_at = \\? WHERE id = \\?").
		WithArgs(now, now, "unknown-user").
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = repo.UpdateLastLoginAt(ctx, "unknown-user", now)
	assert.Error(t, err)
	assert.Equal(t, appErrors.TypeNotFound, appErrors.AsAppError(err).Type)
}

func TestMySQLUserRepository_UpdateLastActivityAt(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := persistence.NewMySQLUserRepository(db)
	ctx := context.Background()
	now := time.Now().UTC()
	userID := "user-activity-update-1"

	// 1. Success
	mock.ExpectExec("UPDATE users SET last_activity_at = \\?, updated_at = \\? WHERE id = \\?").
		WithArgs(now, now, userID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.UpdateLastActivityAt(ctx, userID, now)
	assert.NoError(t, err)

	// 2. Not Found (0 rows affected)
	mock.ExpectExec("UPDATE users SET last_activity_at = \\?, updated_at = \\? WHERE id = \\?").
		WithArgs(now, now, "unknown-user").
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = repo.UpdateLastActivityAt(ctx, "unknown-user", now)
	assert.Error(t, err)
	assert.Equal(t, appErrors.TypeNotFound, appErrors.AsAppError(err).Type)
}
