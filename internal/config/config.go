package config

import (
	"log"
	"os"
	"strings"
)

type Config struct {
	Port                  string
	DatabaseURL           string
	APIKey                string
	JWTSecret             string
	RazorpayKeyID         string
	RazorpayKeySecret     string
	RazorpayWebhookSecret string
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("missing required env var: %s", key)
	}
	return v
}

func envOrDefault(key, fallback string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v != "" {
		return v
	}
	return fallback
}

func Load() *Config {
	return &Config{
		Port:                  envOrDefault("PORT", "8080"),
		DatabaseURL:           envOrDefault("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/vantro?sslmode=disable"),
		APIKey:                envOrDefault("API_KEY", "supersecretapikey"),
		JWTSecret:             envOrDefault("JWT_SECRET", "vantro-dev-jwt-secret"),
		RazorpayKeyID:         envOrDefault("RAZORPAY_KEY_ID", ""),
		RazorpayKeySecret:     envOrDefault("RAZORPAY_KEY_SECRET", ""),
		RazorpayWebhookSecret: envOrDefault("RAZORPAY_WEBHOOK_SECRET", ""),
	}
}
