package users

import (
	"context"
	"database/sql"
	"errors"
	"time"

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

func (s *UserStore) GetExpire(ctx context.Context, code string) (model.AccessKey, error) {
	query := `
		SELECT id, code, created_by, redeemed_by,
		       redeemed_at, expires_at, revoked_at, revoked_by, created_at
		FROM access_keys
		WHERE code = $1
	`

	ctx, cancel := context.WithTimeout(ctx, util.QueryTimeoutDuration)
	defer cancel()

	var AccessKey model.AccessKey

	err := s.db.QueryRowContext(ctx, query, code).Scan(
		&AccessKey.ID,
		&AccessKey.Code,
		&AccessKey.CreatedBy,
		&AccessKey.RedeemedBy,
		&AccessKey.RedeemedAt,
		&AccessKey.ExpiresAt,
		&AccessKey.RevokedAt,
		&AccessKey.RevokedBy,
		&AccessKey.CreatedAt,
	)

	if err != nil {
		switch {
		case err == sql.ErrNoRows:
			return model.AccessKey{}, util.ErrNotFound
		default:
			return model.AccessKey{}, err
		}
	}

	return AccessKey, nil

}

func (s *UserStore) setExpire(ctx context.Context, expires_at time.Time, userID uuid.UUID) error {
	query := `
		UPDATE users
		SET plan_expires_at = $1
		WHERE id = $2
	`

	ctx, cancel := context.WithTimeout(ctx, util.QueryTimeoutDuration)
	defer cancel()

	res, err := s.db.ExecContext(ctx, query, expires_at, userID)

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
