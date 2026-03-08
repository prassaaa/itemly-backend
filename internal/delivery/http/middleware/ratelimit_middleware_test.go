package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// --- mock ---

type mockRateLimiter struct {
	allowFn func(ip string) bool
}

func (m *mockRateLimiter) Allow(ip string) bool {
	return m.allowFn(ip)
}

func TestRateLimit_Allowed(t *testing.T) {
	limiter := &mockRateLimiter{allowFn: func(ip string) bool { return true }}

	r := gin.New()
	r.GET("/test", RateLimit(limiter), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRateLimit_Blocked(t *testing.T) {
	limiter := &mockRateLimiter{allowFn: func(ip string) bool { return false }}

	r := gin.New()
	r.GET("/test", RateLimit(limiter), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)
}
