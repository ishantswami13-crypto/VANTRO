package main

import (
	"log"

	"github.com/joho/godotenv"

	"fintech-backend/internal/config"
	"fintech-backend/internal/db"
	"fintech-backend/internal/router"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, reading from environment")
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	pool, err := db.NewPool(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db error: %v", err)
	}
	defer pool.Close()

	app := router.New(cfg, pool)

	log.Printf("Server starting on :%s ...", cfg.Port)
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
