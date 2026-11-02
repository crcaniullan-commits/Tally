package categories

import (
	"context"
	"database/sql"
	"errors"

	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/internal/util"
	"github.com/google/uuid"
)

type StoreCategorie struct {
	db *sql.DB
}

func NewStoreCategorie(db *sql.DB) *StoreCategorie {
	return &StoreCategorie{db}
}

func (s *StoreCategorie) GetCategoriasPorEmpresa(ctx context.Context, empresaID uuid.UUID) ([]model.Categorie, error) {
	query := `
		SELECT id, nombre, empresa_id
		FROM categories
		WHERE empresa_id = $1
		ORDER BY nombre;
	`

	ctx, cancel := context.WithTimeout(ctx, util.QueryTimeoutDuration)
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

func (s *StoreCategorie) GetCategoriasVisibles(ctx context.Context, empresaID uuid.UUID) ([]model.Categorie, error) {
	query := `
		SELECT id, nombre, empresa_id
		FROM categories
		WHERE empresa_id IS NULL OR empresa_id = $1
		ORDER BY empresa_id NULLS LAST, nombre;
	`

	ctx, cancel := context.WithTimeout(ctx, util.QueryTimeoutDuration)
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

func (s *StoreCategorie) Create(ctx context.Context, categoria *model.Categorie) error {
	query := `
		INSERT INTO categories (nombre, empresa_id)
		VALUES ($1, $2) RETURNING id
	`
	ctx, cancel := context.WithTimeout(ctx, util.QueryTimeoutDuration)
	defer cancel()

	err := s.db.QueryRowContext(ctx, query, categoria.Nombre, categoria.EmpresaID).Scan(
		&categoria.ID,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return util.ErrNotFound
		}

		return err
	}

	return nil
}

func (s *StoreCategorie) CambiarNombre(ctx context.Context, nombre string, categoriaID uuid.UUID, empresaID uuid.UUID) error {
	query := `
		UPDATE categories
		SET nombre = $1
		WHERE id = $2 and empresa_id = $3
	`

	ctx, cancel := context.WithTimeout(ctx, util.QueryTimeoutDuration)
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
		return util.ErrNotFound
	}

	return nil
}

func (s *StoreCategorie) BorrarCategoria(ctx context.Context, categoriaID uuid.UUID, empresaID uuid.UUID) error {
	query := `
		DELETE FROM categories
		WHERE id = $1 and empresa_id = $2
	`

	ctx, cancel := context.WithTimeout(ctx, util.QueryTimeoutDuration)
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
		return util.ErrNotFound
	}

	return nil
}
