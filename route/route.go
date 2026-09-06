package route

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"api-students/app/service"
	"api-students/middleware"
)

// SetupRoute mendaftarkan seluruh route dan middleware pada aplikasi Fiber.
func SetupRoute(app *fiber.App, pool *pgxpool.Pool, studentService *service.StudentService) {
	// Global Middleware
	app.Use(middleware.RequestLogger())

	// Route root /
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Welcome to Student API",
			"status":  "running",
		})
	})

	// Endpoint /health untuk memeriksa kondisi server dan koneksi basis data
	app.Get("/health", func(c *fiber.Ctx) error {
		pingCtx, cancel := context.WithTimeout(c.Context(), 2*time.Second)
		defer cancel()

		if err := pool.Ping(pingCtx); err != nil {
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
	})

	// Rute Mahasiswa (Students CRUD)
	students := app.Group("/students")
	students.Get("/", studentService.GetAllStudents)
	students.Get("/:id", studentService.GetStudentByID)

	// Mutasi data dilindungi middleware RequireJSON
	students.Post("/", middleware.RequireJSON(), studentService.CreateStudent)
	students.Put("/:id", middleware.RequireJSON(), studentService.UpdateStudent)
	students.Patch("/:id", middleware.RequireJSON(), studentService.PatchStudent)
	students.Delete("/:id", studentService.DeleteStudent)
}
