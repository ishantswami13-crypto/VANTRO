package main

import (
	"log"
	"os"
	"time"

	"github.com/getsentry/sentry-go"
	sentryfiber "github.com/getsentry/sentry-go/fiber"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"

	"fintech-backend/internal/config"
	"fintech-backend/internal/db"
	"fintech-backend/internal/router"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, reading from environment")
	}

	cfg := config.Load()

	err := sentry.Init(sentry.ClientOptions{
		Dsn:              os.Getenv("SENTRY_DSN"),
		Environment:      "development",
		TracesSampleRate: 1.0,
	})
	if err != nil {
		log.Fatalf("sentry.Init: %s", err)
	}
	defer sentry.Flush(2 * time.Second)

	pool, err := db.NewPool(cfg.DatabaseURL)
	if err != nil {
		log.Printf("db connection failed, running in in-memory mode: %v", err)
	} else {
		defer pool.Close()
	}

	app := router.New(cfg, pool, sentryfiber.New(sentryfiber.Options{}))

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	log.Printf("API running on http://localhost:%s", cfg.Port)
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
