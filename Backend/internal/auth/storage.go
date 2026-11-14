package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/crcaniullan-commits/Tally/internal/bdtx"
	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/internal/users"
	"github.com/crcaniullan-commits/Tally/internal/utilidades"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type AlmacenAuth struct {
	db *sql.DB
}

func NuevoAlmacenAuth(db *sql.DB) *AlmacenAuth {
	return &AlmacenAuth{db}
}

func (s *AlmacenAuth) ObtenerPorEmail(ctx context.Context, email string) (*model.User, error) {
	query := `
		SELECT id, email, nombre, rut, password_hash
		FROM users
		WHERE email = $1;
	`

	ctx, cancel := context.WithTimeout(ctx, utilidades.DuracionTimeoutConsulta)
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

func (s *AlmacenAuth) Create(ctx context.Context, user *model.User) error {
	query := `
		INSERT INTO users (email, password_hash, nombre, role, rut)
		VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at;
	`

	ctx, cancel := context.WithTimeout(ctx, utilidades.DuracionTimeoutConsulta)
	defer cancel()

	role := user.Role
	if role == "" {
		role = "usuario"
	}

	err := bdtx.FromContext(ctx, s.db).QueryRowContext(ctx, query,
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
		if esDuplicateKey(err, "users_email_key") {
			return users.ErrDuplicateEmail
		} else if esDuplicateKey(err, "users_rut_key") {
			return users.ErrDuplicateRut
		}
		return err
	}
	return nil
}

// esDuplicateKey dice si el error es una violación de unicidad sobre la
// constraint indicada.
//
// No se puede comparar el texto del error: lib/pq devuelve
// `pq: duplicate key value violates unique constraint "users_email_key" (23505)`
// y el sufijo del código SQL rompe cualquier comparación exacta. Por eso se
// mira el error tipado de pq (código 23505 + nombre de la constraint) y, si no
// es de pq, se cae a un Contains del nombre de la constraint.
func esDuplicateKey(err error, constraint string) bool {
	var pqErr *pq.Error

	if errors.As(err, &pqErr) {
		return pqErr.Code == pq.ErrorCode("23505") && pqErr.Constraint == constraint
	}

	return strings.Contains(err.Error(), constraint)
}

func (s *AlmacenAuth) establecerVencimiento(ctx context.Context, expires_at time.Time, userID uuid.UUID) error {
	query := `
		UPDATE users
		SET plan_expires_at = $1
		WHERE id = $2
	`

	exec := bdtx.FromContext(ctx, s.db)

	ctx, cancel := context.WithTimeout(ctx, utilidades.DuracionTimeoutConsulta)
	defer cancel()

	_, err := exec.ExecContext(ctx, query, expires_at, userID)

	if err != nil {
		return fmt.Errorf("actualizando plan_expires_at: %w", err)
	}

	return nil
}

// CrearCodigoRecuperacion invalida los códigos previos del usuario y guarda el
// nuevo. Solo el último código emitido sirve: pedir uno nuevo deja el anterior
// inservible.
func (s *AlmacenAuth) CrearCodigoRecuperacion(ctx context.Context, codigo *model.CodigoRecuperacion) error {
	ctx, cancel := context.WithTimeout(ctx, utilidades.DuracionTimeoutConsulta)
	defer cancel()

	invalidar := `
		UPDATE password_reset_codes
		SET usado_en = now()
		WHERE user_id = $1 AND usado_en IS NULL
	`

	if _, err := s.db.ExecContext(ctx, invalidar, codigo.UserID); err != nil {
		return fmt.Errorf("invalidando los codigos de recuperacion previos: %w", err)
	}

	query := `
		INSERT INTO password_reset_codes (user_id, codigo, expira_en)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`

	err := s.db.QueryRowContext(ctx, query, codigo.UserID, codigo.Codigo, codigo.ExpiraEn).Scan(
		&codigo.ID,
		&codigo.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("creando el codigo de recuperacion: %w", err)
	}

	return nil
}

// ObtenerCodigoRecuperacion devuelve el último código emitido para el usuario.
func (s *AlmacenAuth) ObtenerCodigoRecuperacion(ctx context.Context, userID uuid.UUID) (*model.CodigoRecuperacion, error) {
	query := `
		SELECT id, user_id, codigo, expira_en, verificado_en, usado_en, created_at
		FROM password_reset_codes
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	ctx, cancel := context.WithTimeout(ctx, utilidades.DuracionTimeoutConsulta)
	defer cancel()

	codigo := &model.CodigoRecuperacion{}

	err := s.db.QueryRowContext(ctx, query, userID).Scan(
		&codigo.ID,
		&codigo.UserID,
		&codigo.Codigo,
		&codigo.ExpiraEn,
		&codigo.VerificadoEn,
		&codigo.UsadoEn,
		&codigo.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, utilidades.ErrorNoEncontrado
		}
		return nil, fmt.Errorf("obteniendo el codigo de recuperacion: %w", err)
	}

	return codigo, nil
}

// MarcarCodigoVerificado deja el código verificado por el usuario.
func (s *AlmacenAuth) MarcarCodigoVerificado(ctx context.Context, codigoID uuid.UUID) error {
	query := `
		UPDATE password_reset_codes
		SET verificado_en = now()
		WHERE id = $1 AND verificado_en IS NULL
	`

	ctx, cancel := context.WithTimeout(ctx, utilidades.DuracionTimeoutConsulta)
	defer cancel()

	if _, err := s.db.ExecContext(ctx, query, codigoID); err != nil {
		return fmt.Errorf("marcando el codigo de recuperacion como verificado: %w", err)
	}

	return nil
}

// MarcarCodigoUsado consume el código: no se puede volver a usar.
func (s *AlmacenAuth) MarcarCodigoUsado(ctx context.Context, codigoID uuid.UUID) error {
	query := `
		UPDATE password_reset_codes
		SET usado_en = now()
		WHERE id = $1 AND usado_en IS NULL
	`

	ctx, cancel := context.WithTimeout(ctx, utilidades.DuracionTimeoutConsulta)
	defer cancel()

	if _, err := s.db.ExecContext(ctx, query, codigoID); err != nil {
		return fmt.Errorf("marcando el codigo de recuperacion como usado: %w", err)
	}

	return nil
}

// ActualizarPassword cambia el hash de la contraseña del usuario.
func (s *AlmacenAuth) ActualizarPassword(ctx context.Context, userID uuid.UUID, hash []byte) error {
	query := `
		UPDATE users
		SET password_hash = $1, updated_at = now()
		WHERE id = $2
	`

	ctx, cancel := context.WithTimeout(ctx, utilidades.DuracionTimeoutConsulta)
	defer cancel()

	res, err := s.db.ExecContext(ctx, query, hash, userID)

	if err != nil {
		return fmt.Errorf("actualizando el password del usuario: %w", err)
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
