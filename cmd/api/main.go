package main

import (
	"log"

	"github.com/prassaaa/itemly-backend/config"
	"github.com/prassaaa/itemly-backend/database"
	httpdelivery "github.com/prassaaa/itemly-backend/internal/delivery/http"
	"github.com/prassaaa/itemly-backend/internal/delivery/http/handler"
	"github.com/prassaaa/itemly-backend/internal/repository"
	"github.com/prassaaa/itemly-backend/internal/usecase"
	jwtutil "github.com/prassaaa/itemly-backend/pkg/jwt"

	_ "github.com/prassaaa/itemly-backend/docs"
)

// @title           Itemly API
// @version         1.0
// @description     Inventory management backend API
// @host            localhost:8080
// @BasePath        /

// @securityDefinitions.apikey BearerAuth
// @in                         header
// @name                       Authorization
// @description                Enter your bearer token in the format: Bearer {token}

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := database.NewPostgresDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	jwtService := jwtutil.NewJWTService(cfg.JWTSecret, cfg.JWTExpiryHrs)

	userRepo := repository.NewUserRepository(db)
	authUsecase := usecase.NewAuthUsecase(userRepo, jwtService)
	authHandler := handler.NewAuthHandler(authUsecase)
	generalHandler := handler.NewGeneralHandler()

	router := httpdelivery.NewRouter(authHandler, generalHandler, jwtService)

	log.Printf("Server starting on port %s", cfg.AppPort)
	if err := router.Run(":" + cfg.AppPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
