package model

import (
	"github.com/google/uuid"
)

type Categorie struct {
	ID     uuid.UUID  `json:"id"`
	Nombre string     `json:"nombre"`
	UserID *uuid.UUID `json:"user_id"`
}
