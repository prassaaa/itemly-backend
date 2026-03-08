package usecase

import (
	"errors"

	"github.com/google/uuid"
	"github.com/prassaaa/itemly-backend/internal/model"
	"github.com/prassaaa/itemly-backend/internal/repository"
	"gorm.io/gorm"
)

var (
	ErrInvalidRole = errors.New("invalid role")
)

type AdminUsecase interface {
	AssignRole(targetUserID uuid.UUID, role string) (*model.User, error)
}

type adminUsecase struct {
	userRepo repository.UserRepository
}

func NewAdminUsecase(userRepo repository.UserRepository) AdminUsecase {
	return &adminUsecase{userRepo: userRepo}
}

func (uc *adminUsecase) AssignRole(targetUserID uuid.UUID, role string) (*model.User, error) {
	newRole := model.Role(role)
	if newRole != model.RoleAdmin && newRole != model.RoleManager && newRole != model.RoleStaff {
		return nil, ErrInvalidRole
	}

	user, err := uc.userRepo.FindByID(targetUserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	if err := uc.userRepo.UpdateRole(targetUserID, newRole); err != nil {
		return nil, err
	}

	user.Role = newRole
	return user, nil
}
