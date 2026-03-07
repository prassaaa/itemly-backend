package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prassaaa/itemly-backend/internal/delivery/http/dto"
	"github.com/prassaaa/itemly-backend/internal/usecase"
)

type AuthHandler struct {
	authUsecase usecase.AuthUsecase
}

func NewAuthHandler(authUsecase usecase.AuthUsecase) *AuthHandler {
	return &AuthHandler{authUsecase: authUsecase}
}

// Register godoc
// @Summary      Register a new user
// @Description  Create a new user account with username, email, and password
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body  body      dto.RegisterRequest  true  "Register request"
// @Success      201   {object}  dto.AuthResponse
// @Failure      400   {object}  dto.ErrorResponse
// @Failure      409   {object}  dto.ErrorResponse
// @Failure      500   {object}  dto.ErrorResponse
// @Router       /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	user, token, err := h.authUsecase.Register(req.Username, req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrEmailAlreadyExists):
			c.JSON(http.StatusConflict, dto.ErrorResponse{Error: err.Error()})
		case errors.Is(err, usecase.ErrUsernameAlreadyExists):
			c.JSON(http.StatusConflict, dto.ErrorResponse{Error: err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal server error"})
		}
		return
	}

	c.JSON(http.StatusCreated, dto.AuthResponse{
		Token: token,
		User:  dto.ToUserResponse(user),
	})
}

// Login godoc
// @Summary      Login user
// @Description  Authenticate with email and password, returns JWT token
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body  body      dto.LoginRequest  true  "Login request"
// @Success      200   {object}  dto.AuthResponse
// @Failure      400   {object}  dto.ErrorResponse
// @Failure      401   {object}  dto.ErrorResponse
// @Failure      500   {object}  dto.ErrorResponse
// @Router       /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	user, token, err := h.authUsecase.Login(req.Email, req.Password)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, dto.AuthResponse{
		Token: token,
		User:  dto.ToUserResponse(user),
	})
}

// GetProfile godoc
// @Summary      Get current user profile
// @Description  Returns the profile of the authenticated user
// @Tags         User
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  dto.UserResponse
// @Failure      401  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /api/v1/profile [get]
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "user not authenticated"})
		return
	}

	id, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "invalid user ID"})
		return
	}

	user, err := h.authUsecase.GetProfile(id)
	if err != nil {
		if errors.Is(err, usecase.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, dto.ToUserResponse(user))
}
