package incomes

import (
	"database/sql"
	"time"

	"github.com/crcaniullan-commits/Tally/internal/util"
	"github.com/google/uuid"
)

type IncomeStorage struct {
	ID            uuid.UUID
	UserID        uuid.UUID
	Monto         float64
	PaymentMethod util.PaymentMethod
	Descripcion   *string
	Fecha         time.Time
	CreatedAt     time.Time
}

type StoreIncome struct {
	db *sql.DB
}
