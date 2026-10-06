package service

import (
	"strings"
	"unicode"

	"api-students/app/model"
)

const minPasswordLength = 8

// ValidateLoginRequest memvalidasi POST /api/v1/auth/login.
func ValidateLoginRequest(req model.LoginRequest) map[string]string {
	errs := map[string]string{}
	if !isValidEmail(req.Email) {
		errs["email"] = "format email tidak valid"
	}
	if req.Password == "" {
		errs["password"] = "wajib diisi"
	} else if len(req.Password) < minPasswordLength {
		errs["password"] = "minimal 8 karakter"
	}
	return errs
}

// ValidateRegister dipertahankan untuk kompatibilitas.
func ValidateRegister(req model.RegisterRequest) map[string]string {
	errs := map[string]string{}
	username := strings.TrimSpace(req.Username)

	switch {
	case username == "":
		errs["username"] = "wajib diisi"
	case len(username) < 3:
		errs["username"] = "minimal 3 karakter"
	case !isValidUsername(username):
		errs["username"] = "hanya boleh huruf, angka, titik, dan garis bawah"
	}

	if !isValidEmail(req.Email) {
		errs["email"] = "format email tidak valid"
	}
	if msg := checkPasswordStrength(req.Password); msg != "" {
		errs["password"] = msg
	}
	return errs
}

// ValidateLogin dipertahankan untuk kompatibilitas.
func ValidateLogin(req model.LoginRequest) map[string]string {
	return ValidateLoginRequest(req)
}

func checkPasswordStrength(password string) string {
	if len(password) < minPasswordLength {
		return "minimal 8 karakter"
	}
	var hasLetter, hasDigit bool
	for _, r := range password {
		switch {
		case unicode.IsLetter(r):
			hasLetter = true
		case unicode.IsDigit(r):
			hasDigit = true
		}
	}
	if !hasLetter || !hasDigit {
		return "harus memuat huruf dan angka"
	}
	weak := map[string]bool{
		"password1": true, "12345678": true, "qwerty123": true,
		"admin123": true, "password123": true,
	}
	if weak[strings.ToLower(password)] {
		return "password terlalu umum"
	}
	return ""
}

func isValidUsername(username string) bool {
	for _, r := range username {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '.' && r != '_' {
			return false
		}
	}
	return true
}
