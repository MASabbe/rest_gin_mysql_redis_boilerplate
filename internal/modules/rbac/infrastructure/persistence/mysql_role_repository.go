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

type MySQLRoleRepository struct {
	db *sql.DB
}

// NewMySQLRoleRepository creates a new MySQL Role repository.
func NewMySQLRoleRepository(db *sql.DB) repository.RoleRepository {
	return &MySQLRoleRepository{db: db}
}

func (r *MySQLRoleRepository) Create(ctx context.Context, role *entity.Role) error {
	query := `
		INSERT INTO roles (id, name, description, is_system, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query,
		role.ID,
		role.Name,
		role.Description,
		role.IsSystem,
		role.CreatedAt,
		role.UpdatedAt,
	)

	if err != nil {
		var mysqlErr *mysql.MySQLError
		if (errors.As(err, &mysqlErr) && mysqlErr.Number == 1062) || strings.Contains(err.Error(), "Duplicate entry") {
			return appErrors.NewConflictError(fmt.Sprintf("role with name '%s' already exists", role.Name), err)
		}
		return appErrors.NewInternalError("failed to create role", err)
	}

	return nil
}

func (r *MySQLRoleRepository) FindByID(ctx context.Context, id string) (*entity.Role, error) {
	query := `
		SELECT id, name, description, is_system, created_at, updated_at
		FROM roles
		WHERE id = ?
		LIMIT 1
	`
	var role entity.Role
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&role.ID,
		&role.Name,
		&role.Description,
		&role.IsSystem,
		&role.CreatedAt,
		&role.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, appErrors.NewNotFoundError("role not found", err)
		}
		return nil, appErrors.NewInternalError("failed to find role by id", err)
	}

	return &role, nil
}

func (r *MySQLRoleRepository) FindByName(ctx context.Context, name string) (*entity.Role, error) {
	query := `
		SELECT id, name, description, is_system, created_at, updated_at
		FROM roles
		WHERE name = ?
		LIMIT 1
	`
	var role entity.Role
	err := r.db.QueryRowContext(ctx, query, strings.ToLower(strings.TrimSpace(name))).Scan(
		&role.ID,
		&role.Name,
		&role.Description,
		&role.IsSystem,
		&role.CreatedAt,
		&role.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, appErrors.NewNotFoundError("role not found", err)
		}
		return nil, appErrors.NewInternalError("failed to find role by name", err)
	}

	return &role, nil
}

func (r *MySQLRoleRepository) List(ctx context.Context, p pagination.Pagination) ([]*entity.Role, int64, error) {
	var total int64
	countQuery := `SELECT COUNT(*) FROM roles`
	if err := r.db.QueryRowContext(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, appErrors.NewInternalError("failed to count roles", err)
	}

	query := `
		SELECT id, name, description, is_system, created_at, updated_at
		FROM roles
		ORDER BY created_at ASC
		LIMIT ? OFFSET ?
	`
	rows, err := r.db.QueryContext(ctx, query, p.Limit(), p.Offset())
	if err != nil {
		return nil, 0, appErrors.NewInternalError("failed to list roles", err)
	}
	defer rows.Close()

	var roles []*entity.Role
	for rows.Next() {
		var role entity.Role
		if err := rows.Scan(
			&role.ID,
			&role.Name,
			&role.Description,
			&role.IsSystem,
			&role.CreatedAt,
			&role.UpdatedAt,
		); err != nil {
			return nil, 0, appErrors.NewInternalError("failed to scan role row", err)
		}
		roles = append(roles, &role)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, appErrors.NewInternalError("error iterating roles", err)
	}

	return roles, total, nil
}

func (r *MySQLRoleRepository) Update(ctx context.Context, role *entity.Role) error {
	query := `
		UPDATE roles
		SET name = ?, description = ?, is_system = ?, updated_at = ?
		WHERE id = ?
	`
	result, err := r.db.ExecContext(ctx, query,
		role.Name,
		role.Description,
		role.IsSystem,
		role.UpdatedAt,
		role.ID,
	)

	if err != nil {
		var mysqlErr *mysql.MySQLError
		if (errors.As(err, &mysqlErr) && mysqlErr.Number == 1062) || strings.Contains(err.Error(), "Duplicate entry") {
			return appErrors.NewConflictError(fmt.Sprintf("role with name '%s' already exists", role.Name), err)
		}
		return appErrors.NewInternalError("failed to update role", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return appErrors.NewInternalError("failed to check rows affected", err)
	}
	if rows == 0 {
		return appErrors.NewNotFoundError("role not found for update")
	}

	return nil
}

func (r *MySQLRoleRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM roles WHERE id = ?`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return appErrors.NewInternalError("failed to delete role", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return appErrors.NewInternalError("failed to check rows affected", err)
	}
	if rows == 0 {
		return appErrors.NewNotFoundError("role not found for deletion")
	}

	return nil
}

