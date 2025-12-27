package service

import (
	"context"

	"github.com/Ar1veeee/library-api/internal/dto"
	"github.com/Ar1veeee/library-api/internal/model"
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
		return nil, model.NewAPIError("Member tidak ditemukan", model.ErrCodeNotFound)
	}

	// Get loan history
	loans, err := s.loanRepo.GetByMemberID(ctx, memberID)
	if err != nil {
		return nil, err
	}

	// Convert to DTO
	loanItems := make([]dto.LoanHistoryItem, len(loans))
	for i, loan := range loans {
		var returnedAt *string
		if loan.ReturnedAt != nil {
			formatted := loan.ReturnedAt.Format("2006-01-02 15:04:05")
			returnedAt = &formatted
		}

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

	response := &dto.MemberLoansResponse{
		MemberID:   member.ID,
		MemberName: member.Name,
		TotalLoans: len(loanItems),
		Loans:      loanItems,
	}

	return response, nil
}
