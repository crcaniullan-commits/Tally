package incomes

import (
	"context"
	"database/sql"
	"time"

	"github.com/crcaniullan-commits/Tally/internal/util"
	"github.com/google/uuid"
)

type IncomeStorage struct {
	ID            uuid.UUID          `json:"id"`
	UserID        uuid.UUID          `json:"userID"`
	Monto         int64              `json:"monto"`
	PaymentMethod util.PaymentMethod `json:"payment_method"`
	Descripcion   *string            `json:"descripcion"`
	Fecha         time.Time          `json:"fecha"`
	CreatedAt     time.Time          `json:"created_at"`
}

type StoreIncome struct {
	db *sql.DB
}

func NewStorage(db *sql.DB) *StoreIncome {
	return &StoreIncome{db}
}

func (s *StoreIncome) AddIncome(ctx context.Context, income *IncomeStorage) error {
	query := `
		INSERT INTO incomes (user_id, monto, payment_method, descripcion)
		VALUES ($1, $2, $3, $4)
		RETURNING id, fecha, created_at
	`

	ctx, cancel := context.WithTimeout(ctx, util.QueryTimeoutDuration)
	defer cancel()

	err := s.db.QueryRowContext(ctx, query,
		income.UserID,
		income.Monto,
		income.PaymentMethod,
		income.Descripcion,
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

func (s *StoreIncome) GetAllIncomesOfUser(ctx context.Context, userID uuid.UUID) ([]IncomeStorage, error) {
	query := `
		SELECT id, user_id, monto, payment_method, descripcion, fecha, created_at
		FROM incomes
		WHERE user_id = $1
	`

	ctx, cancel := context.WithTimeout(ctx, util.QueryTimeoutDuration)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query, userID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	incomes := make([]IncomeStorage, 0)

	for rows.Next() {
		var income IncomeStorage
		err := rows.Scan(
			&income.ID,
			&income.UserID,
			&income.Monto,
			&income.PaymentMethod,
			&income.Descripcion,
			&income.Fecha,
			&income.CreatedAt,
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
