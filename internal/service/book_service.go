package service

import (
	"context"

	"github.com/Ar1veeee/library-api/internal/dto"
	"github.com/Ar1veeee/library-api/internal/errors"
	"github.com/Ar1veeee/library-api/internal/model"
	"github.com/Ar1veeee/library-api/internal/repository"
)

type BookService struct {
	bookRepo *repository.BookRepository
}

func NewBookService(bookRepo *repository.BookRepository) *BookService {
	return &BookService{bookRepo: bookRepo}
}

func (s *BookService) GetAllBooks(ctx context.Context) (*dto.BooksListResponse, error) {
	books, err := s.bookRepo.GetAll(ctx)
	if err != nil {
		return nil, err
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

	return &dto.BooksListResponse{
		Total: len(bookResponses),
		Books: bookResponses,
	}, nil
}

func (s *BookService) GetBookByID(ctx context.Context, bookID int) (*model.Book, error) {
	book, err := s.bookRepo.GetByID(ctx, bookID)
	if err != nil {
		return nil, err
	}
	if book == nil {
		return nil, errors.NewAPIError("Buku tidak ditemukan", errors.ErrCodeNotFound)
	}
	return book, nil
}
