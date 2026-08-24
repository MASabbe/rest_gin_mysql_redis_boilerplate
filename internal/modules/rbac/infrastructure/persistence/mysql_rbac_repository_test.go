package persistence_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/rbac/domain/entity"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/rbac/infrastructure/persistence"
	appErrors "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/errors"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/pagination"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMySQLRoleRepository_CRUD(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := persistence.NewMySQLRoleRepository(db)
	ctx := context.Background()
	now := time.Now().UTC()

	role, _ := entity.NewRole("editor", "Editor role")

	// 1. Create Role
	mock.ExpectExec("INSERT INTO roles").
		WithArgs(role.ID, role.Name, role.Description, role.IsSystem, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Create(ctx, role)
	assert.NoError(t, err)

	// 2. FindByID - Found
	rows := sqlmock.NewRows([]string{"id", "name", "description", "is_system", "created_at", "updated_at"}).
		AddRow(role.ID, role.Name, role.Description, role.IsSystem, now, now)

	mock.ExpectQuery("SELECT (.+) FROM roles WHERE id = ?").
		WithArgs(role.ID).
		WillReturnRows(rows)

	found, err := repo.FindByID(ctx, role.ID)
	assert.NoError(t, err)
	assert.Equal(t, role.ID, found.ID)
	assert.Equal(t, "editor", found.Name)

	// 3. FindByName - Found
	mock.ExpectQuery("SELECT (.+) FROM roles WHERE name = ?").
		WithArgs("editor").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "is_system", "created_at", "updated_at"}).
			AddRow(role.ID, "editor", "Editor role", false, now, now))

	foundByName, err := repo.FindByName(ctx, "editor")
	assert.NoError(t, err)
	assert.Equal(t, "editor", foundByName.Name)

	// 4. List Roles
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM roles").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectQuery("SELECT (.+) FROM roles ORDER BY created_at ASC LIMIT \\? OFFSET \\?").
		WithArgs(10, 0).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "is_system", "created_at", "updated_at"}).
			AddRow(role.ID, "editor", "Editor role", false, now, now))

	roles, total, err := repo.List(ctx, pagination.Pagination{Page: 1, PageSize: 10})
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, roles, 1)

	// 5. Update Role
	mock.ExpectExec("UPDATE roles SET").
		WithArgs("senior-editor", "Senior Editor", role.IsSystem, sqlmock.AnyArg(), role.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	role.Name = "senior-editor"
	role.Description = "Senior Editor"
	err = repo.Update(ctx, role)
	assert.NoError(t, err)

	// 6. Delete Role
	mock.ExpectExec("DELETE FROM roles WHERE id = ?").
		WithArgs(role.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Delete(ctx, role.ID)
	assert.NoError(t, err)
}

func TestMySQLPermissionRepository_CRUD(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := persistence.NewMySQLPermissionRepository(db)
	ctx := context.Background()
	now := time.Now().UTC()

	perm, _ := entity.NewPermission("article", "publish", "Publish article")

	// 1. Create Permission
	mock.ExpectExec("INSERT INTO permissions").
		WithArgs(perm.ID, perm.Name, perm.Resource, perm.Action, perm.Description, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Create(ctx, perm)
	assert.NoError(t, err)

	// 2. FindByID - Found
	mock.ExpectQuery("SELECT (.+) FROM permissions WHERE id = ?").
		WithArgs(perm.ID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "resource", "action", "description", "created_at", "updated_at"}).
			AddRow(perm.ID, perm.Name, perm.Resource, perm.Action, perm.Description, now, now))

	found, err := repo.FindByID(ctx, perm.ID)
	assert.NoError(t, err)
	assert.Equal(t, perm.ID, found.ID)
	assert.Equal(t, "article:publish", found.Name)

	// 3. FindByName - Not Found
	mock.ExpectQuery("SELECT (.+) FROM permissions WHERE name = ?").
		WithArgs("article:unknown").
		WillReturnError(sql.ErrNoRows)

	_, err = repo.FindByName(ctx, "article:unknown")
	assert.Error(t, err)
	assert.Equal(t, appErrors.TypeNotFound, appErrors.AsAppError(err).Type)

	// 4. List Permissions
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM permissions").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectQuery("SELECT (.+) FROM permissions ORDER BY resource ASC, action ASC LIMIT \\? OFFSET \\?").
		WithArgs(10, 0).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "resource", "action", "description", "created_at", "updated_at"}).
			AddRow(perm.ID, perm.Name, perm.Resource, perm.Action, perm.Description, now, now))

	perms, total, err := repo.List(ctx, pagination.Pagination{Page: 1, PageSize: 10})
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, perms, 1)

	// 5. Delete Permission
	mock.ExpectExec("DELETE FROM permissions WHERE id = ?").
		WithArgs(perm.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Delete(ctx, perm.ID)
	assert.NoError(t, err)
}
