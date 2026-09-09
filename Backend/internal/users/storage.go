package users

import (
	"context"
	"database/sql"
	"errors"

	"github.com/crcaniullan-commits/Tally/internal/util"
	"github.com/google/uuid"
)

type StoreUser struct {
	db *sql.DB
}

func NewStorage(db *sql.DB) StoreUser {
	return StoreUser{db}
}

func (s *StoreUser) GetByID(ctx context.Context, userID uuid.UUID) (*Users, error) {
	query := `
		SELECT id, email, nombre, rut
		FROM users
		WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	user := &Users{}

	var rawRut string

	err := s.db.QueryRowContext(ctx, query, userID).Scan(
		&user.ID,
		&user.Email,
		&user.Nombre,
		&rawRut,
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

func (s *StoreUser) GetByRut(ctx context.Context, userRut string) (*Users, error) {
	query := `
		SELECT id, email, nombre, rut 
		FROM users
		WHERE rut = $1;
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	user := &Users{}

	var rawRut string

	err := s.db.QueryRowContext(ctx, query, userRut).Scan(
		&user.ID,
		&user.Email,
		&user.Nombre,
		&rawRut,
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

func (s *StoreUser) GetByEmail(ctx context.Context, email string) (*Users, error) {
	query := `
		SELECT id, email, nombre, rut 
		FROM users
		WHERE email = $1;
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	user := &Users{}

	var rawRut string

	err := s.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Nombre,
		&rawRut,
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

func (s *StoreUser) Create(ctx context.Context, user *Users) error {
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
		user.PasswordHash,
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
		}
		return err
	}
	return nil
}

func (s *StoreUser) Update(ctx context.Context, user *Users) error {
	query := `
		UPDATE users
		SET password_hash = $1, nombre = $2, updated_at = NOW()
		WHERE id = $3
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	res, err := s.db.ExecContext(ctx, query, user.PasswordHash, user.Nombre, user.ID)

	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()

	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

func (s *StoreUser) Delete(ctx context.Context, userID uuid.UUID) error {
	query := `
		DELETE FROM users
		WHERE ID = $1
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	res, err := s.db.ExecContext(ctx, query, userID)

	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()

	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrNotFound
	}
	return nil
}
