package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"api-students/config"
)

// NewPool membuat dan mengonfigurasi connection pool ke database PostgreSQL.
// Connection pool memungkinkan penggunaan kembali (reuse) koneksi yang sudah terbuka
// sehingga melayani banyak request secara efisien tanpa overhead pembukaan koneksi baru setiap saat.
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

	// Konfigurasi pool
	cfg.MaxConns = int32(config.GetEnvInt("DB_MAX_CONNS", 10))
	cfg.MinConns = 2
	cfg.MaxConnLifetime = time.Hour
	cfg.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat connection pool: %w", err)
	}

	// Melakukan Ping untuk memastikan server PostgreSQL benar-benar dapat dihubungi
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("gagal terhubung ke database (ping gagal): %w", err)
	}

	return pool, nil
}