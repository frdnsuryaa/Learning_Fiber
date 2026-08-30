package main

import (
	"errors"
	"math"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"api-students/app/repository"
)

// sortableFields adalah whitelist field yang boleh digunakan sebagai parameter sort.
var sortableFields = map[string]bool{
	"name":      true,
	"grade":     true,
	"is_active": true,
}

type StudentHandler struct {
	repo repository.StudentRepository
}

func NewStudentHandler(repo repository.StudentRepository) *StudentHandler {
	return &StudentHandler{repo: repo}
}

// parseStudentQuery mem-parse dan memvalidasi semua query string dari request.
// Mengembalikan StudentQuery yang siap dipakai atau pesan error jika ada parameter tidak valid.
func parseStudentQuery(c *fiber.Ctx) (StudentQuery, string) {
	q := StudentQuery{
		Page:  1,
		Limit: 10,
		Sort:  "name",
		Order: "asc",
	}

	// --- page ---
	if raw := c.Query("page"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 1 {
			return q, "Parameter 'page' harus berupa bilangan bulat positif"
		}
		q.Page = v
	}

	// --- limit (maks 100) ---
	if raw := c.Query("limit"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 1 {
			return q, "Parameter 'limit' harus berupa bilangan bulat positif"
		}
		if v > 100 {
			return q, "Parameter 'limit' tidak boleh melebihi 100"
		}
		q.Limit = v
	}

	// --- search ---
	q.Search = strings.TrimSpace(c.Query("search"))

	// --- sort (whitelist) ---
	if raw := c.Query("sort"); raw != "" {
		if !sortableFields[raw] {
			return q, "Parameter 'sort' hanya boleh berisi: name, grade, is_active"
		}
		q.Sort = raw
	}

	// --- order ---
	if raw := c.Query("order"); raw != "" {
		orderLower := strings.ToLower(raw)
		if orderLower != "asc" && orderLower != "desc" {
			return q, "Parameter 'order' hanya boleh berisi: asc atau desc"
		}
		q.Order = orderLower
	}

	// --- is_active (filter boolean opsional) ---
	if raw := c.Query("is_active"); raw != "" {
		v, err := strconv.ParseBool(raw)
		if err != nil {
			return q, "Parameter 'is_active' hanya boleh berisi: true atau false"
		}
		q.IsActive = &v
	}

	// --- grade_min (filter rentang bawah opsional) ---
	if raw := c.Query("grade_min"); raw != "" {
		v, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return q, "Parameter 'grade_min' harus berupa angka"
		}
		q.GradeMin = &v
	}

	// --- grade_max (filter rentang atas opsional) ---
	if raw := c.Query("grade_max"); raw != "" {
		v, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return q, "Parameter 'grade_max' harus berupa angka"
		}
		q.GradeMax = &v
	}

	return q, ""
}

// GetAllStudents menangani GET /students
// Mendukung paginasi, pencarian nama/NIM, pengurutan, dan filter is_active / rentang grade.
func (h *StudentHandler) GetAllStudents(c *fiber.Ctx) error {
	q, errMsg := parseStudentQuery(c)
	if errMsg != "" {
		return ErrorResponse(c, fiber.StatusBadRequest, errMsg)
	}

	students, total, err := h.repo.FindAll(c.Context(), q)
	if err != nil {
		return ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil data mahasiswa")
	}

	totalPages := int(math.Ceil(float64(total) / float64(q.Limit)))
	if totalPages == 0 {
		totalPages = 1
	}

	meta := Meta{
		Page:       q.Page,
		Limit:      q.Limit,
		Total:      total,
		TotalPages: totalPages,
	}

	return PaginatedSuccessResponse(c, "Students retrieved successfully", meta, students)
}

