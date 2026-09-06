package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// HealthRepository mendefinisikan kontrak pemeriksaan kondisi database.
type HealthRepository interface {
	Ping(ctx context.Context) error
}

type healthPostgresRepository struct {
	pool *pgxpool.Pool
}

// NewHealthRepository membuat instance repository baru untuk pemeriksaan kesehatan database.
func NewHealthRepository(pool *pgxpool.Pool) HealthRepository {
	return &healthPostgresRepository{pool: pool}
}

// Ping melakukan ping ke koneksi database dengan batas waktu timeout 2 detik.
func (r *healthPostgresRepository) Ping(ctx context.Context) error {
	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	return r.pool.Ping(pingCtx)
}
