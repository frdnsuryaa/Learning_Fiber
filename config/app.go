package config

import (
	"github.com/gofiber/fiber/v2"
)

// NewFiberApp membuat instance Fiber dengan konfigurasi standar.
func NewFiberApp() *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:      "Student API",
		ServerHeader: "api-students",
	})

	return app
}
