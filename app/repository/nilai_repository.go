package repository

import (
	"context"
	"errors"
	"fmt"

	"api-students/app/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// NilaiRepository mendefinisikan operasi database untuk tabel nilai.
type NilaiRepository interface {
	FindByStudentID(ctx context.Context, studentID int) ([]model.Nilai, error)
	Create(ctx context.Context, n model.Nilai) (model.Nilai, error)
	Delete(ctx context.Context, idNilai int, studentID int) error
}

type nilaiPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewNilaiRepository(pool *pgxpool.Pool) NilaiRepository {
	return &nilaiPostgresRepository{pool: pool}
}

// FindByStudentID mengambil semua nilai milik student tertentu.
func (r *nilaiPostgresRepository) FindByStudentID(ctx context.Context, studentID int) ([]model.Nilai, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT idnilai, namamatkul, nilai, idstudent, created_at
		FROM nilai
		WHERE idstudent = (SELECT id::text FROM students WHERE id = $1 AND deleted_at IS NULL)
		ORDER BY created_at DESC
	`, studentID)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil nilai: %w", err)
	}
	defer rows.Close()

	var result []model.Nilai
	for rows.Next() {
		var n model.Nilai
		if err := rows.Scan(&n.IDNilai, &n.NamaMatkul, &n.Nilai, &n.IDStudent, &n.CreatedAt); err != nil {
			return nil, fmt.Errorf("gagal membaca baris nilai: %w", err)
		}
		result = append(result, n)
	}
	return result, rows.Err()
}

// Create menyimpan nilai baru.
func (r *nilaiPostgresRepository) Create(ctx context.Context, n model.Nilai) (model.Nilai, error) {
	err := r.pool.QueryRow(ctx, `
		INSERT INTO nilai (namamatkul, nilai, idstudent)
		VALUES ($1, $2, $3)
		RETURNING idnilai, namamatkul, nilai, idstudent, created_at
	`, n.NamaMatkul, n.Nilai, n.IDStudent).
		Scan(&n.IDNilai, &n.NamaMatkul, &n.Nilai, &n.IDStudent, &n.CreatedAt)
	if err != nil {
		return model.Nilai{}, fmt.Errorf("gagal menyimpan nilai: %w", err)
	}
	return n, nil
}

// Delete menghapus nilai berdasarkan idnilai dan memverifikasi kepemilikan student.
func (r *nilaiPostgresRepository) Delete(ctx context.Context, idNilai int, studentID int) error {
	tag, err := r.pool.Exec(ctx, `
		DELETE FROM nilai
		WHERE idnilai = $1
		  AND idstudent = (SELECT id::text FROM students WHERE id = $2 AND deleted_at IS NULL)
	`, idNilai, studentID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return fmt.Errorf("gagal menghapus nilai: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
