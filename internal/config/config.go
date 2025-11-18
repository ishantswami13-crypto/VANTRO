package config

import (
	"errors"
	"os"
)

type Config struct {
	Port        string
	DatabaseURL string
	APIKey      string
}

func Load() (*Config, error) {
	port := os.Getenv("PORT")
	dbURL := os.Getenv("DATABASE_URL")
	apiKey := os.Getenv("API_KEY")

	if port == "" || dbURL == "" || apiKey == "" {
		return nil, errors.New("missing PORT, DATABASE_URL or API_KEY in env")
	}

	return &Config{
		Port:        port,
		DatabaseURL: dbURL,
		APIKey:      apiKey,
	}, nil
}
