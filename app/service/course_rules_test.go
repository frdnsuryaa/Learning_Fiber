package service

import "testing"

func TestParseCourseQuery(t *testing.T) {
	tests := []struct {
		name          string
		semester      string
		search        string
		available     string
		wantSemester  int
		wantSearch    string
		wantAvailable bool
		wantError     string
	}{
		{
			name:          "valid filters",
			semester:      "3",
			search:        "  algoritma ",
			available:     "true",
			wantSemester:  3,
			wantSearch:    "algoritma",
			wantAvailable: true,
		},
		{
			name:      "available false",
			available: "false",
		},
		{
			name:      "invalid semester",
			semester:  "ganjil",
			wantError: "semester",
		},
		{
			name:      "nonpositive semester",
			semester:  "0",
			wantError: "semester",
		},
		{
			name:      "invalid available",
			available: "yes",
			wantError: "available",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query, errs := parseCourseQuery(tt.semester, tt.search, tt.available)
			if tt.wantError != "" {
				if _, ok := errs[tt.wantError]; !ok {
					t.Fatalf("expected error for %q; got %v", tt.wantError, errs)
				}
				return
			}
			if len(errs) != 0 {
				t.Fatalf("unexpected query errors: %v", errs)
			}
			if query.Semester != nil && *query.Semester != tt.wantSemester {
				t.Errorf("semester = %d, want %d", *query.Semester, tt.wantSemester)
			}
			if tt.wantSemester != 0 && query.Semester == nil {
				t.Error("semester should be parsed")
			}
			if query.Search != tt.wantSearch {
				t.Errorf("search = %q, want %q", query.Search, tt.wantSearch)
			}
			if query.Available != tt.wantAvailable {
				t.Errorf("available = %v, want %v", query.Available, tt.wantAvailable)
			}
		})
	}
}
