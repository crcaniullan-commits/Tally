package model

import (
	"time"

	"github.com/google/uuid"
)

type Income struct {
	ID            uuid.UUID     `json:"id"`
	UserID        uuid.UUID     `json:"user_id"`
	Monto         int64         `json:"monto"`
	PaymentMethod PaymentMethod `json:"payment_method"`
	Descripcion   *string       `json:"descripcion"`
	Fecha         time.Time     `json:"fecha"`
	CreatedAt     time.Time     `json:"created_at"`
	CategoryID    *uuid.UUID    `json:"category_id"`
}
