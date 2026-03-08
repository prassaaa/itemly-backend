package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/prassaaa/itemly-backend/config"
	"github.com/prassaaa/itemly-backend/internal/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewPostgresDB(cfg *config.Config) (*gorm.DB, *sql.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	if err := db.AutoMigrate(&model.User{}, &model.Permission{}, &model.RolePermission{}); err != nil {
		return nil, nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, sqlDB, nil
}
