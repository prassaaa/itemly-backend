package repository

import (
	"github.com/prassaaa/itemly-backend/internal/model"
	"gorm.io/gorm"
)

type PermissionRepository interface {
	GetAllRolePermissions() ([]model.RolePermission, error)
	GetPermissionsByRole(role model.Role) ([]model.Permission, error)
}

type permissionRepository struct {
	db *gorm.DB
}

func NewPermissionRepository(db *gorm.DB) PermissionRepository {
	return &permissionRepository{db: db}
}

func (r *permissionRepository) GetAllRolePermissions() ([]model.RolePermission, error) {
	var rolePerms []model.RolePermission
	if err := r.db.Preload("Permission").Find(&rolePerms).Error; err != nil {
		return nil, err
	}
	return rolePerms, nil
}

func (r *permissionRepository) GetPermissionsByRole(role model.Role) ([]model.Permission, error) {
	var permissions []model.Permission
	if err := r.db.
		Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Where("role_permissions.role = ?", role).
		Find(&permissions).Error; err != nil {
		return nil, err
	}
	return permissions, nil
}
