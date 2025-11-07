package domain

import (
	"time"

	"github.com/google/uuid"
)

type Expense struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	AmountCents int
	Category    string
	Mood        string
	Note        string
	SpentAt     time.Time
	CreatedAt   time.Time
}

type SavingPot struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Name        string
	TargetCents int
	SavedCents  int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type CoachPlan struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	WeekStart  time.Time
	Rules      []string
	DailyNudge string
	Health     int
	CreatedAt  time.Time
}
