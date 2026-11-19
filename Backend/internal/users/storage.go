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
	// password_hash viene en el SELECT a propósito: el middleware de
	// autenticación usa este mismo método para armar el usuario del contexto, y
	// PATCH /usuarios reescribe la fila con ese usuario. Si el hash no viniera,
	// la actualización escribiría un hash vacío en una columna NOT NULL.
	query := `
		SELECT id, email, nombre, rut, role, plan_expires_at, password_hash
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
		&user.PasswordHash.Hash,
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

// Actualizar aplica solo los campos indicados en cambios.
//
// El store NO reescribe la fila completa: el CASE WHEN deja intacto el campo
// que no viene en el payload. Importa porque el usuario del contexto puede
// venir del cache (ver internal/cache), y en ese caso PasswordHash llega
// vacío (el cache serializa a JSON y ese campo tiene json:"-"). Si el UPDATE
// escribiera siempre el hash, un PATCH que solo cambia el nombre mandaría un
// password_hash vacío a una columna NOT NULL y Postgres lo rechazaría.
func (s *AlmacenUsuarios) Actualizar(ctx context.Context, cambios CambiosUsuario, userID uuid.UUID) error {
	query := `
		UPDATE users
		SET password_hash = CASE WHEN $1::boolean THEN $2::text ELSE password_hash END,
			nombre = CASE WHEN $3::boolean THEN $4::text ELSE nombre END,
			updated_at = NOW()
		WHERE id = $5
	`

	ctx, cancel := context.WithTimeout(ctx, utilidades.DuracionTimeoutConsulta)
	defer cancel()

	var nombre *string
	if cambios.Nombre != nil {
		nombre = cambios.Nombre
	}

	res, err := s.db.ExecContext(ctx, query,
		cambios.CambiaPassword,
		cambios.PasswordHash,
		cambios.Nombre != nil,
		nombre,
		userID,
	)

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
