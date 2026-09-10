package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"api-students/app/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type StudentRepository interface {
	FindAll(ctx context.Context, q model.StudentQuery) ([]model.Student, int, error)
	FindByID(ctx context.Context, id string) (model.Student, error)
	Create(ctx context.Context, s model.Student) (model.Student, error)
	Update(ctx context.Context, s model.Student) (model.Student, error)
	Patch(ctx context.Context, id string, nim *string, name *string, grade *float64, isActive *bool) (model.Student, error)
	Delete(ctx context.Context, id string) error
}

type studentPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewStudentRepository(pool *pgxpool.Pool) StudentRepository {
	return &studentPostgresRepository{pool: pool}
}

func (r *studentPostgresRepository) Create(ctx context.Context, s model.Student) (model.Student, error) {
	query := `
		INSERT INTO students (id, nim, name, grade, is_active, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		RETURNING id, nim, name, grade, is_active, created_at
	`
	var created model.Student
	err := r.pool.QueryRow(ctx, query, s.ID, s.NIM, s.Name, s.Grade, s.IsActive).Scan(
		&created.ID,
		&created.NIM,
		&created.Name,
		&created.Grade,
		&created.IsActive,
		&created.CreatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // Unique violation
			return model.Student{}, ErrDuplicate
		}
		return model.Student{}, fmt.Errorf("gagal membuat student: %w", err)
	}

	return created, nil
}

func (r *studentPostgresRepository) FindByID(ctx context.Context, id string) (model.Student, error) {
	query := `
		SELECT id, nim, name, grade, is_active, created_at
		FROM students
		WHERE id = $1
	`
	var s model.Student
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&s.ID,
		&s.NIM,
		&s.Name,
		&s.Grade,
		&s.IsActive,
		&s.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Student{}, ErrNotFound
		}
		return model.Student{}, fmt.Errorf("gagal mencari student by id: %w", err)
	}

	return s, nil
}

func (r *studentPostgresRepository) FindAll(ctx context.Context, q model.StudentQuery) ([]model.Student, int, error) {
	var conditions []string
	var args []interface{}
	argIdx := 1

	if q.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(LOWER(name) LIKE $%d OR LOWER(nim) LIKE $%d)", argIdx, argIdx))
		args = append(args, "%"+strings.ToLower(q.Search)+"%")
		argIdx++
	}

	if q.IsActive != nil {
		conditions = append(conditions, fmt.Sprintf("is_active = $%d", argIdx))
		args = append(args, *q.IsActive)
		argIdx++
	}

	if q.GradeMin != nil {
		conditions = append(conditions, fmt.Sprintf("grade >= $%d", argIdx))
		args = append(args, *q.GradeMin)
		argIdx++
	}

	if q.GradeMax != nil {
		conditions = append(conditions, fmt.Sprintf("grade <= $%d", argIdx))
		args = append(args, *q.GradeMax)
		argIdx++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " WHERE " + strings.Join(conditions, " AND ")
	}

	// 1. Hitung total data yang cocok
	countQuery := "SELECT COUNT(*) FROM students" + whereClause
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("gagal menghitung total students: %w", err)
	}

	if total == 0 {
		return []model.Student{}, 0, nil
	}

	// 2. Tentukan pengurutan (whitelist aman)
	orderColumn := "LOWER(name)"
	switch q.Sort {
	case "grade":
		orderColumn = "grade"
	case "is_active":
		orderColumn = "is_active"
	default:
		orderColumn = "LOWER(name)"
	}

	orderDir := "ASC"
	if strings.ToLower(q.Order) == "desc" {
		orderDir = "DESC"
	}

	orderClause := fmt.Sprintf(" ORDER BY %s %s, id ASC", orderColumn, orderDir)

	// 3. Paginasi LIMIT & OFFSET
	offset := (q.Page - 1) * q.Limit
	limitClause := fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	queryArgs := append(args, q.Limit, offset)

	dataQuery := "SELECT id, nim, name, grade, is_active, created_at FROM students" + whereClause + orderClause + limitClause
	rows, err := r.pool.Query(ctx, dataQuery, queryArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("gagal query students: %w", err)
	}
	defer rows.Close()

	students := make([]model.Student, 0)
	for rows.Next() {
		var s model.Student
		if err := rows.Scan(&s.ID, &s.NIM, &s.Name, &s.Grade, &s.IsActive, &s.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("gagal scan student: %w", err)
		}
		students = append(students, s)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterasi rows student: %w", err)
	}

	return students, total, nil
}

func (r *studentPostgresRepository) Update(ctx context.Context, s model.Student) (model.Student, error) {
	query := `
		UPDATE students
		SET nim = $1, name = $2, grade = $3, is_active = $4
		WHERE id = $5
		RETURNING id, nim, name, grade, is_active, created_at
	`
	var updated model.Student
	err := r.pool.QueryRow(ctx, query, s.NIM, s.Name, s.Grade, s.IsActive, s.ID).Scan(
		&updated.ID,
		&updated.NIM,
		&updated.Name,
		&updated.Grade,
		&updated.IsActive,
		&updated.CreatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return model.Student{}, ErrDuplicate
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Student{}, ErrNotFound
		}
		return model.Student{}, fmt.Errorf("gagal update student: %w", err)
	}

	return updated, nil
}

func (r *studentPostgresRepository) Patch(ctx context.Context, id string, nim *string, name *string, grade *float64, isActive *bool) (model.Student, error) {
	existing, err := r.FindByID(ctx, id)
	if err != nil {
		return model.Student{}, err
	}

	if nim != nil {
		existing.NIM = *nim
	}
	if name != nil {
		existing.Name = *name
	}
	if grade != nil {
		existing.Grade = *grade
	}
	if isActive != nil {
		existing.IsActive = *isActive
	}

	return r.Update(ctx, existing)
}

func (r *studentPostgresRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM students WHERE id = $1`
	cmdTag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("gagal delete student: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
