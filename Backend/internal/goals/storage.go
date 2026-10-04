package goals

import (
	"database/sql"
)

type StoreGoal struct {
	db *sql.DB
}
