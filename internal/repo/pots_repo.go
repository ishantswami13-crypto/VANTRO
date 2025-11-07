package repo

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ishantswami13-crypto/vantro/internal/domain"
)

type PotsRepo struct{ db *pgxpool.Pool }

func NewPotsRepo(db *pgxpool.Pool) *PotsRepo { return &PotsRepo{db: db} }

func (r *PotsRepo) Insert(ctx context.Context, p domain.SavingPot) error {
	_, err := r.db.Exec(ctx, `INSERT INTO saving_pots (id,user_id,name,target_cents,saved_cents)
		VALUES ($1,$2,$3,$4,$5)`, p.ID, p.UserID, p.Name, p.TargetCents, p.SavedCents)
	return err
}

func (r *PotsRepo) AddCents(ctx context.Context, id uuid.UUID, inc int) error {
	_, err := r.db.Exec(ctx, `UPDATE saving_pots SET saved_cents = saved_cents + $1 WHERE id=$2`, inc, id)
	return err
}

func (r *PotsRepo) List(ctx context.Context, user uuid.UUID) ([]domain.SavingPot, error) {
	rows, err := r.db.Query(ctx, `SELECT id,user_id,name,target_cents,saved_cents,created_at,updated_at
		FROM saving_pots WHERE user_id=$1 ORDER BY created_at DESC`, user)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.SavingPot
	for rows.Next() {
		var p domain.SavingPot
		if err := rows.Scan(&p.ID, &p.UserID, &p.Name, &p.TargetCents, &p.SavedCents, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}
