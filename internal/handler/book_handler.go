package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/Ar1veeee/library-api/internal/model"
	"github.com/Ar1veeee/library-api/internal/service"
	"github.com/gorilla/mux"
)

type BookHandler struct {
	bookService *service.BookService
}

func NewBookHandler(bookService *service.BookService) *BookHandler {
	return &BookHandler{bookService: bookService}
}

func (h *BookHandler) GetBooks(w http.ResponseWriter, r *http.Request) {
	books, err := h.bookService.GetAllBooks(r.Context())
	if err != nil {
		respondError(w, model.NewAPIError("Gagal mengambil data buku", model.ErrCodeTxFailed), http.StatusInternalServerError)
		return
	}

	respondSuccess(w, books, http.StatusOK)
}

func (h *BookHandler) GetBookByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bookID, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondError(w, model.NewAPIError("Invalid book ID", model.ErrCodeInvalidInput), http.StatusBadRequest)
		return
	}

	book, err := h.bookService.GetBookByID(r.Context(), bookID)
	if err != nil {
		// MENGAPA menggunakan errors.As?
		// - Kita perlu tahu apakah error dari business logic (APIError)
		// - APIError = user-facing error dengan custom code
		var apiErr model.APIError
		if errors.As(err, &apiErr) {
			respondError(w, apiErr, http.StatusNotFound)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(book)
}
