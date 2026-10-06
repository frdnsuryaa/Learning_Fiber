package helper

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"api-students/app/model"
)

// ParseStudentQuery mem-parse parameter query string untuk GET /api/v1/students.
func ParseStudentQuery(c *fiber.Ctx) model.StudentQuery {
	q := model.StudentQuery{
		Page:    1,
		PerPage: 10,
		Sort:    "nama",
	}

	if v := c.QueryInt("page", 1); v >= 1 {
		q.Page = v
	}

	if v := c.QueryInt("per_page", 10); v >= 1 {
		if v > 50 {
			v = 50
		}
		q.PerPage = v
	}

	q.Search = strings.TrimSpace(c.Query("search"))
	q.Prodi = strings.TrimSpace(c.Query("prodi"))

	if raw := c.Query("angkatan"); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil {
			q.Angkatan = &v
		}
	}

	if raw := c.Query("sort"); raw == "nama" || raw == "-ipk_terakhir" {
		q.Sort = raw
	}

	return q
}

