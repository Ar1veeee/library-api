package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Ar1veeee/library-api/internal/dto"
	errorStruct "github.com/Ar1veeee/library-api/internal/errors"
	"github.com/Ar1veeee/library-api/internal/repository"
)

type LoanService struct {
	db         *sql.DB
	bookRepo   *repository.BookRepository
	memberRepo *repository.MemberRepository
	loanRepo   *repository.LoanRepository
}

func NewLoanService(
	db *sql.DB,
	bookRepo *repository.BookRepository,
	memberRepo *repository.MemberRepository,
	loanRepo *repository.LoanRepository,
) *LoanService {
	return &LoanService{
		db:         db,
		bookRepo:   bookRepo,
		memberRepo: memberRepo,
		loanRepo:   loanRepo,
	}
}

func (s *LoanService) BorrowBook(ctx context.Context, memberID, bookID int) (*dto.LoanDetail, error) {
	// Alasan memilih sql.LevelReadCommitted:
	// - Mencegah dirty read (melihat data yang belum di-commit).
	// - Masih mengizinkan non-repeatable read, yang aman untuk use case ini karena kita menggunakan row-level locking (FOR UPDATE) pada query kritis.
	// - Lebih ringan daripada REPEATABLE READ atau SERIALIZABLE, mengurangi risiko deadlock dan contention pada concurrency sedang-tinggi.
	// - Untuk sistem perpustakaan sederhana, isolation ini memberikan balance terbaik antara konsistensi dan performa.
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, errorStruct.NewAPIError(
			"Gagal memulai transaksi database",
			errorStruct.ErrCodeTxFailed,
		)
	}

	// defer tx.Rollback() diletakkan segera setelah BeginTx berhasil.
	// Alasan: memastikan rollback otomatis jika Commit tidak dipanggil (panic atau error path),
	// menjaga integritas data dan mencegah transaksi "zombie".
	defer tx.Rollback()

	// Validasi check apakah member ada
	member, err := s.memberRepo.GetByID(ctx, memberID)
	if err != nil {
		return nil, errorStruct.NewAPIError(
			fmt.Sprintf("Gagal memeriksa member :%v", err),
			errorStruct.ErrCodeTxFailed,
		)
	}
	if member == nil {
		return nil, errorStruct.NewAPIError(
			"Member tidak ditemukan",
			errorStruct.ErrCodeNotFound,
		)
	}

	// CountActiveLoansByMember menggunakan FOR UPDATE: lock semua row loan aktif member.
	// Alasan: mencegah race condition pada kuota (2 request borrow bersamaan bisa bypass batas 3).
	// Lock ini membuat transaksi kedua menunggu hingga yang pertama commit.
	activeLoans, err := s.loanRepo.CountActiveLoansByMember(ctx, tx, memberID)
	if err != nil {
		return nil, errorStruct.NewAPIError(
			fmt.Sprintf("Gagal memeriksa kuota member %v:", err),
			errorStruct.ErrCodeTxFailed,
		)
	}
	if activeLoans >= 3 {
		return nil, errorStruct.NewAPIError(
			"Member sudah mencapai batas pinjam maksimal yaitu 3 buku",
			errorStruct.ErrCodeQuotaExceeded,
		)
	}

	// GetByIDForUpdate dengan FOR UPDATE → lock row buku.
	// Alasan: mencegah dua transaksi borrow buku yang sama bersamaan sehingga stok menjadi negatif.
	// Kombinasi dengan atomic decrement membuat operasi stok benar-benar aman.
	book, err := s.bookRepo.GetByIDForUpdate(ctx, tx, bookID)
	if err != nil {
		return nil, errorStruct.NewAPIError(
			fmt.Sprintf("Gagal memeriksa buku %v:", err),
			errorStruct.ErrCodeTxFailed,
		)
	}
	if book == nil {
		return nil, errorStruct.NewAPIError(
			"Buku tidak ditemukan",
			errorStruct.ErrCodeNotFound,
		)
	}
	if book.Stock <= 0 {
		return nil, errorStruct.NewAPIError(
			"Stock buku habis",
			errorStruct.ErrCodeStockEmpty,
		)
	}

	// Validasi Check apakah member sudah pinjam buku yang sama
	exists, err := s.loanRepo.CheckActiveLoanExists(ctx, tx, memberID, bookID)
	if err != nil {
		return nil, errorStruct.NewAPIError(
			fmt.Sprintf("Gagal memeriksa status peminjaman %v:", err),
			errorStruct.ErrCodeTxFailed,
		)
	}
	if exists {
		return nil, errorStruct.NewAPIError(
			"Anda sedang meminjam buku ini",
			errorStruct.ErrCodeAlreadyBorrowed,
		)
	}

	// DecrementStock menggunakan atomic UPDATE dengan kondisi stock > 0.
	// Alasan: meskipun sudah cek stock > 0 sebelumnya, tetap gunakan atomic decrement untuk defense in depth
	// (mencegah race condition jika ada bug atau perubahan logika di masa depan).
	if err := s.bookRepo.DecrementStock(ctx, tx, bookID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errorStruct.NewAPIError(
				"Stok buku habis",
				errorStruct.ErrCodeStockEmpty,
			)
		}
		return nil, errorStruct.NewAPIError(
			fmt.Sprintf("Gagal mengurangi stok %v:", err),
			errorStruct.ErrCodeTxFailed,
		)
	}

	loanID, err := s.loanRepo.Create(ctx, tx, memberID, bookID)
	if err != nil {
		return nil, errorStruct.NewAPIError(
			fmt.Sprintf("Gagal mencatat peminjaman: %v", err),
			errorStruct.ErrCodeTxFailed,
		)
	}

	// COMMIT TRANSACTION
	if err := tx.Commit(); err != nil {
		return nil, errorStruct.NewAPIError(
			fmt.Sprintf("Gagal menyimpan transaksi: %v", err),
			errorStruct.ErrCodeTxFailed,
		)
	}

	// Alasan mengembalikan detail loan:
	// - Client langsung mendapat loan ID untuk tracking.
	// - Menampilkan detail buku dan timestamp akurat tanpa perlu query ulang.
	// - Memberikan feedback yang lebih kaya di response API.

	// Alasan menggunakan time.Now().In(time.FixedZone("WIB", 7*3600)):
	// - Timestamp di repository menggunakan NOW() (waktu server DB, biasanya UTC).
	// - Untuk response ke client (Indonesia), lebih user-friendly menampilkan waktu lokal WIB (UTC+7).
	// - FixedZone digunakan karena timezone Indonesia tidak ada DST.
	now := time.Now().In(time.FixedZone("WIB", 7*3600))
	borrowedAtFormatted := now.Format("2006-01-02 15:04:05")

	loanDetail := &dto.LoanDetail{
		LoanID:     int(loanID),
		MemberID:   memberID,
		BookID:     bookID,
		BookTitle:  book.Title,
		BookAuthor: book.Author,
		BorrowedAt: borrowedAtFormatted,
	}

	return loanDetail, nil
}

