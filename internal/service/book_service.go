package service

import (
	"context"

	"github.com/Ar1veeee/library-api/internal/model"
	"github.com/Ar1veeee/library-api/internal/repository"
)

type BookService struct {
	bookRepo *repository.BookRepository
}

func NewBookService(bookRepo *repository.BookRepository) *BookService {
	return &BookService{bookRepo: bookRepo}
}

func (s *BookService) GetAllBooks(ctx context.Context) ([]model.Book, error) {
	return s.bookRepo.GetAll(ctx)
}

func (s *BookService) GetBookByID(ctx context.Context, bookID int) (*model.Book, error) {
	book, err := s.bookRepo.GetByID(ctx, bookID)
	if err != nil {
		return nil, err
	}
	if book == nil {
		return nil, model.NewAPIError("Buku tidak ditemukan", model.ErrCodeNotFound)
	}
	return book, nil
}
