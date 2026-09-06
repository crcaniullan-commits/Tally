package store

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type ExpenseStorage struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Monto       float64
	CategoryID  uuid.UUID
	Descripcion *string
	Fecha       time.Time
	CreatedAt   time.Time
}

type StoreExpense struct {
	db *sql.DB
}