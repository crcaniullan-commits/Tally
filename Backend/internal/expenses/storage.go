package expenses

import (
	"context"
	"database/sql"

	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/internal/pagination"
	"github.com/crcaniullan-commits/Tally/internal/util"
	"github.com/google/uuid"
)

type StoreExpense struct {
	db *sql.DB
}

func NewStorage(db *sql.DB) *StoreExpense {
	return &StoreExpense{db: db}
}

func (s *StoreExpense) AddExpense(ctx context.Context, expense *model.Expense) error {
	query := `
		INSERT INTO expenses (empresa_id, monto, category_id, descripcion, fecha)
		VALUES ($1, $2, $3, $4, COALESCE(NULLIF($5, '0001-01-01'::date), CURRENT_DATE))
		RETURNING id, fecha, created_at
	`

	ctx, cancel := context.WithTimeout(ctx, util.QueryTimeoutDuration)
	defer cancel()

	var fecha sql.NullTime
	var createdAt sql.NullTime

	err := s.db.QueryRowContext(ctx, query, expense.EmpresaID, expense.Monto, expense.CategoryID, expense.Descripcion, expense.Fecha).Scan(
		&expense.ID,
		&fecha,
		&createdAt,
	)

	if err != nil {
		return err
	}

	if fecha.Valid {
		expense.Fecha = fecha.Time
	}
	if createdAt.Valid {
		expense.CreatedAt = createdAt.Time
	}

	return nil
}

func (s *StoreExpense) DeleteExpense(ctx context.Context, expenseID uuid.UUID, empresaID uuid.UUID) error {
	query := `
		DELETE FROM expenses
		WHERE id = $1 and empresa_id = $2
	`

	ctx, cancel := context.WithTimeout(ctx, util.QueryTimeoutDuration)
	defer cancel()

	res, err := s.db.ExecContext(ctx, query, expenseID, empresaID)

	if err != nil {
		return err
	}

	row, err := res.RowsAffected()

	if err != nil {
		return err
	}

	if row == 0 {
		return util.ErrNotFound
	}

	return nil
}

func (s *StoreExpense) GetAllExpensesOfEmpresa(ctx context.Context, empresaID uuid.UUID, fq pagination.IncomePaginationQuery) ([]model.Expense, error) {
	query := `
		SELECT id, empresa_id, monto, category_id, descripcion, fecha, created_at
		FROM expenses
		WHERE empresa_id = $1
		  AND (NULLIF($4, '')::date IS NULL OR fecha >= NULLIF($4, '')::date)
		  AND (NULLIF($5, '')::date IS NULL OR fecha <= NULLIF($5, '')::date)
		  AND (NULLIF($6, '')::uuid IS NULL OR category_id = NULLIF($6, '')::uuid)
		ORDER BY fecha DESC, created_at DESC, id DESC
		LIMIT $2 OFFSET $3
	`

	ctx, cancel := context.WithTimeout(ctx, util.QueryTimeoutDuration)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query, empresaID, fq.Limit, fq.Offset, fq.Since, fq.Until, fq.CategoryID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	expenses := []model.Expense{}

	for rows.Next() {
		var expense model.Expense

		if err := rows.Scan(
			&expense.ID,
			&expense.EmpresaID,
			&expense.Monto,
			&expense.CategoryID,
			&expense.Descripcion,
			&expense.Fecha,
			&expense.CreatedAt,
		); err != nil {
			return nil, err
		}

		expenses = append(expenses, expense)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return expenses, nil
}

// GetCategoryIDByName resuelve el nombre de una categoría al id con el que se
// referencia en expenses. Acepta las categorías globales (empresa_id IS NULL) y
// las propias de la empresa; ante un nombre repetido prefiere la de la empresa,
// por eso el NULLS LAST. La comparación usa util.NormalizarNombre, así que
// "Comida", "comida" y "COMIDA" matchean la misma categoría. Sin fila devuelve
// util.ErrNotFound para que el handler lo traduzca a 400 y no a un 500.
func (s *StoreExpense) GetCategoryIDByName(ctx context.Context, nombre string, empresaID uuid.UUID) (uuid.UUID, error) {
	query := `
		SELECT id, nombre
		FROM categories
		WHERE (empresa_id IS NULL OR empresa_id = $1)
		ORDER BY empresa_id NULLS LAST
	`

	ctx, cancel := context.WithTimeout(ctx, util.QueryTimeoutDuration)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query, empresaID)
	if err != nil {
		return uuid.Nil, err
	}
	defer rows.Close()

	buscado := util.NormalizarNombre(nombre)

	for rows.Next() {
		var (
			categoryID uuid.UUID
			categoria  string
		)

		if err := rows.Scan(&categoryID, &categoria); err != nil {
			return uuid.Nil, err
		}

		if util.NormalizarNombre(categoria) == buscado {
			return categoryID, nil
		}
	}

	if err := rows.Err(); err != nil {
		return uuid.Nil, err
	}

	return uuid.Nil, util.ErrNotFound
}
