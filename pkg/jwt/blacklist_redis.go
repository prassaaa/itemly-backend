package jwtutil

import (
	"context"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

type redisTokenBlacklist struct {
	rdb *redis.Client
}

func NewRedisTokenBlacklist(rdb *redis.Client) TokenBlacklist {
	return &redisTokenBlacklist{rdb: rdb}
}

func (bl *redisTokenBlacklist) Add(jti string, expiresAt time.Time) {
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		return
	}
	ctx := context.Background()
	if err := bl.rdb.Set(ctx, "blacklist:"+jti, "", ttl).Err(); err != nil {
		slog.Error("failed to add token to blacklist", "jti", jti, "error", err)
	}
}

func (bl *redisTokenBlacklist) IsBlacklisted(jti string) bool {
	ctx := context.Background()
	n, err := bl.rdb.Exists(ctx, "blacklist:"+jti).Result()
	if err != nil {
		slog.Error("failed to check blacklist", "jti", jti, "error", err)
		return false // fail open
	}
	return n > 0
}
