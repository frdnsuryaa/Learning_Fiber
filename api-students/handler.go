package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)


var students = []Student{}


// Mengembalikan seluruh daftar siswa.
func GetAllStudents(c *fiber.Ctx) error {
	return SuccessResponse(c, fiber.StatusOK, "Students retrieved successfully", students)
}


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
