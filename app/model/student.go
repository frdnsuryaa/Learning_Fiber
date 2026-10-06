package model

import "time"

// Student merepresentasikan baris pada tabel students.
type Student struct {
	ID          int        `json:"id"`
	UserID      int        `json:"user_id,omitempty"`
	NIM         string     `json:"nim"`
	Nama        string     `json:"nama"`
	Prodi       string     `json:"prodi"`
	Angkatan    int        `json:"angkatan"`
	IPKTerakhir float64    `json:"ipk_terakhir"`
	DeletedAt   *time.Time `json:"-"`
}

// StudentDetail adalah response GET /students/{id} yang menyertakan daftar mata kuliah.
type StudentDetail struct {
	ID          int          `json:"id"`
	UserID      int          `json:"user_id,omitempty"`
	NIM         string       `json:"nim"`
	Nama        string       `json:"nama"`
	Prodi       string       `json:"prodi"`
	Angkatan    int          `json:"angkatan"`
	IPKTerakhir float64      `json:"ipk_terakhir"`
	TotalSKS    int          `json:"total_sks"`
	BatasSKS    int          `json:"batas_sks"`
	MataKuliah  []Enrollment `json:"mata_kuliah"`
}

// StudentQuery berisi parameter query string untuk GET /students.
type StudentQuery struct {
	Page     int
	PerPage  int
	Search   string
	Prodi    string
	Angkatan *int
	Sort     string // "nama" atau "-ipk_terakhir"
}

// CreateStudentRequest untuk POST /api/v1/students.
type CreateStudentRequest struct {
	NIM         string  `json:"nim"`
	Nama        string  `json:"nama"`
	Email       string  `json:"email"`
	Prodi       string  `json:"prodi"`
	Angkatan    int     `json:"angkatan"`
	IPKTerakhir float64 `json:"ipk_terakhir"`
}

// UpdateStudentRequest untuk PUT /api/v1/students/{id}.
type UpdateStudentRequest struct {
	Nama        string  `json:"nama"`
	Prodi       string  `json:"prodi"`
	Angkatan    int     `json:"angkatan"`
	IPKTerakhir float64 `json:"ipk_terakhir"`
}


// ── Struct lama dipertahankan agar test tidak break ────────────────────────
// PatchStudentRequest — dipakai oleh student_rules_test.go.
type PatchStudentRequest struct {
	NIM      *string  `json:"nim,omitempty"`
	Name     *string  `json:"name,omitempty"`
	Grade    *float64 `json:"grade,omitempty"`
	IsActive *bool    `json:"is_active,omitempty"`
}
