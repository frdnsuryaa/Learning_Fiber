package main


type Student struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Grade    float64 `json:"grade"`
	IsActive bool    `json:"is_active"`
}


// Semua field wajib diisi; ID di-generate oleh server.
type CreateStudentRequest struct {
	Name     string  `json:"name"`
	Grade    float64 `json:"grade"`
	IsActive bool    `json:"is_active"`
}

// Merupakan replace penuh — semua field wajib diisi.
type UpdateStudentRequest struct {
	Name     string  `json:"name"`
	Grade    float64 `json:"grade"`
	IsActive bool    `json:"is_active"`
}


// Semua field bersifat opsional; hanya field yang dikirim yang akan diperbarui.
type PatchStudentRequest struct {
	Name     *string  `json:"name,omitempty"`
	Grade    *float64 `json:"grade,omitempty"`
	IsActive *bool    `json:"is_active,omitempty"`
}

// StudentQuery menampung semua parameter query yang sudah diparse dan divalidasi
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
