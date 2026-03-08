package usecase

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/prassaaa/itemly-backend/internal/model"
	"github.com/prassaaa/itemly-backend/internal/testutil"
	"github.com/prassaaa/itemly-backend/pkg/hash"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// --- mocks ---

type mockUserRepository struct {
	createFn       func(user *model.User) error
	findByEmailFn  func(email string) (*model.User, error)
	findByIDFn     func(id uuid.UUID) (*model.User, error)
	updateRoleFn   func(id uuid.UUID, role model.Role) error
}

func (m *mockUserRepository) Create(user *model.User) error {
	return m.createFn(user)
}
func (m *mockUserRepository) FindByEmail(email string) (*model.User, error) {
	return m.findByEmailFn(email)
}
func (m *mockUserRepository) FindByUsername(username string) (*model.User, error) {
	return nil, gorm.ErrRecordNotFound
}
func (m *mockUserRepository) FindByID(id uuid.UUID) (*model.User, error) {
	return m.findByIDFn(id)
}
func (m *mockUserRepository) UpdateRole(id uuid.UUID, role model.Role) error {
	return m.updateRoleFn(id, role)
}

type mockTokenBlacklist struct {
	addFn           func(jti string, expiresAt time.Time)
	isBlacklistedFn func(jti string) bool
}

func (m *mockTokenBlacklist) Add(jti string, expiresAt time.Time) {
	m.addFn(jti, expiresAt)
}
func (m *mockTokenBlacklist) IsBlacklisted(jti string) bool {
	return m.isBlacklistedFn(jti)
}

// --- helpers ---

func newTestUser() *model.User {
	hashed, _ := hash.HashPassword("Test@1234")
	return &model.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "test@example.com",
		Password: hashed,
		Role:     model.RoleStaff,
	}
}

func newAuthUsecase(repo *mockUserRepository, bl *mockTokenBlacklist) AuthUsecase {
	return NewAuthUsecase(repo, testutil.TestJWTService(), bl)
}

// --- Register tests ---

func TestRegister_Success(t *testing.T) {
	repo := &mockUserRepository{
		createFn: func(user *model.User) error {
			user.ID = uuid.New()
			return nil
		},
	}
	bl := &mockTokenBlacklist{}
	uc := newAuthUsecase(repo, bl)

	user, tokens, err := uc.Register("TestUser", " Test@Example.COM ", "Test@1234")
	require.NoError(t, err)
	assert.NotNil(t, tokens)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, model.RoleStaff, user.Role)
	assert.NotEmpty(t, tokens.AccessToken)
	assert.NotEmpty(t, tokens.RefreshToken)
}

func TestRegister_DuplicateEmail(t *testing.T) {
	repo := &mockUserRepository{
		createFn: func(user *model.User) error {
			return &pgconn.PgError{Code: "23505", ConstraintName: "uni_users_email"}
		},
	}
	bl := &mockTokenBlacklist{}
	uc := newAuthUsecase(repo, bl)

	_, _, err := uc.Register("newuser", "dup@example.com", "Test@1234")
	assert.ErrorIs(t, err, ErrEmailAlreadyExists)
}

func TestRegister_DuplicateUsername(t *testing.T) {
	repo := &mockUserRepository{
		createFn: func(user *model.User) error {
			return &pgconn.PgError{Code: "23505", ConstraintName: "uni_users_username"}
		},
	}
	bl := &mockTokenBlacklist{}
	uc := newAuthUsecase(repo, bl)

	_, _, err := uc.Register("dupuser", "new@example.com", "Test@1234")
	assert.ErrorIs(t, err, ErrUsernameAlreadyExists)
}

func TestRegister_DuplicateEmail_FallbackConstraint(t *testing.T) {
	repo := &mockUserRepository{
		createFn: func(user *model.User) error {
			return &pgconn.PgError{Code: "23505", ConstraintName: "some_email_constraint"}
		},
	}
	bl := &mockTokenBlacklist{}
	uc := newAuthUsecase(repo, bl)

	_, _, err := uc.Register("user", "dup@example.com", "Test@1234")
	assert.ErrorIs(t, err, ErrEmailAlreadyExists)
}

