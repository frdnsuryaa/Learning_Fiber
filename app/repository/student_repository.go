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

// StudentRepository mendefinisikan operasi database untuk tabel students.
type StudentRepository interface {
	FindAll(ctx context.Context, q model.StudentQuery) ([]model.Student, int, error)
	FindByID(ctx context.Context, id int) (model.Student, error)
	FindByUserID(ctx context.Context, userID int) (model.Student, error)
	FindDetailByID(ctx context.Context, id int) (model.StudentDetail, error)
	// CreateWithUser membuat user + student dalam satu transaksi.
	CreateWithUser(ctx context.Context, email, passwordHash string, s model.Student) (model.User, model.Student, error)
	Update(ctx context.Context, s model.Student) (model.Student, error)
	SoftDelete(ctx context.Context, id int) error
}

type studentPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewStudentRepository(pool *pgxpool.Pool) StudentRepository {
	return &studentPostgresRepository{pool: pool}
}

// scanStudentRow membaca kolom (id, user_id, nim, nama, prodi, angkatan, ipk_terakhir).
func scanStudentRow(row pgx.Row) (model.Student, error) {
	var s model.Student
	err := row.Scan(&s.ID, &s.UserID, &s.NIM, &s.Nama, &s.Prodi, &s.Angkatan, &s.IPKTerakhir)
	return s, err
}

// CreateWithUser membuat akun user + data mahasiswa dalam satu transaksi.
func (r *studentPostgresRepository) CreateWithUser(
	ctx context.Context,
	email, passwordHash string,
	s model.Student,
) (model.User, model.Student, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return model.User{}, model.Student{}, fmt.Errorf("gagal memulai transaksi: %w", err)
	}
	defer tx.Rollback(ctx)

	// 1. Insert user
	var u model.User
	err = tx.QueryRow(ctx,
		`INSERT INTO users (email, password, role)
		 VALUES ($1,$2,'mahasiswa')
		 RETURNING id, email, role, created_at`,
		email, passwordHash,
	).Scan(&u.ID, &u.Email, &u.Role, &u.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return model.User{}, model.Student{}, ErrDuplicateEmail
		}
		return model.User{}, model.Student{}, fmt.Errorf("gagal insert user: %w", err)
	}

	// 2. Insert student
	student, err := scanStudentRow(tx.QueryRow(ctx,
		`INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir)
		 VALUES ($1,$2,$3,$4,$5,$6)
		 RETURNING id, user_id, nim, nama, prodi, angkatan, ipk_terakhir`,
		u.ID, s.NIM, s.Nama, s.Prodi, s.Angkatan, s.IPKTerakhir,
	))
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return model.User{}, model.Student{}, ErrDuplicateNIM
		}
		return model.User{}, model.Student{}, fmt.Errorf("gagal insert student: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return model.User{}, model.Student{}, fmt.Errorf("gagal commit transaksi: %w", err)
	}
	return u, student, nil
}

func (r *studentPostgresRepository) FindByID(ctx context.Context, id int) (model.Student, error) {
	s, err := scanStudentRow(r.pool.QueryRow(ctx, `
		SELECT id, user_id, nim, nama, prodi, angkatan, ipk_terakhir
		FROM students
		WHERE id = $1 AND deleted_at IS NULL
	`, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Student{}, ErrNotFound
		}
		return model.Student{}, fmt.Errorf("gagal mencari student by id: %w", err)
	}
	return s, nil
}

func (r *studentPostgresRepository) FindByUserID(ctx context.Context, userID int) (model.Student, error) {
	s, err := scanStudentRow(r.pool.QueryRow(ctx, `
		SELECT id, user_id, nim, nama, prodi, angkatan, ipk_terakhir
		FROM students
		WHERE user_id = $1 AND deleted_at IS NULL
	`, userID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Student{}, ErrNotFound
		}
		return model.Student{}, fmt.Errorf("gagal mencari student by user_id: %w", err)
	}
	return s, nil
}

