package jwtutil

import "time"

type TokenBlacklist interface {
	Add(jti string, expiresAt time.Time)
	IsBlacklisted(jti string) bool
	Stop()
}
