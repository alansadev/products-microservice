package middleware

import (
	"github.com/gofiber/fiber/v2"
	"os"
)

func AuthRequired() fiber.Handler {
	return func(c *fiber.Ctx) error {
		apikey := c.Get("X-API-KEY")

		secretKey := os.Getenv("API_SECRET_KEY")

		if apikey == "" || apikey != secretKey {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
		}
		return c.Next()
	}
}
