package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"api-students/config"
)

func NewPool(ctx context.Context) (*pgxpool.Pool, error) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		config.GetEnv("DB_USER", "postgres"),
		config.GetEnv("DB_PASSWORD", ""),
		config.GetEnv("DB_HOST", "localhost"),
		config.GetEnv("DB_PORT", "5432"),
		config.GetEnv("DB_NAME", "praktikum_backend"),
		config.GetEnv("DB_SSLMODE", "disable"),
	)

	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("konfigurasi database tidak valid: %w", err)
	}

	cfg.MaxConns = int32(config.GetEnvInt("DB_MAX_CONNS", 10))
	cfg.MinConns = 2
	cfg.MaxConnLifetime = time.Hour
	cfg.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat connection pool: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("gagal terhubung ke database (ping gagal): %w", err)
	}
	return pool, nil
}

// Migrate membuat semua tabel yang dibutuhkan jika belum ada.
func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	query := `
	-- Tabel users: menyimpan akun admin dan mahasiswa.
	CREATE TABLE IF NOT EXISTS users (
		id         SERIAL PRIMARY KEY,
		email      VARCHAR(255) NOT NULL,
		password   VARCHAR(255) NOT NULL,
		role       VARCHAR(20)  NOT NULL DEFAULT 'mahasiswa',
		created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
	);
	CREATE UNIQUE INDEX IF NOT EXISTS users_email_unique ON users (LOWER(email));

	-- Tabel students: data profil mahasiswa, 1-1 ke users.
	CREATE TABLE IF NOT EXISTS students (
		id           SERIAL PRIMARY KEY,
		user_id      INTEGER      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		nim          VARCHAR(12)  NOT NULL,
		nama         VARCHAR(255) NOT NULL,
		prodi        VARCHAR(255) NOT NULL,
		angkatan     INTEGER      NOT NULL,
		ipk_terakhir NUMERIC(3,2) NOT NULL DEFAULT 0.00,
		deleted_at   TIMESTAMPTZ,
		CONSTRAINT students_nim_unique    UNIQUE (nim),
		CONSTRAINT students_user_id_unique UNIQUE (user_id)
	);
	CREATE INDEX IF NOT EXISTS idx_students_deleted_at ON students (deleted_at);

	-- Tabel courses: mata kuliah.
	CREATE TABLE IF NOT EXISTS courses (
		id       SERIAL PRIMARY KEY,
		kode_mk  VARCHAR(20)  NOT NULL,
		nama_mk  VARCHAR(255) NOT NULL,
		sks      INTEGER      NOT NULL,
		semester INTEGER      NOT NULL,
		kuota    INTEGER      NOT NULL DEFAULT 30,
		CONSTRAINT courses_kode_mk_unique UNIQUE (kode_mk)
	);

	-- Tabel enrollments: KRS mahasiswa.
	CREATE TABLE IF NOT EXISTS enrollments (
		id             SERIAL PRIMARY KEY,
		student_id     INTEGER     NOT NULL REFERENCES students(id) ON DELETE CASCADE,
		course_id      INTEGER     NOT NULL REFERENCES courses(id)  ON DELETE CASCADE,
		tahun_akademik VARCHAR(30) NOT NULL,
		created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		CONSTRAINT enrollments_unique UNIQUE (student_id, course_id, tahun_akademik)
	);
	CREATE INDEX IF NOT EXISTS idx_enrollments_student_id ON enrollments (student_id);
	CREATE INDEX IF NOT EXISTS idx_enrollments_course_id  ON enrollments (course_id);

	-- Tabel nilai: nilai mahasiswa per mata kuliah.
	CREATE TABLE IF NOT EXISTS nilai (
		idnilai    SERIAL PRIMARY KEY,
		namamatkul VARCHAR(100) NOT NULL,
		nilai      VARCHAR(10)  NOT NULL,
		idstudent  INTEGER      NOT NULL,
		created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
		CONSTRAINT fk_nilai_student FOREIGN KEY (idstudent) REFERENCES students(id) ON DELETE CASCADE
	);
	CREATE INDEX IF NOT EXISTS idx_nilai_idstudent ON nilai (idstudent);

	-- Tabel refresh_tokens (autentikasi JWT).
	CREATE TABLE IF NOT EXISTS refresh_tokens (
		id         BIGSERIAL PRIMARY KEY,
		user_id    INTEGER  NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		token_hash TEXT     NOT NULL UNIQUE,
		expires_at TIMESTAMPTZ NOT NULL,
		revoked_at TIMESTAMPTZ,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);
	CREATE INDEX IF NOT EXISTS refresh_tokens_user_id_idx ON refresh_tokens (user_id);
	`
	if _, err := pool.Exec(ctx, query); err != nil {
		return fmt.Errorf("gagal migrasi skema database: %w", err)
	}
	return nil
}

// Seed mengisi data awal: 1 admin, 20 mahasiswa, 10 mata kuliah.
// Idempoten — tidak gagal jika data sudah ada.
func Seed(ctx context.Context, pool *pgxpool.Pool) error {
	return seedAll(ctx, pool)
}
