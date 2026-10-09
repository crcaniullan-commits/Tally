package users

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/crcaniullan-commits/Tally/internal/dbtx"
	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/internal/util"
	"github.com/google/uuid"
)

type UserStore struct {
	db *sql.DB
}

func NewStorage(db *sql.DB) *UserStore {
	return &UserStore{db}
}

func (s *UserStore) GetByID(ctx context.Context, userID uuid.UUID) (*model.User, error) {
	query := `
		SELECT id, email, nombre, rut, role, plan_expires_at
		FROM users
		WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(ctx, util.QueryTimeoutDuration)
	defer cancel()

	user := &model.User{}

	var rawRut string

	err := s.db.QueryRowContext(ctx, query, userID).Scan(
		&user.ID,
		&user.Email,
		&user.Nombre,
		&rawRut,
		&user.Role,
		&user.PlanExpiresAt,
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

func (s *UserStore) GetByRut(ctx context.Context, userRut string) (*model.User, error) {
	query := `
		SELECT id, email, nombre, rut, role, plan_expires_at
		FROM users
		WHERE rut = $1;
	`

	ctx, cancel := context.WithTimeout(ctx, util.QueryTimeoutDuration)
	defer cancel()

	user := &model.User{}

	var rawRut string

	err := s.db.QueryRowContext(ctx, query, userRut).Scan(
		&user.ID,
		&user.Email,
		&user.Nombre,
		&rawRut,
		&user.Role,
		&user.PlanExpiresAt,
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

func (s *UserStore) Update(ctx context.Context, user *model.User) error {
	query := `
		UPDATE users
		SET password_hash = $1, nombre = $2, updated_at = NOW()
		WHERE id = $3
	`

	ctx, cancel := context.WithTimeout(ctx, util.QueryTimeoutDuration)
	defer cancel()

	res, err := s.db.ExecContext(ctx, query, user.PasswordHash.Hash, user.Nombre, user.ID)

	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()

	if err != nil {
		return err
	}

	if rows == 0 {
		return util.ErrNotFound
	}

	return nil
}

func (s *UserStore) Delete(ctx context.Context, userID uuid.UUID) error {
	query := `
		DELETE FROM users
		WHERE ID = $1
	`

	ctx, cancel := context.WithTimeout(ctx, util.QueryTimeoutDuration)
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
		return util.ErrNotFound
	}
	return nil
}

func (s *UserStore) setExpire(ctx context.Context, expires_at time.Time, userID uuid.UUID) error {
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
