# Library Transaction API

RESTful API untuk sistem peminjaman perpustakaan dengan database transaction yang atomic.

## 🎯 Fitur Utama

- **Database Transaction**: Semua operasi borrow menggunakan transaction dengan isolation level `READ COMMITTED`
- **Validasi Atomic**: Stok buku & kuota member divalidasi dalam 1 transaction untuk prevent race condition
- **Custom Error Response**: Format error konsisten dengan `ziyad_error_code` dan `trace_id` untuk debugging
- **Row-Level Locking**: Menggunakan `FOR UPDATE` untuk prevent concurrent issues

## 🛠️ Tech Stack

- **Language**: Go 1.21
- **Database**: MySQL 8.0
- **Router**: Gorilla Mux
- **Containerization**: Docker & Docker Compose

## 📋 Prerequisites

- Docker & Docker Compose
- Port 8080 dan 3306 harus available

## 🚀 Quick Start

### 1. Clone Repository

```bash
git clone <repository-url>
cd library-api
```

### 2. Jalankan dengan Docker Compose

```bash
docker-compose up --build
```

Tunggu hingga muncul log:
```
✅ Database connected successfully
🚀 Server starting on :8080
```

### 3. Test Health Check

```bash
curl http://localhost:8080/health
```

Expected response:
```json
{
  "status": "ok",
  "service": "library-api"
}
```

## 📚 API Endpoints

### 1. Borrow Book (Transaction Logic)

**Endpoint**: `POST /api/v1/borrow`

**Request Body**:
```json
{
  "member_id": 1,
  "book_id": 1
}
```

**Success Response** (201):
```json
{
  "message": "Buku berhasil dipinjam"
}
```

**Error Responses**:

Stok habis (400):
```json
{
  "message": "Stok buku habis",
  "ziyad_error_code": "ZYD-ERR-001",
  "trace_id": "a1b2c3d4e5f6..."
}
```

Kuota member penuh (400):
```json
{
  "message": "Member sudah mencapai batas pinjam maksimal (3 buku)",
  "ziyad_error_code": "ZYD-ERR-002",
  "trace_id": "a1b2c3d4e5f6..."
}
```

Buku sudah dipinjam member (400):
```json
{
  "message": "Anda sedang meminjam buku ini",
  "ziyad_error_code": "ZYD-ERR-003",
  "trace_id": "a1b2c3d4e5f6..."
}
```

### 2. Return Book

**Endpoint**: `POST /api/v1/return`

**Request Body**:
```json
{
  "member_id": 1,
  "book_id": 1
}
```

**Success Response** (200):
```json
{
  "message": "Buku berhasil dikembalikan"
}
```

### 3. Get All Books

**Endpoint**: `GET /api/v1/books`

**Response** (200):
```json
[
  {
    "id": 1,
    "title": "Clean Code",
    "author": "Robert C. Martin",
    "stock": 5
  }
]
```

### 4. Get Book by ID

**Endpoint**: `GET /api/v1/books/{id}`

**Response** (200):
```json
{
  "id": 1,
  "title": "Clean Code",
  "author": "Robert C. Martin",
  "stock": 5
}
```

### 5. Get Member Loan History

**Endpoint**: `GET /api/v1/members/{id}/loans`

**Response** (200):
```json
[
  {
    "id": 1,
    "member_id": 1,
    "book_id": 1,
    "borrowed_at": "2024-12-15T10:00:00Z",
    "returned_at": "2024-12-22T15:00:00Z",
    "book_title": "Clean Code",
    "book_author": "Robert C. Martin"
  }
]
```

## 🧪 Testing Scenarios

### Test 1: Happy Path - Borrow Book

```bash
curl -X POST http://localhost:8080/api/v1/borrow \
  -H "Content-Type: application/json" \
  -d '{"member_id": 1, "book_id": 4}'
```

### Test 2: Quota Exceeded (setelah pinjam 3 buku)

