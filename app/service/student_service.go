package service

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"api-students/app/model"
	"api-students/app/repository"
	"api-students/helper"
)

// StudentService menangani HTTP request untuk entitas Student.
type StudentService struct {
	repo repository.StudentRepository
}

// NewStudentService membuat instance baru StudentService.
func NewStudentService(repo repository.StudentRepository) *StudentService {
	return &StudentService{repo: repo}
}

// GET /api/v1/students — hanya admin
func (s *StudentService) GetAllStudents(c *fiber.Ctx) error {
	authUser, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
	}
	if authUser.Role != "admin" {
		return helper.Fail(c, fiber.StatusForbidden, "hanya admin yang dapat mengakses endpoint ini")
	}

	q := parseStudentQuery(c)
	students, total, err := s.repo.FindAll(c.Context(), q)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal mengambil data mahasiswa")
	}

	lastPage := (total + q.PerPage - 1) / q.PerPage
	if lastPage == 0 {
		lastPage = 1
	}
	meta := model.Meta{
		CurrentPage: q.Page,
		PerPage:     q.PerPage,
		Total:       total,
		LastPage:    lastPage,
	}
	return helper.SuccessList(c, "data mahasiswa berhasil diambil", students, meta)
}

// GET /api/v1/students/:id — admin atau mahasiswa milik sendiri
func (s *StudentService) GetStudentByID(c *fiber.Ctx) error {
	authUser, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
	}

	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id < 1 {
		return helper.Fail(c, fiber.StatusNotFound, "mahasiswa tidak ditemukan")
	}

	detail, err := s.repo.FindDetailByID(c.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.Fail(c, fiber.StatusNotFound, "mahasiswa tidak ditemukan")
		}
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal mengambil data mahasiswa")
	}

	// Mahasiswa hanya boleh akses data miliknya sendiri
	if authUser.Role == "mahasiswa" && detail.UserID != authUser.UserID {
		return helper.Fail(c, fiber.StatusForbidden, "anda tidak dapat mengakses data mahasiswa lain")
	}

	detail.UserID = 0 // sembunyikan dari response
	return helper.Success(c, fiber.StatusOK, "data mahasiswa berhasil diambil", detail)
}

// POST /api/v1/students — hanya admin
func (s *StudentService) CreateStudent(c *fiber.Ctx) error {
	authUser, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
	}
	if authUser.Role != "admin" {
		return helper.Fail(c, fiber.StatusForbidden, "hanya admin yang dapat menambah mahasiswa")
	}

	var req model.CreateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}
	req.NIM = strings.TrimSpace(req.NIM)
	req.Nama = strings.TrimSpace(req.Nama)
	req.Email = strings.TrimSpace(req.Email)
	req.Prodi = strings.TrimSpace(req.Prodi)

	if errs := ValidateCreateStudentRequest(req); len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}

	// Password awal = NIM
	hashed, err := helper.HashPassword(req.NIM)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal memproses password")
	}

	_, student, err := s.repo.CreateWithUser(c.Context(), req.Email, hashed, model.Student{
		NIM:         req.NIM,
		Nama:        req.Nama,
		Prodi:       req.Prodi,
		Angkatan:    req.Angkatan,
		IPKTerakhir: req.IPKTerakhir,
	})
	if err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			return helper.FailValidation(c, map[string]string{
				"nim":   "NIM atau email sudah terdaftar",
				"email": "NIM atau email sudah terdaftar",
			})
		}
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal membuat data mahasiswa")
	}

	return helper.Success(c, fiber.StatusCreated, "data mahasiswa berhasil dibuat", student)
}

// PUT /api/v1/students/:id — hanya admin
func (s *StudentService) UpdateStudent(c *fiber.Ctx) error {
	authUser, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
	}
	if authUser.Role != "admin" {
		return helper.Fail(c, fiber.StatusForbidden, "hanya admin yang dapat memperbarui data mahasiswa")
	}

	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id < 1 {
		return helper.Fail(c, fiber.StatusNotFound, "mahasiswa tidak ditemukan")
	}

	var req model.UpdateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}
	req.Nama = strings.TrimSpace(req.Nama)
	req.Prodi = strings.TrimSpace(req.Prodi)

	if errs := ValidateUpdateStudentRequest(req); len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}

	// Pastikan ada
	existing, err := s.repo.FindByID(c.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.Fail(c, fiber.StatusNotFound, "mahasiswa tidak ditemukan")
		}
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal mengambil data mahasiswa")
	}

	existing.Nama = req.Nama
	existing.Prodi = req.Prodi
	existing.Angkatan = req.Angkatan
	existing.IPKTerakhir = req.IPKTerakhir

	updated, err := s.repo.Update(c.Context(), existing)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.Fail(c, fiber.StatusNotFound, "mahasiswa tidak ditemukan")
		}
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal memperbarui data mahasiswa")
	}

	return helper.Success(c, fiber.StatusOK, "data mahasiswa berhasil diperbarui", updated)
}

// DELETE /api/v1/students/:id — hanya admin (soft delete)
func (s *StudentService) DeleteStudent(c *fiber.Ctx) error {
	authUser, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
	}
	if authUser.Role != "admin" {
		return helper.Fail(c, fiber.StatusForbidden, "hanya admin yang dapat menghapus mahasiswa")
	}

	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id < 1 {
		return helper.Fail(c, fiber.StatusNotFound, "mahasiswa tidak ditemukan")
	}

	if err := s.repo.SoftDelete(c.Context(), id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.Fail(c, fiber.StatusNotFound, "mahasiswa tidak ditemukan")
		}
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal menghapus data mahasiswa")
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// PatchStudent dipertahankan agar route lama tidak error (tidak dipakai di spec baru)
func (s *StudentService) PatchStudent(c *fiber.Ctx) error {
	return helper.Fail(c, fiber.StatusMethodNotAllowed, "endpoint ini tidak tersedia")
}

// parseStudentQuery mem-parse query string untuk GET /api/v1/students.
func parseStudentQuery(c *fiber.Ctx) model.StudentQuery {
	return helper.ParseStudentQuery(c)
}

