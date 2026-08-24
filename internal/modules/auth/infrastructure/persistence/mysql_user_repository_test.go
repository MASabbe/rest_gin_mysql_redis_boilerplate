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
		WithArgs(user.ID, user.Email, user.PasswordHash, string(user.Status), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Create(ctx, user)
	assert.NoError(t, err)

	// 2. Database Error
	mock.ExpectExec("INSERT INTO users").
		WithArgs(user.ID, user.Email, user.PasswordHash, string(user.Status), sqlmock.AnyArg(), sqlmock.AnyArg()).
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

	// 1. Found
	rows := sqlmock.NewRows([]string{"id", "email", "password_hash", "status", "created_at", "updated_at"}).
		AddRow("user-1", "test@example.com", "hash", "active", now, now)

	mock.ExpectQuery("SELECT (.+) FROM users WHERE email = ?").
		WithArgs("test@example.com").
		WillReturnRows(rows)

	found, err := repo.FindByEmail(ctx, "test@example.com")
	assert.NoError(t, err)
	assert.Equal(t, "user-1", found.ID)
	assert.Equal(t, "test@example.com", found.Email)

	// 2. Not Found
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
	rows := sqlmock.NewRows([]string{"id", "email", "password_hash", "status", "created_at", "updated_at"}).
		AddRow("user-123", "idtest@example.com", "hash", "active", now, now)

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
		WithArgs(user.PasswordHash, string(user.Status), sqlmock.AnyArg(), user.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Update(ctx, user)
	assert.NoError(t, err)

	// 2. Not Found (0 rows affected)
	mock.ExpectExec("UPDATE users SET").
		WithArgs(user.PasswordHash, string(user.Status), sqlmock.AnyArg(), user.ID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = repo.Update(ctx, user)
	assert.Error(t, err)
	appErr := appErrors.AsAppError(err)
	assert.Equal(t, appErrors.TypeNotFound, appErr.Type)
}