```bash
# Pinjam buku ke-1
curl -X POST http://localhost:8080/api/v1/borrow \
  -H "Content-Type: application/json" \
  -d '{"member_id": 4, "book_id": 1}'

# Pinjam buku ke-2
curl -X POST http://localhost:8080/api/v1/borrow \
  -H "Content-Type: application/json" \
  -d '{"member_id": 4, "book_id": 4}'

# Pinjam buku ke-3
curl -X POST http://localhost:8080/api/v1/borrow \
  -H "Content-Type: application/json" \
  -d '{"member_id": 4, "book_id": 6}'

# Ini akan ditolak dengan ZYD-ERR-002
curl -X POST http://localhost:8080/api/v1/borrow \
  -H "Content-Type: application/json" \
  -d '{"member_id": 4, "book_id": 7}'
```

### Test 3: Stock Empty

```bash
# Buku ID 5 hanya stock 1
curl -X POST http://localhost:8080/api/v1/borrow \
  -H "Content-Type: application/json" \
  -d '{"member_id": 1, "book_id": 5}'

# Request kedua akan ditolak dengan ZYD-ERR-001
curl -X POST http://localhost:8080/api/v1/borrow \
  -H "Content-Type: application/json" \
  -d '{"member_id": 5, "book_id": 5}'
```

### Test 4: Already Borrowed

```bash
# Pinjam buku
curl -X POST http://localhost:8080/api/v1/borrow \
  -H "Content-Type: application/json" \
  -d '{"member_id": 1, "book_id": 6}'

# Coba pinjam buku yang sama lagi -> ZYD-ERR-003
curl -X POST http://localhost:8080/api/v1/borrow \
  -H "Content-Type: application/json" \
  -d '{"member_id": 1, "book_id": 6}'
```

### Test 5: Return Book

```bash
curl -X POST http://localhost:8080/api/v1/return \
  -H "Content-Type: application/json" \
  -d '{"member_id": 1, "book_id": 6}'
```

## 🔍 Transaction Logic Explanation

### Mengapa Database Transaction Penting?

Tanpa transaction, race condition bisa terjadi:

```
Time │ User A                    │ User B
─────┼───────────────────────────┼──────────────────────────
T1   │ Check stock = 1           │
T2   │                           │ Check stock = 1
T3   │ Update stock = 0          │
T4   │                           │ Update stock = -1 ❌
```

Dengan transaction + `FOR UPDATE`:

```
Time │ User A                    │ User B
─────┼───────────────────────────┼──────────────────────────
T1   │ BEGIN + SELECT...FOR UPDATE (LOCK)
T2   │                           │ BEGIN + SELECT...FOR UPDATE (WAIT 🔒)
T3   │ stock = 1 ✅              │
T4   │ UPDATE stock = 0          │
T5   │ INSERT loan               │
T6   │ COMMIT (release lock)     │
T7   │                           │ stock = 0 ❌ -> ROLLBACK
```

### Flow dalam `BorrowBook` Service

1. **BEGIN TRANSACTION** dengan isolation `READ COMMITTED`
2. **Validasi Member**: Pastikan member exist
3. **Check Kuota** (dengan `FOR UPDATE`): Max 3 buku per member
4. **Check & Lock Stok** (dengan `FOR UPDATE`): Pastikan stok > 0
5. **Check Double Borrow**: Pastikan member belum pinjam buku ini
6. **Decrement Stock**: Kurangi stok buku
7. **Insert Loan**: Catat peminjaman
8. **COMMIT**: Simpan semua perubahan

Jika ada 1 step yang gagal, semua perubahan di-rollback.

## 📁 Project Structure

