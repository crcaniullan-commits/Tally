package model

import (
	"github.com/google/uuid"
)

type Empresa struct {
	ID            uuid.UUID `json:"id"`
	NombreEmpresa string    `json:"nombre_empresa"`
	UserID        uuid.UUID `json:"user_id"`
}
