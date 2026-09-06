package model

import "time"

// Student merepresentasikan entitas mahasiswa di tabel students.
type Student struct {
	ID        string    `json:"id"`
	NIM       string    `json:"nim"`
	Name      string    `json:"name"`
	Grade     float64   `json:"grade"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at,omitempty"`
}

// StudentQuery berisi parameter query string untuk GET /students.
type StudentQuery struct {
	Page     int
	Limit    int
	Search   string
	Sort     string
	Order    string
	IsActive *bool
	GradeMin *float64
	GradeMax *float64
}

// ── Student request structs ──

// CreateStudentRequest digunakan untuk POST /students.
// Semua field wajib diisi; ID dan created_at di-generate oleh server.
type CreateStudentRequest struct {
	NIM      string  `json:"nim"`
	Name     string  `json:"name"`
	Grade    float64 `json:"grade"`
	IsActive bool    `json:"is_active"`
}

// UpdateStudentRequest digunakan untuk PUT /students/:id.
// Merupakan replace penuh — semua field wajib diisi.
type UpdateStudentRequest struct {
	NIM      string  `json:"nim"`
	Name     string  `json:"name"`
	Grade    float64 `json:"grade"`
	IsActive bool    `json:"is_active"`
}

// PatchStudentRequest digunakan untuk PATCH /students/:id.
// Semua field bersifat opsional; hanya field yang dikirim yang akan diperbarui.
type PatchStudentRequest struct {
	NIM      *string  `json:"nim,omitempty"`
	Name     *string  `json:"name,omitempty"`
	Grade    *float64 `json:"grade,omitempty"`
	IsActive *bool    `json:"is_active,omitempty"`
}
