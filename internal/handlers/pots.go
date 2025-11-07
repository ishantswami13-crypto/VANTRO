package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/ishantswami13-crypto/vantro/internal/dto"
)

func RegisterPotRoutes(g fiber.Router, svcDeps any) {
	s := svcDeps.(*Deps).Pots

	g.Post("/pots", func(c *fiber.Ctx) error {
		var in dto.PotCreate
		if err := c.BodyParser(&in); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "bad json"})
		}
		if in.Name == "" || in.TargetCents <= 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid"})
		}

		user := uuid.New() // TODO auth
		id, err := s.Create(c.Context(), user, in.Name, in.TargetCents)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": id})
	})

	g.Patch("/pots/:id", func(c *fiber.Ctx) error {
		id, err := uuid.Parse(c.Params("id"))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "bad id"})
		}
		var in dto.PotUpdate
		if err := c.BodyParser(&in); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "bad json"})
		}
		if in.AddCents == nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "add_cents required"})
		}
		if *in.AddCents < 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "must be >= 0"})
		}

		if err := s.Add(c.Context(), id, *in.AddCents); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"status": "updated"})
	})

	g.Get("/pots", func(c *fiber.Ctx) error {
		user := uuid.New() // TODO
		list, err := s.List(c.Context(), user)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(list)
	})
}
