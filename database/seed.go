package database

import (
	"log/slog"

	"github.com/prassaaa/itemly-backend/internal/model"
	"gorm.io/gorm"
)

type permissionSeed struct {
	Name        string
	Description string
}

type rolePermissionSeed struct {
	Role       model.Role
	Permission string
}

func SeedPermissions(db *gorm.DB) error {
	permissions := []permissionSeed{
		{model.PermDashboardView, "View dashboard"},
		{model.PermUsersView, "View users list"},
		{model.PermUsersManage, "Manage users (assign roles)"},
		{model.PermInventoryView, "View inventory items"},
		{model.PermInventoryCreate, "Create inventory items"},
		{model.PermInventoryUpdate, "Update inventory items"},
		{model.PermInventoryDelete, "Delete inventory items"},
	}

	for _, p := range permissions {
		perm := model.Permission{Name: p.Name, Description: p.Description}
		if err := db.Unscoped().Where("name = ?", p.Name).FirstOrCreate(&perm).Error; err != nil {
			return err
		}
	}

	rolePerms := []rolePermissionSeed{
		// admin — all permissions
		{model.RoleAdmin, model.PermDashboardView},
		{model.RoleAdmin, model.PermUsersView},
		{model.RoleAdmin, model.PermUsersManage},
		{model.RoleAdmin, model.PermInventoryView},
		{model.RoleAdmin, model.PermInventoryCreate},
		{model.RoleAdmin, model.PermInventoryUpdate},
		{model.RoleAdmin, model.PermInventoryDelete},
		// manager
		{model.RoleManager, model.PermDashboardView},
		{model.RoleManager, model.PermUsersView},
		{model.RoleManager, model.PermInventoryView},
		{model.RoleManager, model.PermInventoryCreate},
		{model.RoleManager, model.PermInventoryUpdate},
		{model.RoleManager, model.PermInventoryDelete},
		// staff
		{model.RoleStaff, model.PermDashboardView},
		{model.RoleStaff, model.PermInventoryView},
	}

	for _, rp := range rolePerms {
		var perm model.Permission
		if err := db.Where("name = ?", rp.Permission).First(&perm).Error; err != nil {
			return err
		}
		rolePermission := model.RolePermission{Role: rp.Role, PermissionID: perm.ID}
		if err := db.Where("role = ? AND permission_id = ?", rp.Role, perm.ID).FirstOrCreate(&rolePermission).Error; err != nil {
			return err
		}
	}

	slog.Info("permissions seeded successfully")
	return nil
}
