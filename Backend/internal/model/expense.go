package model

import (
	"time"

	"github.com/google/uuid"
)

type Expense struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	Monto       int64     `json:"monto"`
	CategoryID  uuid.UUID `json:"category_id"`
	Descripcion *string   `json:"descripcion"`
	Fecha       time.Time `json:"fecha"`
	CreatedAt   time.Time `json:"created_at"`
}
