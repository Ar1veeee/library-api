package handler

import (
	"net/http"
	"strconv"

	"github.com/Ar1veeee/library-api/internal/dto"
	"github.com/Ar1veeee/library-api/internal/http/mapper"
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
		mapper.HandleHTTPError(w, err)
		return
	}

	bookResponses := make([]dto.BookResponse, len(books))
	for i, book := range books {
		bookResponses[i] = dto.BookResponse{
			ID:     book.ID,
			Title:  book.Title,
			Author: book.Author,
			Stock:  book.Stock,
		}
	}

	booksData := dto.BooksListResponse{
		Total: len(bookResponses),
		Books: bookResponses,
	}

	response := dto.SuccessResponse{
		Message: "Berhasil mengambil data buku",
		Data:    booksData,
	}

	mapper.RespondSuccess(w, response, http.StatusOK)
}

func (h *BookHandler) GetBookByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bookID, err := strconv.Atoi(vars["id"])
	if err != nil {
		mapper.HandleHTTPError(w, err)
		return
	}

	book, err := h.bookService.GetBookByID(r.Context(), bookID)
	if err != nil {
		mapper.HandleHTTPError(w, err)
		return
	}

	bookData := dto.BookResponse{
		ID:     book.ID,
		Title:  book.Title,
		Author: book.Author,
		Stock:  book.Stock,
	}

	response := dto.SuccessResponse{
		Message: "Berhasil mengambil data detail buku",
		Data:    bookData,
	}

	mapper.RespondSuccess(w, response, http.StatusOK)
}