```
library-api/
├── cmd/
│   └── api/
│       └── main.go              # Entry point
├── internal/
│   ├── config/
│   │   └── config.go            # Database & environment config
│   ├── handler/
│   │   ├── loan_handler.go      # HTTP handlers untuk loan
│   │   ├── book_handler.go      # HTTP handlers untuk book
│   │   └── member_handler.go    # HTTP handlers untuk member
│   ├── service/
│   │   ├── loan_service.go      # ⭐ CORE TRANSACTION LOGIC
│   │   ├── book_service.go
│   │   └── member_service.go
│   ├── repository/
│   │   ├── book_repository.go   # Data access untuk books
│   │   ├── member_repository.go # Data access untuk members
│   │   └── loan_repository.go   # Data access untuk loans
│   └── model/
│       └── models.go             # Data models & error types
├── migrations/
│   └── init.sql                  # Database schema & seed data
├── docker-compose.yml
├── Dockerfile
├── go.mod
├── go.sum
└── README.md
```

## 🗄️ Database Schema

### Table: books
```sql
CREATE TABLE books (
    id INT PRIMARY KEY AUTO_INCREMENT,
    title VARCHAR(255) NOT NULL,
    author VARCHAR(255) NOT NULL,
    stock INT NOT NULL DEFAULT 0,
    INDEX idx_stock (stock)
);
```

### Table: members
```sql
CREATE TABLE members (
    id INT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL
);
```

### Table: loans
```sql
CREATE TABLE loans (
    id INT PRIMARY KEY AUTO_INCREMENT,
    member_id INT NOT NULL,
    book_id INT NOT NULL,
    borrowed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    returned_at TIMESTAMP NULL,
    
    FOREIGN KEY (member_id) REFERENCES members(id),
    FOREIGN KEY (book_id) REFERENCES books(id),
    
    INDEX idx_member_active (member_id, returned_at),
    INDEX idx_member_book_active (member_id, book_id, returned_at)
);
```

## 🐛 Troubleshooting

### Port Already in Use

```bash
# Check port 8080
sudo lsof -i :8080
# Kill process if needed
sudo kill -9 <PID>

# Check port 3306
sudo lsof -i :3306
sudo kill -9 <PID>
```

### Docker Permission Denied

```bash
sudo usermod -aG docker $USER
newgrp docker
```

### Container Logs

```bash
# API logs
docker logs library_api

# Database logs
docker logs library_db
```

### Restart Containers

```bash
docker-compose down
docker-compose up --build
```

## 📊 Error Codes Reference

| Code | Message | Description |
|------|---------|-------------|
| ZYD-ERR-001 | Stok buku habis | Book stock is empty |
| ZYD-ERR-002 | Kuota member habis | Member reached max loan limit (3 books) |
| ZYD-ERR-003 | Buku sedang dipinjam member | Member already borrowed this book |
| ZYD-ERR-004 | Database transaction failed | Internal transaction error |
| ZYD-ERR-005 | Resource not found | Book/Member not found |
| ZYD-ERR-006 | Invalid input data | Request validation failed |
| ZYD-ERR-007 | Buku sudah dikembalikan | Book already returned |

## 👨‍💻 Development

### Run Locally (Without Docker)

1. Install MySQL dan buat database:
```bash
mysql -u root -p
CREATE DATABASE library_db;
```

2. Run migrations:
```bash
mysql -u root -p library_db < migrations/init.sql
```

3. Set environment variables:
```bash
export DB_HOST=localhost
export DB_PORT=3306
export DB_USER=root
export DB_PASSWORD=your_password
export DB_NAME=library_db
export SERVER_PORT=8080
```

4. Run application:
```bash
go run cmd/api/main.go
```

## 📝 Notes

- **Kuota Member**: Maximum 3 buku per member
- **Transaction Isolation**: `READ COMMITTED` untuk balance antara consistency & performance
- **Row Locking**: `FOR UPDATE` digunakan untuk prevent race conditions
- **Trace ID**: Setiap error response memiliki unique trace_id untuk debugging

## 🎓 Key Learning Points

1. **Database Transaction adalah MUST** untuk operasi yang melibatkan multiple tables
2. **Row-Level Locking** mencegah race condition dalam concurrent scenario
3. **Custom Error Format** memudahkan client handling dan debugging
4. **Index Strategy** penting untuk performance query dengan WHERE dan JOIN
5. **Defer Rollback Pattern** mencegah transaction leak

---

**Dibuat untuk Test Teknis Backend Engineer**