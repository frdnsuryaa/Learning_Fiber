package main

import "api-students/app/model"

type Student = model.Student
type StudentQuery = model.StudentQuery

// Semua field wajib diisi; ID dan created_at di-generate oleh server.
type CreateStudentRequest struct {
	NIM      string  `json:"nim"`
	Name     string  `json:"name"`
	Grade    float64 `json:"grade"`
	IsActive bool    `json:"is_active"`
}

// Merupakan replace penuh — semua field wajib diisi.
type UpdateStudentRequest struct {
	NIM      string  `json:"nim"`
	Name     string  `json:"name"`
	Grade    float64 `json:"grade"`
	IsActive bool    `json:"is_active"`
}

// Semua field bersifat opsional; hanya field yang dikirim yang akan diperbarui.
type PatchStudentRequest struct {
	NIM      *string  `json:"nim,omitempty"`
	Name     *string  `json:"name,omitempty"`
	Grade    *float64 `json:"grade,omitempty"`
	IsActive *bool    `json:"is_active,omitempty"`
}
