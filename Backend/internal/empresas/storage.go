package empresas

import (
	"context"
	"database/sql"

	"github.com/crcaniullan-commits/Tally/internal/dbtx"
	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/internal/util"
)

type StoreEmpresa struct {
	db *sql.DB
}

func NewStorage(db *sql.DB) *StoreEmpresa {
	return &StoreEmpresa{db}
}

func (s StoreEmpresa) Create(ctx context.Context, empresa *model.Empresa) error {
	query := `
		INSERT INTO empresas (nombre_empresa, user_id)
		VALUES ($1, $2) RETURNING id
	`

	exec := dbtx.FromContext(ctx, s.db)

	ctx, cancel := context.WithTimeout(ctx, util.QueryTimeoutDuration)
	defer cancel()

	row := exec.QueryRowContext(ctx, query, empresa.NombreEmpresa, empresa.UserID)

	if err := row.Scan(&empresa.ID); err != nil {
		return err
	}

	return nil
}
