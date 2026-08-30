package main

import (
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// sortableFields adalah whitelist field yang boleh digunakan sebagai parameter sort.
var sortableFields = map[string]bool{
	"name":      true,
	"grade":     true,
	"is_active": true,
}

// students adalah penyimpanan in-memory sederhana sebagai pengganti database.
var students = []Student{}

// parseStudentQuery mem-parse dan memvalidasi semua query string dari request.
// Mengembalikan StudentQuery yang siap dipakai atau pesan error jika ada parameter tidak valid.
func parseStudentQuery(c *fiber.Ctx) (StudentQuery, string) {
	q := StudentQuery{
		Page:  1,
		Limit: 10,
		Sort:  "name",
		Order: "asc",
	}

	// --- page ---
	if raw := c.Query("page"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 1 {
			return q, "Parameter 'page' harus berupa bilangan bulat positif"
		}
		q.Page = v
	}

	// --- limit (maks 100) ---
	if raw := c.Query("limit"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 1 {
			return q, "Parameter 'limit' harus berupa bilangan bulat positif"
		}
		if v > 100 {
			return q, "Parameter 'limit' tidak boleh melebihi 100"
		}
		q.Limit = v
	}

	// --- search ---
	q.Search = strings.TrimSpace(c.Query("search"))

	// --- sort (whitelist) ---
	if raw := c.Query("sort"); raw != "" {
		if !sortableFields[raw] {
			return q, "Parameter 'sort' hanya boleh berisi: name, grade, is_active"
		}
		q.Sort = raw
	}

	// --- order ---
	if raw := c.Query("order"); raw != "" {
		if raw != "asc" && raw != "desc" {
			return q, "Parameter 'order' hanya boleh berisi: asc atau desc"
		}
		q.Order = raw
	}

	// --- is_active (filter boolean opsional) ---
	if raw := c.Query("is_active"); raw != "" {
		v, err := strconv.ParseBool(raw)
		if err != nil {
			return q, "Parameter 'is_active' hanya boleh berisi: true atau false"
		}
		q.IsActive = &v
	}

	// --- grade_min (filter rentang bawah opsional) ---
	if raw := c.Query("grade_min"); raw != "" {
		v, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return q, "Parameter 'grade_min' harus berupa angka"
		}
		q.GradeMin = &v
	}

	// --- grade_max (filter rentang atas opsional) ---
	if raw := c.Query("grade_max"); raw != "" {
		v, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return q, "Parameter 'grade_max' harus berupa angka"
		}
		q.GradeMax = &v
	}

	return q, ""
}

// GetAllStudents menangani GET /students
// Mendukung paginasi, pencarian nama, pengurutan, dan filter is_active / rentang grade.
func GetAllStudents(c *fiber.Ctx) error {
	q, errMsg := parseStudentQuery(c)
	if errMsg != "" {
		return ErrorResponse(c, fiber.StatusBadRequest, errMsg)
	}

	// 1. Filter
	filtered := []Student{}
	search := strings.ToLower(q.Search)

	for _, s := range students {
		// filter search nama (case-insensitive)
		if search != "" && !strings.Contains(strings.ToLower(s.Name), search) {
			continue
		}
		// filter is_active
		if q.IsActive != nil && s.IsActive != *q.IsActive {
			continue
		}
		// filter grade_min
		if q.GradeMin != nil && s.Grade < *q.GradeMin {
			continue
		}
		// filter grade_max
		if q.GradeMax != nil && s.Grade > *q.GradeMax {
			continue
		}
		filtered = append(filtered, s)
	}

	// 2. Sort
	sort.SliceStable(filtered, func(i, j int) bool {
		var less bool
		switch q.Sort {
		case "grade":
			less = filtered[i].Grade < filtered[j].Grade
		case "is_active":
			// false < true, sehingga asc = non-aktif duluan
			less = !filtered[i].IsActive && filtered[j].IsActive
		default: // "name"
			less = strings.ToLower(filtered[i].Name) < strings.ToLower(filtered[j].Name)
		}
		if q.Order == "desc" {
			return !less
		}
		return less
	})

	// 3. Hitung meta sebelum paginate
	total := len(filtered)
	totalPages := int(math.Ceil(float64(total) / float64(q.Limit)))
	if totalPages == 0 {
		totalPages = 1
	}

	// 4. Paginate
	start := (q.Page - 1) * q.Limit
	if start >= total {
		start = total
	}
	end := start + q.Limit
	if end > total {
		end = total
	}
	paginated := filtered[start:end]

	meta := Meta{
		Page:       q.Page,
		Limit:      q.Limit,
		Total:      total,
		TotalPages: totalPages,
	}

	return PaginatedSuccessResponse(c, "Students retrieved successfully", meta, paginated)
}

