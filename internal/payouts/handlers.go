package payouts

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/ishantswami13-crypto/vantro/internal/types"
	"github.com/jackc/pgx/v5"
)

type Handler struct {
	Svc           *Service
	DB            *pgx.Conn
	WebhookSecret string
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req types.CreatePayoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	in := CreateInput{
		Amount:    req.Amount,
		Currency:  req.Currency,
		Method:    req.Method,
		Reference: req.Reference,
	}
	if req.Method == "upi" && req.UPI != nil {
		in.VPA, in.Name = req.UPI.VPA, req.UPI.Name
	}
	if req.Method == "bank" && req.Bank != nil {
		in.Account, in.IFSC, in.Name = req.Bank.Account, req.Bank.IFSC, req.Bank.Name
	}

	id, err := h.Svc.Create(r.Context(), in)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := types.CreatePayoutResponse{
		PayoutID:           id,
		Status:             "processing",
		Reference:          req.Reference,
		ExpectedSettlement: time.Now().Add(3 * time.Minute).UTC().Format(time.RFC3339),
	}
	writeJSON(w, http.StatusCreated, resp)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var status, method, utr string
	var amountCents int64
	err := h.DB.QueryRow(r.Context(),
		`SELECT status, method, COALESCE(utr,''), amount_cents FROM payouts WHERE id=$1`, id).
		Scan(&status, &method, &utr, &amountCents)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	resp := types.Payout{
		ID:          id,
		Status:      status,
		Method:      method,
		UTR:         utr,
		Amount:      float64(amountCents) / 100.0,
		ProcessedAt: time.Now().UTC().Format(time.RFC3339),
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) Ledger(w http.ResponseWriter, r *http.Request) {
	limit := 50
	if s := r.URL.Query().Get("limit"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 && n <= 200 {
			limit = n
		}
	}
	rows, err := h.DB.Query(r.Context(),
		`SELECT id, status, method, COALESCE(utr,''), amount_cents, created_at FROM payouts
		 ORDER BY created_at DESC LIMIT $1`, limit)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type row struct {
		ID        string  `json:"payout_id"`
		Status    string  `json:"status"`
		Method    string  `json:"method"`
		UTR       string  `json:"utr"`
		Amount    float64 `json:"amount"`
		CreatedAt string  `json:"created_at"`
	}
	var out []row
	for rows.Next() {
		var r0 row
		var cents int64
		var t time.Time
		if err := rows.Scan(&r0.ID, &r0.Status, &r0.Method, &r0.UTR, &cents, &t); err != nil {
			continue
		}
		r0.Amount = float64(cents) / 100.0
		r0.CreatedAt = t.UTC().Format(time.RFC3339)
		out = append(out, r0)
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *Handler) WebhookReplay(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var payload []byte
	err := h.DB.QueryRow(r.Context(),
		`SELECT payload FROM payout_events WHERE payout_id=$1 ORDER BY created_at DESC LIMIT 1`, id).
		Scan(&payload)
	if err != nil {
		http.Error(w, "no events", http.StatusNotFound)
		return
	}
	writeJSONRaw(w, http.StatusOK, payload)
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) Version(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"name": "VANTRO", "service": "payouts", "env": "dev"})
}

func (h *Handler) Ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := h.DB.Ping(ctx); err != nil {
		http.Error(w, "not ready", http.StatusServiceUnavailable)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"ready": "true"})
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
func writeJSONRaw(w http.ResponseWriter, code int, b []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, _ = w.Write(b)
}
