package middleware

import (
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// RequestLogger mencatat setiap request HTTP sebagai satu baris JSON.
// Format mencakup: request_id, method, path, status, dan duration_ms.
func RequestLogger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		requestID := uuid.NewString()

		// Simpan request_id di locals agar bisa dipakai handler lain
		c.Locals("request_id", requestID)

		// Lanjutkan ke handler berikutnya
		err := c.Next()

		duration := time.Since(start)

		slog.Info("http_request",
			"request_id", requestID,
			"method", c.Method(),
			"path", c.OriginalURL(),
			"status", c.Response().StatusCode(),
			"duration_ms", duration.Milliseconds(),
		)

		return err
	}
}
