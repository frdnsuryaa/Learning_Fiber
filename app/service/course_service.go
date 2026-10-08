package service

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"api-students/app/model"
	"api-students/app/repository"
	"api-students/helper"
)

// CourseService menangani HTTP request untuk entitas Course.
type CourseService struct {
	repo repository.CourseRepository
}

// NewCourseService membuat instance baru CourseService.
func NewCourseService(repo repository.CourseRepository) *CourseService {
	return &CourseService{repo: repo}
}

// GET /api/v1/courses — semua role yang sudah login
func (s *CourseService) GetAllCourses(c *fiber.Ctx) error {
	_, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
	}

	q, errs := parseCourseQuery(
		c.Query("semester"),
		c.Query("search"),
		c.Query("available"),
	)
	if len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}

	courses, err := s.repo.FindAll(c.Context(), q)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal mengambil data mata kuliah")
	}

	return helper.Success(c, fiber.StatusOK, "data mata kuliah berhasil diambil", courses)
}

func parseCourseQuery(semester, search, available string) (model.CourseQuery, map[string]string) {
	q := model.CourseQuery{Search: strings.TrimSpace(search)}
	errs := make(map[string]string)

	if raw := strings.TrimSpace(semester); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 1 {
			errs["semester"] = "harus berupa angka positif"
		} else {
			q.Semester = &value
		}
	}

	switch strings.ToLower(strings.TrimSpace(available)) {
	case "", "false":
	case "true":
		q.Available = true
	default:
		errs["available"] = "harus bernilai true atau false"
	}

	return q, errs
}
