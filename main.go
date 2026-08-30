package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
)

func main() {
    app := fiber.New()

    // Route root /
    app.Get("/", func(c *fiber.Ctx) error {
        return c.JSON(fiber.Map{
            "message": "Welcome to Student API",
            "status":  "running",
        })
    })

    // Rute siswa
    app.Get("/students", GetAllStudents)
    app.Get("/students/:id", GetStudentByID)
    app.Post("/students", CreateStudent)
    app.Put("/students/:id", UpdateStudent)
    app.Patch("/students/:id", PatchStudent)
    app.Delete("/students/:id", DeleteStudent)

    log.Fatal(app.Listen(":3000"))
}

