package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prassaaa/itemly-backend/internal/delivery/http/dto"
	"golang.org/x/time/rate"
)

type ipLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type RateLimiterStore struct {
	mu       sync.RWMutex
	limiters map[string]*ipLimiter
	rps      rate.Limit
	burst    int
	stopCh   chan struct{}
}

func NewRateLimiterStore(rps float64, burst int) *RateLimiterStore {
	s := &RateLimiterStore{
		limiters: make(map[string]*ipLimiter),
		rps:      rate.Limit(rps),
		burst:    burst,
		stopCh:   make(chan struct{}),
	}
	go s.cleanup()
	return s
}

func (s *RateLimiterStore) getLimiter(ip string) *rate.Limiter {
	s.mu.Lock()
	defer s.mu.Unlock()

	if entry, exists := s.limiters[ip]; exists {
		entry.lastSeen = time.Now()
		return entry.limiter
	}

	limiter := rate.NewLimiter(s.rps, s.burst)
	s.limiters[ip] = &ipLimiter{limiter: limiter, lastSeen: time.Now()}
	return limiter
}

func (s *RateLimiterStore) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			s.mu.Lock()
			cutoff := time.Now().Add(-10 * time.Minute)
			for ip, entry := range s.limiters {
				if entry.lastSeen.Before(cutoff) {
					delete(s.limiters, ip)
				}
			}
			s.mu.Unlock()
		case <-s.stopCh:
			return
		}
	}
}

func (s *RateLimiterStore) Stop() {
	close(s.stopCh)
}

func RateLimit(store *RateLimiterStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		limiter := store.getLimiter(ip)
		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, dto.ErrorResponse{
				Error: "too many requests, please try again later",
			})
			return
		}
		c.Next()
	}
}
