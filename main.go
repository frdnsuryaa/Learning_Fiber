package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"time"

	"api-students/app/repository"
	"api-students/app/service"
	"api-students/config"
	"api-students/database"
	"api-students/helper"
	"api-students/route"
)

const minSecretLength = 32

func main() {
	config.LoadEnv()
	logger := config.NewLogger()

	jwtSecret := config.GetEnv("JWT_SECRET", "")
	if len(jwtSecret) < minSecretLength {
		logger.Error("JWT_SECRET tidak diisi atau terlalu pendek",
			slog.Int("minimal_karakter", minSecretLength))
		os.Exit(1)
	}

	ctx := context.Background()
	pool, err := database.NewPool(ctx)
	if err != nil {
		logger.Error("gagal terhubung ke database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	// Jalankan migrasi
	if err := database.Migrate(ctx, pool); err != nil {
		logger.Error("gagal migrasi database", "error", err)
		os.Exit(1)
	}

	// Jalankan seeder (idempoten)
	if err := database.Seed(ctx, pool); err != nil {
		logger.Error("gagal menjalankan seeder", "error", err)
		// Tidak exit — seeder gagal tidak seharusnya menghentikan server
	}

	jwtManager := helper.NewJWTManager(
		jwtSecret,
		config.GetEnv("JWT_ISSUER", "praktikum-backend"),
		time.Duration(config.GetEnvInt("JWT_ACCESS_TTL_MINUTES", 15))*time.Minute,
	)

	// Repositories
	userRepository := repository.NewUserRepository(pool)
	tokenRepository := repository.NewTokenRepository(pool)
	studentRepository := repository.NewStudentRepository(pool)
	courseRepository := repository.NewCourseRepository(pool)
	enrollmentRepository := repository.NewEnrollmentRepository(pool)
	nilaiRepository := repository.NewNilaiRepository(pool)
	healthRepository := repository.NewHealthRepository(pool)

	// Services
	systemService := service.NewSystemService(healthRepository)
	authService := service.NewAuthService(
		userRepository,
		tokenRepository,
		studentRepository,
		jwtManager,
		time.Duration(config.GetEnvInt("JWT_REFRESH_TTL_DAYS", 7))*24*time.Hour,
	)
	studentService := service.NewStudentService(studentRepository)
	courseService := service.NewCourseService(courseRepository)
	enrollmentService := service.NewEnrollmentService(enrollmentRepository, studentRepository, courseRepository)
	nilaiService := service.NewNilaiService(nilaiRepository, studentRepository)

	app := config.NewFiberApp()
	route.SetupRoute(app, route.Dependencies{
		JWT:               jwtManager,
		AuthService:       authService,
		StudentService:    studentService,
		CourseService:     courseService,
		EnrollmentService: enrollmentService,
		NilaiService:      nilaiService,
		SystemService:     systemService,
	})

	port := config.GetEnv("APP_PORT", "3000")
	slog.Info("Server berjalan", "port", port, "url", "http://localhost:"+port)
	log.Fatal(app.Listen(":" + port))
}
