# API Students

REST API sederhana untuk manajemen data mahasiswa, dibangun dengan [Go Fiber](https://gofiber.io/).

## Menjalankan Server

```bash
go run .
# Server berjalan di http://localhost:3000


---

## Amplop Respons

Semua endpoint mengembalikan JSON dengan struktur seragam.

**Sukses (non-list):**
```json
{
  "success": true,
  "message": "...",
  "data": { }
}
```

**Sukses (list — GET /students):**
```json
{
  "success": true,
  "message": "...",
  "meta": {
    "page": 1,
    "limit": 10,
    "total": 42,
    "total_pages": 5
  },
  "data": [ ]
}
```

**Gagal:**
```json
{
  "success": false,
  "message": "Pesan error",
  "data": null
}
```

---

## Kontrak API

| Metode | Endpoint | Query String / Parameter | Contoh Body Permintaan | Status yang Mungkin | Contoh Respons |
|--------|----------|--------------------------|------------------------|---------------------|----------------|
| `GET` | `/students` | Lihat tabel query di bawah | — | `200` | [↓ Lihat](#get-students) |
| `GET` | `/students/:id` | `:id` — UUID mahasiswa | — | `200`, `404` | [↓ Lihat](#get-studentsid) |
| `POST` | `/students` | — | [↓ Lihat](#post-students) | `201`, `400`, `422` | [↓ Lihat](#post-students) |
| `PUT` | `/students/:id` | `:id` — UUID mahasiswa | [↓ Lihat](#put-studentsid) | `200`, `400`, `404`, `422` | [↓ Lihat](#put-studentsid) |
| `PATCH` | `/students/:id` | `:id` — UUID mahasiswa | [↓ Lihat](#patch-studentsid) | `200`, `400`, `404` | [↓ Lihat](#patch-studentsid) |
| `DELETE` | `/students/:id` | `:id` — UUID mahasiswa | — | `200`, `404` | [↓ Lihat](#delete-studentsid) |

---

## Query String — GET /students

| Parameter | Tipe | Default | Validasi | Keterangan |
|-----------|------|---------|----------|------------|
| `page` | `int` | `1` | ≥ 1 | Nomor halaman |
| `limit` | `int` | `10` | 1–100 | Jumlah item per halaman. Dibatasi 100 untuk menjaga ukuran payload tetap wajar. |
| `search` | `string` | `""` | — | Pencarian nama mahasiswa, tidak membedakan huruf besar/kecil |
| `sort` | `string` | `name` | `name`, `grade`, `is_active` | Field pengurutan. Nilai di luar whitelist → `400`. |
| `order` | `string` | `asc` | `asc`, `desc` | Arah pengurutan |
| `is_active` | `bool` | *(tidak difilter)* | `true`, `false` | Filter status aktif mahasiswa |
| `grade_min` | `float64` | *(tidak difilter)* | Angka valid | Filter nilai ≥ grade_min |
| `grade_max` | `float64` | *(tidak difilter)* | Angka valid | Filter nilai ≤ grade_max |

---

## Contoh per Endpoint

### GET /students

**Request:**
```
GET /students?page=1&limit=5&search=budi&sort=grade&order=desc&is_active=true&grade_min=70&grade_max=95
```

**Response `200`:**
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
    { "id": "abc-123", "name": "Budi Santoso", "grade": 88.5, "is_active": true },
    { "id": "def-456", "name": "Budi Prasetyo", "grade": 75.0, "is_active": true }
  ]
}
```

**Response `400` (sort tidak valid):**
```json
{
  "success": false,
  "message": "Parameter 'sort' hanya boleh berisi: name, grade, is_active",
  "data": null
}
```

---

### GET /students/:id

**Request:**
```
GET /students/abc-123
```

**Response `200`:**
```json
{
  "success": true,
  "message": "Student retrieved successfully",
  "data": {
    "id": "abc-123",
    "name": "Budi Santoso",
    "grade": 88.5,
    "is_active": true
  }
}
```

**Response `404`:**
```json
{
  "success": false,
  "message": "Student not found",
  "data": null
}
```

---

### POST /students

**Body (semua field wajib):**
```json
{
  "name": "Andi Wijaya",
  "grade": 82.0,
  "is_active": true
}
```

**Response `201`:**
```json
{
  "success": true,
  "message": "Student created successfully",
  "data": {
    "id": "ghi-789",
    "name": "Andi Wijaya",
    "grade": 82.0,
    "is_active": true
  }
}
```

**Response `422` (name kosong):**
```json
{
  "success": false,
  "message": "Field 'name' is required",
  "data": null
}
```

---

### PUT /students/:id

**Body (semua field wajib — replace penuh):**
```json
{
  "name": "Andi Wijaya",
  "grade": 90.0,
  "is_active": false
}
```

**Response `200`:**
```json
{
  "success": true,
  "message": "Student updated successfully",
  "data": {
    "id": "ghi-789",
    "name": "Andi Wijaya",
    "grade": 90.0,
    "is_active": false
  }
}
```

**Response `404`:**
```json
{
  "success": false,
  "message": "Student not found",
  "data": null
}
```

---

### PATCH /students/:id

**Body (semua field opsional — hanya field yang dikirim yang diubah):**
```json
{
  "is_active": false
}
```

**Response `200`:**
```json
{
  "success": true,
  "message": "Student patched successfully",
  "data": {
    "id": "ghi-789",
    "name": "Andi Wijaya",
    "grade": 90.0,
    "is_active": false
  }
}
```

**Response `404`:**
```json
{
  "success": false,
  "message": "Student not found",
  "data": null
}
```

---

### DELETE /students/:id

**Request:**
```
DELETE /students/ghi-789
```

**Response `200`:**
```json
{
  "success": true,
  "message": "Student deleted successfully",
  "data": null
}
```

**Response `404`:**
```json
{
  "success": false,
  "message": "Student not found",
  "data": null
}
```

---

## Struktur Proyek

```
api-students/
├── main.go       # Entry point & routing
├── model.go      # Student, request structs, StudentQuery
├── helper.go     # Response envelope & helper functions
├── handler.go    # Semua handler CRUD
├── go.mod
└── go.sum
```
