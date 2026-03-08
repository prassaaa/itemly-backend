package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/prassaaa/itemly-backend/internal/model"
	"github.com/prassaaa/itemly-backend/internal/usecase"
	jwtutil "github.com/prassaaa/itemly-backend/pkg/jwt"
	pwdvalidator "github.com/prassaaa/itemly-backend/pkg/validator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		_ = v.RegisterValidation("password", pwdvalidator.PasswordStrength)
	}
}

// --- mock ---

type mockAuthUsecase struct {
	registerFn     func(username, email, password string) (*model.User, *jwtutil.TokenPair, error)
	loginFn        func(email, password string) (*model.User, *jwtutil.TokenPair, error)
	getProfileFn   func(userID uuid.UUID) (*model.User, error)
	refreshTokenFn func(refreshTokenStr string) (*jwtutil.TokenPair, error)
	logoutFn       func(accessJTI string, accessExpiresAt time.Time) error
}

func (m *mockAuthUsecase) Register(username, email, password string) (*model.User, *jwtutil.TokenPair, error) {
	return m.registerFn(username, email, password)
}
func (m *mockAuthUsecase) Login(email, password string) (*model.User, *jwtutil.TokenPair, error) {
	return m.loginFn(email, password)
}
func (m *mockAuthUsecase) GetProfile(userID uuid.UUID) (*model.User, error) {
	return m.getProfileFn(userID)
}
func (m *mockAuthUsecase) RefreshToken(refreshTokenStr string) (*jwtutil.TokenPair, error) {
	return m.refreshTokenFn(refreshTokenStr)
}
func (m *mockAuthUsecase) Logout(accessJTI string, accessExpiresAt time.Time) error {
	return m.logoutFn(accessJTI, accessExpiresAt)
}

// --- helpers ---

func newTestRouter(h *AuthHandler) *gin.Engine {
	r := gin.New()
	return r
}

func postJSON(router *gin.Engine, path string, body interface{}) *httptest.ResponseRecorder {
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func getWithContext(router *gin.Engine, path string, ctxSetup func(c *gin.Context)) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func testUser() *model.User {
	return &model.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "test@example.com",
		Role:     model.RoleStaff,
	}
}

func testTokenPair() *jwtutil.TokenPair {
	return &jwtutil.TokenPair{
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
	}
}

// --- Register tests ---

func TestRegister_201(t *testing.T) {
	user := testUser()
	uc := &mockAuthUsecase{
		registerFn: func(username, email, password string) (*model.User, *jwtutil.TokenPair, error) {
			return user, testTokenPair(), nil
		},
	}
	h := NewAuthHandler(uc)
	r := gin.New()
	r.POST("/register", h.Register)

	w := postJSON(r, "/register", map[string]string{
		"username": "testuser",
		"email":    "test@example.com",
		"password": "Test@1234",
	})

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "access-token", resp["access_token"])
	assert.NotNil(t, resp["user"])
}

