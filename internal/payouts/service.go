package payouts

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/ishantswami13-crypto/vantro/internal/storage"
)

type Provider interface {
	CreatePayout(ctx context.Context, amountCents int64, method string, dest map[string]string) (providerRef string, initialStatus string, err error)
	Resolve(ctx context.Context, providerRef string) (finalStatus, utr string)
}

type Service struct {
	DB       *storage.Store
	Provider Provider
}

func NewService(db *storage.Store, p Provider) *Service {
	return &Service{DB: db, Provider: p}
}

func rupeesToCents(amt float64) int64 { return int64(amt * 100.0) }

type CreateInput struct {
	Amount    float64
	Currency  string
	Method    string
	VPA       string
	Name      string
	Account   string
	IFSC      string
	Reference string
}

func (s *Service) Create(ctx context.Context, req CreateInput) (string, error) {
	if req.Currency == "" { req.Currency = "INR" }
	if req.Currency != "INR" { return "", errors.New("only INR supported in sandbox") }
	if req.Method != "upi" && req.Method != "bank" { return "", errors.New("invalid method") }

	dest := map[string]string{}
	switch req.Method {
	case "upi":
		if req.VPA == "" { return "", errors.New("upi.vpa required") }
		dest["vpa"] = req.VPA
		dest["name"] = req.Name
	case "bank":
		if req.Account == "" || req.IFSC == "" { return "", errors.New("bank.account & bank.ifsc required") }
		dest["account"] = req.Account
		dest["ifsc"] = req.IFSC
		dest["name"] = req.Name
	}

	amountCents := rupeesToCents(req.Amount)
	providerRef, initialStatus, err := s.Provider.CreatePayout(ctx, amountCents, req.Method, dest)
	if err != nil { return "", err }

	id := "po_" + uuid.NewString()

	tx, err := s.DB.Conn.Begin(ctx)
	if err != nil { return "", err }
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		INSERT INTO payouts (id, reference_id, amount_cents, currency, method, dest_vpa, dest_name, dest_account, dest_ifsc, status, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10, now(), now())
	`, id, req.Reference, amountCents, req.Currency, req.Method, req.VPA, req.Name, req.Account, req.IFSC, initialStatus)
	if err != nil { return "", err }

	evt := map[string]any{"payout_id": id, "status": initialStatus}
	b, _ := json.Marshal(evt)
	_, _ = tx.Exec(ctx, `INSERT INTO payout_events (id, payout_id, event, payload) VALUES ($1,$2,$3,$4)`,
		"evt_"+uuid.NewString(), id, "payout.processing", b)

	if err = tx.Commit(ctx); err != nil { return "", err }

	// simple background resolver (in prod: queue/worker)
	go s.resolveAsync(id, providerRef)

	return id, nil
}

func (s *Service) resolveAsync(id, providerRef string) {
	ctx := context.Background()
	time.Sleep(2 * time.Second) // simulate processing delay

	final, utr := s.Provider.Resolve(ctx, providerRef)

	var err error
	var res pgx.Rows
	switch final {
	case "success":
		res, err = s.DB.Conn.Query(ctx, `UPDATE payouts SET status='success', utr=$1, updated_at=now() WHERE id=$2 RETURNING id`, utr, id)
	default:
		res, err = s.DB.Conn.Query(ctx, `UPDATE payouts SET status='failed', updated_at=now(), error=$1 WHERE id=$2 RETURNING id`, "mock_failure", id)
	}
	if err == nil { res.Close() }

	evt := map[string]any{"payout_id": id, "status": final, "utr": utr}
	b, _ := json.Marshal(evt)
	_, _ = s.DB.Conn.Exec(ctx, `INSERT INTO payout_events (id, payout_id, event, payload) VALUES ($1,$2,$3,$4)`,
		"evt_"+uuid.NewString(), id, "payout."+final, b)

	// TODO: signed webhooks in next patch
}
