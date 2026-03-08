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
