package excel

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/crcaniullan-commits/Tally/internal/storage"
	"github.com/google/uuid"
)

type fileStore interface {
	Insert(context.Context, *File) error
	List(context.Context, uuid.UUID, int, int) ([]File, error)
	Get(ctx context.Context, empresaID, id uuid.UUID) (File, error)
	Delete(ctx context.Context, empresaID, id uuid.UUID) (string, error)
}

type Service struct {
	store   fileStore
	storage FileStorage
}

func (s *Service) Save(ctx context.Context, userID, empresaID uuid.UUID,
	nombre string, body io.ReadSeeker, size int64) (*File, error) {

	id := uuid.New()
	f := &File{
		ID:         id,
		EmpresaID:  empresaID,
		Nombre:     nombre,
		SizeBytes:  size,
		StorageKey: fmt.Sprintf("%s/%s/%s.xlsx", userID, empresaID, id),
	}

	// 1) Primero el objeto, 2) luego el registro
	if err := s.storage.Put(ctx, f.StorageKey, body, size, xlsxType); err != nil {
		return nil, fmt.Errorf("guardando archivo: %w", err)
	}

	if err := s.store.Insert(ctx, f); err != nil {
		_ = s.storage.Delete(context.WithoutCancel(ctx), f.StorageKey) //compensar

		return nil, err
	}
	return f, nil
}

func (s *Service) List(ctx context.Context, empresaID uuid.UUID, limit, offset int) ([]File, error) {
	return s.store.List(ctx, empresaID, limit, offset)
}

func (s *Service) Open(ctx context.Context, empresaID, id uuid.UUID) (*File, io.ReadCloser, error) {
	f, err := s.store.Get(ctx, empresaID, id)

	if err != nil {
		return nil, nil, err
	}

	rc, err := s.storage.Get(ctx, f.StorageKey)

	if errors.Is(err, storage.ErrNotFound) {
		return nil, nil, ErrNotFound
	}

	if err != nil {
		return nil, nil, err
	}

	return &f, rc, nil
}

func (s *Service) Delete(ctx context.Context, empresaID, id uuid.UUID) error {
	key, err := s.store.Delete(ctx, empresaID, id)
	if err != nil {
		return nil
	}

	_ = s.storage.Delete(ctx, key) //si falla queda un objeto huérfano, sin registro
	return nil
}
