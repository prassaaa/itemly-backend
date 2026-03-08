package jwtutil

import (
	"sync"
	"time"
)

type blacklistEntry struct {
	expiresAt time.Time
}

type TokenBlacklist struct {
	mu      sync.RWMutex
	entries map[string]blacklistEntry
	stopCh  chan struct{}
}

func NewTokenBlacklist() *TokenBlacklist {
	bl := &TokenBlacklist{
		entries: make(map[string]blacklistEntry),
		stopCh:  make(chan struct{}),
	}
	go bl.cleanup()
	return bl
}

func (bl *TokenBlacklist) Add(jti string, expiresAt time.Time) {
	bl.mu.Lock()
	defer bl.mu.Unlock()
	bl.entries[jti] = blacklistEntry{expiresAt: expiresAt}
}

func (bl *TokenBlacklist) IsBlacklisted(jti string) bool {
	bl.mu.RLock()
	defer bl.mu.RUnlock()
	_, exists := bl.entries[jti]
	return exists
}

func (bl *TokenBlacklist) cleanup() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			bl.mu.Lock()
			now := time.Now()
			for jti, entry := range bl.entries {
				if now.After(entry.expiresAt) {
					delete(bl.entries, jti)
				}
			}
			bl.mu.Unlock()
		case <-bl.stopCh:
			return
		}
	}
}

func (bl *TokenBlacklist) Stop() {
	close(bl.stopCh)
}
