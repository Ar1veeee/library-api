package dto

type BorrowBookRequest struct {
	MemberID int `json:"member_id" validate:"required,gt=0"`
	BookID   int `json:"book_id" validate:"required,gt=0"`
}

// ReturnBookRequest represents request body untuk POST /return
type ReturnBookRequest struct {
	MemberID int `json:"member_id" validate:"required,gt=0"`
	BookID   int `json:"book_id" validate:"required,gt=0"`
}

// LoanDetail represents detail peminjaman
type LoanDetail struct {
	LoanID     int    `json:"loan_id"`
	MemberID   int    `json:"member_id"`
	BookID     int    `json:"book_id"`
	BookTitle  string `json:"book_title"`
	BookAuthor string `json:"book_author"`
	BorrowedAt string `json:"borrowed_at"`
}