func (r *studentPostgresRepository) FindDetailByID(ctx context.Context, id int) (model.StudentDetail, error) {
	s, err := r.FindByID(ctx, id)
	if err != nil {
		return model.StudentDetail{}, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT e.id, e.course_id, c.kode_mk, c.nama_mk, c.sks, e.tahun_akademik, e.created_at
		FROM enrollments e
		JOIN courses c ON c.id = e.course_id
		WHERE e.student_id = $1
		ORDER BY e.created_at
	`, id)
	if err != nil {
		return model.StudentDetail{}, fmt.Errorf("gagal query enrollments: %w", err)
	}
	defer rows.Close()

	var enrollments []model.Enrollment
	totalSKS := 0
	for rows.Next() {
		var e model.Enrollment
		if err := rows.Scan(&e.ID, &e.CourseID, &e.KodeMK, &e.NamaMK, &e.SKS, &e.TahunAkademik, &e.CreatedAt); err != nil {
			return model.StudentDetail{}, fmt.Errorf("gagal scan enrollment: %w", err)
		}
		totalSKS += e.SKS
		enrollments = append(enrollments, e)
	}
	if rows.Err() != nil {
		return model.StudentDetail{}, rows.Err()
	}
	if enrollments == nil {
		enrollments = []model.Enrollment{}
	}

	batasSKS := sksBatas(s.IPKTerakhir)
	return model.StudentDetail{
		ID:          s.ID,
		UserID:      s.UserID,
		NIM:         s.NIM,
		Nama:        s.Nama,
		Prodi:       s.Prodi,
		Angkatan:    s.Angkatan,
		IPKTerakhir: s.IPKTerakhir,
		TotalSKS:    totalSKS,
		BatasSKS:    batasSKS,
		MataKuliah:  enrollments,
	}, nil
}

// sksBatas mengembalikan batas SKS berdasarkan IPK sesuai business rule.
func sksBatas(ipk float64) int {
	switch {
	case ipk >= 3.00:
		return 24
	case ipk >= 2.50:
		return 21
	default:
		return 18
	}
}

func (r *studentPostgresRepository) FindAll(ctx context.Context, q model.StudentQuery) ([]model.Student, int, error) {
	var conditions []string
	var args []interface{}
	argIdx := 1

	conditions = append(conditions, "deleted_at IS NULL")

	if q.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(LOWER(nim) LIKE $%d OR LOWER(nama) LIKE $%d)", argIdx, argIdx))
		args = append(args, "%"+strings.ToLower(q.Search)+"%")
		argIdx++
	}
	if q.Prodi != "" {
		conditions = append(conditions, fmt.Sprintf("LOWER(prodi) = LOWER($%d)", argIdx))
		args = append(args, q.Prodi)
		argIdx++
	}
	if q.Angkatan != nil {
		conditions = append(conditions, fmt.Sprintf("angkatan = $%d", argIdx))
		args = append(args, *q.Angkatan)
		argIdx++
	}

	whereClause := " WHERE " + strings.Join(conditions, " AND ")

	var total int
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM students"+whereClause, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("gagal menghitung total students: %w", err)
	}
	if total == 0 {
		return []model.Student{}, 0, nil
	}

	orderClause := " ORDER BY LOWER(nama) ASC"
	if q.Sort == "-ipk_terakhir" {
		orderClause = " ORDER BY ipk_terakhir DESC"
	}

	offset := (q.Page - 1) * q.PerPage
	limitClause := fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	queryArgs := append(args, q.PerPage, offset)

	rows, err := r.pool.Query(ctx,
		"SELECT id, user_id, nim, nama, prodi, angkatan, ipk_terakhir FROM students"+whereClause+orderClause+limitClause,
		queryArgs...,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("gagal query students: %w", err)
	}
	defer rows.Close()

	students := make([]model.Student, 0)
	for rows.Next() {
		var s model.Student
		if err := rows.Scan(&s.ID, &s.UserID, &s.NIM, &s.Nama, &s.Prodi, &s.Angkatan, &s.IPKTerakhir); err != nil {
			return nil, 0, fmt.Errorf("gagal scan student: %w", err)
		}
		s.UserID = 0 // jangan tampilkan user_id di list
		students = append(students, s)
	}
	return students, total, rows.Err()
}

func (r *studentPostgresRepository) Update(ctx context.Context, s model.Student) (model.Student, error) {
	updated, err := scanStudentRow(r.pool.QueryRow(ctx, `
		UPDATE students
		SET nama = $1, prodi = $2, angkatan = $3, ipk_terakhir = $4
		WHERE id = $5 AND deleted_at IS NULL
		RETURNING id, user_id, nim, nama, prodi, angkatan, ipk_terakhir
	`, s.Nama, s.Prodi, s.Angkatan, s.IPKTerakhir, s.ID))
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

func (r *studentPostgresRepository) SoftDelete(ctx context.Context, id int) error {
	cmdTag, err := r.pool.Exec(ctx,
		`UPDATE students SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`, id)
	if err != nil {
		return fmt.Errorf("gagal soft delete student: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
