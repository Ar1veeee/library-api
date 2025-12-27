# Library Transaction API

RESTful API untuk sistem peminjaman perpustakaan dengan database transaction yang atomic dan clean architecture.

## 🎯 Fitur Utama

- **Clean Architecture**: Separation of concerns dengan DTO, Service, Repository, dan Handler layers
- **Database Transaction**: Semua operasi borrow menggunakan transaction dengan isolation level `READ COMMITTED`
- **Validasi Atomic**: Stok buku & kuota member divalidasi dalam 1 transaction untuk prevent race condition
- **Custom Error Response**: Format error konsisten dengan `ziyad_error_code` dan `trace_id` untuk debugging
- **Row-Level Locking**: Menggunakan `FOR UPDATE` untuk prevent concurrent issues
- **Consistent Response Format**: Semua endpoint return format yang konsisten dengan `SuccessResponse` wrapper

## 🛠️ Tech Stack

- **Language**: Go 1.21
- **Database**: MySQL 8.0
- **Router**: Gorilla Mux
- **Containerization**: Docker & Docker Compose
- **Architecture**: Clean Architecture dengan DTO Layer

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
docker compose up --build
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
  "message": "Buku berhasil dipinjam",
  "data": {
    "loan_id": 123,
    "member_id": 1,
    "book_id": 1,
    "book_title": "Clean Code",
    "book_author": "Robert C. Martin",
    "borrowed_at": "2024-12-27 14:30:45"
  }
}
```

**Error Responses**:

Buku tidak ditemukan (404):

```json
{
  "message": "Buku tidak ditemukan",
  "ziyad_error_code": "ZYD-ERR-005",
  "trace_id": "a1b2c3d4e5f6..."
}
```

Stok habis (409):

```json
{
  "message": "Stok buku habis",
  "ziyad_error_code": "ZYD-ERR-001",
  "trace_id": "a1b2c3d4e5f6..."
}
```

Kuota member penuh (409):

```json
{
  "message": "Member sudah mencapai batas pinjam maksimal yaitu 3 buku",
  "ziyad_error_code": "ZYD-ERR-002",
  "trace_id": "a1b2c3d4e5f6..."
}
```

Buku sudah dipinjam member (409):

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

**Error Responses**:

Tidak sedang meminjam (400):

```json
{
  "message": "Anda tidak sedang meminjam buku ini",
  "ziyad_error_code": "ZYD-ERR-005",
  "trace_id": "a1b2c3d4e5f6..."
}
```

Sudah dikembalikan (409):

```json
{
  "message": "Buku sudah dikembalikan",
  "ziyad_error_code": "ZYD-ERR-007",
  "trace_id": "a1b2c3d4e5f6..."
}
```

### 3. Get All Books

**Endpoint**: `GET /api/v1/books`

**Response** (200):

```json
{
  "message": "Berhasil mengambil daftar buku",
  "data": {
    "total": 8,
    "books": [
      {
        "id": 1,
        "title": "Clean Code",
        "author": "Robert C. Martin",
        "stock": 5
      },
      {
        "id": 2,
        "title": "The Pragmatic Programmer",
        "author": "Andrew Hunt",
        "stock": 3
      }
    ]
  }
}
```

### 4. Get Book by ID

**Endpoint**: `GET /api/v1/books/{id}`

**Response** (200):

```json
{
  "message": "Berhasil mengambil detail buku",
  "data": {
    "id": 1,
    "title": "Clean Code",
    "author": "Robert C. Martin",
    "stock": 5
  }
}
```

**Error Response** (404):

```json
{
  "message": "Buku tidak ditemukan",
  "ziyad_error_code": "ZYD-ERR-005",
  "trace_id": "a1b2c3d4e5f6..."
}
```

### 5. Get Member Loan History

**Endpoint**: `GET /api/v1/members/{id}/loans`

**Response** (200):

```json
{
  "message": "Berhasil mengambil riwayat peminjaman member",
  "data": {
    "member_id": 1,
    "member_name": "John Doe",
    "total_loans": 3,
    "loans": [
      {
        "loan_id": 1,
        "book_id": 1,
        "book_title": "Clean Code",
        "book_author": "Robert C. Martin",
        "borrowed_at": "2024-12-20 10:00:00",
        "returned_at": "2024-12-25 15:30:00",
        "status": "returned"
      },
      {
        "loan_id": 2,
        "book_id": 5,
        "book_title": "Head First Design Patterns",
        "book_author": "Eric Freeman",
        "borrowed_at": "2024-12-27 14:30:45",
        "returned_at": null,
        "status": "active"
      }
    ]
  }
}
```

**Error Response** (404):

```json
{
  "message": "Member tidak ditemukan",
  "ziyad_error_code": "ZYD-ERR-005",
  "trace_id": "a1b2c3d4e5f6..."
}
```

## 🧪 Testing Scenarios

### Test 1: Happy Path - Borrow Book

```bash
curl -X POST http://localhost:8080/api/v1/borrow \
  -H "Content-Type: application/json" \
  -d '{"member_id": 1, "book_id": 4}'
