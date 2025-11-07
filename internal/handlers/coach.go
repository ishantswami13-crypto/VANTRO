package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/ishantswami13-crypto/vantro/internal/dto"
)

func RegisterCoachRoutes(g fiber.Router, svcDeps any) {
	s := svcDeps.(*Deps).Coach
	exp := svcDeps.(*Deps).Expenses

	g.Post("/coach/plan", func(c *fiber.Ctx) error {
		var in dto.CoachInput
		if err := c.BodyParser(&in); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "bad json"})
		}
		user := uuid.New() // TODO auth

		now := time.Now().UTC()
		from := now.Add(-14 * 24 * time.Hour)

		list, err := exp.List(c.Context(), user, from, now)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		var total14 int64
		for _, e := range list {
			total14 += int64(e.AmountCents)
		}

		weekStart, rules, nudge, score, err := s.GeneratePlan(c.Context(), user, in.IncomeCents, in.RentCents, in.Goal, total14)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(fiber.Map{
			"week_start":   weekStart.Format("2006-01-02"),
			"rules":        rules,
			"daily_nudge":  nudge,
			"health_score": score,
		})
	})

	g.Get("/coach/plan", func(c *fiber.Ctx) error {
		user := uuid.New()
		week := mondayOfThisWeek(time.Now().UTC())
		rules, nudge, score, found, err := s.GetPlan(c.Context(), user, week)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		if !found {
			return c.JSON(fiber.Map{
				"week_start": week.Format("2006-01-02"), "rules": []string{}, "daily_nudge": "", "health_score": 0,
			})
		}
		return c.JSON(fiber.Map{
			"week_start": week.Format("2006-01-02"), "rules": rules, "daily_nudge": nudge, "health_score": score,
		})
	})
}

func mondayOfThisWeek(t time.Time) time.Time {
	wd := int(t.Weekday())
	if wd == 0 {
		wd = 7
	}
	return time.Date(t.Year(), t.Month(), t.Day()-wd+1, 0, 0, 0, 0, time.UTC)
}
