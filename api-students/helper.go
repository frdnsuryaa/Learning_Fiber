package main

import "github.com/gofiber/fiber/v2"

// Response adalah amplop (envelope) seragam untuk semua respons API.
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}
// SuccessResponse mengirim respons sukses dengan amplop seragam.
func SuccessResponse(c *fiber.Ctx, statusCode int, message string, data interface{}) error {
	return c.Status(statusCode).JSON(Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}
// ErrorResponse mengirim respons gagal dengan amplop seragam (data selalu null).
func ErrorResponse(c *fiber.Ctx, statusCode int, message string) error {
	return c.Status(statusCode).JSON(Response{
		Success: false,
		Message: message,
		Data:    nil,
	})
}