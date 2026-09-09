package categories

import (
	"database/sql"

	"github.com/google/uuid"
)

type CategorieStorage struct {
	ID     uuid.UUID
	Nombre string
	UserID *uuid.UUID
}

type StoreCategorie struct {
	db *sql.DB
}
