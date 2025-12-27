package errors

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

// APIError represents custom error response format
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
