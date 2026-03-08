package jwtutil

import (
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func newTestBlacklist(t *testing.T) (TokenBlacklist, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	return NewRedisTokenBlacklist(rdb), mr
}

func TestRedisBlacklist_AddAndCheck(t *testing.T) {
	bl, _ := newTestBlacklist(t)

	jti := "test-jti-123"
	bl.Add(jti, time.Now().Add(time.Hour))

	assert.True(t, bl.IsBlacklisted(jti))
}

func TestRedisBlacklist_NotBlacklisted(t *testing.T) {
	bl, _ := newTestBlacklist(t)

	assert.False(t, bl.IsBlacklisted("unknown-jti"))
}

func TestRedisBlacklist_ExpiredTTL(t *testing.T) {
	bl, mr := newTestBlacklist(t)

	jti := "expiring-jti"
	bl.Add(jti, time.Now().Add(10*time.Second))
	assert.True(t, bl.IsBlacklisted(jti))

	mr.FastForward(11 * time.Second)
	assert.False(t, bl.IsBlacklisted(jti))
}

func TestRedisBlacklist_PastExpiry(t *testing.T) {
	bl, _ := newTestBlacklist(t)

	jti := "past-jti"
	bl.Add(jti, time.Now().Add(-time.Hour))

	assert.False(t, bl.IsBlacklisted(jti))
}
