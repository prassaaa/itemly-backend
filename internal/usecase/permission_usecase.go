package usecase

import (
	"sync"

	"github.com/prassaaa/itemly-backend/internal/repository"
)

type PermissionUsecase interface {
	HasPermission(role string, permissionName string) bool
	GetPermissionsByRole(role string) []string
	LoadPermissions() error
}

type permissionUsecase struct {
	permRepo repository.PermissionRepository
	mu       sync.RWMutex
	cache    map[string]map[string]bool // role -> permission -> allowed
}

func NewPermissionUsecase(permRepo repository.PermissionRepository) PermissionUsecase {
	return &permissionUsecase{
		permRepo: permRepo,
		cache:    make(map[string]map[string]bool),
	}
}

func (u *permissionUsecase) LoadPermissions() error {
	rolePerms, err := u.permRepo.GetAllRolePermissions()
	if err != nil {
		return err
	}

	newCache := make(map[string]map[string]bool)
	for _, rp := range rolePerms {
		role := string(rp.Role)
		if newCache[role] == nil {
			newCache[role] = make(map[string]bool)
		}
		newCache[role][rp.Permission.Name] = true
	}

	u.mu.Lock()
	u.cache = newCache
	u.mu.Unlock()
	return nil
}

func (u *permissionUsecase) HasPermission(role string, permissionName string) bool {
	u.mu.RLock()
	defer u.mu.RUnlock()

	perms, ok := u.cache[role]
	if !ok {
		return false
	}
	return perms[permissionName]
}

func (u *permissionUsecase) GetPermissionsByRole(role string) []string {
	u.mu.RLock()
	defer u.mu.RUnlock()

	perms, ok := u.cache[role]
	if !ok {
		return nil
	}

	result := make([]string, 0, len(perms))
	for p := range perms {
		result = append(result, p)
	}
	return result
}
