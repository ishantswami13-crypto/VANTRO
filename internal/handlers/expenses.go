package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/ishantswami13-crypto/vantro/internal/dto"
)

func RegisterExpenseRoutes(g fiber.Router, svcDeps any) {
	s := svcDeps.(*Deps).Expenses

	g.Post("/expenses", func(c *fiber.Ctx) error {
		var in dto.ExpenseCreate
		if err := c.BodyParser(&in); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "bad json"})
		}
		if in.AmountCents < 0 || in.Category == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid"})
		}

		user := uuid.New() // TODO replace with authenticated user id
		id, err := s.Create(c.Context(), user, in.AmountCents, in.Category, in.Mood, in.Note)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": id})
	})

	g.Get("/expenses", func(c *fiber.Ctx) error {
		user := uuid.New() // TODO
		fromStr := c.Query("from")
		toStr := c.Query("to")
		if fromStr == "" || toStr == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "from/to required (YYYY-MM-DD)"})
		}
		from, err1 := time.Parse("2006-01-02", fromStr)
		to, err2 := time.Parse("2006-01-02", toStr)
		if err1 != nil || err2 != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "bad date"})
		}

		list, err := s.List(c.Context(), user, from, to.Add(24*time.Hour))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(list)
	})
}