func (s *LoanService) ReturnBook(ctx context.Context, memberID, bookID int) error {
	// Isolation level sama dengan BorrowBook untuk konsistensi behavior transaksi.
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return errorStruct.NewAPIError(
			"Gagal memulai transaksi database",
			errorStruct.ErrCodeTxFailed,
		)
	}

	defer tx.Rollback()

	// GetActiveLoanByMemberAndBook menggunakan FOR UPDATE.
	// Alasan: lock row loan untuk mencegah concurrent return pada loan yang sama.
	// Juga berguna jika nanti ada logika tambahan seperti denda atau perpanjangan.
	loan, err := s.loanRepo.GetActiveLoanByMemberAndBook(ctx, tx, memberID, bookID)
	if err != nil {
		return errorStruct.NewAPIError(
			fmt.Sprintf("Gagal memeriksa peminjaman: %v", err),
			errorStruct.ErrCodeTxFailed,
		)
	}
	if loan == nil {
		return errorStruct.NewAPIError(
			"Anda tidak sedang meminjam buku ini",
			errorStruct.ErrCodeNotFound,
		)
	}

	if loan.ReturnedAt != nil {
		return errorStruct.NewAPIError(
			"Buku sudah dikembalikan",
			errorStruct.ErrCodeAlreadyReturned,
		)
	}

	// MarkAsReturned dan IncrementStock dilakukan dalam satu transaksi.
	// Alasan: menjaga atomicity — stok hanya bertambah jika pengembalian berhasil tercatat.
	if err := s.loanRepo.MarkAsReturned(ctx, tx, loan.ID); err != nil {
		return errorStruct.NewAPIError(
			fmt.Sprintf("Gagal mencatat pengembalian: %v", err),
			errorStruct.ErrCodeTxFailed,
		)
	}

	// IncrementStock tanpa kondisi khusus karena yakin stok sebelumnya sudah dikurangi.
	// Alasan: simplifikasi, dan race condition tidak mungkin karena return hanya bisa sekali per loan.
	if err := s.bookRepo.IncrementStock(ctx, tx, bookID); err != nil {
		return errorStruct.NewAPIError(
			fmt.Sprintf("Gagal menambah stok: %v", err),
			errorStruct.ErrCodeTxFailed,
		)
	}

	if err := tx.Commit(); err != nil {
		return errorStruct.NewAPIError(
			fmt.Sprintf("Gagal menyimpan transaksi: %v", err),
			errorStruct.ErrCodeTxFailed,
		)
	}

	return nil
}
