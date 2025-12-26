package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Ar1veeee/library-api/internal/model"
)

type BookRepository struct {
	db *sql.DB
}

func NewBookRepository(db *sql.DB) *BookRepository {
	return &BookRepository{db: db}
}

// getByID mengambil data buku berdasarkan ID.
// Mendukung row-level locking opsional via forUpdate.
// Digunakan secara internal oleh GetByID (read-only) dan GetByIDForUpdate (with lock).
func (r *BookRepository) getByID(ctx context.Context, tx *sql.Tx, bookID int, forUpdate bool) (*model.Book, error) {
	query := `SELECT id, title, author, stock FROM books WHERE id = ?`
	if forUpdate {
		query += ` FOR UPDATE`
	}

	var book model.Book
	var err error

	if tx != nil {
		err = tx.QueryRowContext(ctx, query, bookID).Scan(
			&book.ID, &book.Title, &book.Author, &book.Stock,
		)
	} else {
		err = r.db.QueryRowContext(ctx, query, bookID).Scan(
			&book.ID, &book.Title, &book.Author, &book.Stock,
		)
	}

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	return &book, err
}

// GetByID mengambil detail buku (tanpa locking).
func (r *BookRepository) GetByID(ctx context.Context, bookID int) (*model.Book, error) {
	return r.getByID(ctx, nil, bookID, false)
}

// GetByIDForUpdate mengambil buku dengan row lock, khusus untuk update stock dalam transaksi.
func (r *BookRepository) GetByIDForUpdate(ctx context.Context, tx *sql.Tx, bookID int) (*model.Book, error) {
	return r.getByID(ctx, tx, bookID, true)
}

func (r *BookRepository) GetAll(ctx context.Context) ([]model.Book, error) {
	query := `SELECT id, title, author, stock FROM books ORDER BY title`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var books []model.Book

	for rows.Next() {
		var book model.Book

		if err := rows.Scan(&book.ID, &book.Title, &book.Author, &book.Stock); err != nil {
			return nil, err
		}

		books = append(books, book)
	}

	return books, rows.Err()
}

// adjustStock mengubah stok buku secara atomic (+ untuk tambah, - untuk kurang).
// Pendekatan UPDATE langsung lebih aman dari race condition daripada SELECT lalu UPDATE.
func (r *BookRepository) adjustStock(ctx context.Context, tx *sql.Tx, bookID int, amount int) error {
	if amount == 0 {
		return fmt.Errorf("jumlah harus lebih besar dari 0")
	}

	// Membuat query dinamis
	query := `UPDATE books SET stock = stock + ? WHERE id = ?`
	params := []interface{}{amount, bookID}

	if amount < 0 {
		query = `UPDATE books SET stock = stock + ? WHERE id = ? AND stock > 0`
	}

	var result sql.Result
	var err error

	result, err = tx.ExecContext(ctx, query, params...)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("buku dengan ID %d tidak ditemukan: %w", bookID, sql.ErrNoRows)
	}

	return nil
}

// DecrementStock mengurangi stok buku dalam transaction peminjaman
func (r *BookRepository) DecrementStock(ctx context.Context, tx *sql.Tx, bookID int) error {
	return r.adjustStock(ctx, tx, bookID, -1)
}

// IncrementStock mengurangi stok buku saat pengembalian
func (r *BookRepository) IncrementStock(ctx context.Context, tx *sql.Tx, bookID int) error {
	return r.adjustStock(ctx, tx, bookID, +1)
}
