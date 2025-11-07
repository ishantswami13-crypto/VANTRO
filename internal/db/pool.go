package db

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func MustPool(ctx context.Context, url string) *pgxpool.Pool {
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		log.Fatalf("parse DATABASE_URL: %v", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		log.Fatalf("pgxpool connect: %v", err)
	}
	return pool
}
