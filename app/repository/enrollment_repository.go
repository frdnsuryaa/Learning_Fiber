package repository

import (
	"context"
	"errors"
	"fmt"

	"api-students/app/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrKuotaPenuh dikembalikan ketika kuota mata kuliah sudah habis.
var ErrKuotaPenuh = errors.New("kuota mata kuliah sudah penuh")

// ErrSKSMelebihi dikembalikan ketika total SKS melebihi batas.
var ErrSKSMelebihi = errors.New("total SKS melebihi batas")

// EnrollmentRepository mendefinisikan operasi database untuk tabel enrollments.
type EnrollmentRepository interface {
	Create(ctx context.Context, e model.Enrollment, batasSKS int) (model.Enrollment, error)
	FindByID(ctx context.Context, id int) (model.Enrollment, error)
	Delete(ctx context.Context, id int, studentID int) error
}

type enrollmentPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewEnrollmentRepository(pool *pgxpool.Pool) EnrollmentRepository {
	return &enrollmentPostgresRepository{pool: pool}
}

// Create melakukan insert enrollment di dalam transaction dengan row locking.
// Memeriksa: duplikasi, kuota penuh, batas SKS.
func (r *enrollmentPostgresRepository) Create(ctx context.Context, e model.Enrollment, batasSKS int) (model.Enrollment, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return model.Enrollment{}, fmt.Errorf("gagal memulai transaksi: %w", err)
	}
	defer tx.Rollback(ctx)

	// 1. Ambil data course dengan row lock untuk cek kuota
	var kuota int
	var sks int
	err = tx.QueryRow(ctx, `
		SELECT c.kuota, c.sks,
		       (SELECT COUNT(*) FROM enrollments WHERE course_id = c.id) AS terisi
		FROM courses c
		WHERE c.id = $1
		FOR UPDATE
	`, e.CourseID).Scan(&kuota, &sks, new(int))
	// Kita butuh terisi secara terpisah
	var terisi int
	err = tx.QueryRow(ctx, `
		SELECT c.kuota, c.sks,
		       COALESCE((SELECT COUNT(*) FROM enrollments WHERE course_id = c.id), 0)
		FROM courses c
		WHERE c.id = $1
		FOR UPDATE
	`, e.CourseID).Scan(&kuota, &sks, &terisi)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Enrollment{}, ErrNotFound
		}
		return model.Enrollment{}, fmt.Errorf("gagal mengambil data course: %w", err)
	}

	// 2. Cek kuota
	if terisi >= kuota {
		return model.Enrollment{}, ErrKuotaPenuh
	}

	// 3. Hitung total SKS mahasiswa pada tahun akademik ini
	var totalSKS int
	tx.QueryRow(ctx, `
		SELECT COALESCE(SUM(c.sks), 0)
		FROM enrollments e
		JOIN courses c ON c.id = e.course_id
		WHERE e.student_id = $1 AND e.tahun_akademik = $2
	`, e.StudentID, e.TahunAkademik).Scan(&totalSKS)

	if totalSKS+sks > batasSKS {
		return model.Enrollment{}, fmt.Errorf("%w: sisa %d SKS, mata kuliah butuh %d SKS",
			ErrSKSMelebihi, batasSKS-totalSKS, sks)
	}

	// 4. Insert
	var created model.Enrollment
	err = tx.QueryRow(ctx, `
		INSERT INTO enrollments (student_id, course_id, tahun_akademik)
		VALUES ($1, $2, $3)
		RETURNING id, student_id, course_id, tahun_akademik, created_at
	`, e.StudentID, e.CourseID, e.TahunAkademik).Scan(
		&created.ID, &created.StudentID, &created.CourseID,
		&created.TahunAkademik, &created.CreatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return model.Enrollment{}, ErrDuplicate
		}
		return model.Enrollment{}, fmt.Errorf("gagal insert enrollment: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return model.Enrollment{}, fmt.Errorf("gagal commit transaksi: %w", err)
	}
	return created, nil
}

func (r *enrollmentPostgresRepository) FindByID(ctx context.Context, id int) (model.Enrollment, error) {
	var e model.Enrollment
	err := r.pool.QueryRow(ctx, `
		SELECT id, student_id, course_id, tahun_akademik, created_at
		FROM enrollments WHERE id = $1
	`, id).Scan(&e.ID, &e.StudentID, &e.CourseID, &e.TahunAkademik, &e.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Enrollment{}, ErrNotFound
		}
		return model.Enrollment{}, fmt.Errorf("gagal mencari enrollment: %w", err)
	}
	return e, nil
}

// Delete menghapus enrollment milik student tertentu.
// Mengembalikan ErrNotFound jika tidak ada, atau error lain jika enrollment bukan milik student.
func (r *enrollmentPostgresRepository) Delete(ctx context.Context, id int, studentID int) error {
	// Cek kepemilikan dulu
	e, err := r.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if e.StudentID != studentID {
		return ErrForbidden
	}

	_, err = r.pool.Exec(ctx, `DELETE FROM enrollments WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("gagal delete enrollment: %w", err)
	}
	return nil
}
