package model

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUser_BeforeCreate_GeneratesID(t *testing.T) {
	u := &User{}
	err := u.BeforeCreate(nil)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, u.ID)
}

func TestUser_BeforeCreate_KeepsExistingID(t *testing.T) {
	existingID := uuid.New()
	u := &User{ID: existingID}
	err := u.BeforeCreate(nil)
	assert.NoError(t, err)
	assert.Equal(t, existingID, u.ID)
}

func TestUser_BeforeCreate_DefaultsToStaff(t *testing.T) {
	u := &User{}
	err := u.BeforeCreate(nil)
	assert.NoError(t, err)
	assert.Equal(t, RoleStaff, u.Role)
}

func TestUser_BeforeCreate_KeepsExistingRole(t *testing.T) {
	u := &User{Role: RoleAdmin}
	err := u.BeforeCreate(nil)
	assert.NoError(t, err)
	assert.Equal(t, RoleAdmin, u.Role)
}

func TestPermission_BeforeCreate_GeneratesID(t *testing.T) {
	p := &Permission{}
	err := p.BeforeCreate(nil)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, p.ID)
}

func TestPermission_BeforeCreate_KeepsExistingID(t *testing.T) {
	existingID := uuid.New()
	p := &Permission{ID: existingID}
	err := p.BeforeCreate(nil)
	assert.NoError(t, err)
	assert.Equal(t, existingID, p.ID)
}

func TestRolePermission_BeforeCreate_GeneratesID(t *testing.T) {
	rp := &RolePermission{}
	err := rp.BeforeCreate(nil)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, rp.ID)
}

func TestRolePermission_BeforeCreate_KeepsExistingID(t *testing.T) {
	existingID := uuid.New()
	rp := &RolePermission{ID: existingID}
	err := rp.BeforeCreate(nil)
	assert.NoError(t, err)
	assert.Equal(t, existingID, rp.ID)
}
