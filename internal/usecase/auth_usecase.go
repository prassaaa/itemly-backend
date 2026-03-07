package usecase

import (
	"errors"

	"github.com/google/uuid"
	"github.com/prassaaa/itemly-backend/internal/model"
	"github.com/prassaaa/itemly-backend/internal/repository"
	"github.com/prassaaa/itemly-backend/pkg/hash"
	jwtutil "github.com/prassaaa/itemly-backend/pkg/jwt"
	"gorm.io/gorm"
)

var (
	ErrEmailAlreadyExists    = errors.New("email already registered")
	ErrUsernameAlreadyExists = errors.New("username already taken")
	ErrInvalidCredentials    = errors.New("invalid email or password")
	ErrUserNotFound          = errors.New("user not found")
)

type AuthUsecase interface {
	Register(username, email, password string) (*model.User, string, error)
	Login(email, password string) (*model.User, string, error)
	GetProfile(userID uuid.UUID) (*model.User, error)
}

type authUsecase struct {
	userRepo   repository.UserRepository
	jwtService *jwtutil.JWTService
}

func NewAuthUsecase(userRepo repository.UserRepository, jwtService *jwtutil.JWTService) AuthUsecase {
	return &authUsecase{
		userRepo:   userRepo,
		jwtService: jwtService,
	}
}

func (uc *authUsecase) Register(username, email, password string) (*model.User, string, error) {
	if _, err := uc.userRepo.FindByEmail(email); err == nil {
		return nil, "", ErrEmailAlreadyExists
	}
	if _, err := uc.userRepo.FindByUsername(username); err == nil {
		return nil, "", ErrUsernameAlreadyExists
	}

	hashedPassword, err := hash.HashPassword(password)
	if err != nil {
		return nil, "", err
	}

	user := &model.User{
		Username: username,
		Email:    email,
		Password: hashedPassword,
		Role:     model.RoleStaff,
	}

	if err := uc.userRepo.Create(user); err != nil {
		return nil, "", err
	}

	token, err := uc.jwtService.GenerateToken(user.ID, user.Username, string(user.Role))
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

func (uc *authUsecase) Login(email, password string) (*model.User, string, error) {
	user, err := uc.userRepo.FindByEmail(email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", ErrInvalidCredentials
		}
		return nil, "", err
	}

	if !hash.CheckPassword(password, user.Password) {
		return nil, "", ErrInvalidCredentials
	}

	token, err := uc.jwtService.GenerateToken(user.ID, user.Username, string(user.Role))
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

func (uc *authUsecase) GetProfile(userID uuid.UUID) (*model.User, error) {
	user, err := uc.userRepo.FindByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}
