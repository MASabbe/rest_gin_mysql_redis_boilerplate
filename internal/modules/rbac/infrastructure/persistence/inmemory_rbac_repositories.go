package persistence

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/rbac/domain/entity"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/modules/rbac/domain/repository"
	appErrors "github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/errors"
	"github.com/MASabbe/rest_gin_mysql_redis_boilerplate/internal/shared/pagination"
)

// InMemoryRoleRepository provides an in-memory thread-safe implementation of RoleRepository.
type InMemoryRoleRepository struct {
	mu              sync.RWMutex
	roles           map[string]*entity.Role
	rolesByName     map[string]*entity.Role
	rolePermissions map[string]map[string]bool // roleID -> permissionID set
	userRoles       map[string]map[string]bool // userID -> roleID set
}

func NewInMemoryRoleRepository() *InMemoryRoleRepository {
	return &InMemoryRoleRepository{
		roles:           make(map[string]*entity.Role),
		rolesByName:     make(map[string]*entity.Role),
		rolePermissions: make(map[string]map[string]bool),
		userRoles:       make(map[string]map[string]bool),
	}
}

func (r *InMemoryRoleRepository) Create(ctx context.Context, role *entity.Role) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	cleanName := strings.ToLower(role.Name)
	if _, exists := r.rolesByName[cleanName]; exists {
		return appErrors.NewConflictError(fmt.Sprintf("role with name '%s' already exists", role.Name))
	}

	r.roles[role.ID] = role
	r.rolesByName[cleanName] = role
	return nil
}

func (r *InMemoryRoleRepository) FindByID(ctx context.Context, id string) (*entity.Role, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	role, exists := r.roles[id]
	if !exists {
		return nil, appErrors.NewNotFoundError("role not found")
	}
	return role, nil
}

func (r *InMemoryRoleRepository) FindByName(ctx context.Context, name string) (*entity.Role, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	role, exists := r.rolesByName[strings.ToLower(name)]
	if !exists {
		return nil, appErrors.NewNotFoundError("role not found")
	}
	return role, nil
}

func (r *InMemoryRoleRepository) List(ctx context.Context, p pagination.Pagination) ([]*entity.Role, int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var all []*entity.Role
	for _, role := range r.roles {
		all = append(all, role)
	}

	total := int64(len(all))
	offset := p.Offset()
	limit := p.Limit()

	if offset >= len(all) {
		return []*entity.Role{}, total, nil
	}

	end := offset + limit
	if end > len(all) {
		end = len(all)
	}

	return all[offset:end], total, nil
}

func (r *InMemoryRoleRepository) Update(ctx context.Context, role *entity.Role) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, exists := r.roles[role.ID]
	if !exists {
		return appErrors.NewNotFoundError("role not found")
	}

	cleanName := strings.ToLower(role.Name)
	if other, nameTaken := r.rolesByName[cleanName]; nameTaken && other.ID != role.ID {
		return appErrors.NewConflictError(fmt.Sprintf("role with name '%s' already exists", role.Name))
	}

	delete(r.rolesByName, strings.ToLower(existing.Name))
	r.roles[role.ID] = role
	r.rolesByName[cleanName] = role
	return nil
}

func (r *InMemoryRoleRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	role, exists := r.roles[id]
	if !exists {
		return appErrors.NewNotFoundError("role not found")
	}

	delete(r.roles, id)
	delete(r.rolesByName, strings.ToLower(role.Name))
	delete(r.rolePermissions, id)

	// Clean up user_roles referencing this role
	for uid, roles := range r.userRoles {
		delete(roles, id)
		if len(roles) == 0 {
			delete(r.userRoles, uid)
		}
	}

	return nil
}

func (r *InMemoryRoleRepository) AssignPermissions(ctx context.Context, roleID string, permissionIDs []string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.roles[roleID]; !exists {
		return appErrors.NewNotFoundError("role not found")
	}

	permSet := make(map[string]bool)
	for _, pid := range permissionIDs {
		permSet[pid] = true
	}
	r.rolePermissions[roleID] = permSet
	return nil
}

func (r *InMemoryRoleRepository) GetRolePermissionIDs(ctx context.Context, roleID string) ([]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	permSet, exists := r.rolePermissions[roleID]
	if !exists {
		return []string{}, nil
	}

	var pids []string
	for pid := range permSet {
		pids = append(pids, pid)
	}
	return pids, nil
}

func (r *InMemoryRoleRepository) AssignUserRoles(ctx context.Context, userID string, roleIDs []string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	roleSet := make(map[string]bool)
	for _, rid := range roleIDs {
		roleSet[rid] = true
	}
	r.userRoles[userID] = roleSet
	return nil
}

func (r *InMemoryRoleRepository) GetUserRoles(ctx context.Context, userID string) ([]*entity.Role, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	roleSet, exists := r.userRoles[userID]
	if !exists {
		return []*entity.Role{}, nil
	}

	var list []*entity.Role
	for rid := range roleSet {
		if role, ok := r.roles[rid]; ok {
			list = append(list, role)
		}
	}
	return list, nil
}

func (r *InMemoryRoleRepository) GetUserIDsByRoleID(ctx context.Context, roleID string) ([]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var userIDs []string
	for uid, roles := range r.userRoles {
		if roles[roleID] {
			userIDs = append(userIDs, uid)
		}
	}
	return userIDs, nil
}

