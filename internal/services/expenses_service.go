package services

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/ishantswami13-crypto/vantro/internal/domain"
	"github.com/ishantswami13-crypto/vantro/internal/repo"
)

type ExpensesService struct{ r *repo.ExpensesRepo }

func NewExpensesService(r *repo.ExpensesRepo) *ExpensesService { return &ExpensesService{r: r} }

func (s *ExpensesService) Create(ctx context.Context, user uuid.UUID, amount int, category, mood, note string) (uuid.UUID, error) {
	id := uuid.New()
	return id, s.r.Insert(ctx, domain.Expense{
		ID: id, UserID: user, AmountCents: amount, Category: category, Mood: mood, Note: note, SpentAt: time.Now().UTC(),
	})
}

func (s *ExpensesService) List(ctx context.Context, user uuid.UUID, from, to time.Time) ([]domain.Expense, error) {
	return s.r.ListBetween(ctx, user, from, to)
}
