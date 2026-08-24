package persistence

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/rbac/domain/entity"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/rbac/domain/repository"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/logger"
)

type PermissionDefinition struct {
	Resource    string
	Action      string
	Description string
}

var CorePermissions = []PermissionDefinition{
	// Role management permissions
	{Resource: "role", Action: "read", Description: "View and list roles"},
	{Resource: "role", Action: "create", Description: "Create new roles"},
	{Resource: "role", Action: "update", Description: "Update existing roles and assign permissions"},
	{Resource: "role", Action: "delete", Description: "Delete non-system roles"},

	// Permission management permissions
	{Resource: "permission", Action: "read", Description: "View and list permissions"},
	{Resource: "permission", Action: "create", Description: "Create new permissions"},
	{Resource: "permission", Action: "update", Description: "Update existing permissions"},
	{Resource: "permission", Action: "delete", Description: "Delete permissions"},

	// User RBAC assignment permissions
	{Resource: "user", Action: "assign-role", Description: "Assign and revoke roles for users"},

	// Article module permissions
	{Resource: "article", Action: "read", Description: "View and list articles"},
	{Resource: "article", Action: "create", Description: "Create new articles"},
	{Resource: "article", Action: "update", Description: "Update articles"},
	{Resource: "article", Action: "delete", Description: "Delete articles"},
}

// Seeder handles idempotent database seeding for RBAC foundation data.
type Seeder struct {
	roleRepo       repository.RoleRepository
	permissionRepo repository.PermissionRepository
}

// NewSeeder creates a new RBAC Seeder instance.
func NewSeeder(roleRepo repository.RoleRepository, permissionRepo repository.PermissionRepository) *Seeder {
	return &Seeder{
		roleRepo:       roleRepo,
		permissionRepo: permissionRepo,
	}
}

// Seed ensures that core permissions, the system admin role, and their assignments exist idempotently.
func (s *Seeder) Seed(ctx context.Context) error {
	log := logger.Get()
	log.Info("Running RBAC idempotent database seeder...")

	// 1. Seed Core Permissions
	permissionIDs := make([]string, 0, len(CorePermissions))
	for _, def := range CorePermissions {
		permName := fmt.Sprintf("%s:%s", def.Resource, def.Action)
		existing, err := s.permissionRepo.FindByName(ctx, permName)
		if err == nil && existing != nil {
			permissionIDs = append(permissionIDs, existing.ID)
			continue
		}

		perm, err := entity.NewPermission(def.Resource, def.Action, def.Description)
		if err != nil {
			return fmt.Errorf("failed to create permission entity for '%s': %w", permName, err)
		}

		if err := s.permissionRepo.Create(ctx, perm); err != nil {
			return fmt.Errorf("failed to persist seeded permission '%s': %w", permName, err)
		}
		permissionIDs = append(permissionIDs, perm.ID)
	}

	// 2. Seed System Admin Role
	adminRole, err := s.roleRepo.FindByName(ctx, "admin")
	if err != nil || adminRole == nil {
		adminRole, err = entity.NewSystemRole("admin", "System Administrator with full permissions")
		if err != nil {
			return fmt.Errorf("failed to create admin role entity: %w", err)
		}

		if err := s.roleRepo.Create(ctx, adminRole); err != nil {
			return fmt.Errorf("failed to persist seeded admin role: %w", err)
		}
	}

	// 3. Assign all core permissions to Admin Role
	if err := s.roleRepo.AssignPermissions(ctx, adminRole.ID, permissionIDs); err != nil {
		return fmt.Errorf("failed to assign seeded permissions to admin role: %w", err)
	}

	log.Info("RBAC seeder completed successfully",
		slog.Int("permissions_seeded", len(permissionIDs)),
		slog.String("admin_role_id", adminRole.ID),
	)

	return nil
}
