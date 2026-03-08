package usecase

import (
	"errors"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/prassaaa/itemly-backend/internal/model"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- mock ---

type mockPermissionRepository struct {
	getAllRolePermissionsFn func() ([]model.RolePermission, error)
	getPermissionsByRoleFn func(role model.Role) ([]model.Permission, error)
}

func (m *mockPermissionRepository) GetAllRolePermissions() ([]model.RolePermission, error) {
	return m.getAllRolePermissionsFn()
}
func (m *mockPermissionRepository) GetPermissionsByRole(role model.Role) ([]model.Permission, error) {
	return m.getPermissionsByRoleFn(role)
}

// --- helpers ---

func setupRedisPermission(t *testing.T, repo *mockPermissionRepository) (PermissionUsecase, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	uc := NewRedisPermissionUsecase(repo, rdb)
	return uc, mr
}

func sampleRolePermissions() []model.RolePermission {
	return []model.RolePermission{
		{
			ID:   uuid.New(),
			Role: model.RoleAdmin,
			Permission: model.Permission{
				ID:   uuid.New(),
				Name: model.PermDashboardView,
			},
		},
		{
			ID:   uuid.New(),
			Role: model.RoleAdmin,
			Permission: model.Permission{
				ID:   uuid.New(),
				Name: model.PermUsersManage,
			},
		},
		{
			ID:   uuid.New(),
			Role: model.RoleStaff,
			Permission: model.Permission{
				ID:   uuid.New(),
				Name: model.PermDashboardView,
			},
		},
	}
}

func TestRedisPermission_LoadAndHasPermission(t *testing.T) {
	repo := &mockPermissionRepository{
		getAllRolePermissionsFn: func() ([]model.RolePermission, error) {
			return sampleRolePermissions(), nil
		},
	}
	uc, _ := setupRedisPermission(t, repo)

	err := uc.LoadPermissions()
	require.NoError(t, err)

	assert.True(t, uc.HasPermission("admin", model.PermDashboardView))
	assert.True(t, uc.HasPermission("admin", model.PermUsersManage))
	assert.True(t, uc.HasPermission("staff", model.PermDashboardView))
	assert.False(t, uc.HasPermission("staff", model.PermUsersManage))
	assert.False(t, uc.HasPermission("manager", model.PermDashboardView))
}

func TestRedisPermission_GetPermissionsByRole(t *testing.T) {
	repo := &mockPermissionRepository{
		getAllRolePermissionsFn: func() ([]model.RolePermission, error) {
			return sampleRolePermissions(), nil
		},
	}
	uc, _ := setupRedisPermission(t, repo)

	err := uc.LoadPermissions()
	require.NoError(t, err)

	adminPerms := uc.GetPermissionsByRole("admin")
	assert.Len(t, adminPerms, 2)
	assert.Contains(t, adminPerms, model.PermDashboardView)
	assert.Contains(t, adminPerms, model.PermUsersManage)

	staffPerms := uc.GetPermissionsByRole("staff")
	assert.Len(t, staffPerms, 1)
	assert.Contains(t, staffPerms, model.PermDashboardView)
}

func TestRedisPermission_LoadClearsOldKeys(t *testing.T) {
	callCount := 0
	repo := &mockPermissionRepository{
		getAllRolePermissionsFn: func() ([]model.RolePermission, error) {
			callCount++
			if callCount == 1 {
				return sampleRolePermissions(), nil
			}
			// Second load: only staff has dashboard:view
			return []model.RolePermission{
				{
					ID:   uuid.New(),
					Role: model.RoleStaff,
					Permission: model.Permission{
						ID:   uuid.New(),
						Name: model.PermDashboardView,
					},
				},
			}, nil
		},
	}
	uc, _ := setupRedisPermission(t, repo)

	require.NoError(t, uc.LoadPermissions())
	assert.True(t, uc.HasPermission("admin", model.PermDashboardView))

	// Reload with reduced permissions
	require.NoError(t, uc.LoadPermissions())
	assert.False(t, uc.HasPermission("admin", model.PermDashboardView))
	assert.True(t, uc.HasPermission("staff", model.PermDashboardView))
}

func TestRedisPermission_LoadRepoError(t *testing.T) {
	repoErr := errors.New("db down")
	repo := &mockPermissionRepository{
		getAllRolePermissionsFn: func() ([]model.RolePermission, error) {
			return nil, repoErr
		},
	}
	uc, _ := setupRedisPermission(t, repo)

	err := uc.LoadPermissions()
	assert.ErrorIs(t, err, repoErr)
}
