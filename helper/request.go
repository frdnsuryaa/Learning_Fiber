package helper

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"api-students/app/model"
)

// RequestContext memberi timeout untuk setiap operasi database.
func RequestContext(c *fiber.Ctx) (context.Context, context.CancelFunc) {
	return context.WithTimeout(c.UserContext(), 5*time.Second)
}

// ParamID membaca parameter :id dari jalur dan memastikan bentuknya benar.
func ParamID(c *fiber.Ctx) (int, bool) {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id < 1 {
		return 0, false
	}
	return id, true
}

// ── User query parser ──

var allowedSort = map[string]bool{
	"id": true, "username": true, "email": true, "created_at": true,
}

// ParseListQuery membaca query string untuk endpoint list users dan memberi nilai bawaan yang aman.
func ParseListQuery(c *fiber.Ctx) model.ListQuery {
	q := model.ListQuery{
		Page:  c.QueryInt("page", 1),
		Limit: c.QueryInt("limit", 10),
		Search: strings.TrimSpace(c.Query("search")),
		Sort:  c.Query("sort", "id"),
		Order: strings.ToLower(c.Query("order", "asc")),
	}
	if q.Page < 1 {
		q.Page = 1
	}
	if q.Limit < 1 {
		q.Limit = 10
	}
	if q.Limit > 100 {
		q.Limit = 100
	}
	if !allowedSort[q.Sort] {
		q.Sort = "id"
	}
	if q.Order != "desc" {
		q.Order = "asc"
	}
	if raw := c.Query("is_active"); raw != "" {
		if v, err := strconv.ParseBool(raw); err == nil {
			q.IsActive = &v
		}
	}
	return q
}

// ── Student query parser ──

// sortableStudentFields adalah whitelist field yang boleh digunakan sebagai parameter sort.
var sortableStudentFields = map[string]bool{
	"name":      true,
	"grade":     true,
	"is_active": true,
}

// ParseStudentQuery mem-parse dan memvalidasi semua query string untuk GET /students.
// Mengembalikan StudentQuery yang siap dipakai atau pesan error jika ada parameter tidak valid.
func ParseStudentQuery(c *fiber.Ctx) (model.StudentQuery, string) {
	q := model.StudentQuery{
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
		if !sortableStudentFields[raw] {
			return q, "Parameter 'sort' hanya boleh berisi: name, grade, is_active"
		}
		q.Sort = raw
	}

	// --- order ---
	if raw := c.Query("order"); raw != "" {
		orderLower := strings.ToLower(raw)
		if orderLower != "asc" && orderLower != "desc" {
			return q, "Parameter 'order' hanya boleh berisi: asc atau desc"
		}
		q.Order = orderLower
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