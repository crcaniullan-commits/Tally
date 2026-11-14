package excel

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

const xlsxType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"

var ErrNotFound = errors.New("excel: archivo no econtrado")

type File struct {
	ID         uuid.UUID `json:"id"`
	EmpresaID  uuid.UUID `json:"empresa_id"`
	Nombre     string    `json:"nombre"`
	SizeBytes  int64     `json:"size_bytes"`
	CreatedAt  time.Time `json:"created_at"`
	StorageKey string    `json:"-"`
}
