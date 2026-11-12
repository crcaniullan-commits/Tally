package users

import (
	"context"
	"database/sql"
	"errors"

	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/internal/utilidades"
	"github.com/google/uuid"
)

type AlmacenUsuarios struct {
	db *sql.DB
}

func NuevoAlmacenUsuarios(db *sql.DB) *AlmacenUsuarios {
	return &AlmacenUsuarios{db}
}

func (s *AlmacenUsuarios) ObtenerPorID(ctx context.Context, userID uuid.UUID) (*model.User, error) {
	query := `
		SELECT id, email, nombre, rut, role, plan_expires_at
		FROM users
		WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(ctx, utilidades.DuracionTimeoutConsulta)
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
			return nil, utilidades.ErrorNoEncontrado
		}
		return nil, err
	}

	parsedRut, err := utilidades.ParseRUT(rawRut)

	if err != nil {
		return nil, err
	}

	user.Rut = parsedRut

	return user, nil
}

func (s *AlmacenUsuarios) ObtenerPorRUT(ctx context.Context, userRut string) (*model.User, error) {
	query := `
		SELECT id, email, nombre, rut, role, plan_expires_at
		FROM users
		WHERE rut = $1;
	`

	ctx, cancel := context.WithTimeout(ctx, utilidades.DuracionTimeoutConsulta)
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
			return nil, utilidades.ErrorNoEncontrado
		}
		return nil, err
	}

	parsedRut, err := utilidades.ParseRUT(rawRut)

	if err != nil {
		return nil, err
	}

	user.Rut = parsedRut

	return user, nil
}

func (s *AlmacenUsuarios) Actualizar(ctx context.Context, user *model.User) error {
	query := `
		UPDATE users
		SET password_hash = $1, nombre = $2, updated_at = NOW()
		WHERE id = $3
	`

	ctx, cancel := context.WithTimeout(ctx, utilidades.DuracionTimeoutConsulta)
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
		return utilidades.ErrorNoEncontrado
	}

	return nil
}

func (s *AlmacenUsuarios) Eliminar(ctx context.Context, userID uuid.UUID) error {
	query := `
		DELETE FROM users
		WHERE ID = $1
	`

	ctx, cancel := context.WithTimeout(ctx, utilidades.DuracionTimeoutConsulta)
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
		return utilidades.ErrorNoEncontrado
	}
	return nil
}
