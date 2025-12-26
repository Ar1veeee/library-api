package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Ar1veeee/library-api/internal/model"
)

type MemberRepository struct {
	db *sql.DB
}

func NewMemberRepository(db *sql.DB) *MemberRepository {
	return &MemberRepository{db: db}
}

func (r *MemberRepository) GetByID(ctx context.Context, memberID int) (*model.Member, error) {
	query := `SELECT id, name, email FROM members WHERE id = ?`

	var member model.Member
	err := r.db.QueryRowContext(ctx, query, memberID).Scan(
		&member.ID, &member.Name, &member.Email,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	return &member, err
}
