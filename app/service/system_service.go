package service

import (
	"github.com/gofiber/fiber/v2"

	"api-students/app/repository"
)

// SystemService menangani permintaan HTTP tingkat sistem (root dan health check).
type SystemService struct {
	healthRepo repository.HealthRepository
}

// NewSystemService membuat instance baru SystemService.
func NewSystemService(healthRepo repository.HealthRepository) *SystemService {
	return &SystemService{healthRepo: healthRepo}
}

// Index menangani GET / (halaman pembuka/root).
func (s *SystemService) Index(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"message": "Welcome to Student API",
		"status":  "running",
	})
}

// Health menangani GET /health untuk memeriksa kesiapan server dan database.
func (s *SystemService) Health(c *fiber.Ctx) error {
	if err := s.healthRepo.Ping(c.Context()); err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"status":   "error",
			"database": "disconnected",
			"message":  "Koneksi ke basis data gagal",
			"error":    err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":   "ok",
		"database": "connected",
		"message":  "Server dan basis data berjalan dengan baik",
	})
}
