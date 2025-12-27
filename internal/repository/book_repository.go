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

	// Alasan memisahkan fungsi internal ini: menghindari duplikasi query dan logika scanning,
	// sekaligus memungkinkan penggunaan yang sama baik dengan maupun tanpa transaksi serta locking.
	// Pendekatan ini menjaga konsistensi query dan memudahkan maintenance jika kolom tabel berubah.
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

	// Alasan mengembalikan (nil, nil) bukannya error khusus saat sql.ErrNoRows:
	// - Memudahkan service layer untuk membedakan "tidak ditemukan" (bisa return 404) dari "error server" tanpa wrapping error tambahan.
	// - Mengurangi boilerplate di service: cukup cek jika result == nil maka not found.
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

	// defer rows.Close() diletakkan segera setelah QueryContext berhasil.
	// Alasan: memastikan resource (koneksi database) selalu dibebaskan bahkan jika terjadi error di bawahnya.
	defer rows.Close()

	var books []model.Book

	for rows.Next() {
		var book model.Book

		// Scan langsung ke field struct tanpa pointer sementara.
		// Alasan: lebih ringkas dan performanya cukup baik untuk jumlah data yang relatif kecil (katalog buku perpustakaan).
		if err := rows.Scan(&book.ID, &book.Title, &book.Author, &book.Stock); err != nil {
			return nil, err
		}

		books = append(books, book)
	}

	// Alasan: rows.Err() dapat mengembalikan error yang terjadi selama iterasi (misalnya koneksi terputus di tengah scan).
	return books, rows.Err()
}

// adjustStock mengubah stok buku secara atomic (+ untuk tambah, - untuk kurang).
// Pendekatan UPDATE langsung lebih aman dari race condition daripada SELECT lalu UPDATE.
func (r *BookRepository) adjustStock(ctx context.Context, tx *sql.Tx, bookID int, amount int) error {
	if amount == 0 {
		return fmt.Errorf("jumlah harus lebih besar dari 0")
	}

	// Alasan menggunakan UPDATE langsung dengan stock = stock + ?:
	// - Menghindari race condition yang bisa terjadi pada pola "read-then-write".
	// - Operasi menjadi atomic pada level database, sehingga aman untuk concurrency tinggi.
	// - Mengurangi round-trip ke database (hanya 1 statement).
	query := `UPDATE books SET stock = stock + ? WHERE id = ?`

	// Ketika mengurangi stok, ditambahkan kondisi stock > 0 untuk mencegah stock menjadi negatif.
	// Alasan memilih kondisi di WHERE daripada CHECK constraint di tabel:
	// - Memberikan kontrol error yang lebih eksplisit di aplikasi (bisa membedakan "not found" vs "insufficient stock").
	// - Memudahkan handling error yang lebih informatif ke layer service.
	if amount < 0 {
		query = `UPDATE books SET stock = stock + ? WHERE id = ? AND stock > 0`
	}

	var result sql.Result
	var err error
	params := []interface{}{amount, bookID}

	result, err = tx.ExecContext(ctx, query, params...)
	if err != nil {
		return err
	}

	// Memeriksa RowsAffected untuk mendeteksi kasus buku tidak ditemukan atau stock tidak cukup.
	// Alasan tidak mengandalkan hanya error dari DB:
	// - Beberapa driver/engine MySQL tidak mengembalikan error pada kondisi WHERE tidak terpenuhi,
	//   hanya rows affected = 0. Jadi pengecekan ini lebih portabel dan reliable.
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
// (Catatan: komentar fungsi salah, seharusnya "menambah stok")
func (r *BookRepository) IncrementStock(ctx context.Context, tx *sql.Tx, bookID int) error {
	return r.adjustStock(ctx, tx, bookID, +1)
}
