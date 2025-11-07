package repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CoachRepo struct{ db *pgxpool.Pool }

func NewCoachRepo(db *pgxpool.Pool) *CoachRepo { return &CoachRepo{db: db} }

func (r *CoachRepo) UpsertPlan(ctx context.Context, id, user uuid.UUID, weekStart time.Time, rules []string, nudge string, score int) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO coach_plans (id,user_id,week_start,rules,daily_nudge,health_score)
		VALUES ($1,$2,$3,$4,$5,$6)
		ON CONFLICT (user_id, week_start)
		DO UPDATE SET rules=EXCLUDED.rules, daily_nudge=EXCLUDED.daily_nudge, health_score=EXCLUDED.health_score
	`, id, user, weekStart.Format("2006-01-02"), rules, nudge, score)
	return err
}

func (r *CoachRepo) GetPlan(ctx context.Context, user uuid.UUID, weekStart time.Time) (rules []string, nudge string, score int, found bool, err error) {
	var n *string
	err = r.db.QueryRow(ctx, `SELECT rules, daily_nudge, health_score FROM coach_plans WHERE user_id=$1 AND week_start=$2`,
		user, weekStart.Format("2006-01-02")).Scan(&rules, &n, &score)
	if err != nil {
		return nil, "", 0, false, err
	}
	if n != nil {
		nudge = *n
	}
	return rules, nudge, score, true, nil
}