// InMemoryPermissionRepository provides an in-memory thread-safe implementation of PermissionRepository.
type InMemoryPermissionRepository struct {
	mu                sync.RWMutex
	permissions       map[string]*entity.Permission
	permissionsByName map[string]*entity.Permission
	roleRepo          *InMemoryRoleRepository
}

func NewInMemoryPermissionRepository(roleRepo *InMemoryRoleRepository) *InMemoryPermissionRepository {
	return &InMemoryPermissionRepository{
		permissions:       make(map[string]*entity.Permission),
		permissionsByName: make(map[string]*entity.Permission),
		roleRepo:          roleRepo,
	}
}

func (r *InMemoryPermissionRepository) Create(ctx context.Context, perm *entity.Permission) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	cleanName := strings.ToLower(perm.Name)
	if _, exists := r.permissionsByName[cleanName]; exists {
		return appErrors.NewConflictError(fmt.Sprintf("permission with name '%s' already exists", perm.Name))
	}

	r.permissions[perm.ID] = perm
	r.permissionsByName[cleanName] = perm
	return nil
}

func (r *InMemoryPermissionRepository) FindByID(ctx context.Context, id string) (*entity.Permission, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	perm, exists := r.permissions[id]
	if !exists {
		return nil, appErrors.NewNotFoundError("permission not found")
	}
	return perm, nil
}

func (r *InMemoryPermissionRepository) FindByName(ctx context.Context, name string) (*entity.Permission, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	perm, exists := r.permissionsByName[strings.ToLower(name)]
	if !exists {
		return nil, appErrors.NewNotFoundError("permission not found")
	}
	return perm, nil
}

func (r *InMemoryPermissionRepository) List(ctx context.Context, p pagination.Pagination) ([]*entity.Permission, int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var all []*entity.Permission
	for _, perm := range r.permissions {
		all = append(all, perm)
	}

	total := int64(len(all))
	offset := p.Offset()
	limit := p.Limit()

	if offset >= len(all) {
		return []*entity.Permission{}, total, nil
	}

	end := offset + limit
	if end > len(all) {
		end = len(all)
	}

	return all[offset:end], total, nil
}

func (r *InMemoryPermissionRepository) Update(ctx context.Context, perm *entity.Permission) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, exists := r.permissions[perm.ID]
	if !exists {
		return appErrors.NewNotFoundError("permission not found")
	}

	cleanName := strings.ToLower(perm.Name)
	if other, nameTaken := r.permissionsByName[cleanName]; nameTaken && other.ID != perm.ID {
		return appErrors.NewConflictError(fmt.Sprintf("permission with name '%s' already exists", perm.Name))
	}

	delete(r.permissionsByName, strings.ToLower(existing.Name))
	r.permissions[perm.ID] = perm
	r.permissionsByName[cleanName] = perm
	return nil
}

func (r *InMemoryPermissionRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	perm, exists := r.permissions[id]
	if !exists {
		return appErrors.NewNotFoundError("permission not found")
	}

	delete(r.permissions, id)
	delete(r.permissionsByName, strings.ToLower(perm.Name))

	// Clean up role_permissions
	if r.roleRepo != nil {
		r.roleRepo.mu.Lock()
		for _, permSet := range r.roleRepo.rolePermissions {
			delete(permSet, id)
		}
		r.roleRepo.mu.Unlock()
	}

	return nil
}

func (r *InMemoryPermissionRepository) FindByRoleID(ctx context.Context, roleID string) ([]*entity.Permission, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.roleRepo == nil {
		return []*entity.Permission{}, nil
	}

	pids, err := r.roleRepo.GetRolePermissionIDs(ctx, roleID)
	if err != nil {
		return nil, err
	}

	var list []*entity.Permission
	for _, pid := range pids {
		if p, ok := r.permissions[pid]; ok {
			list = append(list, p)
		}
	}
	return list, nil
}

func (r *InMemoryPermissionRepository) FindEffectivePermissionsByUserID(ctx context.Context, userID string) ([]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.roleRepo == nil {
		return []string{}, nil
	}

	roles, err := r.roleRepo.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, err
	}

	permNamesMap := make(map[string]bool)
	for _, role := range roles {
		pids, err := r.roleRepo.GetRolePermissionIDs(ctx, role.ID)
		if err != nil {
			return nil, err
		}
		for _, pid := range pids {
			if perm, ok := r.permissions[pid]; ok {
				permNamesMap[perm.Name] = true
			}
		}
	}

	var result []string
	for name := range permNamesMap {
		result = append(result, name)
	}
	return result, nil
}

func (r *InMemoryPermissionRepository) GetAffectedUserIDsByPermissionID(ctx context.Context, permissionID string) ([]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.roleRepo == nil {
		return []string{}, nil
	}

	userMap := make(map[string]bool)
	r.roleRepo.mu.RLock()
	defer r.roleRepo.mu.RUnlock()

	for roleID, perms := range r.roleRepo.rolePermissions {
		if perms[permissionID] {
			for uid, roles := range r.roleRepo.userRoles {
				if roles[roleID] {
					userMap[uid] = true
				}
			}
		}
	}

	var userIDs []string
	for uid := range userMap {
		userIDs = append(userIDs, uid)
	}
	return userIDs, nil
}

var _ repository.RoleRepository = (*InMemoryRoleRepository)(nil)
var _ repository.PermissionRepository = (*InMemoryPermissionRepository)(nil)
