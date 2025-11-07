package repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ishantswami13-crypto/vantro/internal/domain"
)

type ExpensesRepo struct{ db *pgxpool.Pool }

func NewExpensesRepo(db *pgxpool.Pool) *ExpensesRepo { return &ExpensesRepo{db: db} }

func (r *ExpensesRepo) Insert(ctx context.Context, e domain.Expense) error {
	_, err := r.db.Exec(ctx, `INSERT INTO expenses 
		(id, user_id, amount_cents, category, mood, note, spent_at) 
		VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		e.ID, e.UserID, e.AmountCents, e.Category, e.Mood, e.Note, e.SpentAt)
	return err
}

func (r *ExpensesRepo) ListBetween(ctx context.Context, user uuid.UUID, from, to time.Time) ([]domain.Expense, error) {
	rows, err := r.db.Query(ctx, `SELECT id,user_id,amount_cents,category,mood,note,spent_at,created_at
		FROM expenses WHERE user_id=$1 AND spent_at BETWEEN $2 AND $3
		ORDER BY spent_at DESC`, user, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Expense
	for rows.Next() {
		var x domain.Expense
		if err := rows.Scan(&x.ID, &x.UserID, &x.AmountCents, &x.Category, &x.Mood, &x.Note, &x.SpentAt, &x.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}