func TestRegister_400_MissingFields(t *testing.T) {
	uc := &mockAuthUsecase{}
	h := NewAuthHandler(uc)
	r := gin.New()
	r.POST("/register", h.Register)

	w := postJSON(r, "/register", map[string]string{})

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRegister_400_WeakPassword(t *testing.T) {
	uc := &mockAuthUsecase{}
	h := NewAuthHandler(uc)
	r := gin.New()
	r.POST("/register", h.Register)

	w := postJSON(r, "/register", map[string]string{
		"username": "testuser",
		"email":    "test@example.com",
		"password": "abc",
	})

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRegister_409_DuplicateEmail(t *testing.T) {
	uc := &mockAuthUsecase{
		registerFn: func(username, email, password string) (*model.User, *jwtutil.TokenPair, error) {
			return nil, nil, usecase.ErrEmailAlreadyExists
		},
	}
	h := NewAuthHandler(uc)
	r := gin.New()
	r.POST("/register", h.Register)

	w := postJSON(r, "/register", map[string]string{
		"username": "testuser",
		"email":    "dup@example.com",
		"password": "Test@1234",
	})

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestRegister_409_DuplicateUsername(t *testing.T) {
	uc := &mockAuthUsecase{
		registerFn: func(username, email, password string) (*model.User, *jwtutil.TokenPair, error) {
			return nil, nil, usecase.ErrUsernameAlreadyExists
		},
	}
	h := NewAuthHandler(uc)
	r := gin.New()
	r.POST("/register", h.Register)

	w := postJSON(r, "/register", map[string]string{
		"username": "dupuser",
		"email":    "new@example.com",
		"password": "Test@1234",
	})

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestRegister_500(t *testing.T) {
	uc := &mockAuthUsecase{
		registerFn: func(username, email, password string) (*model.User, *jwtutil.TokenPair, error) {
			return nil, nil, errors.New("unexpected")
		},
	}
	h := NewAuthHandler(uc)
	r := gin.New()
	r.POST("/register", h.Register)

	w := postJSON(r, "/register", map[string]string{
		"username": "testuser",
		"email":    "test@example.com",
		"password": "Test@1234",
	})

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// --- Login tests ---

func TestLogin_200(t *testing.T) {
	user := testUser()
	uc := &mockAuthUsecase{
		loginFn: func(email, password string) (*model.User, *jwtutil.TokenPair, error) {
			return user, testTokenPair(), nil
		},
	}
	h := NewAuthHandler(uc)
	r := gin.New()
	r.POST("/login", h.Login)

	w := postJSON(r, "/login", map[string]string{
		"email":    "test@example.com",
		"password": "Test@1234",
	})

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "access-token", resp["access_token"])
}

func TestLogin_400(t *testing.T) {
	uc := &mockAuthUsecase{}
	h := NewAuthHandler(uc)
	r := gin.New()
	r.POST("/login", h.Login)

	w := postJSON(r, "/login", map[string]string{})

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLogin_401(t *testing.T) {
	uc := &mockAuthUsecase{
		loginFn: func(email, password string) (*model.User, *jwtutil.TokenPair, error) {
			return nil, nil, usecase.ErrInvalidCredentials
		},
	}
	h := NewAuthHandler(uc)
	r := gin.New()
	r.POST("/login", h.Login)

	w := postJSON(r, "/login", map[string]string{
		"email":    "test@example.com",
		"password": "Wrong@1234",
	})

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestLogin_500(t *testing.T) {
	uc := &mockAuthUsecase{
		loginFn: func(email, password string) (*model.User, *jwtutil.TokenPair, error) {
			return nil, nil, errors.New("db error")
		},
	}
	h := NewAuthHandler(uc)
	r := gin.New()
	r.POST("/login", h.Login)

	w := postJSON(r, "/login", map[string]string{
		"email":    "test@example.com",
		"password": "Test@1234",
	})

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// --- Refresh tests ---

func TestRefresh_200(t *testing.T) {
	uc := &mockAuthUsecase{
		refreshTokenFn: func(refreshTokenStr string) (*jwtutil.TokenPair, error) {
			return testTokenPair(), nil
		},
	}
	h := NewAuthHandler(uc)
	r := gin.New()
	r.POST("/refresh", h.Refresh)

	w := postJSON(r, "/refresh", map[string]string{
		"refresh_token": "valid-refresh-token",
	})

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRefresh_400(t *testing.T) {
	uc := &mockAuthUsecase{}
	h := NewAuthHandler(uc)
	r := gin.New()
	r.POST("/refresh", h.Refresh)

	w := postJSON(r, "/refresh", map[string]string{})

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRefresh_401(t *testing.T) {
	uc := &mockAuthUsecase{
		refreshTokenFn: func(refreshTokenStr string) (*jwtutil.TokenPair, error) {
			return nil, usecase.ErrInvalidRefreshToken
		},
	}
	h := NewAuthHandler(uc)
	r := gin.New()
	r.POST("/refresh", h.Refresh)

	w := postJSON(r, "/refresh", map[string]string{
		"refresh_token": "expired-token",
	})

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRefresh_401_UserNotFound(t *testing.T) {
	uc := &mockAuthUsecase{
		refreshTokenFn: func(refreshTokenStr string) (*jwtutil.TokenPair, error) {
			return nil, usecase.ErrUserNotFound
		},
	}
	h := NewAuthHandler(uc)
	r := gin.New()
	r.POST("/refresh", h.Refresh)

	w := postJSON(r, "/refresh", map[string]string{
		"refresh_token": "valid-token",
	})

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRefresh_500(t *testing.T) {
	uc := &mockAuthUsecase{
		refreshTokenFn: func(refreshTokenStr string) (*jwtutil.TokenPair, error) {
			return nil, errors.New("unexpected")
		},
	}
	h := NewAuthHandler(uc)
	r := gin.New()
	r.POST("/refresh", h.Refresh)

	w := postJSON(r, "/refresh", map[string]string{
		"refresh_token": "valid-token",
	})

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// --- Logout tests ---

func TestLogout_200(t *testing.T) {
	uc := &mockAuthUsecase{
		logoutFn: func(accessJTI string, accessExpiresAt time.Time) error {
			return nil
		},
	}
	h := NewAuthHandler(uc)
	r := gin.New()
	r.POST("/logout", func(c *gin.Context) {
		c.Set("jti", "test-jti")
		c.Set("tokenExpiresAt", time.Now().Add(time.Hour))
		c.Next()
	}, h.Logout)

	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLogout_401_NoExpiresAt(t *testing.T) {
	uc := &mockAuthUsecase{}
	h := NewAuthHandler(uc)
	r := gin.New()
	r.POST("/logout", func(c *gin.Context) {
		c.Set("jti", "test-jti")
		// no tokenExpiresAt
		c.Next()
	}, h.Logout)

	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestLogout_500(t *testing.T) {
	uc := &mockAuthUsecase{
		logoutFn: func(accessJTI string, accessExpiresAt time.Time) error {
			return errors.New("blacklist error")
		},
	}
	h := NewAuthHandler(uc)
	r := gin.New()
	r.POST("/logout", func(c *gin.Context) {
		c.Set("jti", "test-jti")
		c.Set("tokenExpiresAt", time.Now().Add(time.Hour))
		c.Next()
	}, h.Logout)

	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestLogout_401_NoJTI(t *testing.T) {
	uc := &mockAuthUsecase{}
	h := NewAuthHandler(uc)
	r := gin.New()
	r.POST("/logout", h.Logout)

	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// --- GetProfile tests ---

func TestGetProfile_200(t *testing.T) {
	user := testUser()
	uc := &mockAuthUsecase{
		getProfileFn: func(userID uuid.UUID) (*model.User, error) {
			return user, nil
		},
	}
	h := NewAuthHandler(uc)
	r := gin.New()
	r.GET("/profile", func(c *gin.Context) {
		c.Set("userID", user.ID)
		c.Next()
	}, h.GetProfile)

	req := httptest.NewRequest(http.MethodGet, "/profile", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, user.Username, resp["username"])
}

func TestGetProfile_401_NoUserID(t *testing.T) {
	uc := &mockAuthUsecase{}
	h := NewAuthHandler(uc)
	r := gin.New()
	r.GET("/profile", h.GetProfile)

	req := httptest.NewRequest(http.MethodGet, "/profile", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetProfile_500_InvalidUserIDType(t *testing.T) {
	uc := &mockAuthUsecase{}
	h := NewAuthHandler(uc)
	r := gin.New()
	r.GET("/profile", func(c *gin.Context) {
		c.Set("userID", "not-a-uuid-type")
		c.Next()
	}, h.GetProfile)

	req := httptest.NewRequest(http.MethodGet, "/profile", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetProfile_500_InternalError(t *testing.T) {
	uc := &mockAuthUsecase{
		getProfileFn: func(userID uuid.UUID) (*model.User, error) {
			return nil, errors.New("db error")
		},
	}
	h := NewAuthHandler(uc)
	r := gin.New()
	r.GET("/profile", func(c *gin.Context) {
		c.Set("userID", uuid.New())
		c.Next()
	}, h.GetProfile)

	req := httptest.NewRequest(http.MethodGet, "/profile", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetProfile_404(t *testing.T) {
	uc := &mockAuthUsecase{
		getProfileFn: func(userID uuid.UUID) (*model.User, error) {
			return nil, usecase.ErrUserNotFound
		},
	}
	h := NewAuthHandler(uc)
	r := gin.New()
	r.GET("/profile", func(c *gin.Context) {
		c.Set("userID", uuid.New())
		c.Next()
	}, h.GetProfile)

	req := httptest.NewRequest(http.MethodGet, "/profile", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
