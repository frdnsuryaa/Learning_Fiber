package config

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitLogger(t *testing.T) {
	f, err := InitLogger()
	if err != nil {
		t.Fatalf("InitLogger gagal: %v", err)
	}
	defer f.Close()

	slog.Info("test_log_entry",
		"request_id", "test-uuid-1234",
		"method", "GET",
		"path", "/students",
		"status", 200,
		"duration_ms", 12,
	)

	// Verifikasi file logs/app.log terbuat dan terisi
	content, err := os.ReadFile(filepath.Join(logDir, logFile))
	if err != nil {
		t.Fatalf("Gagal membaca file log: %v", err)
	}

	if !strings.Contains(string(content), "test-uuid-1234") {
		t.Fatalf("File log tidak memuat data yang diharapkan: %s", string(content))
	}
}
