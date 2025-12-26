package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

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

func (s *LoanService) BorrowBook(ctx context.Context, memberID, bookID int) error {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return model.NewAPIError("Gagal memulai transaksi database", model.ErrCodeTxFailed)
	}

	defer tx.Rollback()

	// Validasi check apakah member ada
	member, err := s.memberRepo.GetByID(ctx, memberID)
	if err != nil {
		return model.NewAPIError(
			fmt.Sprintf("Gagal memeriksa member :%v", err),
			model.ErrCodeTxFailed,
		)
	}
	if member == nil {
		return model.NewAPIError(
			"Member tidak ditemukan",
			model.ErrCodeNotFound,
		)
	}

	// Validasi check kuota member (saya buat MAX 3 BUKU)
	activeLoans, err := s.loanRepo.CountActiveLoansByMember(ctx, tx, memberID)
	if err != nil {
		return model.NewAPIError(
			fmt.Sprintf("Gagal memeriksa kuota member %v:", err),
			model.ErrCodeTxFailed,
		)
	}
	if activeLoans >= 3 {
		return model.NewAPIError(
			"Member sudah mencapai batas pinjam maksimal yaitu 3 buku",
			model.ErrCodeQuotaExceeded,
		)
	}

	// Validasi check & lock stok buku
	book, err := s.bookRepo.GetByIDForUpdate(ctx, tx, bookID)
	if err != nil {
		return model.NewAPIError(
			fmt.Sprintf("Gagal memeriksa buku %v:", err),
			model.ErrCodeTxFailed,
		)
	}
	if book == nil {
		return model.NewAPIError(
			"Buku tidak ditemukan",
			model.ErrCodeNotFound,
		)
	}
	if book.Stock <= 0 {
		return model.NewAPIError(
			"Stock buku habis",
			model.ErrCodeStockEmpty,
		)
	}

	// Validasi Check apakah member sudah pinjam buku yang sama
	exists, err := s.loanRepo.CheckActiveLoanExists(ctx, tx, memberID, bookID)
	if err != nil {
		return model.NewAPIError(
			fmt.Sprintf("Gagal memeriksa status peminjaman %v:", err),
			model.ErrCodeTxFailed,
		)
	}
	if exists {
		return model.NewAPIError(
			"Anda sedang meminjam buku ini",
			model.ErrCodeAlreadyBorrowed,
		)
	}

	// Aksi Kurangi stok buku
	if err := s.bookRepo.DecrementStock(ctx, tx, bookID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.NewAPIError(
				"Stok buku habis",
				model.ErrCodeStockEmpty,
			)
		}
		return model.NewAPIError(
			fmt.Sprintf("Gagal mengurangi stok %v:", err),
			model.ErrCodeTxFailed,
		)
	}

	_, err = s.loanRepo.Create(ctx, tx, memberID, bookID)
	if err != nil {
		return model.NewAPIError(
			fmt.Sprintf("Gagal mencatat peminjaman %v:", err),
			model.ErrCodeTxFailed,
		)
	}

	if err := tx.Commit(); err != nil {
		return model.NewAPIError(
			fmt.Sprintf("Gagal menyimpan transaksi %v:", err),
			model.ErrCodeTxFailed,
		)
	}

	return nil
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
