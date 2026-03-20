package business

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	Repo *Repo
}

func NewHandler(repo *Repo) *Handler { return &Handler{Repo: repo} }

func (h *Handler) Create(c *fiber.Ctx) error {
	var req CreateBusinessRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid JSON: " + err.Error()})
	}

	ctx, cancel := context.WithTimeout(c.Context(), 8*time.Second)
	defer cancel()

	resp, err := h.Repo.CreateWithDefaults(ctx, req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}
