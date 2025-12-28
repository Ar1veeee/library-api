package routes

import (
	"net/http"

	handler2 "github.com/Ar1veeee/library-api/internal/http/handler"
	"github.com/gorilla/mux"
)

func RegisterRoutes(router *mux.Router, bookHandler *handler2.BookHandler, memberHandler *handler2.MemberHandler, loanHandler *handler2.LoanHandler) {
	api := router.PathPrefix("/api/v1").Subrouter()

	// Health check
	api.HandleFunc("/health", healthHandler).Methods("GET")

	// Loan
	api.HandleFunc("/borrow", loanHandler.BorrowBook).Methods("POST")
	api.HandleFunc("/return", loanHandler.ReturnBook).Methods("POST")

	// Books
	api.HandleFunc("/books", bookHandler.GetBooks).Methods("GET")
	api.HandleFunc("/books/{id}", bookHandler.GetBookByID).Methods("GET")

	// Members
	api.HandleFunc("/members/{id}/loans", memberHandler.GetMemberLoans).Methods("GET")
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "ok", "service": "library-api"}`))
}
