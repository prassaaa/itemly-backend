package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prassaaa/itemly-backend/internal/delivery/http/dto"
)

type GeneralHandler struct{}

func NewGeneralHandler() *GeneralHandler {
	return &GeneralHandler{}
}

// HealthCheck godoc
// @Summary      Health check
// @Description  Returns OK if the service is running
// @Tags         General
// @Produce      json
// @Success      200  {object}  dto.MessageResponse
// @Router       /api/v1/health [get]
func (h *GeneralHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, dto.MessageResponse{Message: "OK"})
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
func (h *GeneralHandler) AdminDashboard(c *gin.Context) {
	username, _ := c.Get("username")
	usernameStr, _ := username.(string)
	c.JSON(http.StatusOK, dto.AdminDashboardResponse{
		Message: "welcome to admin dashboard",
		User:    usernameStr,
	})
}
