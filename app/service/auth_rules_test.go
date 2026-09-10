package service

import (
	"testing"

	"api-students/app/model"
)

func TestValidateRegisterPasswordTerlaluPendek(t *testing.T) {
	errs := ValidateRegister(model.RegisterRequest{
		Username: "sari", Email: "sari@example.com", Password: "abc123",
	})
	if errs["password"] != "minimal 8 karakter" {
		t.Fatalf("hasil tidak sesuai: %v", errs)
	}
}

func TestValidateRegisterPasswordTanpaAngka(t *testing.T) {
	errs := ValidateRegister(model.RegisterRequest{
		Username: "sari", Email: "sari@example.com", Password: "password",
	})
	if errs["password"] != "harus memuat huruf dan angka" {
		t.Fatalf("hasil tidak sesuai: %v", errs)
	}
}

func TestValidateRegisterPasswordUmum(t *testing.T) {
	errs := ValidateRegister(model.RegisterRequest{
		Username: "sari", Email: "sari@example.com", Password: "password1",
	})
	if errs["password"] != "password terlalu umum" {
		t.Fatalf("hasil tidak sesuai: %v", errs)
	}
}

func TestValidateRegisterPasswordValid(t *testing.T) {
	errs := ValidateRegister(model.RegisterRequest{
		Username: "sari", Email: "sari@example.com", Password: "rahasia123",
	})
	if _, ok := errs["password"]; ok {
		t.Fatalf("password seharusnya valid: %v", errs)
	}
}
