package config

import (
	"log"
	"os"
)

type Config struct {
	Port        string
	DatabaseURL string
	APIKey      string
}

func Load() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	db := os.Getenv("DATABASE_URL")
	if db == "" {
		log.Fatal("DATABASE_URL is required")
	}

	key := os.Getenv("API_KEY")
	if key == "" {
		log.Fatal("API_KEY is required")
	}

	return Config{
		Port:        port,
		DatabaseURL: db,
		APIKey:      key,
	}
}
