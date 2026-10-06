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

	// Serialisasi enrollment mahasiswa agar dua request bersamaan tidak melewati batas SKS.
	var studentID int
	err = tx.QueryRow(ctx, `
		SELECT id
		FROM students
		WHERE id = $1 AND deleted_at IS NULL
		FOR UPDATE
	`, e.StudentID).Scan(&studentID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Enrollment{}, ErrNotFound
		}
		return model.Enrollment{}, fmt.Errorf("gagal mengunci data mahasiswa: %w", err)
	}

	// Mengunci mata kuliah agar request bersamaan tidak melebihi kuota.
	var kuota int
	var sks int
	err = tx.QueryRow(ctx, `
		SELECT kuota, sks
		FROM courses c
		WHERE c.id = $1
		FOR UPDATE
	`, e.CourseID).Scan(&kuota, &sks)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Enrollment{}, ErrNotFound
		}
		return model.Enrollment{}, fmt.Errorf("gagal mengambil data course: %w", err)
	}

	var terisi int
	err = tx.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM enrollments
		WHERE course_id = $1
	`, e.CourseID).Scan(&terisi)
	if err != nil {
		return model.Enrollment{}, fmt.Errorf("gagal menghitung enrollment mata kuliah: %w", err)
	}

	if terisi >= kuota {
		return model.Enrollment{}, ErrKuotaPenuh
	}

	var totalSKS int
	err = tx.QueryRow(ctx, `
		SELECT COALESCE(SUM(c.sks), 0)
		FROM enrollments e
		JOIN courses c ON c.id = e.course_id
		WHERE e.student_id = $1 AND e.tahun_akademik = $2
	`, e.StudentID, e.TahunAkademik).Scan(&totalSKS)
	if err != nil {
		return model.Enrollment{}, fmt.Errorf("gagal menghitung total SKS mahasiswa: %w", err)
	}

	if totalSKS+sks > batasSKS {
		return model.Enrollment{}, fmt.Errorf("%w: sisa %d SKS, mata kuliah butuh %d SKS",
			ErrSKSMelebihi, batasSKS-totalSKS, sks)
	}

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
