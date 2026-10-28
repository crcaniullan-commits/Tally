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

func (s *StoreCategorie) GetCategoriaPorNombreYEmpresa(ctx context.Context, categoria model.Categorie) (bool, error) {
	query := `
		SELECT id, nombre, empresa_id
		FROM categories
		WHERE empresa_id = $1 and nombre = $2;
	`

	ctx, cancel := context.WithTimeout(ctx, util.QueryTimeoutDuration)
	defer cancel()

	var exist model.Categorie

	err := s.db.QueryRowContext(ctx, query, categoria.EmpresaID, categoria.Nombre).Scan(
		&exist.ID,
		&exist.Nombre,
		&exist.EmpresaID,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}

	if err != nil {
		return true, err
	}

	return true, AlreadyExistError
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
		WHERE id = $2 and $3
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
		WHERE id = $1 and $2
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
