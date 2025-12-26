package model

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

type Book struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Author string `json:"author"`
	Stock  int    `json:"stock"`
}

type Member struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type Loan struct {
	ID         int        `json:"id"`
	MemberID   int        `json:"member_id"`
	BookID     int        `json:"book_id"`
	BorrowedAt time.Time  `json:"borrowed_at"`
	ReturnedAt *time.Time `json:"returned_at,omitempty"`

	// Additional fields untuk response
	BookTitle  string `json:"book_title,omitempty"`
	BookAuthor string `json:"book_author,omitempty"`
}

// APIError represents custom error response format
// MENGAPA struct terpisah untuk error?
// - Konsistensi format error di seluruh API
// - Memudahkan client untuk parsing error
type APIError struct {
	Message      string `json:"message"`
	ZiyadErrCode string `json:"ziyad_err_code"`
	TraceID      string `json:"trace_id"`
}

func (e APIError) Error() string {
	return e.Message
}

func NewAPIError(message, code string) APIError {
	return APIError{
		Message:      message,
		ZiyadErrCode: code,
		TraceID:      GenerateTraceID(),
	}
}

// GenerateTraceID generates random string untuk tracking request
// MENGAPA menggunakan crypto/rand?
// - Lebih secure dan random dibanding math/rand
// - Menghindari collision antar trace_id
// - Format: 16 bytes hex = 32 characters
func GenerateTraceID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return time.Now().Format("20060102150405") + "-fallback"
	}
	return hex.EncodeToString(b)
}

const (
	ErrCodeStockEmpty      = "ZYD-ERR-001" // Stok buku habis
	ErrCodeQuotaExceeded   = "ZYD-ERR-002" // Kuota member habis
	ErrCodeAlreadyBorrowed = "ZYD-ERR-003" // Buku sedang dipinjam member
	ErrCodeTxFailed        = "ZYD-ERR-004" // Database transaction failed
	ErrCodeNotFound        = "ZYD-ERR-005" // Resource not found
	ErrCodeInvalidInput    = "ZYD-ERR-006" // Invalid input data
	ErrCodeAlreadyReturned = "ZYD-ERR-007" // Buku sudah dikembalikan
)
