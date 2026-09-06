package route

import (
	"github.com/gofiber/fiber/v2"

	"api-students/app/service"
	"api-students/middleware"
)

// SetupRoute mendaftarkan seluruh rute dan middleware pada aplikasi Fiber.
// File ini murni pendaftaran rute: tidak ada if untuk validasi maupun business rules.
func SetupRoute(
	app *fiber.App,
	systemService *service.SystemService,
	studentService *service.StudentService,
) {
	// Global Middleware
	app.Use(middleware.RequestLogger())

	// Rute Root & Health
	app.Get("/", systemService.Index)
	app.Get("/health", systemService.Health)

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
