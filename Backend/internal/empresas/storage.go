package empresas

import (
	"context"
	"database/sql"
	"errors"

	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/internal/utilidades"
	"github.com/google/uuid"
)

type AlmacenEmpresas struct {
	db *sql.DB
}

func NuevoAlmacenEmpresas(db *sql.DB) *AlmacenEmpresas {
	return &AlmacenEmpresas{db}
}

func (s AlmacenEmpresas) ObtenerPorID(ctx context.Context, empresaID uuid.UUID, userID uuid.UUID) (*model.Empresa, error) {
	query := `
		SELECT * FROM empresas
		WHERE id = $1 and user_id = $2
	`

	ctx, cancel := context.WithTimeout(ctx, utilidades.DuracionTimeoutConsulta)
	defer cancel()

	var empresa model.Empresa

	err := s.db.QueryRowContext(ctx, query, empresaID, userID).Scan(
		&empresa.ID,
		&empresa.NombreEmpresa,
		&empresa.UserID,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, utilidades.ErrorNoEncontrado
		}

		return nil, err
	}

	return &empresa, nil

}

func (s *AlmacenEmpresas) ObtenerEmpresas(ctx context.Context, userID uuid.UUID) ([]model.Empresa, error) {
	query := `
		SELECT id, nombre_empresa, user_id
		FROM empresas
		WHERE user_id = $1
	`

	ctx, cancel := context.WithTimeout(ctx, utilidades.DuracionTimeoutConsulta)
	defer cancel()

	row, err := s.db.QueryContext(ctx, query, userID)

	if err != nil {
		return nil, err
	}

	defer row.Close()

	var empresas []model.Empresa

	for row.Next() {
		var empresa model.Empresa

		if err := row.Scan(&empresa.ID, &empresa.NombreEmpresa, &empresa.UserID); err != nil {
			return nil, err
		}
		empresas = append(empresas, empresa)
	}

	if err := row.Err(); err != nil {
		return nil, err
	}

	return empresas, nil
}

func (s AlmacenEmpresas) Crear(ctx context.Context, empresa *model.Empresa) error {
	query := `
		INSERT INTO empresas (nombre_empresa, user_id)
		VALUES ($1, $2) RETURNING id
	`

	ctx, cancel := context.WithTimeout(ctx, utilidades.DuracionTimeoutConsulta)
	defer cancel()

	err := s.db.QueryRowContext(ctx, query, empresa.NombreEmpresa, empresa.UserID).Scan(&empresa.ID)

	if err != nil {
		return err
	}

	return nil
}

func (s AlmacenEmpresas) CambiarNombre(ctx context.Context, nombre string, empresaID uuid.UUID, userID uuid.UUID) error {
	query := `
		UPDATE empresas
		SET nombre_empresa = $1
		WHERE id = $2 and user_id = $3
	`

	ctx, cancel := context.WithTimeout(ctx, utilidades.DuracionTimeoutConsulta)
	defer cancel()

	res, err := s.db.ExecContext(ctx, query, nombre, empresaID, userID)

	if err != nil {
		return err
	}

	row, err := res.RowsAffected()

	if err != nil {
		return err
	}

	if row == 0 {
		return utilidades.ErrorNoEncontrado
	}

	return nil
}

func (s AlmacenEmpresas) BorrarEmpresa(ctx context.Context, empresaID uuid.UUID, userID uuid.UUID) error {
	query := `
		DELETE FROM empresas
		WHERE id = $1 and user_id = $2
	`

	ctx, cancel := context.WithTimeout(ctx, utilidades.DuracionTimeoutConsulta)
	defer cancel()

	res, err := s.db.ExecContext(ctx, query, empresaID, userID)

	if err != nil {
		return err
	}

	row, err := res.RowsAffected()

	if err != nil {
		return err
	}

	if row == 0 {
		return utilidades.ErrorNoEncontrado
	}

	return nil
}
