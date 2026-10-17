package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/crcaniullan-commits/Tally/internal/dbtx"
	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/internal/users"
	"github.com/crcaniullan-commits/Tally/internal/util"
	"github.com/google/uuid"
)

type StoreAuth struct {
	db *sql.DB
}

func NewStorage(db *sql.DB) *StoreAuth {
	return &StoreAuth{db}
}

func (s *StoreAuth) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `
		SELECT id, email, nombre, rut, password_hash
		FROM users
		WHERE email = $1;
	`

	ctx, cancel := context.WithTimeout(ctx, util.QueryTimeoutDuration)
	defer cancel()

	user := &model.User{}

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
			return nil, util.ErrNotFound
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

func (s *StoreAuth) Create(ctx context.Context, user *model.User) error {
	query := `
		INSERT INTO users (email, password_hash, nombre, role, rut)
		VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at;
	`

	ctx, cancel := context.WithTimeout(ctx, util.QueryTimeoutDuration)
	defer cancel()

	role := user.Role
	if role == "" {
		role = "usuario"
	}

	err := dbtx.FromContext(ctx, s.db).QueryRowContext(ctx, query,
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
			return users.ErrDuplicateEmail
		} else if err.Error() == `pq: duplicate key value violates unique constraint "users_rut_key"` {
			return users.ErrDuplicateRut
		}
		return err
	}
	return nil
}

func (s *StoreAuth) setExpire(ctx context.Context, expires_at time.Time, userID uuid.UUID) error {
	query := `
		UPDATE users
		SET plan_expires_at = $1
		WHERE id = $2
	`

	exec := dbtx.FromContext(ctx, s.db)

	ctx, cancel := context.WithTimeout(ctx, util.QueryTimeoutDuration)
	defer cancel()

	_, err := exec.ExecContext(ctx, query, expires_at, userID)

	if err != nil {
		return fmt.Errorf("actualizando plan_expires_at: %w", err)
	}

	return nil
}
