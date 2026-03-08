package dto

import "github.com/prassaaa/itemly-backend/internal/model"

type AuthResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	User         UserResponse `json:"user"`
}

type UserResponse struct {
	ID       string     `json:"id"`
	Username string     `json:"username"`
	Email    string     `json:"email"`
	Role     model.Role `json:"role"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type ValidationErrorResponse struct {
	Error  string            `json:"error"`
	Fields map[string]string `json:"fields"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

type AdminDashboardResponse struct {
	Message string `json:"message"`
	User    string `json:"user"`
}

func ToUserResponse(user *model.User) UserResponse {
	return UserResponse{
		ID:       user.ID.String(),
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
	}
}
