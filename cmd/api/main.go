package main

import (
	"log"
	"net/http"

	"github.com/Ar1veeee/library-api/internal/config"
	"github.com/Ar1veeee/library-api/internal/http/handler"
	"github.com/Ar1veeee/library-api/internal/http/routes"
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

	bookHandler := handler.NewBookHandler(bookService)
	memberHandler := handler.NewMemberHandler(memberService)
	loanHandler := handler.NewLoanHandler(loanService)

	router := mux.NewRouter()
	routes.RegisterRoutes(router, bookHandler, memberHandler, loanHandler)

	addr := ":" + cfg.ServerPort
	log.Printf("🚀 Server starting on %s", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
