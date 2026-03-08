package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Permission name constants
const (
	PermDashboardView  = "dashboard:view"
	PermUsersView      = "users:view"
	PermUsersManage    = "users:manage"
	PermInventoryView  = "inventory:view"
	PermInventoryCreate = "inventory:create"
	PermInventoryUpdate = "inventory:update"
	PermInventoryDelete = "inventory:delete"
)

type Permission struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Name        string         `gorm:"uniqueIndex;size:100;not null" json:"name"`
	Description string         `gorm:"size:255" json:"description"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (p *Permission) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

type RolePermission struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	Role         Role       `gorm:"type:varchar(20);not null;uniqueIndex:idx_role_permission" json:"role"`
	PermissionID uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:idx_role_permission" json:"permission_id"`
	Permission   Permission `gorm:"foreignKey:PermissionID" json:"permission"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

func (rp *RolePermission) BeforeCreate(tx *gorm.DB) error {
	if rp.ID == uuid.Nil {
		rp.ID = uuid.New()
	}
	return nil
}
