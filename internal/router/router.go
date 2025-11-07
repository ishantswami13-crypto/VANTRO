package router

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ishantswami13-crypto/vantro/internal/handlers"
)

func New(app *fiber.App, pool *pgxpool.Pool, deps *handlers.Deps) {
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Type("html").SendString(`<!doctype html><html><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>VANTRO API</title><style>
body{margin:0;background:#fff;color:#111;font:500 16px/1.6 -apple-system,BlinkMacSystemFont,Segoe UI,Roboto}
.wrap{max-width:900px;margin:0 auto;padding:56px 20px}.card{border:1px solid #eee;border-radius:16px;padding:24px}
h1{margin:0 0 6px}
.pill{display:inline-block;margin:6px 8px 0 0;padding:6px 10px;border:1px solid #eee;border-radius:999px}
pre{background:#fafafa;border:1px solid #eee;border-radius:12px;padding:12px}
</style></head><body><div class="wrap"><div class="card">
<h1>VANTRO API</h1><p>Where Wealth Meets Wisdom.</p>
<div class="pill">Go Fiber</div><div class="pill">Postgres (Neon)</div><div class="pill">API Key</div>
<pre>GET /health
GET  /api/expenses?from=YYYY-MM-DD&to=YYYY-MM-DD
POST /api/expenses
GET  /api/pots
POST /api/pots
PATCH /api/pots/:id
POST /api/coach/plan
GET  /api/coach/plan
</pre></div></div></body></html>`)
	})

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.Status(http.StatusOK).JSON(fiber.Map{"status": "ok", "service": "vantro"})
	})

	api := app.Group("/api")
	handlers.RegisterExpenseRoutes(api, deps)
	handlers.RegisterPotRoutes(api, deps)
	handlers.RegisterCoachRoutes(api, deps)
	_ = pool // reserved for future DB-aware middleware hooks
}
