package usecase

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/prassaaa/itemly-backend/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// --- mock (reuses mockUserRepository fields from auth_usecase_test.go) ---

type mockAdminUserRepository struct {
	findByIDFn   func(id uuid.UUID) (*model.User, error)
	updateRoleFn func(id uuid.UUID, role model.Role) error
}

func (m *mockAdminUserRepository) Create(user *model.User) error          { return nil }
func (m *mockAdminUserRepository) FindByEmail(email string) (*model.User, error) {
	return nil, gorm.ErrRecordNotFound
}
func (m *mockAdminUserRepository) FindByUsername(username string) (*model.User, error) {
	return nil, gorm.ErrRecordNotFound
}
func (m *mockAdminUserRepository) FindByID(id uuid.UUID) (*model.User, error) {
	return m.findByIDFn(id)
}
func (m *mockAdminUserRepository) UpdateRole(id uuid.UUID, role model.Role) error {
	return m.updateRoleFn(id, role)
}

func TestAssignRole_Success(t *testing.T) {
	userID := uuid.New()
	repo := &mockAdminUserRepository{
		findByIDFn: func(id uuid.UUID) (*model.User, error) {
			return &model.User{ID: userID, Username: "target", Role: model.RoleStaff}, nil
		},
		updateRoleFn: func(id uuid.UUID, role model.Role) error {
			return nil
		},
	}
	uc := NewAdminUsecase(repo)

	user, err := uc.AssignRole(userID, "manager")
	require.NoError(t, err)
	assert.Equal(t, model.RoleManager, user.Role)
}

func TestAssignRole_InvalidRole(t *testing.T) {
	repo := &mockAdminUserRepository{}
	uc := NewAdminUsecase(repo)

	_, err := uc.AssignRole(uuid.New(), "superadmin")
	assert.ErrorIs(t, err, ErrInvalidRole)
}

func TestAssignRole_UserNotFound(t *testing.T) {
	repo := &mockAdminUserRepository{
		findByIDFn: func(id uuid.UUID) (*model.User, error) {
			return nil, gorm.ErrRecordNotFound
		},
	}
	uc := NewAdminUsecase(repo)

	_, err := uc.AssignRole(uuid.New(), "admin")
	assert.ErrorIs(t, err, ErrUserNotFound)
}

func TestAssignRole_UpdateFails(t *testing.T) {
	dbErr := errors.New("update failed")
	repo := &mockAdminUserRepository{
		findByIDFn: func(id uuid.UUID) (*model.User, error) {
			return &model.User{ID: uuid.New(), Role: model.RoleStaff}, nil
		},
		updateRoleFn: func(id uuid.UUID, role model.Role) error {
			return dbErr
		},
	}
	uc := NewAdminUsecase(repo)

	_, err := uc.AssignRole(uuid.New(), "admin")
	assert.ErrorIs(t, err, dbErr)
}
