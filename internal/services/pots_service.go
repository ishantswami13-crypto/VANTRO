package services

import (
	"context"

	"github.com/google/uuid"

	"github.com/ishantswami13-crypto/vantro/internal/domain"
	"github.com/ishantswami13-crypto/vantro/internal/repo"
)

type PotsService struct{ r *repo.PotsRepo }

func NewPotsService(r *repo.PotsRepo) *PotsService { return &PotsService{r: r} }

func (s *PotsService) Create(ctx context.Context, user uuid.UUID, name string, target int) (uuid.UUID, error) {
	id := uuid.New()
	err := s.r.Insert(ctx, domain.SavingPot{
		ID: id, UserID: user, Name: name, TargetCents: target, SavedCents: 0,
	})
	return id, err
}

func (s *PotsService) Add(ctx context.Context, potID uuid.UUID, inc int) error {
	return s.r.AddCents(ctx, potID, inc)
}

func (s *PotsService) List(ctx context.Context, user uuid.UUID) ([]domain.SavingPot, error) {
	return s.r.List(ctx, user)
}
