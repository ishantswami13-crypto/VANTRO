package storage

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNoDatabaseURL = errors.New("DATABASE_URL missing")

func NewPoolFromEnv(ctx context.Context) (*pgxpool.Pool, error) {
	// Trim whitespace or newline characters from env variable
	url := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if url == "" {
		return nil, ErrNoDatabaseURL
	}

	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, err
	}

	return pgxpool.NewWithConfig(ctx, cfg)
}
