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

// EnrollmentService menangani HTTP request untuk entitas Enrollment (KRS).
type EnrollmentService struct {
	enrollRepo  repository.EnrollmentRepository
	studentRepo repository.StudentRepository
	courseRepo  repository.CourseRepository
}

// NewEnrollmentService membuat instance baru EnrollmentService.
func NewEnrollmentService(
	enrollRepo repository.EnrollmentRepository,
	studentRepo repository.StudentRepository,
	courseRepo repository.CourseRepository,
) *EnrollmentService {
	return &EnrollmentService{
		enrollRepo:  enrollRepo,
		studentRepo: studentRepo,
		courseRepo:  courseRepo,
	}
}

// POST /api/v1/enrollments — hanya mahasiswa
func (s *EnrollmentService) CreateEnrollment(c *fiber.Ctx) error {
	authUser, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
	}
	if authUser.Role != "mahasiswa" {
		return helper.Fail(c, fiber.StatusForbidden, "hanya mahasiswa yang dapat mengambil mata kuliah")
	}

	var req model.EnrollmentRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}
	req.TahunAkademik = strings.TrimSpace(req.TahunAkademik)

	if errs := ValidateEnrollmentRequest(req); len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}

	// Ambil data student
	student, err := s.studentRepo.FindByUserID(c.Context(), authUser.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.Fail(c, fiber.StatusNotFound, "data mahasiswa tidak ditemukan")
		}
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal mengambil data mahasiswa")
	}

	// Pastikan course ada
	_, err = s.courseRepo.FindByID(c.Context(), req.CourseID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.FailValidation(c, map[string]string{"course_id": "mata kuliah tidak ditemukan"})
		}
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal mengambil data mata kuliah")
	}

	// Hitung batas SKS
	batasSKS := sksBatas(student.IPKTerakhir)

	enrollment, err := s.enrollRepo.Create(c.Context(), model.Enrollment{
		StudentID:     student.ID,
		CourseID:      req.CourseID,
		TahunAkademik: req.TahunAkademik,
	}, batasSKS)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			return helper.Fail(c, fiber.StatusConflict, "mata kuliah sudah diambil pada tahun akademik ini")
		}
		if errors.Is(err, repository.ErrKuotaPenuh) {
			return helper.FailValidation(c, map[string]string{
				"course_id": "kuota mata kuliah sudah penuh",
			})
		}
		if errors.Is(err, repository.ErrSKSMelebihi) {
			return helper.FailValidation(c, map[string]string{"sks": err.Error()})
		}
		if errors.Is(err, repository.ErrNotFound) {
			return helper.FailValidation(c, map[string]string{"course_id": "mata kuliah tidak ditemukan"})
		}
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal mengambil mata kuliah")
	}

	return helper.Success(c, fiber.StatusCreated, "mata kuliah berhasil diambil", enrollment)
}

// DELETE /api/v1/enrollments/:id — mahasiswa (milik sendiri)
func (s *EnrollmentService) DeleteEnrollment(c *fiber.Ctx) error {
	authUser, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
	}
	if authUser.Role != "mahasiswa" {
		return helper.Fail(c, fiber.StatusForbidden, "hanya mahasiswa yang dapat membatalkan mata kuliah")
	}

	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id < 1 {
		return helper.Fail(c, fiber.StatusNotFound, "enrollment tidak ditemukan")
	}

	// Ambil student untuk cek kepemilikan
	student, err := s.studentRepo.FindByUserID(c.Context(), authUser.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.Fail(c, fiber.StatusNotFound, "data mahasiswa tidak ditemukan")
		}
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal mengambil data mahasiswa")
	}

	if err := s.enrollRepo.Delete(c.Context(), id, student.ID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.Fail(c, fiber.StatusNotFound, "enrollment tidak ditemukan")
		}
		if errors.Is(err, repository.ErrForbidden) {
			return helper.Fail(c, fiber.StatusForbidden, "enrollment bukan milik anda")
		}
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal membatalkan mata kuliah")
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ValidateEnrollmentRequest memvalidasi request POST /api/v1/enrollments.
func ValidateEnrollmentRequest(req model.EnrollmentRequest) map[string]string {
	errs := map[string]string{}

	if req.CourseID < 1 {
		errs["course_id"] = "wajib diisi dan harus berupa ID mata kuliah yang valid"
	}

	ta := strings.TrimSpace(req.TahunAkademik)
	if ta == "" {
		errs["tahun_akademik"] = "wajib diisi"
	} else if !isValidTahunAkademik(ta) {
		errs["tahun_akademik"] = "format harus seperti 2026/2027-Ganjil atau 2026/2027-Genap"
	}

	return errs
}

// isValidTahunAkademik memeriksa format tahun akademik: YYYY/YYYY-Ganjil atau YYYY/YYYY-Genap.
func isValidTahunAkademik(s string) bool {
	if len(s) != 15 && len(s) != 16 {
		return false
	}
	if s[4] != '/' || s[9] != '-' {
		return false
	}
	for _, i := range []int{0, 1, 2, 3, 5, 6, 7, 8} {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	suffix := s[10:]
	return suffix == "Ganjil" || suffix == "Genap"
}
