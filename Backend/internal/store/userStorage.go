package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/crcaniullan-commits/Tally/internal/util"
	"github.com/google/uuid"
)

type Users struct {
	ID           uuid.UUID     `json:"id"`
	Email        string        `json:"email"`
	PasswordHash string        `json:"-"`
	Nombre       string        `json:"nombre"`
	Role         util.UserRole `json:"role"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
}

type StoreUser struct {
	db *sql.DB
}

func (s *StoreUser) GetByID(ctx context.Context, userID uuid.UUID) (*Users, error) {
	return nil, nil
}

func (s *StoreUser) Create(ctx context.Context, user *Users) error {
	query := `
		INSERT INTO users (email, password_hash, nombre, role)
		VALUES ($1, $2, $3, $4) RETURNING id, created_at
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	role := user.Role
	if role == "" {
		role = "usuario"
	}

	err := s.db.QueryRowContext(ctx, query,
		user.Email,
		user.PasswordHash,
		user.Nombre,
		user.Role,
	).Scan(
		&user.ID,
		&user.CreatedAt,
	)
	if err != nil {
		if err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"` {
			return ErrDuplicateEmail
		}
		return err
	}
	return nil
}

func (s *StoreUser) Update(ctx context.Context, userID uuid.UUID) error {
	return nil
}

func (s *StoreUser) Delete(ctx context.Context, userID uuid.UUID) error {
	return nil
}
