package categories

import (
	"context"
	"database/sql"
	"errors"

	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/internal/utilidades"
	"github.com/google/uuid"
)

type AlmacenCategorias struct {
	db *sql.DB
}

func NuevoAlmacenCategorias(db *sql.DB) *AlmacenCategorias {
	return &AlmacenCategorias{db}
}

func (s *AlmacenCategorias) ObtenerCategoriasPorEmpresa(ctx context.Context, empresaID uuid.UUID) ([]model.Categorie, error) {
	query := `
		SELECT id, nombre, empresa_id
		FROM categories
		WHERE empresa_id = $1
		ORDER BY nombre;
	`

	ctx, cancel := context.WithTimeout(ctx, utilidades.DuracionTimeoutConsulta)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query, empresaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categorias []model.Categorie

	for rows.Next() {
		var categoria model.Categorie

		if err := rows.Scan(&categoria.ID, &categoria.Nombre, &categoria.EmpresaID); err != nil {
			return nil, err
		}

		categorias = append(categorias, categoria)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return categorias, nil
}

func (s *AlmacenCategorias) ObtenerCategoriasVisibles(ctx context.Context, empresaID uuid.UUID) ([]model.Categorie, error) {
	query := `
		SELECT id, nombre, empresa_id
		FROM categories
		WHERE empresa_id IS NULL OR empresa_id = $1
		ORDER BY empresa_id NULLS LAST, nombre;
	`

	ctx, cancel := context.WithTimeout(ctx, utilidades.DuracionTimeoutConsulta)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query, empresaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categorias []model.Categorie

	for rows.Next() {
		var categoria model.Categorie

		if err := rows.Scan(&categoria.ID, &categoria.Nombre, &categoria.EmpresaID); err != nil {
			return nil, err
		}

		categorias = append(categorias, categoria)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return categorias, nil
}

func (s *AlmacenCategorias) Crear(ctx context.Context, categoria *model.Categorie) error {
	query := `
		INSERT INTO categories (nombre, empresa_id)
		VALUES ($1, $2) RETURNING id
	`
	ctx, cancel := context.WithTimeout(ctx, utilidades.DuracionTimeoutConsulta)
	defer cancel()

	err := s.db.QueryRowContext(ctx, query, categoria.Nombre, categoria.EmpresaID).Scan(
		&categoria.ID,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return utilidades.ErrorNoEncontrado
		}

		return err
	}

	return nil
}

func (s *AlmacenCategorias) CambiarNombre(ctx context.Context, nombre string, categoriaID uuid.UUID, empresaID uuid.UUID) error {
	query := `
		UPDATE categories
		SET nombre = $1
		WHERE id = $2 and empresa_id = $3
	`

	ctx, cancel := context.WithTimeout(ctx, utilidades.DuracionTimeoutConsulta)
	defer cancel()

	res, err := s.db.ExecContext(ctx, query, nombre, categoriaID, empresaID)

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

func (s *AlmacenCategorias) BorrarCategoria(ctx context.Context, categoriaID uuid.UUID, empresaID uuid.UUID) error {
	query := `
		DELETE FROM categories
		WHERE id = $1 and empresa_id = $2
	`

	ctx, cancel := context.WithTimeout(ctx, utilidades.DuracionTimeoutConsulta)
	defer cancel()

	res, err := s.db.ExecContext(ctx, query, categoriaID, empresaID)

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
