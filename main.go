package main

import (
	"context"
	"log"

	"github.com/gofiber/fiber/v2"

	"api-students/app/repository"
	"api-students/config"
	"api-students/database"
)

func main() {
	// Memuat file konfigurasi environment (.env)
	config.LoadEnv()

	ctx := context.Background()

	// Inisialisasi koneksi pool ke PostgreSQL
	pool, err := database.NewPool(ctx)
	if err != nil {
		log.Fatalf("Gagal inisialisasi koneksi database: %v", err)
	}
	defer pool.Close()

	// Otomatis membuat tabel students jika belum ada
	createTableQuery := `
	CREATE TABLE IF NOT EXISTS students (
		id VARCHAR(36) PRIMARY KEY,
		nim VARCHAR(20) NOT NULL,
		name VARCHAR(255) NOT NULL,
		grade DOUBLE PRECISION NOT NULL DEFAULT 0,
		is_active BOOLEAN NOT NULL DEFAULT TRUE,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		CONSTRAINT students_nim_unique UNIQUE (nim)
	);
	CREATE INDEX IF NOT EXISTS idx_students_name_lower ON students (LOWER(name));
	CREATE INDEX IF NOT EXISTS idx_students_created_at ON students (created_at DESC);
	`
	if _, err := pool.Exec(ctx, createTableQuery); err != nil {
		log.Fatalf("Gagal inisialisasi tabel students: %v", err)
	}

	// Inisialisasi Repository dan Handler
	studentRepo := repository.NewStudentRepository(pool)
	studentHandler := NewStudentHandler(studentRepo)

	app := fiber.New()

	// Route root /
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Welcome to Student API",
			"status":  "running",
		})
	})

	// Rute siswa (Students CRUD)
	app.Get("/students", studentHandler.GetAllStudents)
	app.Get("/students/:id", studentHandler.GetStudentByID)
	app.Post("/students", studentHandler.CreateStudent)
	app.Put("/students/:id", studentHandler.UpdateStudent)
	app.Patch("/students/:id", studentHandler.PatchStudent)
	app.Delete("/students/:id", studentHandler.DeleteStudent)

	port := config.GetEnv("APP_PORT", "3000")
	log.Printf("Server berjalan di http://localhost:%s", port)
	log.Fatal(app.Listen(":" + port))
}
