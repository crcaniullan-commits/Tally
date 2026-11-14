package excel

import (
	"context"
	"io"
)

type FileStorage interface {
	Put(ctx context.Context, key string, body io.ReadSeeker, size int64, contentType string) error
	Get(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
}
