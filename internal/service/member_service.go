package service

import (
	"context"

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

func (s *MemberService) GetMemberLoans(ctx context.Context, memberID int) ([]model.Loan, error) {
	member, err := s.memberRepo.GetByID(ctx, memberID)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, model.NewAPIError("Member tidak ditemukan", model.ErrCodeNotFound)
	}

	return s.loanRepo.GetByMemberID(ctx, memberID)
}
