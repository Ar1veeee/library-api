package handler

import (
	"encoding/json"
	"net/http"

	"github.com/Ar1veeee/library-api/internal/dto"
	"github.com/Ar1veeee/library-api/internal/errors"
	"github.com/Ar1veeee/library-api/internal/http/mapper"
	"github.com/Ar1veeee/library-api/internal/service"
)

type LoanHandler struct {
	loanService *service.LoanService
}

func NewLoanHandler(loanService *service.LoanService) *LoanHandler {
	return &LoanHandler{loanService: loanService}
}

func (h *LoanHandler) BorrowBook(w http.ResponseWriter, r *http.Request) {
	var req dto.BorrowBookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mapper.HandleHTTPError(w, err)
		return
	}

	if req.MemberID <= 0 || req.BookID <= 0 {
		mapper.HandleHTTPError(
			w,
			errors.NewAPIError(
				"member_id dan book_id harus lebih dari 0",
				errors.ErrCodeInvalidInput,
			),
		)
		return
	}

	loanDetail, err := h.loanService.BorrowBook(r.Context(), req.MemberID, req.BookID)
	if err != nil {
		mapper.HandleHTTPError(w, err)
		return
	}

	response := dto.SuccessResponse{
		Message: "Buku berhasil dipinjam",
		Data:    loanDetail,
	}

	mapper.RespondSuccess(w, response, http.StatusCreated)
}

func (h *LoanHandler) ReturnBook(w http.ResponseWriter, r *http.Request) {
	var req dto.ReturnBookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		mapper.HandleHTTPError(w, err)
		return
	}

	if req.MemberID <= 0 || req.BookID <= 0 {
		mapper.HandleHTTPError(
			w,
			errors.NewAPIError(
				"member_id dan book_id harus lebih dari 0",
				errors.ErrCodeInvalidInput,
			),
		)
		return
	}

	if err := h.loanService.ReturnBook(r.Context(), req.MemberID, req.BookID); err != nil {
		mapper.HandleHTTPError(w, err)
		return
	}

	response := dto.SuccessResponse{
		Message: "Buku berhasil dikembalikan",
		Data:    nil,
	}

	mapper.RespondSuccess(w, response, http.StatusOK)
}
