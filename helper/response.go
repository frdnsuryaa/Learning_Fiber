package helper

import "github.com/gofiber/fiber/v2"

func Success(c *fiber.Ctx, status int, message string, data any) error {
	return c.Status(status).JSON(fiber.Map{
		"success": true,
		"message": message,
		"data":    data,
	})
}

func Created(c *fiber.Ctx, message string, data any, location string) error {
	if location != "" {
		c.Set("Location", location)
	}
	return Success(c, fiber.StatusCreated, message, data)
}

func SuccessList(c *fiber.Ctx, message string, data any, meta any) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": message,
		"data":    data,
		"meta":    meta,
	})
}

func Fail(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(fiber.Map{
		"success": false,
		"message": message,
	})
}

// FailValidation mengembalikan 422 dengan format errors map[string][]string sesuai spec.
func FailValidation(c *fiber.Ctx, errs map[string]string) error {
	// Konversi ke map[string][]string sesuai format spec
	converted := make(map[string][]string, len(errs))
	for k, v := range errs {
		converted[k] = []string{v}
	}
	return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
		"success": false,
		"message": "Validasi gagal",
		"errors":  converted,
	})
}
