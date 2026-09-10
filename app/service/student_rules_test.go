package service

import (
	"testing"

	"api-students/app/model"
)

// ── Test ValidateCreateStudent (POST) ──

func TestValidateCreateStudent_OK(t *testing.T) {
	req := model.CreateStudentRequest{
		NIM:  "12345",
		Name: "Budi",
	}
	errs := ValidateCreateStudent(req)
	if len(errs) != 0 {
		t.Fatalf("seharusnya tidak ada error, dapat: %v", errs)
	}
}

func TestValidateCreateStudent_NIM_Kosong(t *testing.T) {
	req := model.CreateStudentRequest{Name: "Budi"}
	errs := ValidateCreateStudent(req)
	if _, ok := errs["nim"]; !ok {
		t.Fatal("seharusnya ada error untuk field nim")
	}
}

func TestValidateCreateStudent_Name_Kosong(t *testing.T) {
	req := model.CreateStudentRequest{NIM: "12345"}
	errs := ValidateCreateStudent(req)
	if _, ok := errs["name"]; !ok {
		t.Fatal("seharusnya ada error untuk field name")
	}
}

func TestValidateCreateStudent_Grade_Melebihi_Batas(t *testing.T) {
	req := model.CreateStudentRequest{NIM: "12345", Name: "Budi", Grade: 5.0}
	errs := ValidateCreateStudent(req)
	if _, ok := errs["grade"]; !ok {
		t.Fatal("seharusnya ada error untuk field grade")
	}
}

// ── Test ValidateUpdateStudent (PUT) ──

func TestValidateUpdateStudent_OK(t *testing.T) {
	req := model.UpdateStudentRequest{
		NIM:   "12345",
		Name:  "Budi",
		Grade: 3.5,
	}
	errs := ValidateUpdateStudent(req)
	if len(errs) != 0 {
		t.Fatalf("seharusnya tidak ada error, dapat: %v", errs)
	}
}

func TestValidateUpdateStudent_NIM_Kosong(t *testing.T) {
	req := model.UpdateStudentRequest{Name: "Budi", Grade: 3.0}
	errs := ValidateUpdateStudent(req)
	if _, ok := errs["nim"]; !ok {
		t.Fatal("seharusnya ada error untuk field nim pada PUT")
	}
}

func TestValidateUpdateStudent_Semua_Kosong(t *testing.T) {
	req := model.UpdateStudentRequest{}
	errs := ValidateUpdateStudent(req)
	if len(errs) < 2 {
		t.Fatalf("seharusnya ada minimal 2 error, dapat: %v", errs)
	}
}

// ── Test ApplyStudentPatch (PATCH) ──

func TestApplyStudentPatch_Partial(t *testing.T) {
	current := model.Student{
		ID:       "abc",
		NIM:      "11111",
		Name:     "Lama",
		Grade:    2.0,
		IsActive: true,
	}
	newName := "Baru"
	req := model.PatchStudentRequest{Name: &newName}

	patched, errs := ApplyStudentPatch(current, req)
	if len(errs) != 0 {
		t.Fatalf("seharusnya tidak ada error, dapat: %v", errs)
	}
	if patched.Name != "Baru" {
		t.Fatalf("nama seharusnya 'Baru', dapat: %s", patched.Name)
	}
	// Field lain tidak berubah
	if patched.NIM != "11111" {
		t.Fatalf("nim seharusnya tetap '11111', dapat: %s", patched.NIM)
	}
}

func TestApplyStudentPatch_Name_Kosong(t *testing.T) {
	current := model.Student{NIM: "11111", Name: "Lama"}
	empty := ""
	req := model.PatchStudentRequest{Name: &empty}

	_, errs := ApplyStudentPatch(current, req)
	if _, ok := errs["name"]; !ok {
		t.Fatal("seharusnya ada error untuk field name yang kosong")
	}
}

func TestApplyStudentPatch_Grade_Invalid(t *testing.T) {
	current := model.Student{NIM: "11111", Name: "Budi", Grade: 3.0}
	badGrade := 5.0
	req := model.PatchStudentRequest{Grade: &badGrade}

	_, errs := ApplyStudentPatch(current, req)
	if _, ok := errs["grade"]; !ok {
		t.Fatal("seharusnya ada error untuk grade di luar batas")
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

// ── Test IsEmptyStudentPatch ──

func TestIsEmptyStudentPatch_True(t *testing.T) {
	req := model.PatchStudentRequest{}
	if !IsEmptyStudentPatch(req) {
		t.Fatal("seharusnya true untuk request kosong")
	}
}

func TestIsEmptyStudentPatch_False(t *testing.T) {
	name := "Tes"
	req := model.PatchStudentRequest{Name: &name}
	if IsEmptyStudentPatch(req) {
		t.Fatal("seharusnya false untuk request dengan field terisi")
	}
}
