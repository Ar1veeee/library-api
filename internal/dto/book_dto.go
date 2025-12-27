package dto

// BookResponse represents single book response
type BookResponse struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Author string `json:"author"`
	Stock  int    `json:"stock"`
}

// BooksListResponse represents list of books
type BooksListResponse struct {
	Total int            `json:"total"`
	Books []BookResponse `json:"books"`
}
