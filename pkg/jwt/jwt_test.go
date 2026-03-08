package jwtutil

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewJWTService(t *testing.T) {
	svc := NewJWTService("secret", 1, 24)
	assert.NotNil(t, svc)
}

func TestGenerateToken(t *testing.T) {
	svc := NewJWTService("secret", 1, 24)
	userID := uuid.New()

	token, err := svc.GenerateToken(userID, "testuser", "staff")
	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestGenerateTokenPair(t *testing.T) {
	svc := NewJWTService("secret", 1, 24)
	userID := uuid.New()

	pair, err := svc.GenerateTokenPair(userID, "testuser", "staff")
	require.NoError(t, err)
	assert.NotEmpty(t, pair.AccessToken)
	assert.NotEmpty(t, pair.RefreshToken)
}

func TestValidateToken_Valid(t *testing.T) {
	svc := NewJWTService("secret", 1, 24)
	userID := uuid.New()

	token, err := svc.GenerateToken(userID, "testuser", "admin")
	require.NoError(t, err)

	claims, err := svc.ValidateToken(token)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, "testuser", claims.Username)
	assert.Equal(t, "admin", claims.Role)
	assert.Equal(t, AccessToken, claims.TokenType)
	assert.NotEmpty(t, claims.ID)
}

func TestValidateToken_InvalidString(t *testing.T) {
	svc := NewJWTService("secret", 1, 24)

	_, err := svc.ValidateToken("not-a-token")
	assert.Error(t, err)
}

func TestValidateToken_WrongSecret(t *testing.T) {
	svc1 := NewJWTService("secret1", 1, 24)
	svc2 := NewJWTService("secret2", 1, 24)

	token, err := svc1.GenerateToken(uuid.New(), "user", "staff")
	require.NoError(t, err)

	_, err = svc2.ValidateToken(token)
	assert.Error(t, err)
}

func TestValidateToken_Expired(t *testing.T) {
	svc := NewJWTService("secret", -1, -1)

	token, err := svc.GenerateToken(uuid.New(), "user", "staff")
	require.NoError(t, err)

	_, err = svc.ValidateToken(token)
	assert.Error(t, err)
}

func TestGenerateTokenPair_TokenTypes(t *testing.T) {
	svc := NewJWTService("secret", 1, 24)
	userID := uuid.New()

	pair, err := svc.GenerateTokenPair(userID, "testuser", "staff")
	require.NoError(t, err)

	accessClaims, err := svc.ValidateToken(pair.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, AccessToken, accessClaims.TokenType)

	refreshClaims, err := svc.ValidateToken(pair.RefreshToken)
	require.NoError(t, err)
	assert.Equal(t, RefreshToken, refreshClaims.TokenType)
}
