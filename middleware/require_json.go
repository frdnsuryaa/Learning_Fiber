package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

// RequireJSON menolak request POST, PUT, dan PATCH jika Content-Type bukan application/json.
func RequireJSON(c *fiber.Ctx) error {
	if c.Method() == fiber.MethodGet || c.Method() == fiber.MethodDelete {
		return c.Next()
	}
	contentType := strings.ToLower(c.Get("Content-Type"))
	if !strings.HasPrefix(contentType, "application/json") {
		return fiber.NewError(fiber.StatusUnsupportedMediaType, "Content-Type harus application/json")
	}
	return c.Next()
}
