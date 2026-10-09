package incomes

import (
	"context"
	"database/sql"

	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/internal/pagination"
	"github.com/crcaniullan-commits/Tally/internal/util"
	"github.com/google/uuid"
)

type StoreIncome struct {
	db *sql.DB
}

func NewStorage(db *sql.DB) *StoreIncome {
	return &StoreIncome{db}
}

func (s *StoreIncome) AddIncome(ctx context.Context, income *model.Income) error {
	query := `
		INSERT INTO incomes (user_id, monto, payment_method, descripcion, category_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, fecha, created_at
	`

	ctx, cancel := context.WithTimeout(ctx, util.QueryTimeoutDuration)
	defer cancel()

	// income.CategoryID es un *uuid.UUID porque la columna es nullable: un nil
	// viaja a la base como NULL, no como un UUID cero que no matchearía ninguna
	// categoría.
	err := s.db.QueryRowContext(ctx, query,
		income.UserID,
		income.Monto,
		income.PaymentMethod,
		income.Descripcion,
		income.CategoryID,
	).Scan(
		&income.ID,
		&income.Fecha,
		&income.CreatedAt,
	)

	if err != nil {
		return err
	}

	return nil
}

func (s *StoreIncome) DeleteIncome(ctx context.Context, incomeID uuid.UUID, userID uuid.UUID) error {
	query := `
		DELETE FROM incomes
		WHERE id = $1 and user_id = $2
	`

	ctx, cancel := context.WithTimeout(ctx, util.QueryTimeoutDuration)
	defer cancel()

	res, err := s.db.ExecContext(ctx, query, incomeID, userID)

	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()

	if err != nil {
		return err
	}

	if rows == 0 {
		return util.ErrNotFound
	}

	return nil
}

func (s *StoreIncome) GetAllIncomesOfUser(ctx context.Context, userID uuid.UUID, fq pagination.IncomePaginationQuery) ([]model.Income, error) {
	// El store es la última línea de defensa de la paginación: un
	// IncomePaginationQuery{} (valor cero) pondría LIMIT 0 y devolvería siempre
	// lista vacía, y un offset negativo hace fallar la query en Postgres.
	if fq.Limit <= 0 {
		fq.Limit = pagination.DefaultLimit
	}
	if fq.Offset < 0 {
		fq.Offset = 0
	}

	// El ORDER BY no es cosmético: sin un orden determinista el LIMIT/OFFSET
	// puede repetir o saltear filas entre páginas. El id desempata los ingresos
	// que comparten fecha.
	//
	// category_id va último en la lista y en el Scan a propósito: mantenerlo al
	// final deja intactas las posiciones de las otras 7 columnas.
	query := `
		SELECT id, user_id, monto, payment_method, descripcion, fecha, created_at, category_id
		FROM incomes
		WHERE user_id = $1
			  AND ($4 = '' OR fecha >= $4::date)
			  AND ($5 = '' OR fecha <= $5::date)
			  AND ($6 = '' OR category_id = $6::uuid)
		ORDER BY fecha DESC, id DESC
		LIMIT $2 OFFSET $3
	`

	ctx, cancel := context.WithTimeout(ctx, util.QueryTimeoutDuration)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query,
		userID,
		fq.Limit,
		fq.Offset,
		fq.Since,
		fq.Until,
		fq.CategoryID,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	incomes := make([]model.Income, 0)

	for rows.Next() {
		var income model.Income
		err := rows.Scan(
			&income.ID,
			&income.UserID,
			&income.Monto,
			&income.PaymentMethod,
			&income.Descripcion,
			&income.Fecha,
			&income.CreatedAt,
			&income.CategoryID,
		)

		if err != nil {
			return nil, err
		}

		incomes = append(incomes, income)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return incomes, nil
}
