package middleware

import (
	"os"

	"github.com/gofiber/fiber/v2"
)

func InternalAuthMiddleware(c *fiber.Ctx) error {
	incomingSecret := c.Get("x-Internal-Secret")

	expectedSecret := os.Getenv("INTERNAL_SECRET")

	if expectedSecret == "" {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    500,
			"status":  "Internal Server Error",
			"message": "internal secret is not configured",
		})
	}

	if incomingSecret == "" || incomingSecret != expectedSecret {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"code":    401,
			"status":  "Unauthorized",
			"message": "invalid or missing internal secret",
		})

	}
	return c.Next()
}
