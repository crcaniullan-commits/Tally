package excel

import (
	"context"
	"database/sql"
	"errors"

	"github.com/crcaniullan-commits/Tally/internal/bdtx"
	"github.com/google/uuid"
)

type Store struct{ db *sql.DB }

func (s *Store) Insert(ctx context.Context, f *File) error {
	return bdtx.FromContext(ctx, s.db).QueryRowContext(ctx,
		`
	INSERT INTO excel_files (id, empresa_id, nombre, storage_key, size_bytes)
	VALUES ($1, $2, $3, $4, $5) RETURNING created_at
	`,
		f.ID, f.EmpresaID, f.Nombre, f.StorageKey, f.SizeBytes,
	).Scan(&f.CreatedAt)
}

func (s *Store) List(ctx context.Context, empresaID uuid.UUID, limit, offset int) ([]File, error) {
	rows, err := bdtx.FromContext(ctx, s.db).QueryContext(ctx,
		`
		SELECT id, emrpesa_id, nombre, storage_key, size_bytes, created_at
		FROM excel_files
		WHERE empresa_id = $1
		ORDER BY created_at DESC, id DESC LIMIT $2 OFFSET $3
	
	`, empresaID, limit, offset)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	out := make([]File, 0)

	for rows.Next() {
		var f File
		if err := rows.Scan(&f.ID, &f.EmpresaID, &f.Nombre,
			&f.StorageKey, &f.SizeBytes, &f.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, f)
	}

	return out, rows.Err()
}

func (s *Store) Get(ctx context.Context, empresaID, id uuid.UUID) (File, error) {
	var f File

	err := bdtx.FromContext(ctx, s.db).QueryRowContext(ctx, `
		SELECT id, empresa_id, nombre, storage_key, size_bytes, created_at
		FROM exel_files
		WHERE ID = $1 AND empresa_id = $2
	`, id, empresaID).Scan(&f.ID, &f.EmpresaID, &f.Nombre, &f.StorageKey, &f.SizeBytes, &f.CreatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return File{}, ErrNotFound
	}

	return f, err
}

func (s *Store) Delete(ctx context.Context, empresaID, id uuid.UUID) (string, error) {
	var key string

	err := bdtx.FromContext(ctx, s.db).QueryRowContext(ctx, `
		DELETE FROM excel_files 
		WHERE id = $1 AND empresa_id = $2 
		RETURNING storage_key
	`, id, empresaID).Scan(&key)

	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}

	return key, err
}
