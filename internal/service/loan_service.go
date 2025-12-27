package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Ar1veeee/library-api/internal/dto"
	"github.com/Ar1veeee/library-api/internal/model"
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
	// MENGAPA menggunakan sql.LevelReadCommitted?
	// - READ UNCOMMITTED: Terlalu loose, bisa dirty read
	// - READ COMMITTED: Balance yang baik, cegah dirty read tapi allow non-repeatable read
	// - REPEATABLE READ: Lebih strict, tapi bisa phantom read & deadlock lebih sering
	// - SERIALIZABLE: Terlalu strict, performance buruk untuk high concurrency
	// Untuk use case library yang tidak butuh strict serialization, READ COMMITTED cukup
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, model.NewAPIError(
			"Gagal memulai transaksi database",
			model.ErrCodeTxFailed,
		)
	}

	defer tx.Rollback()

	// Validasi check apakah member ada
	member, err := s.memberRepo.GetByID(ctx, memberID)
	if err != nil {
		return nil, model.NewAPIError(
			fmt.Sprintf("Gagal memeriksa member :%v", err),
			model.ErrCodeTxFailed,
		)
	}
	if member == nil {
		return nil, model.NewAPIError(
			"Member tidak ditemukan",
			model.ErrCodeNotFound,
		)
	}

	// Validasi check kuota member (saya buat MAX 3 BUKU)
	activeLoans, err := s.loanRepo.CountActiveLoansByMember(ctx, tx, memberID)
	if err != nil {
		return nil, model.NewAPIError(
			fmt.Sprintf("Gagal memeriksa kuota member %v:", err),
			model.ErrCodeTxFailed,
		)
	}
	if activeLoans >= 3 {
		return nil, model.NewAPIError(
			"Member sudah mencapai batas pinjam maksimal yaitu 3 buku",
			model.ErrCodeQuotaExceeded,
		)
	}

	// Validasi check & lock stok buku
	book, err := s.bookRepo.GetByIDForUpdate(ctx, tx, bookID)
	if err != nil {
		return nil, model.NewAPIError(
			fmt.Sprintf("Gagal memeriksa buku %v:", err),
			model.ErrCodeTxFailed,
		)
	}
	if book == nil {
		return nil, model.NewAPIError(
			"Buku tidak ditemukan",
			model.ErrCodeNotFound,
		)
	}
	if book.Stock <= 0 {
		return nil, model.NewAPIError(
			"Stock buku habis",
			model.ErrCodeStockEmpty,
		)
	}

	// Validasi Check apakah member sudah pinjam buku yang sama
	exists, err := s.loanRepo.CheckActiveLoanExists(ctx, tx, memberID, bookID)
	if err != nil {
		return nil, model.NewAPIError(
			fmt.Sprintf("Gagal memeriksa status peminjaman %v:", err),
			model.ErrCodeTxFailed,
		)
	}
	if exists {
		return nil, model.NewAPIError(
			"Anda sedang meminjam buku ini",
			model.ErrCodeAlreadyBorrowed,
		)
	}

	// Aksi Kurangi stok buku
	if err := s.bookRepo.DecrementStock(ctx, tx, bookID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.NewAPIError(
				"Stok buku habis",
				model.ErrCodeStockEmpty,
			)
		}
		return nil, model.NewAPIError(
			fmt.Sprintf("Gagal mengurangi stok %v:", err),
			model.ErrCodeTxFailed,
		)
	}

	loanID, err := s.loanRepo.Create(ctx, tx, memberID, bookID)
	if err != nil {
		return nil, model.NewAPIError(
			fmt.Sprintf("Gagal mencatat peminjaman: %v", err),
			model.ErrCodeTxFailed,
		)
	}

	// COMMIT TRANSACTION
	if err := tx.Commit(); err != nil {
		return nil, model.NewAPIError(
			fmt.Sprintf("Gagal menyimpan transaksi: %v", err),
			model.ErrCodeTxFailed,
		)
	}

	// MENGAPA saya return detail loan?
	// - Client bisa langsung tahu loan_id yang baru dibuat
	// - Client bisa langsung tampilkan detail buku yang dipinjam
	// - Timestamp actual saat transaksi berhasil (bukan "Baru saja")
	// ============================================================

	// MENGAPA time.Now().In(time.FixedZone("WIB", 7*3600))?
	// - Get current time saat request ini diproses
	// - Convert ke timezone WIB (UTC+7) untuk Indonesia
	// - Format: "2006-01-02 15:04:05" adalah Go's reference time format
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
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return model.NewAPIError(
			"Gagal memulai transaksi database",
			model.ErrCodeTxFailed,
		)
	}

	defer tx.Rollback()

	// Mencari peminjaman aktif
	loan, err := s.loanRepo.GetActiveLoanByMemberAndBook(ctx, tx, memberID, bookID)
	if err != nil {
		return model.NewAPIError(
			fmt.Sprintf("Gagal memeriksa peminjaman: %v", err),
			model.ErrCodeTxFailed,
		)
	}
	if loan == nil {
		return model.NewAPIError(
			"Anda tidak sedang meminjam buku ini",
			model.ErrCodeNotFound,
		)
	}

	if loan.ReturnedAt != nil {
		return model.NewAPIError(
			"Buku sudah dikembalikan",
			model.ErrCodeAlreadyReturned,
		)
	}

	// Tandai peminjaman telah dikembalikan
	if err := s.loanRepo.MarkAsReturned(ctx, tx, loan.ID); err != nil {
		return model.NewAPIError(
			fmt.Sprintf("Gagal mencatat pengembalian: %v", err),
			model.ErrCodeTxFailed,
		)
	}

	// Menambah stok buku kembalki
	if err := s.bookRepo.IncrementStock(ctx, tx, bookID); err != nil {
		return model.NewAPIError(
			fmt.Sprintf("Gagal menambah stok: %v", err),
			model.ErrCodeTxFailed,
		)
	}

	if err := tx.Commit(); err != nil {
		return model.NewAPIError(
			fmt.Sprintf("Gagal menyimpan transaksi: %v", err),
			model.ErrCodeTxFailed,
		)
	}

	return nil
}
