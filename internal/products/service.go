package products

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("product not found")

type Service struct {
	db *pgxpool.Pool
}

func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

func (s *Service) Create(ctx context.Context, in CreateInput) (*Product, error) {
	if in.Currency == nil {
		def := "INR"
		in.Currency = &def
	}
	if in.Stock == nil {
		def := 0
		in.Stock = &def
	}
	if in.Active == nil {
		def := true
		in.Active = &def
	}
	if in.Metadata == nil {
		in.Metadata = map[string]any{}
	}

	id := uuid.New()
	row := s.db.QueryRow(ctx, `
		INSERT INTO products (id, name, sku, price_cents, currency, stock, active, metadata)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING id, name, sku, price_cents, currency, stock, active, metadata, created_at, updated_at
	`,
		id, in.Name, in.SKU, in.PriceCents, *in.Currency, *in.Stock, *in.Active, in.Metadata)

	var p Product
	if err := row.Scan(&p.ID, &p.Name, &p.SKU, &p.PriceCents, &p.Currency, &p.Stock, &p.Active, &p.Metadata, &p.CreatedAt, &p.UpdatedAt); err != nil {
		return nil, err
	}
	return &p, nil
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (*Product, error) {
	row := s.db.QueryRow(ctx, `
		SELECT id, name, sku, price_cents, currency, stock, active, metadata, created_at, updated_at
		FROM products WHERE id=$1
	`, id)

	var p Product
	if err := row.Scan(&p.ID, &p.Name, &p.SKU, &p.PriceCents, &p.Currency, &p.Stock, &p.Active, &p.Metadata, &p.CreatedAt, &p.UpdatedAt); err != nil {
		return nil, ErrNotFound
	}
	return &p, nil
}

func (s *Service) List(ctx context.Context, limit, offset int) ([]Product, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	rows, err := s.db.Query(ctx, `
		SELECT id, name, sku, price_cents, currency, stock, active, metadata, created_at, updated_at
		FROM products
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Product
	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.ID, &p.Name, &p.SKU, &p.PriceCents, &p.Currency, &p.Stock, &p.Active, &p.Metadata, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, in UpdateInput) (*Product, error) {
	// Build a simple partial update via COALESCE on each field
	row := s.db.QueryRow(ctx, `
		UPDATE products SET
			name        = COALESCE($2, name),
			sku         = COALESCE($3, sku),
			price_cents = COALESCE($4, price_cents),
			currency    = COALESCE($5, currency),
			stock       = COALESCE($6, stock),
			active      = COALESCE($7, active),
			metadata    = COALESCE($8, metadata)
		WHERE id=$1
		RETURNING id, name, sku, price_cents, currency, stock, active, metadata, created_at, updated_at
	`, id, in.Name, in.SKU, in.PriceCents, in.Currency, in.Stock, in.Active, in.Metadata)

	var p Product
	if err := row.Scan(&p.ID, &p.Name, &p.SKU, &p.PriceCents, &p.Currency, &p.Stock, &p.Active, &p.Metadata, &p.CreatedAt, &p.UpdatedAt); err != nil {
		return nil, ErrNotFound
	}
	return &p, nil
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	ct, err := s.db.Exec(ctx, `DELETE FROM products WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
