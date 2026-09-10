package route

import (
	"github.com/gofiber/fiber/v2"

	"api-students/app/service"
	"api-students/helper"
	"api-students/middleware"
)

type Dependencies struct {
	JWT            *helper.JWTManager
	AuthService    *service.AuthService
	StudentService *service.StudentService
	SystemService  *service.SystemService
}

func SetupRoute(app *fiber.App, deps Dependencies) {
	app.Get("/", deps.SystemService.Index)
	app.Get("/health", deps.SystemService.Health)

	api := app.Group("/api/v1")

	// Endpoint autentikasi.
	auth := api.Group("/auth")
	auth.Post("/register", middleware.RequireJSON, deps.AuthService.Register)
	auth.Post("/login", middleware.RequireJSON, middleware.LoginRateLimiter(), deps.AuthService.Login)
	auth.Post("/refresh", middleware.RequireJSON, deps.AuthService.Refresh)
	auth.Post("/logout", middleware.RequireJSON, deps.AuthService.Logout)
	auth.Get("/me", middleware.RequireAuth(deps.JWT), deps.AuthService.Me)

	// Seluruh endpoint students wajib membawa access token.
	students := api.Group("/students", middleware.RequireAuth(deps.JWT))
	students.Get("/", deps.StudentService.GetAllStudents)
	students.Get("/:id", deps.StudentService.GetStudentByID)
	students.Post("/", middleware.RequireJSON, deps.StudentService.CreateStudent)
	students.Put("/:id", middleware.RequireJSON, deps.StudentService.UpdateStudent)
	students.Patch("/:id", middleware.RequireJSON, deps.StudentService.PatchStudent)
	students.Delete("/:id", deps.StudentService.DeleteStudent)
}
