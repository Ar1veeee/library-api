package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Ar1veeee/library-api/internal/model"
	"github.com/Ar1veeee/library-api/internal/service"
)

type LoanHandler struct {
	loanService *service.LoanService
}

func NewLoanHandler(loanService *service.LoanService) *LoanHandler {
	return &LoanHandler{loanService: loanService}
}

type BorrowRequest struct {
	MemberID int `json:"member_id"`
	BookID   int `json:"book_id"`
}

type ReturnRequest struct {
	MemberID int `json:"member_id"`
	BookID   int `json:"book_id"`
}

func (h *LoanHandler) BorrowBook(w http.ResponseWriter, r *http.Request) {
	var req BorrowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, model.NewAPIError("Invalid request body", model.ErrCodeInvalidInput), http.StatusBadRequest)
		return
	}

	if req.MemberID <= 0 || req.BookID <= 0 {
		respondError(w, model.NewAPIError("member_id dan book_id harus lebih dari 0", model.ErrCodeInvalidInput), http.StatusBadRequest)
		return
	}

	if err := h.loanService.BorrowBook(r.Context(), req.MemberID, req.BookID); err != nil {
		// MENGAPA menggunakan errors.As?
		// - Kita perlu tahu apakah error dari business logic (APIError)
		// - APIError = user-facing error dengan custom code
		var apiErr model.APIError
		if errors.As(err, &apiErr) {
			respondError(w, apiErr, http.StatusBadRequest)
		}
		return
	}

	respondSuccess(w, map[string]string{
		"message": "Buku berhasil dipinjam",
	}, http.StatusCreated)
}

func (h *LoanHandler) ReturnBook(w http.ResponseWriter, r *http.Request) {
	var req ReturnRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, model.NewAPIError("Invalid request body", model.ErrCodeInvalidInput), http.StatusBadRequest)
		return
	}

	if req.MemberID <= 0 || req.BookID <= 0 {
		respondError(w, model.NewAPIError("member_id dan book_id harus lebih dari 0", model.ErrCodeInvalidInput), http.StatusBadRequest)
		return
	}

	if err := h.loanService.ReturnBook(r.Context(), req.MemberID, req.BookID); err != nil {
		// MENGAPA menggunakan errors.As?
		// - Kita perlu tahu apakah error dari business logic (APIError)
		// - APIError = user-facing error dengan custom code
		var apiErr model.APIError
		if errors.As(err, &apiErr) {
			respondError(w, apiErr, http.StatusBadRequest)
		}
		return
	}

	respondSuccess(w, map[string]string{
		"message": "Buku berhasil dikembalikan",
	}, http.StatusOK)
}

// Helper functions untuk consistent response format
func respondError(w http.ResponseWriter, err model.APIError, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(err)
}

func respondSuccess(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}
