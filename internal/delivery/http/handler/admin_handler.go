package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prassaaa/itemly-backend/internal/delivery/http/dto"
	"github.com/prassaaa/itemly-backend/internal/usecase"
)

type AdminHandler struct {
	adminUsecase usecase.AdminUsecase
}

func NewAdminHandler(adminUsecase usecase.AdminUsecase) *AdminHandler {
	return &AdminHandler{adminUsecase: adminUsecase}
}

// AdminDashboard godoc
// @Summary      Admin dashboard
// @Description  Returns admin dashboard info (admin only)
// @Tags         Admin
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  dto.AdminDashboardResponse
// @Failure      401  {object}  dto.ErrorResponse
// @Failure      403  {object}  dto.ErrorResponse
// @Router       /api/v1/admin/dashboard [get]
func (h *AdminHandler) AdminDashboard(c *gin.Context) {
	username, _ := c.Get("username")
	usernameStr, _ := username.(string)
	c.JSON(http.StatusOK, dto.AdminDashboardResponse{
		Message: "welcome to admin dashboard",
		User:    usernameStr,
	})
}

// AssignRole godoc
// @Summary      Assign role to user
// @Description  Update a user's role (admin only)
// @Tags         Admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      string               true  "User ID (UUID)"
// @Param        body  body      dto.AssignRoleRequest true  "Role assignment request"
// @Success      200   {object}  dto.UserResponse
// @Failure      400   {object}  dto.ErrorResponse
// @Failure      401   {object}  dto.ErrorResponse
// @Failure      403   {object}  dto.ErrorResponse
// @Failure      404   {object}  dto.ErrorResponse
// @Failure      500   {object}  dto.ErrorResponse
// @Router       /api/v1/admin/users/{id}/role [put]
func (h *AdminHandler) AssignRole(c *gin.Context) {
	targetID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid user ID"})
		return
	}

	var req dto.AssignRoleRequest
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

	user, err := h.adminUsecase.AssignRole(targetID, req.Role)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrUserNotFound):
			c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: err.Error()})
		case errors.Is(err, usecase.ErrInvalidRole):
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, dto.ToUserResponse(user))
}
