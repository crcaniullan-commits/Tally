package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type DiskStorage struct{ base string }

func NewDisk(base string) *DiskStorage {
	return &DiskStorage{base: filepath.Clean(base)}
}

// path convierte una clave en ruta y evita que escape de la carpeta base.
func (d *DiskStorage) path(key string) (string, error) {
	p := filepath.Join(d.base, filepath.FromSlash(key))

	if !strings.HasPrefix(p, d.base+string(os.PathSeparator)) {
		return "", fmt.Errorf("Storage: clave inválida %q", key)
	}

	return p, nil
}

func (d *DiskStorage) Put(ctx context.Context, key string, body io.ReadSeeker, size int64, contentType string) error {
	p, err := d.path(key)

	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}

	f, err := os.Create(p)
	if err != nil {
		return err
	}

	defer f.Close()

	_, err = io.Copy(f, body)
	return err
}

func (d *DiskStorage) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	p, err := d.path(key)

	if err != nil {
		return nil, err
	}

	f, err := os.Open(p)

	if errors.Is(err, os.ErrNotExist) {
		return nil, ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	return f, nil
}

func (d *DiskStorage) Delete(ctx context.Context, key string) error {
	p, err := d.path(key)

	if err != nil {
		return err
	}

	err = os.Remove(p)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}

	return err
}
