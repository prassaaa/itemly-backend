package usecase

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
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
	ErrInvalidRefreshToken   = errors.New("invalid or expired refresh token")
)

type AuthUsecase interface {
	Register(username, email, password string) (*model.User, *jwtutil.TokenPair, error)
	Login(email, password string) (*model.User, *jwtutil.TokenPair, error)
	GetProfile(userID uuid.UUID) (*model.User, error)
	RefreshToken(refreshTokenStr string) (*jwtutil.TokenPair, error)
	Logout(accessJTI string, accessExpiresAt time.Time) error
}

type authUsecase struct {
	userRepo   repository.UserRepository
	jwtService *jwtutil.JWTService
	blacklist  *jwtutil.TokenBlacklist
}

func NewAuthUsecase(userRepo repository.UserRepository, jwtService *jwtutil.JWTService, blacklist *jwtutil.TokenBlacklist) AuthUsecase {
	return &authUsecase{
		userRepo:   userRepo,
		jwtService: jwtService,
		blacklist:  blacklist,
	}
}

func (uc *authUsecase) Register(username, email, password string) (*model.User, *jwtutil.TokenPair, error) {
	username = strings.TrimSpace(username)
	email = strings.ToLower(strings.TrimSpace(email))

	hashedPassword, err := hash.HashPassword(password)
	if err != nil {
		return nil, nil, err
	}

	user := &model.User{
		Username: username,
		Email:    email,
		Password: hashedPassword,
		Role:     model.RoleStaff,
	}

	if err := uc.userRepo.Create(user); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			switch pgErr.ConstraintName {
			case "uni_users_email", "idx_users_email":
				return nil, nil, ErrEmailAlreadyExists
			case "uni_users_username", "idx_users_username":
				return nil, nil, ErrUsernameAlreadyExists
			default:
				if strings.Contains(pgErr.ConstraintName, "email") {
					return nil, nil, ErrEmailAlreadyExists
				}
				return nil, nil, ErrUsernameAlreadyExists
			}
		}
		return nil, nil, err
	}

	tokenPair, err := uc.jwtService.GenerateTokenPair(user.ID, user.Username, string(user.Role))
	if err != nil {
		return nil, nil, err
	}

	return user, tokenPair, nil
}

func (uc *authUsecase) Login(email, password string) (*model.User, *jwtutil.TokenPair, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	user, err := uc.userRepo.FindByEmail(email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, ErrInvalidCredentials
		}
		return nil, nil, err
	}

	if !hash.CheckPassword(password, user.Password) {
		return nil, nil, ErrInvalidCredentials
	}

	tokenPair, err := uc.jwtService.GenerateTokenPair(user.ID, user.Username, string(user.Role))
	if err != nil {
		return nil, nil, err
	}

	return user, tokenPair, nil
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

func (uc *authUsecase) RefreshToken(refreshTokenStr string) (*jwtutil.TokenPair, error) {
	claims, err := uc.jwtService.ValidateToken(refreshTokenStr)
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}

	if claims.TokenType != jwtutil.RefreshToken {
		return nil, ErrInvalidRefreshToken
	}

	if uc.blacklist.IsBlacklisted(claims.ID) {
		return nil, ErrInvalidRefreshToken
	}

	// Blacklist old refresh token
	uc.blacklist.Add(claims.ID, claims.ExpiresAt.Time)

	// Verify user still exists
	user, err := uc.userRepo.FindByID(claims.UserID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	// Issue new token pair
	tokenPair, err := uc.jwtService.GenerateTokenPair(user.ID, user.Username, string(user.Role))
	if err != nil {
		return nil, err
	}

	return tokenPair, nil
}

func (uc *authUsecase) Logout(accessJTI string, accessExpiresAt time.Time) error {
	uc.blacklist.Add(accessJTI, accessExpiresAt)
	return nil
}
