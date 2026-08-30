package main

import "github.com/gofiber/fiber/v2"

// Response adalah amplop (envelope) seragam untuk semua respons API.
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// Meta berisi informasi paginasi yang disertakan dalam respons list.
type Meta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// PaginatedResponse adalah amplop khusus untuk endpoint list yang menyertakan meta paginasi.
type PaginatedResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Meta    Meta        `json:"meta"`
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

// PaginatedSuccessResponse mengirim respons sukses untuk endpoint list beserta meta paginasi.
func PaginatedSuccessResponse(c *fiber.Ctx, message string, meta Meta, data interface{}) error {
	return c.Status(fiber.StatusOK).JSON(PaginatedResponse{
		Success: true,
		Message: message,
		Meta:    meta,
		Data:    data,
	})
}