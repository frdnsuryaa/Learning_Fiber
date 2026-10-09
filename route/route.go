package route

import (
	"github.com/gofiber/fiber/v2"

	"api-students/app/service"
	"api-students/helper"
	"api-students/middleware"
)

type Dependencies struct {
	JWT               *helper.JWTManager
	AuthService       *service.AuthService
	StudentService    *service.StudentService
	CourseService     *service.CourseService
	EnrollmentService *service.EnrollmentService
	NilaiService      *service.NilaiService
	SystemService     *service.SystemService
}

func SetupRoute(app *fiber.App, deps Dependencies) {
	app.Get("/", deps.SystemService.Index)
	app.Get("/health", deps.SystemService.Health)

	api := app.Group("/api/v1")

	// ── Auth ──────────────────────────────────────────────────────────────────
	auth := api.Group("/auth")
	auth.Post("/login", middleware.RequireJSON, middleware.LoginRateLimiter(), deps.AuthService.Login)
	auth.Get("/me", middleware.RequireAuth(deps.JWT), deps.AuthService.Me)

	// Endpoint lama dipertahankan (tidak terpakai di spec baru, tapi tidak error)
	auth.Post("/refresh", middleware.RequireJSON, deps.AuthService.Refresh)
	auth.Post("/logout", middleware.RequireJSON, deps.AuthService.Logout)

	// ── Students (semua perlu token) ──────────────────────────────────────────
	students := api.Group("/students", middleware.RequireAuth(deps.JWT))
	students.Get("/", deps.StudentService.GetAllStudents)
	students.Post("/", middleware.RequireJSON, deps.StudentService.CreateStudent)
	students.Get("/:id", deps.StudentService.GetStudentByID)
	students.Put("/:id", middleware.RequireJSON, deps.StudentService.UpdateStudent)
	students.Delete("/:id", deps.StudentService.DeleteStudent)

	// ── Courses (semua perlu token) ───────────────────────────────────────────
	courses := api.Group("/courses", middleware.RequireAuth(deps.JWT))
	courses.Get("/", deps.CourseService.GetAllCourses)

	// ── Enrollments (semua perlu token) ──────────────────────────────────────
	enrollments := api.Group("/enrollments", middleware.RequireAuth(deps.JWT))
	enrollments.Post("/", middleware.RequireJSON, deps.EnrollmentService.CreateEnrollment)
	enrollments.Delete("/:id", deps.EnrollmentService.DeleteEnrollment)

	// ── Nilai (semua perlu token) ─────────────────────────────────────────────
	nilai := api.Group("/nilai", middleware.RequireAuth(deps.JWT))
	nilai.Get("/", deps.NilaiService.GetNilai)
	nilai.Post("/", middleware.RequireJSON, deps.NilaiService.CreateNilai)
	nilai.Delete("/:id", deps.NilaiService.DeleteNilai)
}
