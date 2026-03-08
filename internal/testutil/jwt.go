package testutil

import (
	"github.com/google/uuid"
	jwtutil "github.com/prassaaa/itemly-backend/pkg/jwt"
)

const (
	TestSecret             = "test-secret-key-for-testing"
	TestAccessExpiryHours  = 1
	TestRefreshExpiryHours = 24
)

func TestJWTService() *jwtutil.JWTService {
	return jwtutil.NewJWTService(TestSecret, TestAccessExpiryHours, TestRefreshExpiryHours)
}

func TestAccessToken(userID uuid.UUID, username, role string) string {
	svc := TestJWTService()
	token, err := svc.GenerateToken(userID, username, role)
	if err != nil {
		panic("testutil: failed to generate access token: " + err.Error())
	}
	return token
}

func TestRefreshToken(userID uuid.UUID, username, role string) string {
	svc := TestJWTService()
	pair, err := svc.GenerateTokenPair(userID, username, role)
	if err != nil {
		panic("testutil: failed to generate token pair: " + err.Error())
	}
	return pair.RefreshToken
}
