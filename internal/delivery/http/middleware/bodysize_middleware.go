package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prassaaa/itemly-backend/internal/delivery/http/dto"
)

func MaxBodySize(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > maxBytes {
			c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, dto.ErrorResponse{
				Error: "request body too large",
			})
			return
		}
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		c.Next()
	}
}
