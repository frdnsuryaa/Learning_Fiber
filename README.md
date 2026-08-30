# API Students (Go Fiber + PostgreSQL)

REST API untuk manajemen data mahasiswa, dibangun menggunakan [Go Fiber](https://gofiber.io/) dan basis data **PostgreSQL** dengan koneksi pool (`pgxpool`).

---

## 📋 Daftar Isi
- [Prasyarat](#-prasyarat)
- [Variabel Environment](#-variabel-environment)
- [Menyiapkan Basis Data dari Nol](#-menyiapkan-basis-data-dari-nol)
- [Skema Tabel](#-skema-tabel)
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

### 1. Buat berkas `.env`
Salin berkas template `.env.example` menjadi `.env`:

```bash
# Di Windows PowerShell / CMD:
copy .env.example .env

# Di Linux / macOS / Git Bash:
cp .env.example .env
```

### 2. Daftar variabel yang diperlukan

Sesuaikan nilai di dalam `.env` dengan kredensial PostgreSQL lokal Anda:

| Variabel | Tipe | Default | Keterangan |
| :--- | :--- | :--- | :--- |
| `APP_PORT` | `int` | `3000` | Port tempat server HTTP Fiber berjalan |
| `DB_HOST` | `string` | `localhost` | Host/IP server PostgreSQL |
| `DB_PORT` | `int` | `5432` | Port server PostgreSQL |
| `DB_USER` | `string` | `postgres` | Username akun PostgreSQL |
| `DB_PASSWORD` | `string` | *(wajib disesuaikan)*  | Password akun PostgreSQL Anda |
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
psql -U postgres -d praktikum_backend -f migrations/002_create_students.sql
```

---

## 📐 Skema Tabel

Tabel `students` dirancang untuk menyimpan data mahasiswa:

```sql
CREATE TABLE IF NOT EXISTS students (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    grade DOUBLE PRECISION NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index untuk mempercepat pencarian nama case-insensitive
CREATE INDEX IF NOT EXISTS students_name_lower_idx
    ON students (LOWER(name));
```

### Penjelasan Kolom:
- **`id`** (`VARCHAR(36)`): Primary Key berupa string UUID unik (dihasilkan otomatis oleh server).
- **`name`** (`VARCHAR(255)`): Nama lengkap mahasiswa (wajib diisi).
- **`grade`** (`DOUBLE PRECISION`): Nilai akademik mahasiswa (angka desimal / float).
- **`is_active`** (`BOOLEAN`): Status aktif mahasiswa (`true`/`false`).
- **`created_at`** (`TIMESTAMPTZ`): Timestamp waktu data dibuat.

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

3. **Verifikasi server:**
   Buka browser atau kirim request GET ke:
   ```text
   http://localhost:3000/
   ```
   Respons yang diharapkan:
   ```json
   {
     "message": "Welcome to Student API",
     "status": "running"
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
| `GET` | `/` | — | Health check / status API | `200` |
| `GET` | `/students` | Paginasi, Filter, Sort (lihat detail) | Mengambil daftar mahasiswa | `200`, `400` |
| `GET` | `/students/:id` | `:id` (UUID) | Mengambil detail 1 mahasiswa | `200`, `404` |
| `POST` | `/students` | — | Menambahkan mahasiswa baru | `201`, `400`, `422` |
| `PUT` | `/students/:id` | `:id` (UUID) | Mengganti seluruh data mahasiswa | `200`, `400`, `404`, `422` |
| `PATCH` | `/students/:id` | `:id` (UUID) | Memperbarui sebagian data mahasiswa | `200`, `400`, `404` |
| `DELETE` | `/students/:id` | `:id` (UUID) | Menghapus data mahasiswa | `200`, `404` |

---

### Query Parameter — `GET /students`

| Parameter | Tipe | Default | Validasi / Aturan | Keterangan |
| :--- | :--- | :--- | :--- | :--- |
| `page` | `int` | `1` | Bilangan bulat $\ge 1$ | Nomor halaman |
| `limit` | `int` | `10` | Bilangan bulat $1 - 100$ | Jumlah data per halaman |
| `search` | `string` | `""` | Bebas | Pencarian nama (tidak membedakan huruf besar/kecil) |
| `sort` | `string` | `name` | Hanya: `name`, `grade`, `is_active` | Kolom pengurutan data |
| `order` | `string` | `asc` | Hanya: `asc`, `desc` | Arah pengurutan |
| `is_active` | `bool` | *(semua)* | Hanya: `true`, `false` | Filter status aktif |
| `grade_min` | `float` | *(semua)* | Angka valid | Filter nilai minimum ($\ge$) |
| `grade_max` | `float` | *(semua)* | Angka valid | Filter nilai maksimum ($\le$) |

---

### Contoh Request & Response

#### 1. GET `/students` (Dengan Query)
```http
GET /students?page=1&limit=5&search=budi&sort=grade&order=desc&is_active=true&grade_min=70&grade_max=95
```
**Response `200 OK`:**
```json
{
  "success": true,
  "message": "Students retrieved successfully",
  "meta": {
    "page": 1,
    "limit": 5,
    "total": 2,
    "total_pages": 1
  },
  "data": [
    {
      "id": "e6a0d4c8-3c94-4d89-bcf8-3486c91a3291",
      "name": "Budi Santoso",
      "grade": 88.5,
      "is_active": true
    },
    {
      "id": "7b095cf0-0e1f-4ffb-a8d6-724f2b1d3312",
      "name": "Budi Prasetyo",
      "grade": 75.0,
      "is_active": true
    }
  ]
}
```

#### 2. GET `/students/:id`
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
    "name": "Budi Santoso",
    "grade": 88.5,
    "is_active": true
  }
}
```

#### 3. POST `/students`
```http
POST /students
Content-Type: application/json

{
  "name": "Andi Wijaya",
  "grade": 85.5,
  "is_active": true
}
```
**Response `201 Created`:**
```json
{
  "success": true,
  "message": "Student created successfully",
  "data": {
    "id": "c1f74fae-9d22-47d0-8f96-df302919d363",
    "name": "Andi Wijaya",
    "grade": 85.5,
    "is_active": true
  }
}
```

#### 4. PUT `/students/:id`
```http
PUT /students/c1f74fae-9d22-47d0-8f96-df302919d363
Content-Type: application/json

{
  "name": "Andi Wijaya Updated",
  "grade": 90.0,
  "is_active": false
}
```
**Response `200 OK`:**
```json
{
  "success": true,
  "message": "Student updated successfully",
  "data": {
    "id": "c1f74fae-9d22-47d0-8f96-df302919d363",
    "name": "Andi Wijaya Updated",
    "grade": 90.0,
    "is_active": false
  }
}
```

#### 5. PATCH `/students/:id`
```http
PATCH /students/c1f74fae-9d22-47d0-8f96-df302919d363
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
    "id": "c1f74fae-9d22-47d0-8f96-df302919d363",
    "name": "Andi Wijaya Updated",
    "grade": 95.0,
    "is_active": false
  }
}
```

#### 6. DELETE `/students/:id`
```http
DELETE /students/c1f74fae-9d22-47d0-8f96-df302919d363
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
│   └── postgres.go          # Inisialisasi Connection Pool (pgxpool)
├── migrations/
│   ├── 001_create_users.sql    # Skema tabel users
│   └── 002_create_students.sql # Skema tabel students
├── handler.go               # HTTP Handler CRUD Student
├── helper.go                # Response envelope JSON & formatting
├── model.go                 # Request struct & alias model
├── main.go                  # Routing & entry point aplikasi
├── .env.example             # Template konfigurasi environment
├── go.mod
└── go.sum
```
