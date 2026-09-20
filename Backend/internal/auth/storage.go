package auth

import (
	"context"
	"database/sql"
	"errors"

	"github.com/crcaniullan-commits/Tally/internal/users"
	"github.com/crcaniullan-commits/Tally/internal/util"
)

type StoreAuth struct {
	db *sql.DB
}

func NewStorage(db *sql.DB) *StoreAuth {
	return &StoreAuth{db}
}

func (s *StoreAuth) GetByEmail(ctx context.Context, email string) (*users.Users, error) {
	query := `
		SELECT id, email, nombre, rut, password_hash
		FROM users
		WHERE email = $1;
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	user := &users.Users{}

	var rawRut string

	err := s.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Nombre,
		&rawRut,
		&user.PasswordHash.Hash,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNotFound
		}
		return nil, err
	}

	parsedRut, err := util.ParseRUT(rawRut)

	if err != nil {
		return nil, err
	}

	user.Rut = parsedRut

	return user, nil
}

func (s *StoreAuth) Create(ctx context.Context, user *users.Users) error {
	query := `
		INSERT INTO users (email, password_hash, nombre, role, rut)
		VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at;
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	role := user.Role
	if role == "" {
		role = "usuario"
	}

	err := s.db.QueryRowContext(ctx, query,
		user.Email,
		user.PasswordHash.GetHash(),
		user.Nombre,
		user.Role,
		user.Rut.String(),
	).Scan(
		&user.ID,
		&user.CreatedAt,
	)
	if err != nil {
		if err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"` {
			return ErrDuplicateEmail
		} else if err.Error() == `pq: duplicate key value violates unique constraint "users_rut_key"` {
			return ErrDuplicateRut
		}
		return err
	}
	return nil
}
