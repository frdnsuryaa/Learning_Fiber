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
	Page        int `json:"page,omitempty"`
	Limit       int `json:"limit,omitempty"`
	CurrentPage int `json:"current_page,omitempty"`
	PerPage     int `json:"per_page,omitempty"`
	Total       int `json:"total"`
	TotalPages  int `json:"total_pages,omitempty"`
	LastPage    int `json:"last_page,omitempty"`
}

