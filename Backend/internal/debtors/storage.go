package debtors

import (
	"context"
	"database/sql"
	"errors"

	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/internal/util"
	"github.com/google/uuid"
)

type StoreDebtor struct {
	db *sql.DB
}

func NewStorage(db *sql.DB) *StoreDebtor {
	return &StoreDebtor{db: db}
}

// GetDebtor devuelve si la deuda está pagada. Si no existe o no pertenece a la
// empresa, devuelve util.ErrNotFound.
func (s *StoreDebtor) GetDebtor(ctx context.Context, debtorID uuid.UUID, empresaID uuid.UUID) (bool, error) {
	query := `
		SELECT pagado FROM debtors
		WHERE id = $1 and empresa_id = $2
	`

	ctx, cancel := context.WithTimeout(ctx, util.QueryTimeoutDuration)
	defer cancel()

	var pagado bool

	err := s.db.QueryRowContext(ctx, query, debtorID, empresaID).Scan(&pagado)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, util.ErrNotFound
		}
		return false, err
	}

	return pagado, nil
}

func (s *StoreDebtor) Create(ctx context.Context, debtor *model.Debtor) error {
	query := `
		INSERT INTO debtors (empresa_id, nombre, monto, descripcion, fecha_limite)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, pagado, created_at
	`

	ctx, cancel := context.WithTimeout(ctx, util.QueryTimeoutDuration)
	defer cancel()

	return s.db.QueryRowContext(ctx, query,
		debtor.EmpresaID,
		debtor.Nombre,
		debtor.Monto,
		debtor.Descripcion,
		debtor.FechaLimite,
	).Scan(&debtor.ID, &debtor.Pagado, &debtor.CreatedAt)
}

func (s *StoreDebtor) Update(ctx context.Context, debtorID uuid.UUID, empresaID uuid.UUID, debtor model.Debtor) error {
	query := `
		UPDATE debtors
		SET nombre = $1, monto = $2, fecha_limite = $3
		WHERE id = $4 and empresa_id = $5
	`

	ctx, cancel := context.WithTimeout(ctx, util.QueryTimeoutDuration)
	defer cancel()

	result, err := s.db.ExecContext(ctx, query,
		debtor.Nombre,
		debtor.Monto,
		debtor.FechaLimite,
		debtorID,
		empresaID,
	)
	if err != nil {
		return err
	}

	filas, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if filas == 0 {
		return util.ErrNotFound
	}

	return nil
}

func (s *StoreDebtor) PaidDebtor(ctx context.Context, debtorID uuid.UUID, empresaID uuid.UUID) error {
	query := `
		UPDATE debtors
		SET pagado = true
		WHERE id = $1 and empresa_id = $2
	`

	ctx, cancel := context.WithTimeout(ctx, util.QueryTimeoutDuration)
	defer cancel()

	result, err := s.db.ExecContext(ctx, query, debtorID, empresaID)
	if err != nil {
		return err
	}

	filas, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if filas == 0 {
		return util.ErrNotFound
	}

	return nil
}

func (s *StoreDebtor) GetDebtors(ctx context.Context, empresaID uuid.UUID) ([]model.Debtor, error) {
	query := `
		SELECT id, empresa_id, nombre, monto, descripcion, pagado, fecha_limite, created_at
		FROM debtors
		WHERE empresa_id = $1
		ORDER BY created_at DESC, id DESC
	`

	ctx, cancel := context.WithTimeout(ctx, util.QueryTimeoutDuration)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query, empresaID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	debtors := make([]model.Debtor, 0)

	for rows.Next() {
		var debtor model.Debtor

		err := rows.Scan(
			&debtor.ID,
			&debtor.EmpresaID,
			&debtor.Nombre,
			&debtor.Monto,
			&debtor.Descripcion,
			&debtor.Pagado,
			&debtor.FechaLimite,
			&debtor.CreatedAt,
		)

		if err != nil {
			return nil, err
		}

		debtors = append(debtors, debtor)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return debtors, nil
}
