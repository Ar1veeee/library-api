package service

import (
	"context"

	"github.com/Ar1veeee/library-api/internal/dto"
	"github.com/Ar1veeee/library-api/internal/errors"
	"github.com/Ar1veeee/library-api/internal/repository"
)

type MemberService struct {
	memberRepo *repository.MemberRepository
	loanRepo   *repository.LoanRepository
}

func NewMemberService(memberRepo *repository.MemberRepository, loanRepo *repository.LoanRepository) *MemberService {
	return &MemberService{
		memberRepo: memberRepo,
		loanRepo:   loanRepo,
	}
}

func (s *MemberService) GetMemberLoans(ctx context.Context, memberID int) (*dto.MemberLoansResponse, error) {
	member, err := s.memberRepo.GetByID(ctx, memberID)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, errors.NewAPIError("Member tidak ditemukan", errors.ErrCodeNotFound)
	}

	loans, err := s.loanRepo.GetByMemberID(ctx, memberID)
	if err != nil {
		return nil, err
	}

	// Pre-allocate slice dengan panjang pasti untuk menghindari multiple reallocation saat append.
	// Alasan: performa lebih baik dan lebih predictable memory usage.
	loanItems := make([]dto.LoanHistoryItem, len(loans))
	for i, loan := range loans {
		// ReturnedAt di DTO bertipe *string agar bisa null ketika buku belum dikembalikan.
		// Alasan menggunakan pointer daripada string kosong:
		// - Representasi JSON yang benar: null vs "" memiliki makna berbeda di API response.
		// - Client bisa membedakan "belum dikembalikan" (null) vs "dikembalikan tapi timestamp kosong".
		var returnedAt *string
		if loan.ReturnedAt != nil {
			formatted := loan.ReturnedAt.Format("2006-01-02 15:04:05")
			returnedAt = &formatted
		}

		// Status "loan" vs "returned" ditentukan di service layer, bukan di repository.
		// Alasan: status adalah derived data untuk keperluan presentasi (DTO), bukan data domain murni.
		// Menjaga repository tetap fokus pada persistence, sementara service menangani business/presentation logic.
		status := "loan"
		if loan.ReturnedAt != nil {
			status = "returned"
		}

		loanItems[i] = dto.LoanHistoryItem{
			LoanID:     loan.ID,
			BookID:     loan.BookID,
			BookTitle:  loan.BookTitle,
			BookAuthor: loan.BookAuthor,
			BorrowedAt: loan.BorrowedAt.Format("2006-01-02 15:04:05"),
			ReturnedAt: returnedAt,
			Status:     status,
		}
	}

	// TotalLoans dihitung dari len(loanItems) termasuk yang sudah returned.
	// Alasan: memberikan informasi lengkap riwayat peminjaman (bukan hanya loan).
	response := &dto.MemberLoansResponse{
		MemberID:   member.ID,
		MemberName: member.Name,
		TotalLoans: len(loanItems),
		Loans:      loanItems,
	}

	return response, nil
}
