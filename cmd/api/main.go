package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"vantro/internal/httpx"
	"vantro/internal/payouts"
	"vantro/internal/providers/mock"
	"vantro/internal/storage"
)

func main() {
	_ = godotenv.Load()

	port := env("PORT", "8080")
	dbURL := env("DATABASE_URL", "")
	apiKey := env("API_KEY", "")
	if dbURL == "" {
		log.Fatal("DATABASE_URL required")
	}
	if apiKey == "" {
		log.Println("WARN: API_KEY empty; set for auth")
	}

	ctx := context.Background()
	store, err := storage.New(ctx, dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close(ctx)

	prov := mock.New() // sandbox provider
	svc := payouts.NewService(store, prov)
	h := &payouts.Handler{Svc: svc, DB: store.Conn, WebhookSecret: env("WEBHOOK_SECRET", "")}

	router := httpx.Router(h, apiKey)
	log.Printf("VANTRO payouts listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}

func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
