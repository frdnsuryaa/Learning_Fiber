package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"api-students/app/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CourseRepository mendefinisikan operasi database untuk tabel courses.
type CourseRepository interface {
	FindAll(ctx context.Context, q model.CourseQuery) ([]model.Course, error)
	FindByID(ctx context.Context, id int) (model.Course, error)
}

type coursePostgresRepository struct {
	pool *pgxpool.Pool
}

func NewCourseRepository(pool *pgxpool.Pool) CourseRepository {
	return &coursePostgresRepository{pool: pool}
}

func (r *coursePostgresRepository) FindAll(ctx context.Context, q model.CourseQuery) ([]model.Course, error) {
	var conditions []string
	var args []interface{}
	argIdx := 1

	if q.Search != "" {
		conditions = append(conditions, fmt.Sprintf(
			"(LOWER(kode_mk) LIKE $%d OR LOWER(nama_mk) LIKE $%d)", argIdx, argIdx))
		args = append(args, "%"+strings.ToLower(q.Search)+"%")
		argIdx++
	}

	if q.Semester != nil {
		conditions = append(conditions, fmt.Sprintf("semester = $%d", argIdx))
		args = append(args, *q.Semester)
		argIdx++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " WHERE " + strings.Join(conditions, " AND ")
	}

	// Hitung terisi dari enrollments
	query := fmt.Sprintf(`
		SELECT c.id, c.kode_mk, c.nama_mk, c.sks, c.semester, c.kuota,
		       COUNT(e.id) AS terisi,
		       c.kuota - COUNT(e.id) AS sisa_kuota
		FROM courses c
		LEFT JOIN enrollments e ON e.course_id = c.id
		%s
		GROUP BY c.id, c.kode_mk, c.nama_mk, c.sks, c.semester, c.kuota
		HAVING ($%d = FALSE OR (c.kuota - COUNT(e.id)) > 0)
		ORDER BY c.kode_mk
	`, whereClause, argIdx)
	args = append(args, q.Available)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("gagal query courses: %w", err)
	}
	defer rows.Close()

	courses := make([]model.Course, 0)
	for rows.Next() {
		var c model.Course
		if err := rows.Scan(&c.ID, &c.KodeMK, &c.NamaMK, &c.SKS, &c.Semester, &c.Kuota, &c.Terisi, &c.SisaKuota); err != nil {
			return nil, fmt.Errorf("gagal scan course: %w", err)
		}
		courses = append(courses, c)
	}
	return courses, rows.Err()
}

func (r *coursePostgresRepository) FindByID(ctx context.Context, id int) (model.Course, error) {
	var c model.Course
	err := r.pool.QueryRow(ctx, `
		SELECT c.id, c.kode_mk, c.nama_mk, c.sks, c.semester, c.kuota,
		       COUNT(e.id) AS terisi,
		       c.kuota - COUNT(e.id) AS sisa_kuota
		FROM courses c
		LEFT JOIN enrollments e ON e.course_id = c.id
		WHERE c.id = $1
		GROUP BY c.id, c.kode_mk, c.nama_mk, c.sks, c.semester, c.kuota
	`, id).Scan(&c.ID, &c.KodeMK, &c.NamaMK, &c.SKS, &c.Semester, &c.Kuota, &c.Terisi, &c.SisaKuota)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Course{}, ErrNotFound
		}
		return model.Course{}, fmt.Errorf("gagal mencari course by id: %w", err)
	}
	return c, nil
}
