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

// NilaiService menangani HTTP request untuk entitas Nilai.
type NilaiService struct {
	nilaiRepo   repository.NilaiRepository
	studentRepo repository.StudentRepository
}

// NewNilaiService membuat instance baru NilaiService.
func NewNilaiService(nilaiRepo repository.NilaiRepository, studentRepo repository.StudentRepository) *NilaiService {
	return &NilaiService{nilaiRepo: nilaiRepo, studentRepo: studentRepo}
}

// GET /api/v1/nilai — mahasiswa: nilai milik sendiri; admin: wajib ?student_id=
func (s *NilaiService) GetNilai(c *fiber.Ctx) error {
	authUser, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
	}

	var studentID int

	if authUser.Role == "mahasiswa" {
		student, err := s.studentRepo.FindByUserID(c.Context(), authUser.UserID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return helper.Fail(c, fiber.StatusNotFound, "data mahasiswa tidak ditemukan")
			}
			return helper.Fail(c, fiber.StatusInternalServerError, "gagal mengambil data mahasiswa")
		}
		studentID = student.ID
	} else {
		// admin harus sertakan query param student_id
		param := c.Query("student_id")
		if param == "" {
			return helper.Fail(c, fiber.StatusBadRequest, "parameter student_id wajib diisi")
		}
		id, err := strconv.Atoi(param)
		if err != nil || id < 1 {
			return helper.Fail(c, fiber.StatusBadRequest, "student_id tidak valid")
		}
		studentID = id
	}

	nilaiList, err := s.nilaiRepo.FindByStudentID(c.Context(), studentID)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal mengambil data nilai")
	}
	if nilaiList == nil {
		nilaiList = []model.Nilai{}
	}

	return helper.Success(c, fiber.StatusOK, "berhasil mengambil data nilai", nilaiList)
}

// POST /api/v1/nilai — hanya admin
func (s *NilaiService) CreateNilai(c *fiber.Ctx) error {
	authUser, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
	}
	if authUser.Role != "admin" {
		return helper.Fail(c, fiber.StatusForbidden, "hanya admin yang dapat menambah nilai")
	}

	var req struct {
		NamaMatkul string `json:"namamatkul"`
		Nilai      string `json:"nilai"`
		IDStudent  int    `json:"idstudent"`
	}
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}

	req.NamaMatkul = strings.TrimSpace(req.NamaMatkul)
	req.Nilai = strings.TrimSpace(req.Nilai)

	errs := map[string]string{}
	if req.NamaMatkul == "" {
		errs["namamatkul"] = "nama mata kuliah wajib diisi"
	}
	if req.Nilai == "" {
		errs["nilai"] = "nilai wajib diisi"
	}
	if req.IDStudent < 1 {
		errs["idstudent"] = "idstudent wajib diisi"
	}
	if len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}

	_, err := s.studentRepo.FindByID(c.Context(), req.IDStudent)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.FailValidation(c, map[string]string{"idstudent": "mahasiswa tidak ditemukan"})
		}
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal memverifikasi data mahasiswa")
	}

	created, err := s.nilaiRepo.Create(c.Context(), model.Nilai{
		NamaMatkul: req.NamaMatkul,
		Nilai:      req.Nilai,
		IDStudent:  req.IDStudent,
	})
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal menyimpan nilai")
	}

	return helper.Success(c, fiber.StatusCreated, "nilai berhasil ditambahkan", created)
}

// DELETE /api/v1/nilai/:id — hanya admin
func (s *NilaiService) DeleteNilai(c *fiber.Ctx) error {
	authUser, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
	}
	if authUser.Role != "admin" {
		return helper.Fail(c, fiber.StatusForbidden, "hanya admin yang dapat menghapus nilai")
	}

	idNilai, err := strconv.Atoi(c.Params("id"))
	if err != nil || idNilai < 1 {
		return helper.Fail(c, fiber.StatusNotFound, "nilai tidak ditemukan")
	}

	// student_id wajib untuk verifikasi kepemilikan
	studentParam := c.Query("student_id")
	if studentParam == "" {
		return helper.Fail(c, fiber.StatusBadRequest, "parameter student_id wajib diisi")
	}
	studentID, err := strconv.Atoi(studentParam)
	if err != nil || studentID < 1 {
		return helper.Fail(c, fiber.StatusBadRequest, "student_id tidak valid")
	}

	if err := s.nilaiRepo.Delete(c.Context(), idNilai, studentID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.Fail(c, fiber.StatusNotFound, "nilai tidak ditemukan")
		}
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal menghapus nilai")
	}

	return helper.Success(c, fiber.StatusOK, "nilai berhasil dihapus", nil)
}
