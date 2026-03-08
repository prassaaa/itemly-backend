package middleware

import (
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func newTestRateLimiter(t *testing.T, rps, burst int) (RateLimiter, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	rl := NewRedisRateLimiter(rdb, rps, burst)
	return rl, mr
}

func TestRedisRateLimiter_AllowsUnderLimit(t *testing.T) {
	rl, _ := newTestRateLimiter(t, 10, 10)

	assert.True(t, rl.Allow("192.168.1.1"))
}

func TestRedisRateLimiter_BlocksOverLimit(t *testing.T) {
	rl, _ := newTestRateLimiter(t, 1, 1)

	// First request should be allowed
	first := rl.Allow("192.168.1.1")
	assert.True(t, first)

	// Second request should be blocked (burst=1)
	second := rl.Allow("192.168.1.1")
	assert.False(t, second)
}
