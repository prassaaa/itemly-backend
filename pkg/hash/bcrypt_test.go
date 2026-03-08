package hash

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashPassword(t *testing.T) {
	hashed, err := HashPassword("Test@1234")
	require.NoError(t, err)
	assert.NotEmpty(t, hashed)
	assert.NotEqual(t, "Test@1234", hashed)
}

func TestCheckPassword_Correct(t *testing.T) {
	hashed, err := HashPassword("Test@1234")
	require.NoError(t, err)

	assert.True(t, CheckPassword("Test@1234", hashed))
}

func TestCheckPassword_Wrong(t *testing.T) {
	hashed, err := HashPassword("Test@1234")
	require.NoError(t, err)

	assert.False(t, CheckPassword("WrongPass", hashed))
}

func TestCheckPassword_InvalidHash(t *testing.T) {
	assert.False(t, CheckPassword("anything", "not-a-bcrypt-hash"))
}
