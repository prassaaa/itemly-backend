package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	db, sqlDB, err := database.NewPostgresDB(cfg)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}

	jwtService := jwtutil.NewJWTService(cfg.JWTSecret, cfg.JWTExpiryHrs)

	userRepo := repository.NewUserRepository(db)
	permissionRepo := repository.NewPermissionRepository(db)
	authUsecase := usecase.NewAuthUsecase(userRepo, jwtService)
	adminUsecase := usecase.NewAdminUsecase(userRepo)
	permissionUsecase := usecase.NewPermissionUsecase(permissionRepo)
	authHandler := handler.NewAuthHandler(authUsecase)
	generalHandler := handler.NewGeneralHandler()
	adminHandler := handler.NewAdminHandler(adminUsecase)

	if err := database.SeedPermissions(db); err != nil {
		slog.Error("failed to seed permissions", "error", err)
		os.Exit(1)
	}

	if err := permissionUsecase.LoadPermissions(); err != nil {
		slog.Error("failed to load permissions into cache", "error", err)
		os.Exit(1)
	}

	router := httpdelivery.NewRouter(authHandler, generalHandler, adminHandler, jwtService, permissionUsecase)

	srv := &http.Server{
		Addr:    ":" + cfg.AppPort,
		Handler: router,
	}

	go func() {
		slog.Info("server starting", "port", cfg.AppPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server listen error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("server forced to shutdown", "error", err)
	}

	if err := sqlDB.Close(); err != nil {
		slog.Error("failed to close database connection", "error", err)
	}

	slog.Info("server exited")
}
