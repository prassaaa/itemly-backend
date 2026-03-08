package validator

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testStruct struct {
	Password string `validate:"password"`
}

func setupValidator(t *testing.T) *validator.Validate {
	t.Helper()
	v := validator.New()
	require.NoError(t, v.RegisterValidation("password", PasswordStrength))
	return v
}

func TestPasswordStrength_Valid(t *testing.T) {
	v := setupValidator(t)
	err := v.Struct(testStruct{Password: "Test@1234"})
	assert.NoError(t, err)
}

func TestPasswordStrength_NoUppercase(t *testing.T) {
	v := setupValidator(t)
	err := v.Struct(testStruct{Password: "test@1234"})
	assert.Error(t, err)
}

func TestPasswordStrength_NoLowercase(t *testing.T) {
	v := setupValidator(t)
	err := v.Struct(testStruct{Password: "TEST@1234"})
	assert.Error(t, err)
}

func TestPasswordStrength_NoDigit(t *testing.T) {
	v := setupValidator(t)
	err := v.Struct(testStruct{Password: "Test@abcd"})
	assert.Error(t, err)
}

func TestPasswordStrength_NoSpecial(t *testing.T) {
	v := setupValidator(t)
	err := v.Struct(testStruct{Password: "Test1234a"})
	assert.Error(t, err)
}