func TestRegister_DuplicateUsername_FallbackConstraint(t *testing.T) {
	repo := &mockUserRepository{
		createFn: func(user *model.User) error {
			return &pgconn.PgError{Code: "23505", ConstraintName: "some_other_constraint"}
		},
	}
	bl := &mockTokenBlacklist{}
	uc := newAuthUsecase(repo, bl)

	_, _, err := uc.Register("dupuser", "new@example.com", "Test@1234")
	assert.ErrorIs(t, err, ErrUsernameAlreadyExists)
}

// --- Login tests ---

func TestLogin_Success(t *testing.T) {
	testUser := newTestUser()
	repo := &mockUserRepository{
		findByEmailFn: func(email string) (*model.User, error) {
			return testUser, nil
		},
	}
	bl := &mockTokenBlacklist{}
	uc := newAuthUsecase(repo, bl)

	user, tokens, err := uc.Login("test@example.com", "Test@1234")
	require.NoError(t, err)
	assert.Equal(t, testUser.ID, user.ID)
	assert.NotEmpty(t, tokens.AccessToken)
}

func TestLogin_UserNotFound(t *testing.T) {
	repo := &mockUserRepository{
		findByEmailFn: func(email string) (*model.User, error) {
			return nil, gorm.ErrRecordNotFound
		},
	}
	bl := &mockTokenBlacklist{}
	uc := newAuthUsecase(repo, bl)

	_, _, err := uc.Login("noone@example.com", "Test@1234")
	assert.ErrorIs(t, err, ErrInvalidCredentials)
}

func TestLogin_WrongPassword(t *testing.T) {
	testUser := newTestUser()
	repo := &mockUserRepository{
		findByEmailFn: func(email string) (*model.User, error) {
			return testUser, nil
		},
	}
	bl := &mockTokenBlacklist{}
	uc := newAuthUsecase(repo, bl)

	_, _, err := uc.Login("test@example.com", "WrongPass@1")
	assert.ErrorIs(t, err, ErrInvalidCredentials)
}

func TestLogin_RepoError(t *testing.T) {
	repoErr := errors.New("db connection lost")
	repo := &mockUserRepository{
		findByEmailFn: func(email string) (*model.User, error) {
			return nil, repoErr
		},
	}
	bl := &mockTokenBlacklist{}
	uc := newAuthUsecase(repo, bl)

	_, _, err := uc.Login("test@example.com", "Test@1234")
	assert.ErrorIs(t, err, repoErr)
}

// --- GetProfile tests ---

func TestGetProfile_Success(t *testing.T) {
	testUser := newTestUser()
	repo := &mockUserRepository{
		findByIDFn: func(id uuid.UUID) (*model.User, error) {
			return testUser, nil
		},
	}
	bl := &mockTokenBlacklist{}
	uc := newAuthUsecase(repo, bl)

	user, err := uc.GetProfile(testUser.ID)
	require.NoError(t, err)
	assert.Equal(t, testUser.Username, user.Username)
}

func TestGetProfile_NotFound(t *testing.T) {
	repo := &mockUserRepository{
		findByIDFn: func(id uuid.UUID) (*model.User, error) {
			return nil, gorm.ErrRecordNotFound
		},
	}
	bl := &mockTokenBlacklist{}
	uc := newAuthUsecase(repo, bl)

	_, err := uc.GetProfile(uuid.New())
	assert.ErrorIs(t, err, ErrUserNotFound)
}

func TestGetProfile_RepoError(t *testing.T) {
	repoErr := errors.New("db timeout")
	repo := &mockUserRepository{
		findByIDFn: func(id uuid.UUID) (*model.User, error) {
			return nil, repoErr
		},
	}
	bl := &mockTokenBlacklist{}
	uc := newAuthUsecase(repo, bl)

	_, err := uc.GetProfile(uuid.New())
	assert.ErrorIs(t, err, repoErr)
}

