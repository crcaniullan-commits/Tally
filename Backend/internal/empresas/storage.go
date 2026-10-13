package empresas

import (
	"database/sql"
)

type StoreEmpresa struct {
	db *sql.DB
}

func NewStorage(db *sql.DB) *StoreEmpresa {
	return &StoreEmpresa{db}
}
