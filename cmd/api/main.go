package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	// your new endpoints package
	"github.com/ishantswami13-crypto/vantro/internal/api"
)

// --- helpers ---
func mustGetEnv(key string) string {
	val := strings.TrimSpace(os.Getenv(key))
	if val == "" {
		log.Fatalf("%s is required", key)
	}
	return val
}

func newPool(ctx context.Context, url string) *pgxpool.Pool {
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		log.Fatalf("parse DATABASE_URL: %v", err)
	}
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		log.Fatalf("pgxpool connect: %v", err)
	}
	return pool
}

func apiKeyGuard(apiKey string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if c.Request().Method == http.MethodOptions {
				return next(c)
			}
			if c.Path() == "/" || c.Path() == "/health" {
				return next(c)
			}
			if !strings.HasPrefix(c.Path(), "/api") {
				return next(c)
			}
			got := c.Request().Header.Get("API_KEY")
			if got == "" || got != apiKey {
				return c.JSON(http.StatusUnauthorized, echo.Map{"error": "unauthorized"})
			}
			return next(c)
		}
	}
}

func main() {
	// env
	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = "8080"
	}
	dbURL := mustGetEnv("DATABASE_URL") // e.g. postgres://...:5432/neondb?sslmode=require
	apiKey := mustGetEnv("API_KEY")

	// db
	ctx := context.Background()
	pool := newPool(ctx, dbURL)
	defer pool.Close()

	// echo
	e := echo.New()
	e.HideBanner = true
	e.Pre(middleware.RemoveTrailingSlash())
	e.Use(middleware.Recover())
	e.Use(middleware.Secure())

	// public
	e.GET("/", func(c echo.Context) error {
		return c.HTML(http.StatusOK, `<!doctype html><html><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>VANTRO API</title><style>
		body{margin:0;background:#fff;color:#111;font:500 16px/1.6 -apple-system,BlinkMacSystemFont,Segoe UI,Roboto} .wrap{max-width:920px;margin:0 auto;padding:56px 20px}
		.card{border:1px solid #eee;border-radius:16px;padding:24px}
		h1{margin:0 0 8px;font-weight:700} .pill{display:inline-block;margin-right:8px;padding:6px 10px;border:1px solid #eee;border-radius:999px}
		</style></head><body><div class="wrap"><div class="card">
		<h1>VANTRO API</h1><p>Where Wealth Meets Wisdom.</p>
		<div class="pill">DB: Postgres (Neon)</div><div class="pill">Auth: API Key</div>
		<pre>GET /health
GET /api/expenses
POST /api/expenses
GET /api/pots
POST /api/pots
PATCH /api/pots/:id
POST /api/coach/plan
GET /api/coach/plan
</pre></div></div></body></html>`)
	})
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, echo.Map{"status": "ok", "service": "vantro"})
	})

	// secured API group
	apiGroup := e.Group("/api", apiKeyGuard(apiKey))

	// register routes from your files (expenses.go, pots.go, coach.go)
	api.RegisterExpenseRoutes(apiGroup, pool)
	api.RegisterPotRoutes(apiGroup, pool)
	api.RegisterCoachRoutes(apiGroup, pool)

	// server
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      e,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	log.Printf("VANTRO API listening on :%s", port)
	if err := e.StartServer(srv); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}
