# API Students (Go Fiber + PostgreSQL)

REST API untuk manajemen data mahasiswa, dibangun menggunakan [Go Fiber](https://gofiber.io/) dan basis data **PostgreSQL** dengan koneksi pool (`pgxpool`).

---

## 📋 Daftar Isi
- [Prasyarat](#-prasyarat)
- [Variabel Environment](#-variabel-environment)
- [Menyiapkan Basis Data dari Nol](#-menyiapkan-basis-data-dari-nol)
- [Skema Tabel & Migrasi](#-skema-tabel--migrasi)
- [Menjalankan Aplikasi](#-menjalankan-aplikasi)
- [Amplop Respons](#-amplop-respons)
- [Kontrak API & Endpoint](#-kontrak-api--endpoint)
- [Struktur Proyek](#-struktur-proyek)

---

## ⚙️ Prasyarat

Sebelum memulai, pastikan perangkat Anda telah terpasang:
1. **Go** (versi 1.22 atau lebih baru) — [Unduh Go](https://go.dev/dl/)
2. **PostgreSQL** (versi 14 atau lebih baru) — [Unduh PostgreSQL](https://www.postgresql.org/download/)
3. Terminal / Command Prompt / PowerShell
4. API Client seperti **Postman**, **Thunder Client**, atau `curl`

---

## 🔐 Variabel Environment

Aplikasi membaca konfigurasi dari berkas `.env` di direktori utama.

> [!IMPORTANT]
> Berkas `.env` berisi kredensial sensitif dan **tidak boleh di-commit** ke repositori Git (sudah ditambahkan ke `.gitignore`).

### 1. Buat berkas `.env` dari `.env.example`
Repositori menyertakan template `.env.example` dengan nilai kosong:

```env
APP_PORT=
DB_HOST=
DB_PORT=
DB_USER=
DB_PASSWORD=
DB_NAME=
DB_SSLMODE=
DB_MAX_CONNS=
```

Salin template menjadi `.env`:
```bash
# Di Windows PowerShell / CMD:
copy .env.example .env

# Di Linux / macOS / Git Bash:
cp .env.example .env
```

### 2. Daftar variabel yang diperlukan

Isi nilai di dalam `.env` sesuai kredensial PostgreSQL lokal Anda:

| Variabel | Tipe | Contoh Nilai | Keterangan |
| :--- | :--- | :--- | :--- |
| `APP_PORT` | `int` | `3000` | Port tempat server HTTP Fiber berjalan |
| `DB_HOST` | `string` | `localhost` | Host/IP server PostgreSQL |
| `DB_PORT` | `int` | `5432` | Port server PostgreSQL |
| `DB_USER` | `string` | `postgres` | Username akun PostgreSQL |
| `DB_PASSWORD` | `string` | *(password anda)* | Password akun PostgreSQL Anda |
| `DB_NAME` | `string` | `praktikum_backend` | Nama database yang digunakan |
| `DB_SSLMODE` | `string` | `disable` | Mode SSL (`disable` untuk development lokal) |
| `DB_MAX_CONNS` | `int` | `10` | Jumlah koneksi maksimum pada connection pool |

Contoh isi `.env`:
```env
APP_PORT=3000
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password_postgres_anda
DB_NAME=praktikum_backend
DB_SSLMODE=disable
DB_MAX_CONNS=10
```

---

## 🗄️ Menyiapkan Basis Data dari Nol

Ikuti langkah-langkah berikut jika Anda baru pertama kali mengklona repositori ini:

### Langkah 1: Pastikan Service PostgreSQL Sedang Berjalan
Pastikan PostgreSQL service aktif di komputer Anda.

### Langkah 2: Buat Database Baru
Buka terminal dan masuk ke CLI PostgreSQL (`psql`) atau gunakan GUI seperti pgAdmin / DBeaver:

```bash
psql -U postgres
```
*(Masukkan password postgres Anda saat diminta)*

Jalankan perintah SQL untuk membuat database:
```sql
CREATE DATABASE praktikum_backend;
```
Keluar dari psql:
```sql
\q
```

### Langkah 3: Eksekusi Migrasi Tabel (Opsional / Manual)
> **Catatan:** Server Go ini sudah dilengkapi **auto-migration** saat `go run .` dieksekusi pertama kali. Namun jika ingin mengeksekusi skema database secara manual melalui file migrasi:

```bash
psql -U postgres -d praktikum_backend -f migrations/001_create_students.sql
```

---

## 📐 Skema Tabel & Migrasi

Skema tabel didefinisikan dalam berkas `migrations/001_create_students.sql`:

```sql
CREATE TABLE IF NOT EXISTS students (
    id VARCHAR(36) PRIMARY KEY,
    nim VARCHAR(20) NOT NULL,
    name VARCHAR(255) NOT NULL,
    grade DOUBLE PRECISION NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT students_nim_unique UNIQUE (nim)
);

-- Indeks selain primary key:
-- 1. Indeks pada LOWER(name) untuk pencarian nama case-insensitive
CREATE INDEX IF NOT EXISTS idx_students_name_lower
    ON students (LOWER(name));

-- 2. Indeks pada created_at untuk pengurutan data terbaru
CREATE INDEX IF NOT EXISTS idx_students_created_at
    ON students (created_at DESC);
```

### Penjelasan Kolom & Batasan (Constraints):
- **`id`** (`VARCHAR(36)`): Primary Key berupa string UUID unik (dihasilkan otomatis oleh server).
- **`nim`** (`VARCHAR(20) NOT NULL UNIQUE`): Nomor Induk Mahasiswa, wajib unik untuk setiap mahasiswa.
- **`name`** (`VARCHAR(255) NOT NULL`): Nama lengkap mahasiswa.
- **`grade`** (`DOUBLE PRECISION NOT NULL DEFAULT 0`): Nilai akademik mahasiswa (angka desimal / float).
- **`is_active`** (`BOOLEAN NOT NULL DEFAULT TRUE`): Status keaktifan mahasiswa.
- **`created_at`** (`TIMESTAMPTZ NOT NULL DEFAULT NOW()`): Timestamp pencatatan waktu data dibuat.

---

## 💡 Penjelasan Teknis & Desain Database

### 1. Mengapa Keunikan NIM Lebih Baik Dijaga oleh Basis Data daripada Kode Go?
1. **Mencegah *Race Condition* (Konkurensi):**
   Jika dua request POST dengan NIM yang sama masuk secara bersamaan (*concurrent requests*), pengecekan manual di kode Go (`SELECT ... WHERE nim = ?`) pada kedua goroutine bisa sama-sama menghasilkan "belum ada", sehingga keduanya melakukan `INSERT` dan terjadi data ganda. Basis data menjamin keunikan secara *atomic* di tingkat ACID / transaction engine melalui `UNIQUE CONSTRAINT`.
2. **Integritas Data Tunggal (*Single Source of Truth*):**
   Jika basis data diakses oleh beberapa *instance* backend (skala horizontal), microservices lain, script migrasi, atau tool admin (seperti DBeaver / pgAdmin), aturan keunikan tetap terlindungi dan tidak bergantung pada apakah kode aplikasi mengimplementasikan validasi atau tidak.
3. **Efisiensi & Performa:**
   Memeriksa keunikan di Go memerlukan ekstra 1 query `SELECT` sebelum setiap operasi `INSERT`/`UPDATE` (menambah round-trip jaringan). Dengan *database unique constraint*, validasi keunikan dilakukan langsung saat operasi write secara optimal menggunakan indeks internal.

### 2. Mengapa Perlu Menambahkan Indeks Selain Kunci Primer?
1. **Indeks `idx_students_name_lower` (`LOWER(name)`):**
   Endpoint `GET /students?search=...` sering melakukan filter pencarian nama tanpa membedakan huruf besar/kecil (*case-insensitive*). Tanpa indeks ekspresi ini, PostgreSQL harus memindai seluruh tabel baris demi baris (*Full Table Scan* / `Seq Scan`) dan menjalankan fungsi `LOWER()` untuk setiap baris. Dengan B-Tree Index pada `LOWER(name)`, pencarian teks menjadi jauh lebih cepat ($O(\log N)$).
2. **Indeks `idx_students_created_at` (`created_at DESC`):**
   Digunakan untuk mengoptimalkan query pengurutan atau audit log berdasarkan waktu pembuatan data terbaru tanpa perlu melakukan proses sorting di memori (*in-memory sort*).

---

## 🚀 Menjalankan Aplikasi

1. **Unduh seluruh dependency Go:**
   ```bash
   go mod download
   ```

2. **Jalankan server:**
   ```bash
   go run .
   ```

3. **Verifikasi server & koneksi database:**
   Buka browser atau kirim request GET ke:
   ```text
   http://localhost:3000/health
   ```
   Respons yang diharapkan:
   ```json
   {
     "status": "ok",
     "database": "connected",
     "message": "Server dan basis data berjalan dengan baik"
   }
   ```

---

## 📦 Amplop Respons

Semua endpoint mengembalikan JSON dengan amplop (*envelope*) standar dan konsisten.

**1. Sukses (Single item / Mutasi data):**
```json
{
  "success": true,
  "message": "Student created successfully",
  "data": { ... }
}
```

**2. Sukses dengan Paginasi (GET /students):**
```json
{
  "success": true,
  "message": "Students retrieved successfully",
  "meta": {
    "page": 1,
    "limit": 10,
    "total": 42,
    "total_pages": 5
  },
  "data": [ ... ]
}
```

**3. Gagal / Error:**
```json
{
  "success": false,
  "message": "Pesan deskripsi kesalahan",
  "data": null
}
```

---

## 📖 Kontrak API & Endpoint

| Metode | Endpoint | Query Parameter | Deskripsi | Status Code |
| :--- | :--- | :--- | :--- | :--- |
| `GET` | `/` | — | Welcome message | `200` |
| `GET` | `/health` | — | Health check server & koneksi basis data | `200`, `503` |
| `GET` | `/students` | Paginasi, Filter, Sort (lihat detail) | Mengambil daftar mahasiswa | `200`, `400` |
| `GET` | `/students/:id` | `:id` (UUID) | Mengambil detail 1 mahasiswa | `200`, `404` |
| `POST` | `/students` | — | Menambahkan mahasiswa baru | `201`, `400`, `409`, `422` |
| `PUT` | `/students/:id` | `:id` (UUID) | Mengganti seluruh data mahasiswa | `200`, `400`, `409`, `404`, `422` |
| `PATCH` | `/students/:id` | `:id` (UUID) | Memperbarui sebagian data mahasiswa | `200`, `400`, `409`, `404` |
| `DELETE` | `/students/:id` | `:id` (UUID) | Menghapus data mahasiswa | `200`, `404` |

---

### Contoh Request & Response

#### 1. GET `/health` (Health Check)
```http
GET /health
```
**Response `200 OK` (Database Terhubung):**
```json
{
  "status": "ok",
  "database": "connected",
  "message": "Server dan basis data berjalan dengan baik"
}
```
**Response `503 Service Unavailable` (Jika Database Terputus):**
```json
{
  "status": "error",
  "database": "disconnected",
  "message": "Koneksi ke basis data gagal",
  "error": "dial tcp [::1]:5432: connect: connection refused"
}
```

#### 2. POST `/students` (Membuat Mahasiswa Baru)
```http
POST /students
Content-Type: application/json

{
  "nim": "220101001",
  "name": "Budi Santoso",
  "grade": 88.5,
  "is_active": true
}
```
**Response `201 Created`:**
```json
{
  "success": true,
  "message": "Student created successfully",
  "data": {
    "id": "e6a0d4c8-3c94-4d89-bcf8-3486c91a3291",
    "nim": "220101001",
    "name": "Budi Santoso",
    "grade": 88.5,
    "is_active": true
  }
}
```

**Response `409 Conflict` (Jika NIM sudah ada):**
```json
{
  "success": false,
  "message": "NIM sudah terdaftar",
  "data": null
}
```

#### 3. GET `/students` (Dengan Query Paginasi & Filter)
```http
GET /students?page=1&limit=5&search=budi&sort=grade&order=desc
```
**Response `200 OK`:**
```json
{
  "success": true,
  "message": "Students retrieved successfully",
  "meta": {
    "page": 1,
    "limit": 5,
    "total": 1,
    "total_pages": 1
  },
  "data": [
    {
      "id": "e6a0d4c8-3c94-4d89-bcf8-3486c91a3291",
      "nim": "220101001",
      "name": "Budi Santoso",
      "grade": 88.5,
      "is_active": true
    }
  ]
}
```

#### 4. GET `/students/:id`
```http
GET /students/e6a0d4c8-3c94-4d89-bcf8-3486c91a3291
```
**Response `200 OK`:**
```json
{
  "success": true,
  "message": "Student retrieved successfully",
  "data": {
    "id": "e6a0d4c8-3c94-4d89-bcf8-3486c91a3291",
    "nim": "220101001",
    "name": "Budi Santoso",
    "grade": 88.5,
    "is_active": true
  }
}
```

#### 5. PUT `/students/:id`
```http
PUT /students/e6a0d4c8-3c94-4d89-bcf8-3486c91a3291
Content-Type: application/json

{
  "nim": "220101001",
  "name": "Budi Santoso, S.Kom",
  "grade": 92.0,
  "is_active": true
}
```
**Response `200 OK`:**
```json
{
  "success": true,
  "message": "Student updated successfully",
  "data": {
    "id": "e6a0d4c8-3c94-4d89-bcf8-3486c91a3291",
    "nim": "220101001",
    "name": "Budi Santoso, S.Kom",
    "grade": 92.0,
    "is_active": true
  }
}
```

#### 6. PATCH `/students/:id`
```http
PATCH /students/e6a0d4c8-3c94-4d89-bcf8-3486c91a3291
Content-Type: application/json

{
  "grade": 95.0
}
```
**Response `200 OK`:**
```json
{
  "success": true,
  "message": "Student patched successfully",
  "data": {
    "id": "e6a0d4c8-3c94-4d89-bcf8-3486c91a3291",
    "nim": "220101001",
    "name": "Budi Santoso, S.Kom",
    "grade": 95.0,
    "is_active": true
  }
}
```

#### 7. DELETE `/students/:id`
```http
DELETE /students/e6a0d4c8-3c94-4d89-bcf8-3486c91a3291
```
**Response `200 OK`:**
```json
{
  "success": true,
  "message": "Student deleted successfully",
  "data": null
}
```

---

## 🏗️ Struktur Proyek

```
api-students/
├── app/
│   ├── model/
│   │   ├── student.go       # Model data Student & StudentQuery
│   │   └── user.go          # Model data User
│   └── repository/
│       ├── student_repository.go # Operasi database Student (PostgreSQL)
│       └── user_repository.go    # Operasi database User (PostgreSQL)
├── config/
│   └── env.go               # Loader & parser konfigurasi .env
├── database/
│   └── postgres.go          # Inisialisasi Connection Pool (pgxpool) & Ping
├── migrations/
│   ├── 001_create_students.sql # Skema tabel students & indeks
│   └── 001_create_users.sql    # Skema tabel users
├── handler.go               # HTTP Handler CRUD Student
├── helper.go                # Response envelope JSON & formatting
├── model.go                 # Request struct & alias model
├── main.go                  # Routing, health check & entry point aplikasi
├── .env.example             # Template konfigurasi environment (nilai kosong)
├── .gitignore               # Mengabaikan .env dan binary
├── go.mod
└── go.sum
```
