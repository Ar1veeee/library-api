package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Ar1veeee/library-api/internal/model"
)

type MemberRepository struct {
	db *sql.DB
}

func NewMemberRepository(db *sql.DB) *MemberRepository {
	return &MemberRepository{db: db}
}

// GetByID mengambil data member berdasarkan ID.
// Mengembalikan (*model.Member, nil) jika ditemukan, (nil, nil) jika tidak ada, dan error jika terjadi kegagalan query.
func (r *MemberRepository) GetByID(ctx context.Context, memberID int) (*model.Member, error) {
	query := `SELECT id, name, email FROM members WHERE id = ?`

	var member model.Member

	// Alasan menggunakan QueryRowContext langsung tanpa transaksi atau locking:
	// Operasi ini pure read-only dan tidak memerlukan konsistensi transaksional ketat.
	// Menjaga performa tinggi dan overhead rendah untuk operasi yang sering dipanggil (misalnya validasi member saat borrow).
	err := r.db.QueryRowContext(ctx, query, memberID).Scan(
		&member.ID, &member.Name, &member.Email,
	)

	// Alasan mengembalikan (nil, nil) bukannya error khusus saat sql.ErrNoRows:
	// - Memudahkan service layer untuk membedakan "tidak ditemukan" (bisa return 404) dari "error server" tanpa wrapping error tambahan.
	// - Mengurangi boilerplate di service: cukup cek jika result == nil maka not found.
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	return &member, err
}
