package config

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	ServerPort string
}

func Load() *Config {
	return &Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "3306"),
		DBUser:     getEnv("DB_USER", "root"),
		DBPassword: getEnv("DB_PASSWORD", "secret"),
		DBName:     getEnv("DB_NAME", "library_db"),
		ServerPort: getEnv("SERVER_PORT", "8080"),
	}
}

func NewDatabase(cfg *Config) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// MENGAPA mengatur pengaturan pool koneksi?
	// - Untuk mengatur concurrent requests (API diakses banyak user sekaligus) dengan baik
	// - MaxOpenConns: Maksimal jumlah koneksi yang boleh dibuka (hindari overload DB)
	// - MaxIdleConns: Maksimal koneksi yang tidak langsung ditutup (hindari overhead create/destroy dan reuseable)
	// - ConnMaxLifetime: Batas maksimal umur 1 koneksi yang wajib ditutup dan diganti baru (hindari stale connections)
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return defaultValue
}
