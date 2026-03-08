package handler

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
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

func formatValidationErrors(err error) map[string]string {
	fields := make(map[string]string)
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		for _, fe := range ve {
			field := fe.Field()
			switch fe.Tag() {
			case "required":
				fields[field] = fmt.Sprintf("%s is required", field)
			case "email":
				fields[field] = fmt.Sprintf("%s must be a valid email address", field)
			case "min":
				fields[field] = fmt.Sprintf("%s must be at least %s characters", field, fe.Param())
			case "max":
				fields[field] = fmt.Sprintf("%s must be at most %s characters", field, fe.Param())
			case "password":
				fields[field] = "password must contain at least one uppercase letter, one lowercase letter, one digit, and one special character"
			default:
				fields[field] = fmt.Sprintf("%s is invalid", field)
			}
		}
	}
	return fields
}

// Register godoc
// @Summary      Register a new user
// @Description  Create a new user account with username, email, and password
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body  body      dto.RegisterRequest  true  "Register request"
// @Success      201   {object}  dto.AuthResponse
// @Failure      400   {object}  dto.ValidationErrorResponse
// @Failure      409   {object}  dto.ErrorResponse
// @Failure      500   {object}  dto.ErrorResponse
// @Router       /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if fields := formatValidationErrors(err); len(fields) > 0 {
			c.JSON(http.StatusBadRequest, dto.ValidationErrorResponse{
				Error:  "validation failed",
				Fields: fields,
			})
			return
		}
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	user, tokenPair, err := h.authUsecase.Register(req.Username, req.Email, req.Password)
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
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		User:         dto.ToUserResponse(user),
	})
}

// Login godoc
// @Summary      Login user
// @Description  Authenticate with email and password, returns JWT token pair
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body  body      dto.LoginRequest  true  "Login request"
// @Success      200   {object}  dto.AuthResponse
// @Failure      400   {object}  dto.ValidationErrorResponse
// @Failure      401   {object}  dto.ErrorResponse
// @Failure      500   {object}  dto.ErrorResponse
// @Router       /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if fields := formatValidationErrors(err); len(fields) > 0 {
			c.JSON(http.StatusBadRequest, dto.ValidationErrorResponse{
				Error:  "validation failed",
				Fields: fields,
			})
			return
		}
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	user, tokenPair, err := h.authUsecase.Login(req.Email, req.Password)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, dto.AuthResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		User:         dto.ToUserResponse(user),
	})
}

// Refresh godoc
// @Summary      Refresh token pair
// @Description  Exchange a valid refresh token for a new access/refresh token pair
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body  body      dto.RefreshTokenRequest  true  "Refresh token request"
// @Success      200   {object}  dto.AuthResponse
// @Failure      400   {object}  dto.ValidationErrorResponse
// @Failure      401   {object}  dto.ErrorResponse
// @Failure      500   {object}  dto.ErrorResponse
// @Router       /api/v1/auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		if fields := formatValidationErrors(err); len(fields) > 0 {
			c.JSON(http.StatusBadRequest, dto.ValidationErrorResponse{
				Error:  "validation failed",
				Fields: fields,
			})
			return
		}
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	tokenPair, err := h.authUsecase.RefreshToken(req.RefreshToken)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidRefreshToken) {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: err.Error()})
		} else if errors.Is(err, usecase.ErrUserNotFound) {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "user no longer exists"})
		} else {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  tokenPair.AccessToken,
		"refresh_token": tokenPair.RefreshToken,
	})
}

// Logout godoc
// @Summary      Logout user
// @Description  Revoke the current access token
// @Tags         Auth
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  dto.MessageResponse
// @Failure      401  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	jti, exists := c.Get("jti")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "token info not found"})
		return
	}

	expiresAt, exists := c.Get("tokenExpiresAt")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "token info not found"})
		return
	}

	if err := h.authUsecase.Logout(jti.(string), expiresAt.(time.Time)); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal server error"})
		return
	}

	c.JSON(http.StatusOK, dto.MessageResponse{Message: "successfully logged out"})
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
