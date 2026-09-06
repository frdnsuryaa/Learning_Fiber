package service

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"api-students/app/model"
	"api-students/app/repository"
	"api-students/helper"
)

// StudentService menangani HTTP request untuk entitas Student.
// Ia bertindak sebagai penerima fiber.Ctx yang mendelegasikan
// business rules ke student_rules.go dan akses data ke repository.
type StudentService struct {
	repo repository.StudentRepository
}

// NewStudentService membuat instance baru StudentService.
func NewStudentService(repo repository.StudentRepository) *StudentService {
	return &StudentService{repo: repo}
}

// GetAllStudents menangani GET /students
// Mendukung paginasi, pencarian nama/NIM, pengurutan, dan filter is_active / rentang grade.
func (s *StudentService) GetAllStudents(c *fiber.Ctx) error {
	q, errMsg := helper.ParseStudentQuery(c)
	if errMsg != "" {
		return helper.Fail(c, fiber.StatusBadRequest, errMsg)
	}

	students, total, err := s.repo.FindAll(c.Context(), q)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "Gagal mengambil data mahasiswa")
	}

	totalPages := CountTotalPages(total, q.Limit)
	if totalPages == 0 {
		totalPages = 1
	}

	meta := &model.Meta{
		Page:       q.Page,
		Limit:      q.Limit,
		Total:      total,
		TotalPages: totalPages,
	}

	return helper.SuccessList(c, "Students retrieved successfully", students, meta)
}

// GetStudentByID menangani GET /students/:id
// Mengembalikan satu siswa berdasarkan ID-nya.
func (s *StudentService) GetStudentByID(c *fiber.Ctx) error {
	id := c.Params("id")

	student, err := s.repo.FindByID(c.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.Fail(c, fiber.StatusNotFound, "Student not found")
		}
		return helper.Fail(c, fiber.StatusInternalServerError, "Gagal mengambil data mahasiswa")
	}

	return helper.Success(c, fiber.StatusOK, "Student retrieved successfully", student)
}

// CreateStudent menangani POST /students
// Membuat siswa baru dari body request dan menyimpannya di PostgreSQL.
func (s *StudentService) CreateStudent(c *fiber.Ctx) error {
	var req model.CreateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "Invalid request body")
	}

	// Validasi menggunakan business rules murni
	if errs := ValidateCreateStudent(req); len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}

	newStudent := model.Student{
		ID:       uuid.NewString(),
		NIM:      req.NIM,
		Name:     req.Name,
		Grade:    req.Grade,
		IsActive: req.IsActive,
	}

	created, err := s.repo.Create(c.Context(), newStudent)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			return helper.Fail(c, fiber.StatusConflict, "NIM sudah terdaftar")
		}
		return helper.Fail(c, fiber.StatusInternalServerError, "Gagal membuat data mahasiswa")
	}

	return helper.Success(c, fiber.StatusCreated, "Student created successfully", created)
}

// UpdateStudent menangani PUT /students/:id
// Mengganti seluruh data siswa berdasarkan ID.
func (s *StudentService) UpdateStudent(c *fiber.Ctx) error {
	id := c.Params("id")

	var req model.UpdateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "Invalid request body")
	}

	// Validasi menggunakan business rules murni
	if errs := ValidateUpdateStudent(req); len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}

	updateStudent := model.Student{
		ID:       id,
		NIM:      req.NIM,
		Name:     req.Name,
		Grade:    req.Grade,
		IsActive: req.IsActive,
	}

	updated, err := s.repo.Update(c.Context(), updateStudent)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.Fail(c, fiber.StatusNotFound, "Student not found")
		}
		if errors.Is(err, repository.ErrDuplicate) {
			return helper.Fail(c, fiber.StatusConflict, "NIM sudah terdaftar")
		}
		return helper.Fail(c, fiber.StatusInternalServerError, "Gagal memperbarui data mahasiswa")
	}

	return helper.Success(c, fiber.StatusOK, "Student updated successfully", updated)
}

// PatchStudent menangani PATCH /students/:id
// Memperbarui sebagian data siswa, hanya field yang dikirim yang akan diubah.
func (s *StudentService) PatchStudent(c *fiber.Ctx) error {
	id := c.Params("id")

	var req model.PatchStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "Invalid request body")
	}

	patched, err := s.repo.Patch(c.Context(), id, req.NIM, req.Name, req.Grade, req.IsActive)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.Fail(c, fiber.StatusNotFound, "Student not found")
		}
		if errors.Is(err, repository.ErrDuplicate) {
			return helper.Fail(c, fiber.StatusConflict, "NIM sudah terdaftar")
		}
		return helper.Fail(c, fiber.StatusInternalServerError, "Gagal memperbarui data mahasiswa")
	}

	return helper.Success(c, fiber.StatusOK, "Student patched successfully", patched)
}

// DeleteStudent menangani DELETE /students/:id
// Menghapus siswa berdasarkan ID.
func (s *StudentService) DeleteStudent(c *fiber.Ctx) error {
	id := c.Params("id")

	err := s.repo.Delete(c.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.Fail(c, fiber.StatusNotFound, "Student not found")
		}
		return helper.Fail(c, fiber.StatusInternalServerError, "Gagal menghapus data mahasiswa")
	}

	return helper.Success(c, fiber.StatusOK, "Student deleted successfully", nil)
}
