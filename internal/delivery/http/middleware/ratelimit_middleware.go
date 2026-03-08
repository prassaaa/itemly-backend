package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prassaaa/itemly-backend/internal/delivery/http/dto"
)

type RateLimiter interface {
	Allow(ip string) bool
}

func RateLimit(limiter RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !limiter.Allow(ip) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, dto.ErrorResponse{
				Error: "too many requests, please try again later",
			})
			return
		}
		c.Next()
	}
}
