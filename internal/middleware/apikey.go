package middleware

import "github.com/gofiber/fiber/v2"

func APIKeyGuard(expected string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if c.Path() == "/" || c.Path() == "/health" {
			return c.Next()
		}
		if len(c.Path()) < 5 || c.Path()[:5] != "/api/" {
			return c.Next()
		}

		key := c.Get("API_KEY")
		if key == "" || key != expected {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
		}

		return c.Next()
	}
}
