package clavesacceso

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/crcaniullan-commits/Tally/internal/bdtx"
	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/internal/utilidades"
	"github.com/google/uuid"
)

type AlmacenClavesAcceso struct {
	db *sql.DB
}

func NuevoAlmacenClavesAcceso(db *sql.DB) *AlmacenClavesAcceso {
	return &AlmacenClavesAcceso{db}
}

type DatosCodigoRevocacion struct {
	RevocadoPor uuid.UUID
	code        string
}

func (s *AlmacenClavesAcceso) Crear(ctx context.Context, AccessKey *model.AccessKey) error {
	query := `
		INSERT INTO access_keys (code, created_by, expires_at)
		VALUES ($1, $2, $3) RETURNING code, id
	`
	ctx, cancel := context.WithTimeout(ctx, utilidades.DuracionTimeoutConsulta)
	defer cancel()

	err := s.db.QueryRowContext(ctx, query,
		AccessKey.Code,
		AccessKey.CreatedBy,
		AccessKey.ExpiresAt,
	).Scan(
		&AccessKey.Code,
		&AccessKey.ID,
	)

	if err != nil {
		return err
	}

	return nil
}

func (s *AlmacenClavesAcceso) ObtenerPorID(ctx context.Context, accessKeyID uuid.UUID) (model.AccessKey, error) {
	query := `
		SELECT id, code, created_by, redeemed_by,
		       redeemed_at, expires_at, revoked_at, revoked_by, created_at
		FROM access_keys
		WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(ctx, utilidades.DuracionTimeoutConsulta)
	defer cancel()

	AccessKey := model.AccessKey{}

	err := s.db.QueryRowContext(ctx, query, accessKeyID).Scan(
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
		if errors.Is(err, sql.ErrNoRows) {
			return model.AccessKey{}, utilidades.ErrorNoEncontrado
		}
		return model.AccessKey{}, err
	}

	return AccessKey, nil
}

func (s *AlmacenClavesAcceso) Canjear(ctx context.Context, code string, userID uuid.UUID) (*model.AccessKey, error) {
	query := `
		UPDATE access_keys
		SET redeemed_by = $1, redeemed_at = now()
		WHERE code = $2 AND redeemed_at IS NULL
			  AND expires_at > now() AND revoked_at IS NULL
		RETURNING id, expires_at
	`

	exec := bdtx.FromContext(ctx, s.db)

	ctx, cancel := context.WithTimeout(ctx, utilidades.DuracionTimeoutConsulta)
	defer cancel()

	row := exec.QueryRowContext(ctx, query, userID, code)

	var ak model.AccessKey

	if err := row.Scan(&ak.ID, &ak.ExpiresAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrorCodigoNoCanjeable
		}
		return nil, fmt.Errorf("canjeando la clave de acceso: %w", err)
	}

	return &ak, nil
}

func (s *AlmacenClavesAcceso) Revocar(ctx context.Context, revokeData DatosCodigoRevocacion) error {
	query := `
		UPDATE access_keys
		SET revoked_by = $1, revoked_at = now()
		WHERE code = $2
	`

	ctx, cancel := context.WithTimeout(ctx, utilidades.DuracionTimeoutConsulta)
	defer cancel()

	res, err := s.db.ExecContext(ctx, query, revokeData.RevocadoPor, revokeData.code)

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