// GetStudentByID menangani GET /students/:id
// Mengembalikan satu siswa berdasarkan ID-nya.
func GetStudentByID(c *fiber.Ctx) error {
	id := c.Params("id")

	for _, s := range students {
		if s.ID == id {
			return SuccessResponse(c, fiber.StatusOK, "Student retrieved successfully", s)
		}
	}

	return ErrorResponse(c, fiber.StatusNotFound, "Student not found")
}

// CreateStudent menangani POST /students
// Membuat siswa baru dari body request dan menyimpannya.
func CreateStudent(c *fiber.Ctx) error {
	var req CreateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if req.Name == "" {
		return ErrorResponse(c, fiber.StatusUnprocessableEntity, "Field 'name' is required")
	}

	newStudent := Student{
		ID:       uuid.NewString(),
		Name:     req.Name,
		Grade:    req.Grade,
		IsActive: req.IsActive,
	}

	students = append(students, newStudent)

	return SuccessResponse(c, fiber.StatusCreated, "Student created successfully", newStudent)
}

// UpdateStudent menangani PUT /students/:id
// Mengganti seluruh data siswa berdasarkan ID.
func UpdateStudent(c *fiber.Ctx) error {
	id := c.Params("id")

	var req UpdateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if req.Name == "" {
		return ErrorResponse(c, fiber.StatusUnprocessableEntity, "Field 'name' is required")
	}

	for i, s := range students {
		if s.ID == id {
			students[i] = Student{
				ID:       id,
				Name:     req.Name,
				Grade:    req.Grade,
				IsActive: req.IsActive,
			}
			return SuccessResponse(c, fiber.StatusOK, "Student updated successfully", students[i])
		}
	}

	return ErrorResponse(c, fiber.StatusNotFound, "Student not found")
}

// PatchStudent menangani PATCH /students/:id
// Memperbarui sebagian data siswa, hanya field yang dikirim yang akan diubah.
func PatchStudent(c *fiber.Ctx) error {
	id := c.Params("id")

	var req PatchStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return ErrorResponse(c, fiber.StatusBadRequest, "Invalid request body")
	}

	for i, s := range students {
		if s.ID == id {
			if req.Name != nil {
				students[i].Name = *req.Name
			}
			if req.Grade != nil {
				students[i].Grade = *req.Grade
			}
			if req.IsActive != nil {
				students[i].IsActive = *req.IsActive
			}
			return SuccessResponse(c, fiber.StatusOK, "Student patched successfully", students[i])
		}
	}

	return ErrorResponse(c, fiber.StatusNotFound, "Student not found")
}

// DeleteStudent menangani DELETE /students/:id
// Menghapus siswa berdasarkan ID
func DeleteStudent(c *fiber.Ctx) error {
	id := c.Params("id")

	for i, s := range students {
		if s.ID == id {
			students = append(students[:i], students[i+1:]...)
			return SuccessResponse(c, fiber.StatusOK, "Student deleted successfully", nil)
		}
	}

	return ErrorResponse(c, fiber.StatusNotFound, "Student not found")
}
