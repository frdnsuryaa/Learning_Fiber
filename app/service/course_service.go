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

	q := model.CourseQuery{
		Search: strings.TrimSpace(c.Query("search")),
	}

	if raw := c.Query("semester"); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil {
			q.Semester = &v
		}
	}

	if c.Query("available") == "true" {
		q.Available = true
	}

	courses, err := s.repo.FindAll(c.Context(), q)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal mengambil data mata kuliah")
	}

	return helper.Success(c, fiber.StatusOK, "data mata kuliah berhasil diambil", courses)
}
