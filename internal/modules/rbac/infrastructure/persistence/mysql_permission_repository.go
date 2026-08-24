package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/rbac/domain/entity"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/rbac/domain/repository"
	appErrors "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/errors"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/pagination"
	"github.com/go-sql-driver/mysql"
)

type MySQLPermissionRepository struct {
	db *sql.DB
}

// NewMySQLPermissionRepository creates a new MySQL Permission repository.
func NewMySQLPermissionRepository(db *sql.DB) repository.PermissionRepository {
	return &MySQLPermissionRepository{db: db}
}

func (r *MySQLPermissionRepository) Create(ctx context.Context, perm *entity.Permission) error {
	query := `
		INSERT INTO permissions (id, name, resource, action, description, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query,
		perm.ID,
		perm.Name,
		perm.Resource,
		perm.Action,
		perm.Description,
		perm.CreatedAt,
		perm.UpdatedAt,
	)

	if err != nil {
		var mysqlErr *mysql.MySQLError
		if (errors.As(err, &mysqlErr) && mysqlErr.Number == 1062) || strings.Contains(err.Error(), "Duplicate entry") {
			return appErrors.NewConflictError(fmt.Sprintf("permission with name '%s' already exists", perm.Name), err)
		}
		return appErrors.NewInternalError("failed to create permission", err)
	}

	return nil
}

func (r *MySQLPermissionRepository) FindByID(ctx context.Context, id string) (*entity.Permission, error) {
	query := `
		SELECT id, name, resource, action, description, created_at, updated_at
		FROM permissions
		WHERE id = ?
		LIMIT 1
	`
	var perm entity.Permission
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&perm.ID,
		&perm.Name,
		&perm.Resource,
		&perm.Action,
		&perm.Description,
		&perm.CreatedAt,
		&perm.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, appErrors.NewNotFoundError("permission not found", err)
		}
		return nil, appErrors.NewInternalError("failed to find permission by id", err)
	}

	return &perm, nil
}

func (r *MySQLPermissionRepository) FindByName(ctx context.Context, name string) (*entity.Permission, error) {
	query := `
		SELECT id, name, resource, action, description, created_at, updated_at
		FROM permissions
		WHERE name = ?
		LIMIT 1
	`
	var perm entity.Permission
	err := r.db.QueryRowContext(ctx, query, strings.ToLower(strings.TrimSpace(name))).Scan(
		&perm.ID,
		&perm.Name,
		&perm.Resource,
		&perm.Action,
		&perm.Description,
		&perm.CreatedAt,
		&perm.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, appErrors.NewNotFoundError("permission not found", err)
		}
		return nil, appErrors.NewInternalError("failed to find permission by name", err)
	}

	return &perm, nil
}

func (r *MySQLPermissionRepository) List(ctx context.Context, p pagination.Pagination) ([]*entity.Permission, int64, error) {
	var total int64
	countQuery := `SELECT COUNT(*) FROM permissions`
	if err := r.db.QueryRowContext(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, appErrors.NewInternalError("failed to count permissions", err)
	}

	query := `
		SELECT id, name, resource, action, description, created_at, updated_at
		FROM permissions
		ORDER BY resource ASC, action ASC
		LIMIT ? OFFSET ?
	`
	rows, err := r.db.QueryContext(ctx, query, p.Limit(), p.Offset())
	if err != nil {
		return nil, 0, appErrors.NewInternalError("failed to list permissions", err)
	}
	defer rows.Close()

	var perms []*entity.Permission
	for rows.Next() {
		var perm entity.Permission
		if err := rows.Scan(
			&perm.ID,
			&perm.Name,
			&perm.Resource,
			&perm.Action,
			&perm.Description,
			&perm.CreatedAt,
			&perm.UpdatedAt,
		); err != nil {
			return nil, 0, appErrors.NewInternalError("failed to scan permission row", err)
		}
		perms = append(perms, &perm)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, appErrors.NewInternalError("error iterating permissions", err)
	}

	return perms, total, nil
}

func (r *MySQLPermissionRepository) Update(ctx context.Context, perm *entity.Permission) error {
	query := `
		UPDATE permissions
		SET name = ?, resource = ?, action = ?, description = ?, updated_at = ?
		WHERE id = ?
	`
	result, err := r.db.ExecContext(ctx, query,
		perm.Name,
		perm.Resource,
		perm.Action,
		perm.Description,
		perm.UpdatedAt,
		perm.ID,
	)

	if err != nil {
		var mysqlErr *mysql.MySQLError
		if (errors.As(err, &mysqlErr) && mysqlErr.Number == 1062) || strings.Contains(err.Error(), "Duplicate entry") {
			return appErrors.NewConflictError(fmt.Sprintf("permission with name '%s' already exists", perm.Name), err)
		}
		return appErrors.NewInternalError("failed to update permission", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return appErrors.NewInternalError("failed to check rows affected", err)
	}
	if rows == 0 {
		return appErrors.NewNotFoundError("permission not found for update")
	}

	return nil
}

func (r *MySQLPermissionRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM permissions WHERE id = ?`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return appErrors.NewInternalError("failed to delete permission", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return appErrors.NewInternalError("failed to check rows affected", err)
	}
	if rows == 0 {
		return appErrors.NewNotFoundError("permission not found for deletion")
	}

	return nil
}