// GetStudentByID menangani GET /students/:id
// Mengembalikan satu siswa berdasarkan ID-nya.
func (h *StudentHandler) GetStudentByID(c *fiber.Ctx) error {
	id := c.Params("id")

	s, err := h.repo.FindByID(c.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrorResponse(c, fiber.StatusNotFound, "Student not found")
		}
		return ErrorResponse(c, fiber.StatusInternalServerError, "Gagal mengambil data mahasiswa")
	}

	return SuccessResponse(c, fiber.StatusOK, "Student retrieved successfully", s)
}

// CreateStudent menangani POST /students
// Membuat siswa baru dari body request dan menyimpannya di PostgreSQL.
func (h *StudentHandler) CreateStudent(c *fiber.Ctx) error {
	var req CreateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if req.NIM == "" {
		return ErrorResponse(c, fiber.StatusUnprocessableEntity, "Field 'nim' is required")
	}
	if req.Name == "" {
		return ErrorResponse(c, fiber.StatusUnprocessableEntity, "Field 'name' is required")
	}

	newStudent := Student{
		ID:       uuid.NewString(),
		NIM:      req.NIM,
		Name:     req.Name,
		Grade:    req.Grade,
		IsActive: req.IsActive,
	}

	created, err := h.repo.Create(c.Context(), newStudent)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			return ErrorResponse(c, fiber.StatusConflict, "NIM sudah terdaftar")
		}
		return ErrorResponse(c, fiber.StatusInternalServerError, "Gagal membuat data mahasiswa")
	}

	return SuccessResponse(c, fiber.StatusCreated, "Student created successfully", created)
}

// UpdateStudent menangani PUT /students/:id
// Mengganti seluruh data siswa berdasarkan ID.
func (h *StudentHandler) UpdateStudent(c *fiber.Ctx) error {
	id := c.Params("id")

	var req UpdateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if req.NIM == "" {
		return ErrorResponse(c, fiber.StatusUnprocessableEntity, "Field 'nim' is required")
	}
	if req.Name == "" {
		return ErrorResponse(c, fiber.StatusUnprocessableEntity, "Field 'name' is required")
	}

	updateStudent := Student{
		ID:       id,
		NIM:      req.NIM,
		Name:     req.Name,
		Grade:    req.Grade,
		IsActive: req.IsActive,
	}

	updated, err := h.repo.Update(c.Context(), updateStudent)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrorResponse(c, fiber.StatusNotFound, "Student not found")
		}
		if errors.Is(err, repository.ErrDuplicate) {
			return ErrorResponse(c, fiber.StatusConflict, "NIM sudah terdaftar")
		}
		return ErrorResponse(c, fiber.StatusInternalServerError, "Gagal memperbarui data mahasiswa")
	}

	return SuccessResponse(c, fiber.StatusOK, "Student updated successfully", updated)
}

// PatchStudent menangani PATCH /students/:id
// Memperbarui sebagian data siswa, hanya field yang dikirim yang akan diubah.
func (h *StudentHandler) PatchStudent(c *fiber.Ctx) error {
	id := c.Params("id")

	var req PatchStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	patched, err := h.repo.Patch(c.Context(), id, req.NIM, req.Name, req.Grade, req.IsActive)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrorResponse(c, fiber.StatusNotFound, "Student not found")
		}
		if errors.Is(err, repository.ErrDuplicate) {
			return ErrorResponse(c, fiber.StatusConflict, "NIM sudah terdaftar")
		}
		return ErrorResponse(c, fiber.StatusInternalServerError, "Gagal memperbarui data mahasiswa")
	}

	return SuccessResponse(c, fiber.StatusOK, "Student patched successfully", patched)
}

// DeleteStudent menangani DELETE /students/:id
// Menghapus siswa berdasarkan ID
func (h *StudentHandler) DeleteStudent(c *fiber.Ctx) error {
	id := c.Params("id")

	err := h.repo.Delete(c.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrorResponse(c, fiber.StatusNotFound, "Student not found")
		}
		return ErrorResponse(c, fiber.StatusInternalServerError, "Gagal menghapus data mahasiswa")
	}

	return SuccessResponse(c, fiber.StatusOK, "Student deleted successfully", nil)
}
