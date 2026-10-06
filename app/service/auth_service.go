package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"api-students/app/model"
	"api-students/app/repository"
	"api-students/helper"
)

const refreshTokenBytes = 32

type AuthService struct {
	users      repository.UserRepository
	tokens     repository.TokenRepository
	students   repository.StudentRepository
	jwt        *helper.JWTManager
	refreshTTL time.Duration
}

func NewAuthService(
	users repository.UserRepository,
	tokens repository.TokenRepository,
	students repository.StudentRepository,
	jwtManager *helper.JWTManager,
	refreshTTL time.Duration,
) *AuthService {
	return &AuthService{
		users:      users,
		tokens:     tokens,
		students:   students,
		jwt:        jwtManager,
		refreshTTL: refreshTTL,
	}
}

// POST /api/v1/auth/login
func (s *AuthService) Login(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	var req model.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}
	req.Email = strings.TrimSpace(req.Email)

	if errs := ValidateLoginRequest(req); len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}

	user, err := s.users.FindByEmail(ctx, req.Email)
	if err != nil {
		helper.VerifyDummyPassword(req.Password) // anti-timing attack
		return helper.Fail(c, fiber.StatusUnauthorized, "email atau password salah")
	}
	if !helper.VerifyPassword(user.Password, req.Password) {
		return helper.Fail(c, fiber.StatusUnauthorized, "email atau password salah")
	}

	// Mahasiswa yang sudah di-soft delete tidak boleh login
	if user.Role == "mahasiswa" {
		if _, err := s.students.FindByUserID(ctx, user.ID); err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return helper.Fail(c, fiber.StatusUnauthorized, "akun mahasiswa tidak aktif")
			}
		}
	}

	accessToken, err := s.jwt.GenerateAccess(user)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal membuat token")
	}

	resp := model.LoginResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   int(s.jwt.AccessTTL().Seconds()),
		User: model.UserInfo{
			ID:    user.ID,
			Email: user.Email,
			Role:  user.Role,
		},
	}
	return helper.Success(c, fiber.StatusOK, "login berhasil", resp)
}

// GET /api/v1/auth/me
func (s *AuthService) Me(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	authUser, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
	}

	user, err := s.users.FindByID(ctx, authUser.UserID)
	if err != nil {
		return helper.Fail(c, fiber.StatusUnauthorized, "user tidak ditemukan")
	}

	// Untuk mahasiswa, sertakan data students
	if user.Role == "mahasiswa" {
		student, err := s.students.FindByUserID(ctx, user.ID)
		if err == nil {
			return helper.Success(c, fiber.StatusOK, "profil berhasil diambil", fiber.Map{
				"id":    user.ID,
				"email": user.Email,
				"role":  user.Role,
				"student": fiber.Map{
					"nim":      student.NIM,
					"nama":     student.Nama,
					"prodi":    student.Prodi,
					"angkatan": student.Angkatan,
				},
			})
		}
	}

	return helper.Success(c, fiber.StatusOK, "profil berhasil diambil", fiber.Map{
		"id":    user.ID,
		"email": user.Email,
		"role":  user.Role,
	})
}

// issueTokenPair masih digunakan oleh Refresh/Logout (dipertahankan untuk kompatibilitas)
func (s *AuthService) issueTokenPair(ctx context.Context, user model.User) (model.TokenPair, error) {
	accessToken, err := s.jwt.GenerateAccess(user)
	if err != nil {
		return model.TokenPair{}, err
	}
	refreshToken, err := helper.RandomToken(refreshTokenBytes)
	if err != nil {
		return model.TokenPair{}, err
	}
	if err = s.tokens.Save(ctx, model.RefreshToken{
		UserID: user.ID, TokenHash: helper.SHA256Hex(refreshToken),
		ExpiresAt: time.Now().Add(s.refreshTTL),
	}); err != nil {
		return model.TokenPair{}, err
	}
	return model.TokenPair{
		AccessToken: accessToken, RefreshToken: refreshToken,
		TokenType: "Bearer", ExpiresIn: int(s.jwt.AccessTTL().Seconds()),
	}, nil
}

// Refresh dan Logout dipertahankan agar route lama tidak break
func (s *AuthService) Refresh(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	var req model.RefreshRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}
	if strings.TrimSpace(req.RefreshToken) == "" {
		return helper.Fail(c, fiber.StatusBadRequest, "refresh_token wajib diisi")
	}

	hash := helper.SHA256Hex(req.RefreshToken)
	stored, err := s.tokens.FindActive(ctx, hash)
	if err != nil {
		return helper.Fail(c, fiber.StatusUnauthorized, "refresh token tidak valid atau sudah kedaluwarsa")
	}

	user, err := s.users.FindByID(ctx, stored.UserID)
	if err != nil {
		return helper.Fail(c, fiber.StatusUnauthorized, "akun tidak dapat dipakai")
	}

	if err := s.tokens.Revoke(ctx, hash); err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal memperbarui token")
	}

	pair, err := s.issueTokenPair(ctx, user)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal membuat token")
	}
	return helper.Success(c, fiber.StatusOK, "token berhasil diperbarui", pair)
}

func (s *AuthService) Logout(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	var req model.RefreshRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}
	if strings.TrimSpace(req.RefreshToken) != "" {
		_ = s.tokens.Revoke(ctx, helper.SHA256Hex(req.RefreshToken))
	}
	return helper.Success(c, fiber.StatusOK, "logout berhasil", nil)
}
