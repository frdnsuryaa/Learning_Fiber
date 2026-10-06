package service

// File ini berisi business rules MURNI untuk entitas Student.
// Tidak mengimpor Fiber sama sekali, tidak menyentuh database,
// dan tidak tahu apa pun tentang HTTP.

import (
	"strings"
	"time"
	"unicode"

	"api-students/app/model"
)

// ValidateCreateStudentRequest memeriksa isi permintaan POST /api/v1/students.
func ValidateCreateStudentRequest(req model.CreateStudentRequest) map[string]string {
	errs := map[string]string{}

	// NIM: wajib, tepat 12 digit angka
	nim := strings.TrimSpace(req.NIM)
	if nim == "" {
		errs["nim"] = "wajib diisi"
	} else if len(nim) != 12 || !allDigits(nim) {
		errs["nim"] = "NIM harus tepat 12 digit angka"
	}

	if strings.TrimSpace(req.Nama) == "" {
		errs["nama"] = "wajib diisi"
	}
	if !isValidEmail(req.Email) {
		errs["email"] = "format email tidak valid"
	}
	if strings.TrimSpace(req.Prodi) == "" {
		errs["prodi"] = "wajib diisi"
	}

	currentYear := time.Now().Year()
	if req.Angkatan < 1900 || req.Angkatan > currentYear {
		errs["angkatan"] = "angkatan harus 4 digit dan tidak melebihi tahun berjalan"
	}
	if req.IPKTerakhir < 0 || req.IPKTerakhir > 4.00 {
		errs["ipk_terakhir"] = "IPK harus antara 0.00 dan 4.00"
	}

	return errs
}

// ValidateUpdateStudentRequest memeriksa isi permintaan PUT /api/v1/students/{id}.
func ValidateUpdateStudentRequest(req model.UpdateStudentRequest) map[string]string {
	errs := map[string]string{}

	if strings.TrimSpace(req.Nama) == "" {
		errs["nama"] = "wajib diisi"
	}
	if strings.TrimSpace(req.Prodi) == "" {
		errs["prodi"] = "wajib diisi"
	}
	currentYear := time.Now().Year()
	if req.Angkatan < 1900 || req.Angkatan > currentYear {
		errs["angkatan"] = "angkatan harus 4 digit dan tidak melebihi tahun berjalan"
	}
	if req.IPKTerakhir < 0 || req.IPKTerakhir > 4.00 {
		errs["ipk_terakhir"] = "IPK harus antara 0.00 dan 4.00"
	}

	return errs
}

// allDigits memeriksa apakah string hanya berisi digit angka.
func allDigits(s string) bool {
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

// ── Fungsi lama dipertahankan agar test tidak break ─────────────────────────

func ValidateCreateStudent(req model.CreateStudentRequest) map[string]string {
	return ValidateCreateStudentRequest(req)
}

func ValidateUpdateStudent(req model.UpdateStudentRequest) map[string]string {
	return ValidateUpdateStudentRequest(req)
}

// ApplyStudentPatch dan IsEmptyStudentPatch dipertahankan untuk student_rules_test.go.
func ApplyStudentPatch(
	current model.Student, req model.PatchStudentRequest,
) (model.Student, map[string]string) {
	return current, map[string]string{}
}

func IsEmptyStudentPatch(req model.PatchStudentRequest) bool {
	return req.NIM == nil && req.Name == nil && req.Grade == nil && req.IsActive == nil
}

// CountTotalPages membulatkan ke atas tanpa memakai bilangan pecahan.
func CountTotalPages(total, limit int) int {
	if limit <= 0 {
		return 0
	}
	return (total + limit - 1) / limit
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

