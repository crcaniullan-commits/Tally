package expenses

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/internal/paginacion"
	"github.com/crcaniullan-commits/Tally/internal/utilidades"
	"github.com/google/uuid"
)

type AlmacenGastos struct {
	db *sql.DB
}

func NuevoAlmacenGastos(db *sql.DB) *AlmacenGastos {
	return &AlmacenGastos{db: db}
}

func (s *AlmacenGastos) AgregarGasto(ctx context.Context, expense *model.Expense) error {
	query := `
		INSERT INTO expenses (empresa_id, monto, category_id, descripcion)
		VALUES ($1, $2, $3, $4)
		RETURNING id, fecha, created_at
	`

	ctx, cancel := context.WithTimeout(ctx, utilidades.DuracionTimeoutConsulta)
	defer cancel()

	var fecha sql.NullTime
	var createdAt sql.NullTime

	err := s.db.QueryRowContext(ctx, query, expense.EmpresaID, expense.Monto, expense.CategoryID, expense.Descripcion).Scan(
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

// ActualizarGasto modifica los campos indicados en cambios y devuelve el gasto
// ya actualizado.
//
// El WHERE lleva empresa_id: un gasto de otra empresa es indistinguible de uno
// inexistente (404), igual que en el borrado.
func (s *AlmacenGastos) ActualizarGasto(ctx context.Context, cambios CambiosGasto, gastoID uuid.UUID, empresaID uuid.UUID) (model.Expense, error) {
	query := `
		UPDATE expenses
		SET monto = COALESCE($1::bigint, monto),
			category_id = COALESCE($2::uuid, category_id),
			descripcion = CASE WHEN $3::boolean THEN $4::text ELSE descripcion END,
			fecha = COALESCE($5::date, fecha)
		WHERE id = $6 AND empresa_id = $7
		RETURNING id, empresa_id, monto, category_id, descripcion, fecha, created_at
	`

	ctx, cancel := context.WithTimeout(ctx, utilidades.DuracionTimeoutConsulta)
	defer cancel()

	var (
		monto       *int64
		categoriaID *uuid.UUID
		descripcion *string
		fecha       *string
	)

	if cambios.Monto != nil {
		monto = cambios.Monto
	}

	if cambios.CategoryID != nil {
		categoriaID = cambios.CategoryID
	}

	descripcion = cambios.Descripcion

	if cambios.Fecha != nil {
		// La fecha viaja como texto AAAA-MM-DD para no depender de la zona
		// horaria de la sesión al castearla a date.
		valor := cambios.Fecha.Format(time.DateOnly)
		fecha = &valor
	}

	expense := model.Expense{}

	err := s.db.QueryRowContext(ctx, query,
		monto,
		categoriaID,
		cambios.Descripcion != nil,
		descripcion,
		fecha,
		gastoID,
		empresaID,
	).Scan(
		&expense.ID,
		&expense.EmpresaID,
		&expense.Monto,
		&expense.CategoryID,
		&expense.Descripcion,
		&expense.Fecha,
		&expense.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Expense{}, utilidades.ErrorNoEncontrado
		}

		return model.Expense{}, err
	}

	return expense, nil
}

func (s *AlmacenGastos) EliminarGasto(ctx context.Context, gastoID uuid.UUID, empresaID uuid.UUID) error {
	query := `
		DELETE FROM expenses
		WHERE id = $1 and empresa_id = $2
	`

	ctx, cancel := context.WithTimeout(ctx, utilidades.DuracionTimeoutConsulta)
	defer cancel()

	res, err := s.db.ExecContext(ctx, query, gastoID, empresaID)

	if err != nil {
		return err
	}

	row, err := res.RowsAffected()

	if err != nil {
		return err
	}

	if row == 0 {
		return utilidades.ErrorNoEncontrado
	}

	return nil
}

func (s *AlmacenGastos) ObtenerGastosDeEmpresa(ctx context.Context, empresaID uuid.UUID, fq paginacion.ConsultaPaginacionIngresos) ([]model.Expense, error) {
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

	ctx, cancel := context.WithTimeout(ctx, utilidades.DuracionTimeoutConsulta)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query, empresaID, fq.Limite, fq.Desplazamiento, fq.Desde, fq.Hasta, fq.CategoryID)

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

// CategoriaExiste verifica que la categoría exista y sea visible para la
// empresa: global (empresa_id IS NULL) o propia. Una categoría de otra empresa
// no pasa el filtro y el service la trata como inexistente.
func (s *AlmacenGastos) CategoriaExiste(ctx context.Context, categoryID uuid.UUID, empresaID uuid.UUID) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM categories
			WHERE id = $1 AND (empresa_id IS NULL OR empresa_id = $2)
		)
	`

	ctx, cancel := context.WithTimeout(ctx, utilidades.DuracionTimeoutConsulta)
	defer cancel()

	var existe bool

	if err := s.db.QueryRowContext(ctx, query, categoryID, empresaID).Scan(&existe); err != nil {
		return false, err
	}

	return existe, nil
}
