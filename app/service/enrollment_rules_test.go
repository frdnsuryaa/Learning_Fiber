package service

import (
	"testing"

	"api-students/app/model"
)

func TestValidateEnrollmentRequestAcademicYear(t *testing.T) {
	tests := []struct {
		name         string
		academicYear string
		wantValid    bool
	}{
		{name: "ganjil with required separator", academicYear: "2026/2027-Ganjil", wantValid: true},
		{name: "genap with required separator", academicYear: "2026/2027-Genap", wantValid: true},
		{name: "surrounding whitespace", academicYear: " 2026/2027-Ganjil ", wantValid: true},
		{name: "missing separator", academicYear: "2026/2027Ganjil"},
		{name: "invalid year digits", academicYear: "202A/2027-Ganjil"},
		{name: "invalid semester", academicYear: "2026/2027-Remedial"},
		{name: "missing second year", academicYear: "2026/-Ganjil"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateEnrollmentRequest(model.EnrollmentRequest{
				CourseID:      1,
				TahunAkademik: tt.academicYear,
			})
			_, invalid := errs["tahun_akademik"]
			if gotValid := !invalid; gotValid != tt.wantValid {
				t.Fatalf("year %q valid = %v, want %v (errors: %v)",
					tt.academicYear, gotValid, tt.wantValid, errs)
			}
		})
	}
}
