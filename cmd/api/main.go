package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	apphttp "github.com/ishantswami13-crypto/vantro/internal/http"
	"github.com/ishantswami13-crypto/vantro/internal/payouts"
	products "github.com/ishantswami13-crypto/vantro/internal/products"
	"github.com/ishantswami13-crypto/vantro/internal/providers/mock"
	"github.com/ishantswami13-crypto/vantro/internal/storage"
	"github.com/joho/godotenv"
)

func main() {
	// load .env locally; harmless on Render
	_ = godotenv.Load()

	// connect DB (forces using DATABASE_URL, avoids local /tmp/.s.PGSQL.5432)
	pool, err := storage.NewPoolFromEnv(context.Background())
	if err != nil {
		log.Fatalf("db connect failed: %v", err)
	}
	defer pool.Close()

	prov := mock.New() // sandbox provider
	svc := payouts.NewService(pool, prov)
	h := &payouts.Handler{
		Svc:           svc,
		DB:            pool,
		WebhookSecret: env("WEBHOOK_SECRET", ""),
	}
	prodSvc := products.NewService(pool)
	prodH := products.NewHandler(prodSvc)

	apiKey := env("API_KEY", "")
	if apiKey == "" {
		log.Println("WARN: API_KEY empty; set for auth")
	}
	mux := apphttp.NewRouter(h, prodH, apiKey)

	addr := ":" + env("PORT", "10000") // Render injects PORT
	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	log.Printf("VANTRO up on %s", addr)
	log.Fatal(srv.ListenAndServe())
}

func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
