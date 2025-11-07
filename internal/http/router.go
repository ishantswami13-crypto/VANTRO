package httpx

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/ishantswami13-crypto/vantro/internal/payouts"
)

func NewRouter(h *payouts.Handler, apiKey string) http.Handler {
	r := chi.NewRouter()

	// Public routes
	r.Get("/", landingPageHandler)
	r.Get("/health", healthHandler)

	// Authenticated API
	auth := APIKeyAuth(apiKey)
	r.Route("/api", func(api chi.Router) {
		api.Use(auth)

		api.Get("/health", h.Health)
		api.Get("/version", h.Version)
		api.Get("/ready", h.Ready)

		api.Post("/payouts", h.Create)
		api.Get("/payouts/{id}", h.Get)
		api.Get("/payouts/ledger", h.Ledger)
		api.Post("/payouts/{id}/webhook/replay", h.WebhookReplay)
	})

	return r
}

func landingPageHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(landingHTML))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"status":"ok","service":"VANTRO","version":"1.0"}`))
}

const landingHTML = `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8" />
<meta name="viewport" content="width=device-width, initial-scale=1" />
<title>VANTRO API</title>
<style>
  :root{--bg:#0b0b10;--card:#11131a;--muted:#9aa3b2;--accent:#7dd3fc;--text:#e6edf3;--ok:#22c55e}
  *{box-sizing:border-box} body{margin:0;background:radial-gradient(1000px 600px at 10% -10%,#111a23 0,#0b0b10 40%),var(--bg);color:var(--text);font:500 15px/1.6 ui-sans-serif,system-ui,-apple-system,"Segoe UI",Roboto}
  .wrap{max-width:960px;margin:0 auto;padding:56px 20px}
  .card{background:linear-gradient(180deg,rgba(255,255,255,.03),rgba(255,255,255,.01));border:1px solid rgba(255,255,255,.06);border-radius:16px;padding:28px}
  h1{font-size:36px;letter-spacing:.3px;margin:0 0 8px}
  p{color:var(--muted);margin:0 0 18px}
  .row{display:flex;flex-wrap:wrap;gap:14px;margin-top:18px}
  .pill{padding:8px 12px;border-radius:999px;background:#0f1320;border:1px solid rgba(125,211,252,.25)}
  code,pre{font-family:ui-monospace,SFMono-Regular,Menlo,Consolas,monospace}
  pre{background:#0f1117;border:1px solid rgba(255,255,255,.06);padding:14px;border-radius:12px;overflow:auto}
  .cta{display:inline-block;margin-top:14px;padding:10px 14px;border-radius:10px;border:1px solid rgba(125,211,252,.4);background:linear-gradient(180deg,rgba(125,211,252,.12),rgba(125,211,252,.06));color:#c8eefc;text-decoration:none}
  .ok{color:var(--ok)}
</style>
</head>
<body>
  <div class="wrap">
    <div class="card">
      <h1>VANTRO API</h1>
      <p>Your payouts API is live. Root is public; authenticated endpoints are under <code>/api</code>.</p>
      <div class="row">
        <span class="pill">Status: <span class="ok">Healthy</span></span>
        <span class="pill">Region: Render</span>
        <span class="pill">DB: Neon (Postgres)</span>
      </div>
      <h3 style="margin:22px 0 10px">Quick checks</h3>
      <pre><code>GET /health
GET /api/payouts            // requires API_KEY header
</code></pre>
      <a class="cta" href="/health">Open /health</a>
    </div>
  </div>
</body>
</html>`
