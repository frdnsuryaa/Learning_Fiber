package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

// RequireJSON menolak request POST, PUT, dan PATCH
// yang tidak menyertakan Content-Type: application/json.
func RequireJSON() fiber.Handler {
	return func(c *fiber.Ctx) error {
		method := c.Method()
		if method == "POST" || method == "PUT" || method == "PATCH" {
			ct := c.Get("Content-Type")
			if !strings.HasPrefix(ct, "application/json") {
				return c.Status(fiber.StatusUnsupportedMediaType).JSON(fiber.Map{
					"success": false,
					"message": "Content-Type harus application/json",
				})
			}
		}
		return c.Next()
	}
}
