package service

import (
	"testing"

	"api-students/app/model"
)

// ── Test ValidateCreateStudentRequest (POST) ──

func TestValidateCreateStudentRequest_OK(t *testing.T) {
	req := model.CreateStudentRequest{
		NIM:         "123456789012",
		Nama:        "Budi Santoso",
		Email:       "budi@example.com",
		Prodi:       "Informatika",
		Angkatan:    2023,
		IPKTerakhir: 3.75,
	}
	errs := ValidateCreateStudentRequest(req)
	if len(errs) != 0 {
		t.Fatalf("seharusnya tidak ada error, dapat: %v", errs)
	}
}

func TestValidateCreateStudentRequest_NIM_Kosong(t *testing.T) {
	req := model.CreateStudentRequest{
		Nama:        "Budi Santoso",
		Email:       "budi@example.com",
		Prodi:       "Informatika",
		Angkatan:    2023,
		IPKTerakhir: 3.75,
	}
	errs := ValidateCreateStudentRequest(req)
	if _, ok := errs["nim"]; !ok {
		t.Fatal("seharusnya ada error untuk field nim")
	}
}

func TestValidateCreateStudentRequest_NIM_Bukan12Digit(t *testing.T) {
	req := model.CreateStudentRequest{
		NIM:         "12345",
		Nama:        "Budi Santoso",
		Email:       "budi@example.com",
		Prodi:       "Informatika",
		Angkatan:    2023,
		IPKTerakhir: 3.75,
	}
	errs := ValidateCreateStudentRequest(req)
	if _, ok := errs["nim"]; !ok {
		t.Fatal("seharusnya ada error untuk nim bukan 12 digit")
	}
}

func TestValidateCreateStudentRequest_Nama_Kosong(t *testing.T) {
	req := model.CreateStudentRequest{
		NIM:         "123456789012",
		Email:       "budi@example.com",
		Prodi:       "Informatika",
		Angkatan:    2023,
		IPKTerakhir: 3.75,
	}
	errs := ValidateCreateStudentRequest(req)
	if _, ok := errs["nama"]; !ok {
		t.Fatal("seharusnya ada error untuk field nama")
	}
}

func TestValidateCreateStudentRequest_Email_Invalid(t *testing.T) {
	req := model.CreateStudentRequest{
		NIM:         "123456789012",
		Nama:        "Budi Santoso",
		Email:       "invalid-email",
		Prodi:       "Informatika",
		Angkatan:    2023,
		IPKTerakhir: 3.75,
	}
	errs := ValidateCreateStudentRequest(req)
	if _, ok := errs["email"]; !ok {
		t.Fatal("seharusnya ada error untuk email invalid")
	}
}

func TestValidateCreateStudentRequest_IPK_Melebihi_Batas(t *testing.T) {
	req := model.CreateStudentRequest{
		NIM:         "123456789012",
		Nama:        "Budi Santoso",
		Email:       "budi@example.com",
		Prodi:       "Informatika",
		Angkatan:    2023,
		IPKTerakhir: 4.50,
	}
	errs := ValidateCreateStudentRequest(req)
	if _, ok := errs["ipk_terakhir"]; !ok {
		t.Fatal("seharusnya ada error untuk field ipk_terakhir")
	}
}

// ── Test ValidateUpdateStudentRequest (PUT) ──

func TestValidateUpdateStudentRequest_OK(t *testing.T) {
	req := model.UpdateStudentRequest{
		Nama:        "Budi Santoso",
		Prodi:       "Informatika",
		Angkatan:    2023,
		IPKTerakhir: 3.80,
	}
	errs := ValidateUpdateStudentRequest(req)
	if len(errs) != 0 {
		t.Fatalf("seharusnya tidak ada error, dapat: %v", errs)
	}
}

func TestValidateUpdateStudentRequest_Nama_Kosong(t *testing.T) {
	req := model.UpdateStudentRequest{
		Prodi:       "Informatika",
		Angkatan:    2023,
		IPKTerakhir: 3.80,
	}
	errs := ValidateUpdateStudentRequest(req)
	if _, ok := errs["nama"]; !ok {
		t.Fatal("seharusnya ada error untuk field nama pada PUT")
	}
}

func TestValidateUpdateStudentRequest_Semua_Kosong(t *testing.T) {
	req := model.UpdateStudentRequest{}
	errs := ValidateUpdateStudentRequest(req)
	if len(errs) < 2 {
		t.Fatalf("seharusnya ada minimal 2 error, dapat: %v", errs)
	}
}

// ── Test sksBatas ──

func TestSksBatas(t *testing.T) {
	cases := []struct {
		ipk  float64
		want int
	}{
		{4.00, 24},
		{3.50, 24},
		{3.00, 24},
		{2.99, 21},
		{2.50, 21},
		{2.49, 18},
		{1.00, 18},
		{0.00, 18},
	}
	for _, tc := range cases {
		got := sksBatas(tc.ipk)
		if got != tc.want {
			t.Errorf("sksBatas(%.2f) = %d, want %d", tc.ipk, got, tc.want)
		}
	}
}

// ── Test CountTotalPages ──

func TestCountTotalPages(t *testing.T) {
	cases := []struct {
		total, limit, want int
	}{
		{0, 10, 0},
		{10, 10, 1},
		{11, 10, 2},
		{100, 10, 10},
		{1, 10, 1},
		{0, 0, 0}, // limit nol
	}
	for _, tc := range cases {
		got := CountTotalPages(tc.total, tc.limit)
		if got != tc.want {
			t.Errorf("CountTotalPages(%d, %d) = %d, want %d",
				tc.total, tc.limit, got, tc.want)
		}
	}
}

