package main

import (
	"context"
	"log"
	"log/slog"

	"api-students/app/repository"
	"api-students/app/service"
	"api-students/config"
	"api-students/database"
	"api-students/route"
)

func main() {
	// Memuat file konfigurasi environment (.env)
	config.LoadEnv()

	// Inisialisasi Logger dengan file logs/app.log dan rotasi
	logFile, err := config.InitLogger()
	if err != nil {
		log.Printf("Peringatan: gagal inisialisasi file logger: %v", err)
	} else {
		defer logFile.Close()
	}

	ctx := context.Background()

	// Inisialisasi koneksi pool ke PostgreSQL (termasuk Ping verifikasi)
	pool, err := database.NewPool(ctx)
	if err != nil {
		slog.Error("Gagal inisialisasi koneksi database", "error", err)
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
		slog.Error("Gagal inisialisasi tabel students", "error", err)
		log.Fatalf("Gagal inisialisasi tabel students: %v", err)
	}

	// Inisialisasi Layer Repository dan Service
	studentRepo := repository.NewStudentRepository(pool)
	studentService := service.NewStudentService(studentRepo)

	// Inisialisasi Aplikasi Fiber
	app := config.NewFiberApp()

	// Daftarkan Route
	route.SetupRoute(app, pool, studentService)

	port := config.GetEnv("APP_PORT", "3000")
	slog.Info("Server berjalan", "port", port, "url", "http://localhost:"+port)
	log.Fatal(app.Listen(":" + port))
}
