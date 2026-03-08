package dto

import (
	"testing"

	"github.com/google/uuid"
	"github.com/prassaaa/itemly-backend/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestToUserResponse(t *testing.T) {
	user := &model.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "test@example.com",
		Role:     model.RoleAdmin,
	}

	resp := ToUserResponse(user)

	assert.Equal(t, user.ID.String(), resp.ID)
	assert.Equal(t, "testuser", resp.Username)
	assert.Equal(t, "test@example.com", resp.Email)
	assert.Equal(t, model.RoleAdmin, resp.Role)
}