```

Expected Response (201):

```json
{
  "message": "Buku berhasil dipinjam",
  "data": {
    "loan_id": 4,
    "member_id": 1,
    "book_id": 4,
    "book_title": "Refactoring",
    "book_author": "Martin Fowler",
    "borrowed_at": "2024-12-27 14:30:45"
  }
}
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

Expected Error (409):

```json
{
  "message": "Member sudah mencapai batas pinjam maksimal yaitu 3 buku",
  "ziyad_error_code": "ZYD-ERR-002",
  "trace_id": "..."
}
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

Expected Error (409):

```json
{
  "message": "Stok buku habis",
  "ziyad_error_code": "ZYD-ERR-001",
  "trace_id": "..."
}
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

Expected Error (409):

```json
{
  "message": "Anda sedang meminjam buku ini",
  "ziyad_error_code": "ZYD-ERR-003",
  "trace_id": "..."
}
```

### Test 5: Return Book

```bash
curl -X POST http://localhost:8080/api/v1/return \
  -H "Content-Type: application/json" \
  -d '{"member_id": 1, "book_id": 6}'
```

Expected Response (200):

```json
{
  "message": "Buku berhasil dikembalikan"
}
```

### Test 6: Get All Books

```bash
curl http://localhost:8080/api/v1/books
```

### Test 7: Get Book Detail

```bash
curl http://localhost:8080/api/v1/books/1
```

### Test 8: Get Member Loan History

```bash
curl http://localhost:8080/api/v1/members/1/loans
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

## 🏗️ Clean Architecture

### Project Structure

```
library-api/
├── cmd/
│   └── api/
│       └── main.go              # Entry point - Dependency injection
├── internal/
│   ├── config/
│   │   └── config.go            # Database & environment config
│   ├── dto/                     # Data Transfer Objects
│   │   ├── loan_dto.go          # Request/Response untuk Loan
│   │   ├── book_dto.go          # Response untuk Book
│   │   ├── member_dto.go        # Response untuk Member
│   │   └── common_dto.go        # Success & Error response format
│   ├── model/
│   │   └── models.go            # Domain entities & error types
│   ├── repository/              # Data Access Layer
│   │   ├── book_repository.go   # Database operations - Books
│   │   ├── member_repository.go # Database operations - Members
│   │   └── loan_repository.go   # Database operations - Loans
│   ├── service/                 # Business Logic Layer
│   │   ├── loan_service.go      # CORE TRANSACTION LOGIC
│   │   ├── book_service.go      # Book business logic
│   │   └── member_service.go    # Member business logic
│   └── handler/                 # HTTP Handler Layer
│       ├── loan_handler.go      # HTTP endpoints - Loans
│       ├── book_handler.go      # HTTP endpoints - Books
│       └── member_handler.go    # HTTP endpoints - Members
├── migrations/
│   └── init.sql                 # Database schema & seed data
├── docker-compose.yml
├── Dockerfile
├── go.mod
├── go.sum
└── README.md
```

### Architecture Layers

```
┌─────────────────────────────────────────────┐
│              HTTP Request                    │
└─────────────────┬───────────────────────────┘
                  │
         ┌────────▼────────┐
         │   Handler Layer  │ ← Parse request, validate input
         │   (HTTP Logic)   │   Return HTTP response
         └────────┬────────┘
                  │
         ┌────────▼────────┐
         │    DTO Layer     │ ← Request/Response structures
         │  (Data Transfer) │   API contract definition
         └────────┬────────┘
                  │
         ┌────────▼────────┐
         │  Service Layer   │ ← Business logic & transactions
         │ (Business Logic) │   Coordinate repositories
         └────────┬────────┘
                  │
         ┌────────▼────────┐
         │ Repository Layer │ ← Database operations
         │  (Data Access)   │   SQL queries & CRUD
         └────────┬────────┘
                  │
         ┌────────▼────────┐
         │    Model Layer   │ ← Domain entities
         │ (Domain Objects) │   Pure data structures
         └─────────────────┘
