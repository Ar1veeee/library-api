package main

import (
	"log"
	"net/http"

	"github.com/Ar1veeee/library-api/internal/config"
	handler2 "github.com/Ar1veeee/library-api/internal/http/handler"
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
	log.Println("✅ Database connected successfully")

	bookRepo := repository.NewBookRepository(db)
	memberRepo := repository.NewMemberRepository(db)
	loanRepo := repository.NewLoanRepository(db)

	bookService := service.NewBookService(bookRepo)
	memberService := service.NewMemberService(memberRepo, loanRepo)
	loanService := service.NewLoanService(db, bookRepo, memberRepo, loanRepo)

	bookHandler := handler2.NewBookHandler(bookService)
	memberHandler := handler2.NewMemberHandler(memberService)
	loanHandler := handler2.NewLoanHandler(loanService)

	router := mux.NewRouter()

	// API routes prefix
	api := router.PathPrefix("/api/v1").Subrouter()

	api.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok", "service": "library-api"}`))
	}).Methods("GET")

	// Loan endpoints
	api.HandleFunc("/borrow", loanHandler.BorrowBook).Methods("POST")
	api.HandleFunc("/return", loanHandler.ReturnBook).Methods("POST")

	// Book endpoints
	api.HandleFunc("/books", bookHandler.GetBooks).Methods("GET")
	api.HandleFunc("/books/{id}", bookHandler.GetBookByID).Methods("GET")

	// Member endpoints
	api.HandleFunc("/members/{id}/loans", memberHandler.GetMemberLoans).Methods("GET")

	addr := ":" + cfg.ServerPort
	log.Printf("🚀 Server starting on %s", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
