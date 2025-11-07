package storage

import (
	"context"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPoolFromEnv connects using DATABASE_URL. sslmode is already in your Neon URL.
func NewPoolFromEnv(ctx context.Context) (*pgxpool.Pool, error) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		return nil, ErrMissingURL
	}
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, err
	}
	return pgxpool.NewWithConfig(ctx, cfg)
}

// tiny sentinel error so failures are obvious in logs
var ErrMissingURL = &missingURLError{}

type missingURLError struct{}

func (e *missingURLError) Error() string { return "DATABASE_URL is not set" }