// --- RefreshToken tests ---

func TestRefreshToken_Success(t *testing.T) {
	testUser := newTestUser()
	var blacklistedJTI string

	repo := &mockUserRepository{
		findByIDFn: func(id uuid.UUID) (*model.User, error) {
			return testUser, nil
		},
	}
	bl := &mockTokenBlacklist{
		isBlacklistedFn: func(jti string) bool { return false },
		addFn:           func(jti string, exp time.Time) { blacklistedJTI = jti },
	}
	uc := newAuthUsecase(repo, bl)

	refreshToken := testutil.TestRefreshToken(testUser.ID, testUser.Username, string(testUser.Role))

	tokens, err := uc.RefreshToken(refreshToken)
	require.NoError(t, err)
	assert.NotEmpty(t, tokens.AccessToken)
	assert.NotEmpty(t, tokens.RefreshToken)
	assert.NotEmpty(t, blacklistedJTI)
}

func TestRefreshToken_InvalidToken(t *testing.T) {
	repo := &mockUserRepository{}
	bl := &mockTokenBlacklist{}
	uc := newAuthUsecase(repo, bl)

	_, err := uc.RefreshToken("garbage-token")
	assert.ErrorIs(t, err, ErrInvalidRefreshToken)
}

func TestRefreshToken_AccessTokenRejected(t *testing.T) {
	testUser := newTestUser()
	repo := &mockUserRepository{}
	bl := &mockTokenBlacklist{
		isBlacklistedFn: func(jti string) bool { return false },
	}
	uc := newAuthUsecase(repo, bl)

	accessToken := testutil.TestAccessToken(testUser.ID, testUser.Username, string(testUser.Role))

	_, err := uc.RefreshToken(accessToken)
	assert.ErrorIs(t, err, ErrInvalidRefreshToken)
}

func TestRefreshToken_BlacklistedToken(t *testing.T) {
	testUser := newTestUser()
	repo := &mockUserRepository{}
	bl := &mockTokenBlacklist{
		isBlacklistedFn: func(jti string) bool { return true },
	}
	uc := newAuthUsecase(repo, bl)

	refreshToken := testutil.TestRefreshToken(testUser.ID, testUser.Username, string(testUser.Role))

	_, err := uc.RefreshToken(refreshToken)
	assert.ErrorIs(t, err, ErrInvalidRefreshToken)
}

func TestRefreshToken_UserDeleted(t *testing.T) {
	testUser := newTestUser()
	repo := &mockUserRepository{
		findByIDFn: func(id uuid.UUID) (*model.User, error) {
			return nil, gorm.ErrRecordNotFound
		},
	}
	bl := &mockTokenBlacklist{
		isBlacklistedFn: func(jti string) bool { return false },
		addFn:           func(jti string, exp time.Time) {},
	}
	uc := newAuthUsecase(repo, bl)

	refreshToken := testutil.TestRefreshToken(testUser.ID, testUser.Username, string(testUser.Role))

	_, err := uc.RefreshToken(refreshToken)
	assert.ErrorIs(t, err, ErrUserNotFound)
}

// --- Logout tests ---

func TestLogout_Success(t *testing.T) {
	var addCalled bool
	repo := &mockUserRepository{}
	bl := &mockTokenBlacklist{
		addFn: func(jti string, exp time.Time) { addCalled = true },
	}
	uc := newAuthUsecase(repo, bl)

	err := uc.Logout("some-jti", time.Now().Add(time.Hour))
	require.NoError(t, err)
	assert.True(t, addCalled)
}

// --- Verify generic repo error propagation ---

func TestRegister_RepoError(t *testing.T) {
	repoErr := errors.New("db connection lost")
	repo := &mockUserRepository{
		createFn: func(user *model.User) error { return repoErr },
	}
	bl := &mockTokenBlacklist{}
	uc := newAuthUsecase(repo, bl)

	_, _, err := uc.Register("user", "a@b.com", "Test@1234")
	assert.ErrorIs(t, err, repoErr)
}
