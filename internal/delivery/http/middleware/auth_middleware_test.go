package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prassaaa/itemly-backend/internal/model"
	"github.com/prassaaa/itemly-backend/internal/testutil"
	jwtutil "github.com/prassaaa/itemly-backend/pkg/jwt"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// --- mocks ---

type mockTokenBlacklist struct {
	isBlacklistedFn func(jti string) bool
}

func (m *mockTokenBlacklist) Add(jti string, expiresAt time.Time) {}
func (m *mockTokenBlacklist) IsBlacklisted(jti string) bool {
	return m.isBlacklistedFn(jti)
}

type mockPermissionUsecase struct {
	hasPermissionFn func(role string, perm string) bool
}

func (m *mockPermissionUsecase) HasPermission(role string, permissionName string) bool {
	return m.hasPermissionFn(role, permissionName)
}
func (m *mockPermissionUsecase) GetPermissionsByRole(role string) []string { return nil }
func (m *mockPermissionUsecase) LoadPermissions() error                    { return nil }

// --- helpers ---

func setupJWTRouter(jwtSvc *jwtutil.JWTService, bl jwtutil.TokenBlacklist) *gin.Engine {
	r := gin.New()
	r.GET("/protected", JWTAuth(jwtSvc, bl), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"userID":   c.MustGet("userID"),
			"username": c.MustGet("username"),
			"role":     c.MustGet("role"),
			"jti":      c.MustGet("jti"),
		})
	})
	return r
}

// --- JWTAuth tests ---

func TestJWTAuth_ValidToken(t *testing.T) {
	jwtSvc := testutil.TestJWTService()
	bl := &mockTokenBlacklist{isBlacklistedFn: func(jti string) bool { return false }}

	userID := uuid.New()
	token := testutil.TestAccessToken(userID, "testuser", "staff")

	r := setupJWTRouter(jwtSvc, bl)
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestJWTAuth_NoHeader(t *testing.T) {
	jwtSvc := testutil.TestJWTService()
	bl := &mockTokenBlacklist{isBlacklistedFn: func(jti string) bool { return false }}

	r := setupJWTRouter(jwtSvc, bl)
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestJWTAuth_InvalidFormat(t *testing.T) {
	jwtSvc := testutil.TestJWTService()
	bl := &mockTokenBlacklist{isBlacklistedFn: func(jti string) bool { return false }}

	r := setupJWTRouter(jwtSvc, bl)
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "InvalidFormat")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestJWTAuth_ExpiredToken(t *testing.T) {
	// Create JWT service with -1h expiry to produce an already-expired token
	expiredSvc := jwtutil.NewJWTService(testutil.TestSecret, -1, -1)
	bl := &mockTokenBlacklist{isBlacklistedFn: func(jti string) bool { return false }}

	token, _ := expiredSvc.GenerateToken(uuid.New(), "expired", "staff")

	jwtSvc := testutil.TestJWTService()
	r := setupJWTRouter(jwtSvc, bl)
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestJWTAuth_RefreshTokenRejected(t *testing.T) {
	jwtSvc := testutil.TestJWTService()
	bl := &mockTokenBlacklist{isBlacklistedFn: func(jti string) bool { return false }}

	refreshToken := testutil.TestRefreshToken(uuid.New(), "testuser", "staff")

	r := setupJWTRouter(jwtSvc, bl)
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+refreshToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestJWTAuth_BlacklistedToken(t *testing.T) {
	jwtSvc := testutil.TestJWTService()
	bl := &mockTokenBlacklist{isBlacklistedFn: func(jti string) bool { return true }}

	token := testutil.TestAccessToken(uuid.New(), "testuser", "staff")

	r := setupJWTRouter(jwtSvc, bl)
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// --- RoleAuth tests ---

func TestRoleAuth_NoRoleInContext(t *testing.T) {
	r := gin.New()
	r.GET("/admin", RoleAuth(model.RoleAdmin), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestRoleAuth_Allowed(t *testing.T) {
	r := gin.New()
	r.GET("/admin", func(c *gin.Context) {
		c.Set("role", "admin")
		c.Next()
	}, RoleAuth(model.RoleAdmin), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRoleAuth_Forbidden(t *testing.T) {
	r := gin.New()
	r.GET("/admin", func(c *gin.Context) {
		c.Set("role", "staff")
		c.Next()
	}, RoleAuth(model.RoleAdmin), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestRoleAuth_MultipleRoles(t *testing.T) {
	r := gin.New()
	r.GET("/manage", func(c *gin.Context) {
		c.Set("role", "manager")
		c.Next()
	}, RoleAuth(model.RoleAdmin, model.RoleManager), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/manage", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// --- PermissionAuth tests ---

func TestPermissionAuth_NoRoleInContext(t *testing.T) {
	perm := &mockPermissionUsecase{
		hasPermissionFn: func(role string, p string) bool { return true },
	}

	r := gin.New()
	r.GET("/test", PermissionAuth(perm, "dashboard:view"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestPermissionAuth_Allowed(t *testing.T) {
	perm := &mockPermissionUsecase{
		hasPermissionFn: func(role string, p string) bool { return true },
	}

	r := gin.New()
	r.GET("/dashboard", func(c *gin.Context) {
		c.Set("role", "admin")
		c.Next()
	}, PermissionAuth(perm, "dashboard:view"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestPermissionAuth_Denied(t *testing.T) {
	perm := &mockPermissionUsecase{
		hasPermissionFn: func(role string, p string) bool { return false },
	}

	r := gin.New()
	r.GET("/dashboard", func(c *gin.Context) {
		c.Set("role", "staff")
		c.Next()
	}, PermissionAuth(perm, "users:manage"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}
