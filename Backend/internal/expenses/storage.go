package expenses

import (
	"database/sql"
)

type StoreExpense struct {
	db *sql.DB
}
