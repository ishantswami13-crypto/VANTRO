package httpx

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/ishantswami13-crypto/vantro/internal/payouts"
)

func Router(h *payouts.Handler, apiKey string) http.Handler {
	r := chi.NewRouter()

	// Health
	r.Get("/health", h.Health)
	r.Get("/version", h.Version)
	r.Get("/ready", h.Ready)

	// Authenticated API
	auth := APIKeyAuth(apiKey)
	r.Group(func(r chi.Router) {
		r.Use(auth)
		r.Post("/v1/payouts", h.Create)
		r.Get("/v1/payouts/{id}", h.Get)
		r.Get("/v1/payouts/ledger", h.Ledger)
		r.Post("/v1/payouts/{id}/webhook/replay", h.WebhookReplay)
	})

	return r
}
