package middleware

import (
	"log"
	"strings"

	"fintech-backend/internal/config"

	"github.com/gofiber/fiber/v2"
)

func APIKeyAuth(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Keep health unprotected
		if c.Path() == "/health" {
			return c.Next()
		}

		expected := strings.TrimSpace(cfg.APIKey)
		var key string

		// Prefer Authorization: Bearer <key>
		if auth := strings.TrimSpace(c.Get("Authorization")); strings.HasPrefix(strings.ToLower(auth), "bearer ") {
			key = strings.TrimSpace(auth[len("bearer "):])
		}

		// Fallback to X-API-Key (any casing)
		if key == "" {
			for _, h := range []string{"X-API-Key", "X-Api-Key", "x-api-key"} {
				key = strings.TrimSpace(c.Get(h))
				if key != "" {
					break
				}
			}
		}

		if expected == "" || key == "" || key != expected {
			log.Println("auth failed: missing or invalid api key")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid_api_key",
			})
		}

		return c.Next()
	}
}
