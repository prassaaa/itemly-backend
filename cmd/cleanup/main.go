package main

import (
	"fmt"
	"log"
	"os"

	"github.com/prassaaa/itemly-backend/config"
	"github.com/prassaaa/itemly-backend/database"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("failed to load config:", err)
	}

	db, sqlDB, err := database.NewPostgresDB(cfg)
	if err != nil {
		log.Fatal("failed to connect to database:", err)
	}
	defer sqlDB.Close()

	result := db.Exec("DELETE FROM users WHERE email LIKE '%@loadtest.com'")
	if result.Error != nil {
		log.Fatal("failed to delete loadtest users:", result.Error)
	}

	fmt.Printf("Deleted %d loadtest users\n", result.RowsAffected)
	os.Exit(0)
}
