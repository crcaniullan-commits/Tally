package goals

import (
	"database/sql"
	"time"

	"github.com/crcaniullan-commits/Tally/internal/util"
	"github.com/google/uuid"
)

type GoalStorage struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Nombre    string
	Periodo   util.GoalPeriod
	MontoMeta float64
	CreatedAt time.Time
}

type StoreGoal struct {
	db *sql.DB
}