// AssignPermissions assigns a list of permissions to a role within a single transaction.
func (r *MySQLRoleRepository) AssignPermissions(ctx context.Context, roleID string, permissionIDs []string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return appErrors.NewInternalError("failed to begin transaction", err)
	}
	defer tx.Rollback()

	// 1. Delete existing associations
	deleteQuery := `DELETE FROM role_permissions WHERE role_id = ?`
	if _, err := tx.ExecContext(ctx, deleteQuery, roleID); err != nil {
		return appErrors.NewInternalError("failed to clear existing role permissions", err)
	}

	// 2. Insert new associations
	if len(permissionIDs) > 0 {
		insertQuery := `INSERT INTO role_permissions (role_id, permission_id) VALUES (?, ?)`
		stmt, err := tx.PrepareContext(ctx, insertQuery)
		if err != nil {
			return appErrors.NewInternalError("failed to prepare insert statement", err)
		}
		defer stmt.Close()

		for _, pid := range permissionIDs {
			if strings.TrimSpace(pid) == "" {
				continue
			}
			if _, err := stmt.ExecContext(ctx, roleID, pid); err != nil {
				return appErrors.NewInternalError("failed to associate permission to role", err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return appErrors.NewInternalError("failed to commit transaction", err)
	}

	return nil
}

func (r *MySQLRoleRepository) GetRolePermissionIDs(ctx context.Context, roleID string) ([]string, error) {
	query := `SELECT permission_id FROM role_permissions WHERE role_id = ?`
	rows, err := r.db.QueryContext(ctx, query, roleID)
	if err != nil {
		return nil, appErrors.NewInternalError("failed to query role permission IDs", err)
	}
	defer rows.Close()

	var permissionIDs []string
	for rows.Next() {
		var pid string
		if err := rows.Scan(&pid); err != nil {
			return nil, appErrors.NewInternalError("failed to scan permission ID", err)
		}
		permissionIDs = append(permissionIDs, pid)
	}

	return permissionIDs, nil
}

// AssignUserRoles assigns roles to a user within a single transaction.
func (r *MySQLRoleRepository) AssignUserRoles(ctx context.Context, userID string, roleIDs []string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return appErrors.NewInternalError("failed to begin transaction", err)
	}
	defer tx.Rollback()

	// 1. Delete existing user roles
	deleteQuery := `DELETE FROM user_roles WHERE user_id = ?`
	if _, err := tx.ExecContext(ctx, deleteQuery, userID); err != nil {
		return appErrors.NewInternalError("failed to clear existing user roles", err)
	}

	// 2. Insert new user roles
	if len(roleIDs) > 0 {
		insertQuery := `INSERT INTO user_roles (user_id, role_id) VALUES (?, ?)`
		stmt, err := tx.PrepareContext(ctx, insertQuery)
		if err != nil {
			return appErrors.NewInternalError("failed to prepare insert statement", err)
		}
		defer stmt.Close()

		for _, rid := range roleIDs {
			if strings.TrimSpace(rid) == "" {
				continue
			}
			if _, err := stmt.ExecContext(ctx, userID, rid); err != nil {
				return appErrors.NewInternalError("failed to associate role to user", err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return appErrors.NewInternalError("failed to commit transaction", err)
	}

	return nil
}

func (r *MySQLRoleRepository) GetUserRoles(ctx context.Context, userID string) ([]*entity.Role, error) {
	query := `
		SELECT r.id, r.name, r.description, r.is_system, r.created_at, r.updated_at
		FROM roles r
		JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = ?
		ORDER BY r.name ASC
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, appErrors.NewInternalError("failed to get user roles", err)
	}
	defer rows.Close()

	var roles []*entity.Role
	for rows.Next() {
		var role entity.Role
		if err := rows.Scan(
			&role.ID,
			&role.Name,
			&role.Description,
			&role.IsSystem,
			&role.CreatedAt,
			&role.UpdatedAt,
		); err != nil {
			return nil, appErrors.NewInternalError("failed to scan user role", err)
		}
		roles = append(roles, &role)
	}

	return roles, nil
}

func (r *MySQLRoleRepository) GetUserIDsByRoleID(ctx context.Context, roleID string) ([]string, error) {
	query := `SELECT user_id FROM user_roles WHERE role_id = ?`
	rows, err := r.db.QueryContext(ctx, query, roleID)
	if err != nil {
		return nil, appErrors.NewInternalError("failed to get user IDs for role", err)
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

var _ repository.RoleRepository = (*MySQLRoleRepository)(nil)
