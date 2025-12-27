package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Ar1veeee/library-api/internal/model"
)

type LoanRepository struct {
	db *sql.DB
}

func NewLoanRepository(db *sql.DB) *LoanRepository {
	return &LoanRepository{db: db}
}

// CountActiveLoansByMember menghitung jumlah buku yang sedang dipinjam
// MENGAPA menggunakan FOR UPDATE?
//   - Lock semua row di tabel loans yang memenuhi kondisi WHERE (member_id = ? AND returned_at IS NULL).
//   - Alasan memilih row-level lock di sini daripada table lock atau lock di tabel members:
//     Member hanya memiliki kuota maksimal 3 buku aktif. Tanpa lock, dua request borrow bersamaan bisa sama-sama membaca count = 2,
//     lalu keduanya berhasil insert → total menjadi 4 (race condition).
//   - FOR UPDATE pada query COUNT memastikan transaksi kedua menunggu hingga transaksi pertama commit/rollback,
//     sehingga kuota selalu konsisten bahkan pada concurrency tinggi.
func (r *LoanRepository) CountActiveLoansByMember(ctx context.Context, tx *sql.Tx, memberID int) (int, error) {
	query := `
       SELECT count(*)
       FROM loans
       WHERE member_id = ? AND returned_at IS NULL
       FOR UPDATE
    `

	var count int
	err := tx.QueryRowContext(ctx, query, memberID).Scan(&count)
	return count, err
}

// CheckActiveLoanExists cek apakah member sedang meminjam buku tertentu
// Tidak menggunakan FOR UPDATE karena fungsi ini hanya read-only untuk validasi duplikat.
// Locking tidak diperlukan karena tidak mengubah data dan hasilnya hanya untuk pencegahan logika bisnis,
// bukan untuk menjaga integritas kuota/stock.
func (r *LoanRepository) CheckActiveLoanExists(ctx context.Context, tx *sql.Tx, memberID, bookID int) (bool, error) {
	query := `
       SELECT EXISTS(
          SELECT 1
          FROM loans
          WHERE member_id = ? AND book_id = ? AND returned_at IS NULL
       )
    `

	// Alasan menggunakan SELECT EXISTS daripada COUNT atau SELECT 1 LIMIT 1:
	// - Lebih efisien: database bisa berhenti segera setelah menemukan satu baris yang cocok.
	// - Semantik lebih jelas dan idiomatic untuk pengecekan keberadaan record.
	var exists bool
	err := tx.QueryRowContext(ctx, query, memberID, bookID).Scan(&exists)
	return exists, err
}

// Create membuat record peminjaman baru
func (r *LoanRepository) Create(ctx context.Context, tx *sql.Tx, memberID, bookID int) (int64, error) {
	query := `INSERT INTO loans (member_id, book_id, borrowed_at) VALUES (?, ?, NOW())`

	// Alasan menggunakan NOW() di sisi database:
	// - Konsistensi waktu: semua server menggunakan waktu database yang sama, menghindari perbedaan clock antar instance.
	// - Atomic dengan insert, sehingga tidak ada race pada timestamp.
	result, err := tx.ExecContext(ctx, query, memberID, bookID)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func (r *LoanRepository) GetActiveLoanByMemberAndBook(ctx context.Context, tx *sql.Tx, memberID, bookID int) (*model.Loan, error) {
	query := `
       SELECT id, member_id, book_id, borrowed_at, returned_at
       FROM loans
       WHERE member_id = ? AND book_id = ?
       FOR UPDATE
    `

	// Alasan menambahkan FOR UPDATE:
	// - Digunakan saat proses return buku untuk mencegah dua request return yang sama secara bersamaan.
	// - Meskipun jarang terjadi, lock ini menjamin integritas jika ada retry atau concurrent call.
	// - Juga berguna jika nanti ditambahkan logika lain (misalnya perpanjangan pinjaman) yang memerlukan ekslusif access ke row loan.
	var loan model.Loan
	err := tx.QueryRowContext(ctx, query, memberID, bookID).Scan(
		&loan.ID, &loan.MemberID, &loan.BookID, &loan.BorrowedAt, &loan.ReturnedAt,
	)

	// Alasan mengembalikan (nil, nil) bukannya error khusus saat sql.ErrNoRows:
	// - Memudahkan service layer untuk membedakan "tidak ditemukan" (bisa return 404) dari "error server" tanpa wrapping error tambahan.
	// - Mengurangi boilerplate di service: cukup cek jika result == nil maka not found.
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	return &loan, err
}

func (r *LoanRepository) MarkAsReturned(ctx context.Context, tx *sql.Tx, loanID int) error {
	query := `UPDATE loans SET returned_at = NOW() WHERE id = ?`

	// Alasan menggunakan NOW() di database dan tidak menyertakan returned_at IS NULL di WHERE:
	// - Jika loan sudah returned, update tetap berhasil tapi tidak mengubah apa-apa.
	// - Menghindari error "not found" yang tidak perlu. Operasi return bersifat idempotent dan aman diulang.
	_, err := tx.ExecContext(ctx, query, loanID)
	return err
}

func (r *LoanRepository) GetByMemberID(ctx context.Context, memberID int) ([]model.Loan, error) {
	query := `
          SELECT l.id, l.member_id, l.book_id, l.borrowed_at, l.returned_at, b.title, b.author
          FROM loans l
          JOIN books b ON l.book_id = b.id
          WHERE l.member_id = ?
          ORDER BY l.borrowed_at DESC
       `

	// Alasan JOIN dengan books dan mengambil title+author:
	// - Mengurangi kebutuhan N+1 query di service layer (tidak perlu fetch detail buku terpisah).
	// - Data langsung lengkap untuk response history peminjaman member.
	// - ORDER BY DESC agar pinjaman terbaru muncul paling atas.
	rows, err := r.db.QueryContext(ctx, query, memberID)
	if err != nil {
		return nil, err
	}

	// defer rows.Close() diletakkan segera setelah QueryContext berhasil.
	// Alasan: memastikan resource (koneksi database) selalu dibebaskan bahkan jika terjadi error di bawahnya.
	defer rows.Close()

	var loans []model.Loan
	for rows.Next() {
		var loan model.Loan
		if err := rows.Scan(
			&loan.ID, &loan.MemberID, &loan.BookID, &loan.BorrowedAt, &loan.ReturnedAt, &loan.BookTitle, &loan.BookAuthor,
		); err != nil {
			return nil, err
		}
		loans = append(loans, loan)
	}

	// Alasan: rows.Err() dapat mengembalikan error yang terjadi selama iterasi (misalnya koneksi terputus di tengah scan).
	return loans, rows.Err()
}
