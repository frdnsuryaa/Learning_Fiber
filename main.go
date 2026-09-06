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
	// 1. Memuat konfigurasi environment (.env)
	config.LoadEnv()

	// 2. Inisialisasi Logger (output terminal + file logs/app.log)
	logFile, err := config.InitLogger()
	if err != nil {
		log.Printf("Peringatan: gagal inisialisasi file logger: %v", err)
	} else {
		defer logFile.Close()
	}

	ctx := context.Background()

	// 3. Inisialisasi koneksi database PostgreSQL pool
	pool, err := database.NewPool(ctx)
	if err != nil {
		slog.Error("Gagal inisialisasi koneksi database", "error", err)
		log.Fatalf("Gagal inisialisasi koneksi database: %v", err)
	}
	defer pool.Close()

	// 4. Migrasi skema database tabel students
	if err := database.Migrate(ctx, pool); err != nil {
		slog.Error("Gagal inisialisasi tabel students", "error", err)
		log.Fatalf("Gagal inisialisasi tabel students: %v", err)
	}

	// 5. Inisialisasi Layer Repository
	healthRepo := repository.NewHealthRepository(pool)
	studentRepo := repository.NewStudentRepository(pool)

	// 6. Inisialisasi Layer Service
	systemService := service.NewSystemService(healthRepo)
	studentService := service.NewStudentService(studentRepo)

	// 7. Inisialisasi Aplikasi Web Fiber
	app := config.NewFiberApp()

	// 8. Pendaftaran Rute
	route.SetupRoute(app, systemService, studentService)

	// 9. Menjalankan Server
	port := config.GetEnv("APP_PORT", "3000")
	slog.Info("Server berjalan", "port", port, "url", "http://localhost:"+port)
	log.Fatal(app.Listen(":" + port))
}
