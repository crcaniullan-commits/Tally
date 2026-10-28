package model

import (
	"time"

	"github.com/google/uuid"
)

type Goal struct {
	ID        uuid.UUID  `json:"id"`
	EmpresaID uuid.UUID  `json:"empresa_id"`
	Nombre    string     `json:"nombre"`
	Periodo   GoalPeriod `json:"periodo"`
	MontoMeta int64      `json:"monto_meta"`
	CreatedAt time.Time  `json:"created_at"`
}
