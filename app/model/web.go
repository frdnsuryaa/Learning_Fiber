package model

// WebResponse adalah amplop (envelope) seragam untuk semua respons API.
type WebResponse struct {
	Success bool              `json:"success"`
	Message string            `json:"message"`
	Data    any               `json:"data,omitempty"`
	Meta    *Meta             `json:"meta,omitempty"`
	Errors  map[string]string `json:"errors,omitempty"`
}

// Meta berisi informasi paginasi yang disertakan dalam respons list.
type Meta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// ── User request/response structs ──

// CreateUserRequest digunakan untuk POST /users.
type CreateUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// ReplaceUserRequest digunakan untuk PUT /users/:id.
type ReplaceUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
}

// PatchUserRequest digunakan untuk PATCH /users/:id.
// Semua field bersifat opsional; hanya yang dikirim yang diperbarui.
type PatchUserRequest struct {
	Username *string `json:"username,omitempty"`
	Email    *string `json:"email,omitempty"`
	IsActive *bool   `json:"is_active,omitempty"`
}
