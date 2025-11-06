package storage

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type Store struct{ Conn *pgx.Conn }

func New(ctx context.Context, url string) (*Store, error) {
	conn, err := pgx.Connect(ctx, url)
	if err != nil {
		return nil, err
	}
	return &Store{Conn: conn}, nil
}

func (s *Store) Close(ctx context.Context) { _ = s.Conn.Close(ctx) }
