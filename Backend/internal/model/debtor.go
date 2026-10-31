package model

import (
	"time"

	"github.com/google/uuid"
)

type Debtor struct {
	ID          uuid.UUID  `json:"id"`
	EmpresaID   uuid.UUID  `json:"empresa_id"`
	Nombre      string     `json:"nombre"`
	Monto       int64      `json:"monto"`
	Descripcion *string    `json:"descripcion"`
	Pagado      bool       `json:"pagado"`
	FechaLimite *time.Time `json:"fecha_limite"`
	CreatedAt   time.Time  `json:"created_at"`
}