func (r *MySQLPermissionRepository) FindByRoleID(ctx context.Context, roleID string) ([]*entity.Permission, error) {
	query := `
		SELECT p.id, p.name, p.resource, p.action, p.description, p.created_at, p.updated_at
		FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		WHERE rp.role_id = ?
		ORDER BY p.name ASC
	`
	rows, err := r.db.QueryContext(ctx, query, roleID)
	if err != nil {
		return nil, appErrors.NewInternalError("failed to query permissions by role ID", err)
	}
	defer rows.Close()

	var perms []*entity.Permission
	for rows.Next() {
		var perm entity.Permission
		if err := rows.Scan(
			&perm.ID,
			&perm.Name,
			&perm.Resource,
			&perm.Action,
			&perm.Description,
			&perm.CreatedAt,
			&perm.UpdatedAt,
		); err != nil {
			return nil, appErrors.NewInternalError("failed to scan permission", err)
		}
		perms = append(perms, &perm)
	}

	return perms, nil
}

func (r *MySQLPermissionRepository) FindEffectivePermissionsByUserID(ctx context.Context, userID string) ([]string, error) {
	query := `
		SELECT DISTINCT p.name
		FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		JOIN user_roles ur ON rp.role_id = ur.role_id
		WHERE ur.user_id = ?
		ORDER BY p.name ASC
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, appErrors.NewInternalError("failed to query effective user permissions", err)
	}
	defer rows.Close()

	var permNames []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, appErrors.NewInternalError("failed to scan permission name", err)
		}
		permNames = append(permNames, name)
	}

	return permNames, nil
}

func (r *MySQLPermissionRepository) GetAffectedUserIDsByPermissionID(ctx context.Context, permissionID string) ([]string, error) {
	query := `
		SELECT DISTINCT ur.user_id
		FROM user_roles ur
		JOIN role_permissions rp ON ur.role_id = rp.role_id
		WHERE rp.permission_id = ?
	`
	rows, err := r.db.QueryContext(ctx, query, permissionID)
	if err != nil {
		return nil, appErrors.NewInternalError("failed to get affected user IDs for permission", err)
	}
	defer rows.Close()

	var userIDs []string
	for rows.Next() {
		var uid string
		if err := rows.Scan(&uid); err != nil {
			return nil, appErrors.NewInternalError("failed to scan user ID", err)
		}
		userIDs = append(userIDs, uid)
	}

	return userIDs, nil
}

var _ repository.PermissionRepository = (*MySQLPermissionRepository)(nil)
