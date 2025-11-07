package services

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/ishantswami13-crypto/vantro/internal/repo"
)

type CoachService struct{ r *repo.CoachRepo }

func NewCoachService(r *repo.CoachRepo) *CoachService { return &CoachService{r: r} }

func mondayOfThisWeek(t time.Time) time.Time {
	wd := int(t.Weekday())
	if wd == 0 {
		wd = 7
	}
	return time.Date(t.Year(), t.Month(), t.Day()-wd+1, 0, 0, 0, 0, time.UTC)
}

func (s *CoachService) GeneratePlan(ctx context.Context, user uuid.UUID, income, rent int, goal string, last14Total int64) (weekStart time.Time, rules []string, nudge string, score int, err error) {
	weekStart = mondayOfThisWeek(time.Now().UTC())

	if income <= 0 {
		income = 600000
	}
	expected14 := int64(float64(income) * 0.45)

	switch r := float64(last14Total) / float64(expected14); {
	case r <= 0.6:
		score = 90
	case r <= 1.0:
		score = 80
	case r <= 1.4:
		score = 70
	default:
		score = 60
	}

	rules = []string{
		"No food delivery Mon–Thu.",
		"Pause impulse buys for 7 days.",
		"Round-up each purchase to ₹50 into your top saving pot.",
	}
	if goal != "" {
		rules[2] = "Round-up each purchase to ₹50 into your '" + goal + "' pot."
	}
	if rent > 0 && income > 0 && float64(rent)/float64(income) > 0.35 {
		rules = append(rules, "Aim to keep fixed costs under 50% this month.")
	}
	nudge = "Your money is a mirror. Keep it clear."

	err = s.r.UpsertPlan(ctx, uuid.New(), user, weekStart, rules, nudge, score)
	return
}

func (s *CoachService) GetPlan(ctx context.Context, user uuid.UUID, weekStart time.Time) (rules []string, nudge string, score int, found bool, err error) {
	return s.r.GetPlan(ctx, user, weekStart)
}