```

### Response Format Consistency

Semua success response menggunakan format yang konsisten:

```json
{
  "message": "Success message here",
  "data": {
    ...
  }
  // Optional, tergantung endpoint
}
```

Semua error response menggunakan format:

```json
{
  "message": "Error message here",
  "ziyad_error_code": "ZYD-ERR-XXX",
  "trace_id": "unique-tracking-id"
}
```

## 🗄️ Database Schema

### Table: books

```sql
CREATE TABLE books
(
    id         INT PRIMARY KEY AUTO_INCREMENT,
    title      VARCHAR(255) NOT NULL,
    author     VARCHAR(255) NOT NULL,
    stock      INT          NOT NULL DEFAULT 0,
    created_at TIMESTAMP             DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP             DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    -- Index untuk query WHERE stock > 0
    INDEX      idx_stock (stock)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

### Table: members

```sql
CREATE TABLE members
(
    id         INT PRIMARY KEY AUTO_INCREMENT,
    name       VARCHAR(255)        NOT NULL,
    email      VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

### Table: loans

```sql
CREATE TABLE loans
(
    id          INT PRIMARY KEY AUTO_INCREMENT,
    member_id   INT NOT NULL,
    book_id     INT NOT NULL,
    borrowed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    returned_at TIMESTAMP NULL,

    FOREIGN KEY (member_id) REFERENCES members (id) ON DELETE CASCADE,
    FOREIGN KEY (book_id) REFERENCES books (id) ON DELETE CASCADE,

    -- Index untuk query active loans per member
    INDEX       idx_member_active (member_id, returned_at),

    -- Index untuk check duplicate borrow
    INDEX       idx_member_book_active (member_id, book_id, returned_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

### Seed Data

Database otomatis terisi dengan sample data:

- **8 Buku** dengan stok bervariasi (1-6)
- **5 Member** dengan data lengkap
- **3 Sample loans** untuk testing history

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

# Follow logs (realtime)
docker logs -f library_api
```

### Restart Containers

```bash
# Stop dan hapus containers
docker compose down

# Rebuild dan start
docker compose up --build

# Start di background
docker compose up -d --build
```

### Reset Database

```bash
# Stop containers dan hapus volumes
docker compose down -v

# Start ulang (akan recreate database)
docker compose up --build
```

## 📊 Error Codes Reference

| Code        | Message                     | HTTP Status | Description                             |
|-------------|-----------------------------|-------------|-----------------------------------------|
| ZYD-ERR-001 | Stok buku habis             | 409         | Book stock is empty                     |
| ZYD-ERR-002 | Kuota member habis          | 409         | Member reached max loan limit (3 books) |
| ZYD-ERR-003 | Buku sedang dipinjam member | 409         | Member already borrowed this book       |
| ZYD-ERR-004 | Database transaction failed | 400/500     | Internal transaction error              |
| ZYD-ERR-005 | Resource not found          | 404         | Book/Member not found                   |
| ZYD-ERR-006 | Invalid input data          | 400         | Request validation failed               |
| ZYD-ERR-007 | Buku sudah dikembalikan     | 409         | Book is already returned                |

