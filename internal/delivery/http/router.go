package http

import (
	"github.com/gin-gonic/gin"
	"github.com/prassaaa/itemly-backend/internal/delivery/http/handler"
	"github.com/prassaaa/itemly-backend/internal/delivery/http/middleware"
	"github.com/prassaaa/itemly-backend/internal/model"
	jwtutil "github.com/prassaaa/itemly-backend/pkg/jwt"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func NewRouter(authHandler *handler.AuthHandler, generalHandler *handler.GeneralHandler, jwtService *jwtutil.JWTService) *gin.Engine {
	r := gin.Default()

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	v1 := r.Group("/api/v1")
	{
		v1.GET("/health", generalHandler.HealthCheck)

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
				admin.GET("/dashboard", generalHandler.AdminDashboard)
			}
		}
	}

	return r
}
