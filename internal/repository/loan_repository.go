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
// - Lock row member untuk prevent race condition
// - Jika 2 request borrow bersamaan, yang kedua harus tunggu yang pertama selesai
// - Tanpa lock: bisa bypass kuota 3 buku jika request bersamaan
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
func (r *LoanRepository) CheckActiveLoanExists(ctx context.Context, tx *sql.Tx, memberID, bookID int) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1
			FROM loans
			WHERE member_id = ? AND book_id = ? AND returned_at IS NULL
		)
	`

	var exists bool
	err := tx.QueryRowContext(ctx, query, memberID, bookID).Scan(&exists)
	return exists, err
}

// Create membuat record peminjaman baru
func (r *LoanRepository) Create(ctx context.Context, tx *sql.Tx, memberID, bookID int) (int64, error) {
	query := `INSERT INTO loans (member_id, book_id, borrowed_at) VALUES (?, ?, NOW())`

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

	var loan model.Loan
	err := tx.QueryRowContext(ctx, query, memberID, bookID).Scan(
		&loan.ID, &loan.MemberID, &loan.BookID, &loan.BorrowedAt, &loan.ReturnedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	return &loan, err
}

func (r *LoanRepository) MarkAsReturned(ctx context.Context, tx *sql.Tx, loanID int) error {
	query := `UPDATE loans SET returned_at = NOW() WHERE id = ?`
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

	rows, err := r.db.QueryContext(ctx, query, memberID)
	if err != nil {
		return nil, err
	}

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

	return loans, nil
}
