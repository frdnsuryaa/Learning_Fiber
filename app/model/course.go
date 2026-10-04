package model

import "time"

// Course merepresentasikan baris pada tabel courses.
type Course struct {
	ID        int    `json:"id"`
	KodeMK    string `json:"kode_mk"`
	NamaMK    string `json:"nama_mk"`
	SKS       int    `json:"sks"`
	Semester  int    `json:"semester"`
	Kuota     int    `json:"kuota"`
	Terisi    int    `json:"terisi"`
	SisaKuota int    `json:"sisa_kuota"`
}

// CourseQuery berisi parameter query string untuk GET /courses.
type CourseQuery struct {
	Semester  *int
	Search    string
	Available bool // hanya yang sisa_kuota > 0
}

// Enrollment merepresentasikan baris pada tabel enrollments.
type Enrollment struct {
	ID            int       `json:"id"`
	StudentID     int       `json:"student_id,omitempty"`
	CourseID      int       `json:"course_id"`
	KodeMK        string    `json:"kode_mk,omitempty"`
	NamaMK        string    `json:"nama_mk,omitempty"`
	SKS           int       `json:"sks,omitempty"`
	TahunAkademik string    `json:"tahun_akademik"`
	CreatedAt     time.Time `json:"created_at"`
}

// EnrollmentRequest untuk POST /api/v1/enrollments.
type EnrollmentRequest struct {
	CourseID      int    `json:"course_id"`
	TahunAkademik string `json:"tahun_akademik"`
}
