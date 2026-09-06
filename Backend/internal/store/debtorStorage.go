package store

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type DebtorStorage struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Nombre      string
	Monto       float64
	Descripcion *string
	Pagado      bool
	FechaLimite *time.Time
	CreatedAt   time.Time
}

type StoreDebtor struct {
	db *sql.DB
}