package main

import (
	"log"
	"net/http"

	"github.com/Ar1veeee/library-api/internal/config"
	"github.com/Ar1veeee/library-api/internal/handler"
	"github.com/Ar1veeee/library-api/internal/repository"
	"github.com/Ar1veeee/library-api/internal/service"
	"github.com/gorilla/mux"
)

func main() {
	cfg := config.Load()

	db, err := config.NewDatabase(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database %v:", err)
	}

	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database %v:", err)
	}
	log.Println("Database is ready")

	bookRepo := repository.NewBookRepository(db)
	memberRepo := repository.NewMemberRepository(db)
	loanRepo := repository.NewLoanRepository(db)

	bookService := service.NewBookService(bookRepo)
	memberService := service.NewMemberService(memberRepo, loanRepo)
	loanService := service.NewLoanService(db, bookRepo, memberRepo, loanRepo)

	bookHandler := handler.NewBookHandler(bookService)
	memberHandler := handler.NewMemberHandler(memberService)
	loanHandler := handler.NewLoanHandler(loanService)

	router := mux.NewRouter()

	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok", "service": "library-api"}`))
	}).Methods("GET")

	// API routes prefix
	api := router.PathPrefix("/api/v1").Subrouter()

	// Loan endpoints
	api.HandleFunc("/borrow", loanHandler.BorrowBook).Methods("POST")
	api.HandleFunc("/return", loanHandler.ReturnBook).Methods("POST")

	// Book endpoints
	api.HandleFunc("/books", bookHandler.GetBooks).Methods("GET")
	api.HandleFunc("/books/{id}", bookHandler.GetBookByID).Methods("GET")

	// Member endpoints
	api.HandleFunc("/members/{id}/loans", memberHandler.GetMemberLoans).Methods("GET")

	addr := ":" + cfg.ServerPort
	log.Printf("Server starting on %s", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
