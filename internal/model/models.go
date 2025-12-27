package model

import (
	"time"
)

type Book struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Author string `json:"author"`
	Stock  int    `json:"stock"`
}

type Member struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type Loan struct {
	ID         int        `json:"id"`
	MemberID   int        `json:"member_id"`
	BookID     int        `json:"book_id"`
	BorrowedAt time.Time  `json:"borrowed_at"`
	ReturnedAt *time.Time `json:"returned_at,omitempty"`

	// Additional fields untuk response
	BookTitle  string `json:"book_title,omitempty"`
	BookAuthor string `json:"book_author,omitempty"`
}
