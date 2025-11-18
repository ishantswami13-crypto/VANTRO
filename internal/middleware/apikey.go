package middleware

import (
	"fintech-backend/internal/config"

	"github.com/gofiber/fiber/v2"
)

func APIKeyAuth(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		apiKey := c.Get("X-API-Key")
		if apiKey == "" || apiKey != cfg.APIKey {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid or missing API key",
			})
		}
		return c.Next()
	}
}
