package config

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

const (
	logDir       = "logs"
	logFile      = "app.log"
	maxLogSize   = 5 * 1024 * 1024 // 5 MB
)

// InitLogger menginisialisasi structured JSON logger yang menulis ke
// layar (stdout) sekaligus ke file logs/app.log.
// Jika file log melebihi maxLogSize, file lama di-rename dengan timestamp (rotasi).
func InitLogger() (*os.File, error) {
	// Pastikan folder logs/ ada
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return nil, fmt.Errorf("gagal membuat folder logs: %w", err)
	}

	logPath := filepath.Join(logDir, logFile)

	// Rotasi sederhana: jika file sudah besar, rename dengan timestamp
	if info, err := os.Stat(logPath); err == nil && info.Size() > maxLogSize {
		rotated := filepath.Join(logDir,
			fmt.Sprintf("app_%s.log", time.Now().Format("20060102_150405")))
		_ = os.Rename(logPath, rotated)
	}

	// Buka (atau buat) file log dalam mode append
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, fmt.Errorf("gagal membuka file log: %w", err)
	}

	// MultiWriter: stdout + file
	multi := io.MultiWriter(os.Stdout, f)

	// Ganti default slog handler dengan JSON handler
	handler := slog.NewJSONHandler(multi, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	slog.SetDefault(slog.New(handler))

	return f, nil
}
