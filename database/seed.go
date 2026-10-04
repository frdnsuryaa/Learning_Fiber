package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

// hashPw membuat bcrypt hash password dengan cost 12.
func hashPw(plain string) string {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), 12)
	if err != nil {
		panic(fmt.Sprintf("gagal hash password seeder: %v", err))
	}
	return string(b)
}

func seedAll(ctx context.Context, pool *pgxpool.Pool) error {
	// ── 1. Admin ──────────────────────────────────────────────────────────────
	var adminID int
	err := pool.QueryRow(ctx,
		`INSERT INTO users (email, password, role)
		 VALUES ($1, $2, 'admin')
		 ON CONFLICT DO NOTHING
		 RETURNING id`,
		"admin@siakad.ac.id", hashPw("Admin@1234"),
	).Scan(&adminID)
	if err != nil {
		// Sudah ada — ambil ID-nya
		pool.QueryRow(ctx,
			`SELECT id FROM users WHERE LOWER(email) = LOWER($1)`,
			"admin@siakad.ac.id",
		).Scan(&adminID)
	}

	// ── 2. Mata Kuliah (10) ───────────────────────────────────────────────────
	courses := []struct {
		kode, nama string
		sks, sem, kuota int
	}{
		{"IF101", "Algoritma dan Pemrograman", 3, 1, 40},
		{"IF102", "Matematika Diskrit", 3, 1, 35},
		{"IF201", "Struktur Data", 3, 3, 35},
		{"IF202", "Basis Data", 3, 3, 40},
		{"IF203", "Pemrograman Berorientasi Objek", 3, 3, 35},
		{"IF301", "Pemrograman Web", 3, 5, 30},
		{"IF302", "Jaringan Komputer", 3, 5, 30},
		{"IF303", "Rekayasa Perangkat Lunak", 3, 5, 30},
		{"IF401", "Kecerdasan Buatan", 3, 7, 25},
		{"IF402", "Keamanan Informasi", 3, 7, 25},
	}
	for _, c := range courses {
		pool.Exec(ctx,
			`INSERT INTO courses (kode_mk, nama_mk, sks, semester, kuota)
			 VALUES ($1,$2,$3,$4,$5)
			 ON CONFLICT (kode_mk) DO NOTHING`,
			c.kode, c.nama, c.sks, c.sem, c.kuota,
		)
	}

	// ── 3. Mahasiswa (20) ─────────────────────────────────────────────────────
	mahasiswas := []struct {
		nim, nama, prodi, email string
		angkatan                int
		ipk                     float64
	}{
		{"202210001001", "Rina Putri", "Sistem Informasi", "rina.putri@mhs.ac.id", 2022, 3.45},
		{"202210001002", "Budi Santoso", "Sistem Informasi", "budi.santoso@mhs.ac.id", 2022, 2.80},
		{"202210001003", "Cahya Dewi", "Teknik Informatika", "cahya.dewi@mhs.ac.id", 2022, 3.75},
		{"202210001004", "Dani Pratama", "Teknik Informatika", "dani.pratama@mhs.ac.id", 2022, 2.40},
		{"202110001001", "Eka Wulandari", "Sistem Informasi", "eka.wulandari@mhs.ac.id", 2021, 3.10},
		{"202110001002", "Fajar Nugroho", "Sistem Informasi", "fajar.nugroho@mhs.ac.id", 2021, 2.65},
		{"202110001003", "Gita Rahayu", "Teknik Informatika", "gita.rahayu@mhs.ac.id", 2021, 3.90},
		{"202110001004", "Hadi Kurniawan", "Teknik Informatika", "hadi.kurniawan@mhs.ac.id", 2021, 2.20},
		{"202010001001", "Indah Lestari", "Sistem Informasi", "indah.lestari@mhs.ac.id", 2020, 3.55},
		{"202010001002", "Joko Widodo", "Sistem Informasi", "joko.widodo@mhs.ac.id", 2020, 2.75},
		{"202010001003", "Kartika Sari", "Teknik Informatika", "kartika.sari@mhs.ac.id", 2020, 3.20},
		{"202010001004", "Lukman Hakim", "Teknik Informatika", "lukman.hakim@mhs.ac.id", 2020, 2.50},
		{"202310001001", "Maya Anggraeni", "Sistem Informasi", "maya.anggraeni@mhs.ac.id", 2023, 3.85},
		{"202310001002", "Niko Prasetyo", "Sistem Informasi", "niko.prasetyo@mhs.ac.id", 2023, 2.90},
		{"202310001003", "Olivia Tanaka", "Teknik Informatika", "olivia.tanaka@mhs.ac.id", 2023, 3.60},
		{"202310001004", "Putra Halim", "Teknik Informatika", "putra.halim@mhs.ac.id", 2023, 2.30},
		{"202410001001", "Qori Maharani", "Sistem Informasi", "qori.maharani@mhs.ac.id", 2024, 3.00},
		{"202410001002", "Rizky Firmansyah", "Sistem Informasi", "rizky.firmansyah@mhs.ac.id", 2024, 2.55},
		{"202410001003", "Sari Puspitasari", "Teknik Informatika", "sari.puspitasari@mhs.ac.id", 2024, 3.70},
		{"202410001004", "Tono Wibisono", "Teknik Informatika", "tono.wibisono@mhs.ac.id", 2024, 2.48},
	}

	for _, m := range mahasiswas {
		var uid int
		// Password awal = NIM
		err := pool.QueryRow(ctx,
			`INSERT INTO users (email, password, role)
			 VALUES ($1,$2,'mahasiswa')
			 ON CONFLICT DO NOTHING
			 RETURNING id`,
			m.email, hashPw(m.nim),
		).Scan(&uid)
		if err != nil {
			pool.QueryRow(ctx,
				`SELECT id FROM users WHERE LOWER(email) = LOWER($1)`,
				m.email,
			).Scan(&uid)
		}
		if uid == 0 {
			continue
		}
		pool.Exec(ctx,
			`INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir)
			 VALUES ($1,$2,$3,$4,$5,$6)
			 ON CONFLICT (nim) DO NOTHING`,
			uid, m.nim, m.nama, m.prodi, m.angkatan, m.ipk,
		)
	}
	return nil
}
