package dto

// MemberLoansResponse represents member loan history
type MemberLoansResponse struct {
	MemberID   int               `json:"member_id"`
	MemberName string            `json:"member_name"`
	TotalLoans int               `json:"total_loans"`
	Loans      []LoanHistoryItem `json:"loans"`
}

// LoanHistoryItem represents single loan in history
type LoanHistoryItem struct {
	LoanID     int     `json:"loan_id"`
	BookID     int     `json:"book_id"`
	BookTitle  string  `json:"book_title"`
	BookAuthor string  `json:"book_author"`
	BorrowedAt string  `json:"borrowed_at"`
	ReturnedAt *string `json:"returned_at,omitempty"`
	Status     string  `json:"status"`
}
