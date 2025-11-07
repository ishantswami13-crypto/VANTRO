package main

import (
	"context"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"

	"github.com/ishantswami13-crypto/vantro/internal/config"
	"github.com/ishantswami13-crypto/vantro/internal/db"
	"github.com/ishantswami13-crypto/vantro/internal/handlers"
	"github.com/ishantswami13-crypto/vantro/internal/middleware"
	"github.com/ishantswami13-crypto/vantro/internal/repo"
	"github.com/ishantswami13-crypto/vantro/internal/router"
	"github.com/ishantswami13-crypto/vantro/internal/services"
)

func main() {
	_ = godotenv.Load()

	cfg := config.Load()
	pool := db.MustPool(context.Background(), cfg.DatabaseURL)
	defer pool.Close()

	expRepo := repo.NewExpensesRepo(pool)
	potRepo := repo.NewPotsRepo(pool)
	coachRepo := repo.NewCoachRepo(pool)

	expSvc := services.NewExpensesService(expRepo)
	potSvc := services.NewPotsService(potRepo)
	coachSvc := services.NewCoachService(coachRepo)

	deps := &handlers.Deps{
		Expenses: expSvc,
		Pots:     potSvc,
		Coach:    coachSvc,
	}

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Use(middleware.APIKeyGuard(cfg.APIKey))

	router.New(app, pool, deps)

	log.Printf("VANTRO Fiber API listening on :%s", cfg.Port)
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
