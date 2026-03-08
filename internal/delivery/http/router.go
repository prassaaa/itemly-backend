package http

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prassaaa/itemly-backend/config"
	"github.com/prassaaa/itemly-backend/internal/delivery/http/handler"
	"github.com/prassaaa/itemly-backend/internal/delivery/http/middleware"
	"github.com/prassaaa/itemly-backend/internal/model"
	"github.com/prassaaa/itemly-backend/internal/usecase"
	jwtutil "github.com/prassaaa/itemly-backend/pkg/jwt"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func NewRouter(
	authHandler *handler.AuthHandler,
	generalHandler *handler.GeneralHandler,
	adminHandler *handler.AdminHandler,
	jwtService *jwtutil.JWTService,
	permUsecase usecase.PermissionUsecase,
	blacklist *jwtutil.TokenBlacklist,
	rateLimiterStore *middleware.RateLimiterStore,
	cfg *config.Config,
) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.SecurityHeaders())
	r.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORSOriginsList(),
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	r.Use(middleware.MaxBodySize(cfg.MaxBodySize))
	r.Use(middleware.RequestLogger())

	if !cfg.IsProduction() {
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	v1 := r.Group("/api/v1")
	{
		v1.GET("/health", generalHandler.HealthCheck)

		auth := v1.Group("/auth")
		auth.Use(middleware.RateLimit(rateLimiterStore))
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.Refresh)
		}

		protected := v1.Group("")
		protected.Use(middleware.JWTAuth(jwtService, blacklist))
		{
			protected.GET("/profile", authHandler.GetProfile)
			protected.POST("/auth/logout", authHandler.Logout)
			protected.GET("/dashboard", middleware.PermissionAuth(permUsecase, model.PermDashboardView), adminHandler.AdminDashboard)
			protected.PUT("/users/:id/role", middleware.PermissionAuth(permUsecase, model.PermUsersManage), adminHandler.AssignRole)
		}
	}

	return r
}
