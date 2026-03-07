package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prassaaa/itemly-backend/internal/delivery/http/dto"
	"github.com/prassaaa/itemly-backend/internal/delivery/http/handler"
	"github.com/prassaaa/itemly-backend/internal/delivery/http/middleware"
	"github.com/prassaaa/itemly-backend/internal/model"
	jwtutil "github.com/prassaaa/itemly-backend/pkg/jwt"
)

func NewRouter(authHandler *handler.AuthHandler, jwtService *jwtutil.JWTService) *gin.Engine {
	r := gin.Default()

	v1 := r.Group("/api/v1")
	{
		v1.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, dto.MessageResponse{Message: "OK"})
		})

		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		protected := v1.Group("")
		protected.Use(middleware.JWTAuth(jwtService))
		{
			protected.GET("/profile", authHandler.GetProfile)

			admin := protected.Group("/admin")
			admin.Use(middleware.RoleAuth(model.RoleAdmin))
			{
				admin.GET("/dashboard", func(c *gin.Context) {
					username, _ := c.Get("username")
					c.JSON(http.StatusOK, gin.H{
						"message": "welcome to admin dashboard",
						"user":    username,
					})
				})
			}
		}
	}

	return r
}
