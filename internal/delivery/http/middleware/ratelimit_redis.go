package middleware

import (
	"context"
	"log/slog"
	"time"

	"github.com/go-redis/redis_rate/v10"
	"github.com/redis/go-redis/v9"
)

type redisRateLimiter struct {
	limiter *redis_rate.Limiter
	limit   redis_rate.Limit
}

func NewRedisRateLimiter(rdb *redis.Client, rps int, burst int) RateLimiter {
	return &redisRateLimiter{
		limiter: redis_rate.NewLimiter(rdb),
		limit:   redis_rate.Limit{Rate: rps, Burst: burst, Period: time.Second},
	}
}

func (rl *redisRateLimiter) Allow(ip string) bool {
	ctx := context.Background()
	res, err := rl.limiter.Allow(ctx, "ratelimit:"+ip, rl.limit)
	if err != nil {
		slog.Error("rate limiter error", "ip", ip, "error", err)
		return true // fail open
	}
	return res.Allowed > 0
}

func (rl *redisRateLimiter) Stop() {
	// no-op: Redis client lifecycle managed in main.go
}
