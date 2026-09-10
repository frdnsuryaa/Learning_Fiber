package config

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"

	"api-students/middleware"
)

func NewFiberApp() *fiber.App {
	logger := NewLogger()
	app := fiber.New(fiber.Config{
		AppName:      GetEnv("APP_NAME", "Praktikum Backend Lanjut"),
		BodyLimit:    1 * 1024 * 1024,
		ErrorHandler: newErrorHandler(logger),
	})
	middleware.Register(app, logger, GetEnv("ALLOWED_ORIGINS", ""))
	return app
}

func newErrorHandler(logger *slog.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		code := fiber.StatusInternalServerError
		if e, ok := err.(*fiber.Error); ok {
			code = e.Code
		}
		if logger != nil {
			logger.Error("request error", "error", err, "status", code)
		}
		return c.Status(code).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}
}
