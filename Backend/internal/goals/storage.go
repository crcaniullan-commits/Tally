package goals

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/internal/util"
	"github.com/google/uuid"
)

type StoreGoal struct {
	db *sql.DB
}

func NewStoreGoal(db *sql.DB) *StoreGoal {
	return &StoreGoal{db: db}
}

func (s *StoreGoal) CreateGoal(ctx context.Context, goal *model.Goal) error {
	query := `
		INSERT INTO goals (empresa_id, nombre, periodo, monto_meta)
		VALUES ($1, $2, $3, $4) RETURNING id, created_at
	`

	ctx, cancel := context.WithTimeout(ctx, util.QueryTimeoutDuration)
	defer cancel()

	err := s.db.QueryRowContext(ctx, query, goal.EmpresaID, goal.Nombre, goal.Periodo, goal.MontoMeta).Scan(
		&goal.ID,
		&goal.CreatedAt,
	)

	if err != nil {
		return err
	}

	return nil
}

func (s *StoreGoal) DeleteGoal(ctx context.Context, goalID uuid.UUID, empresaID uuid.UUID) error {
	query := `
		DELETE FROM goals
		WHERE id = $1 and empresa_id = $2
	`

	ctx, cancel := context.WithTimeout(ctx, util.QueryTimeoutDuration)
	defer cancel()

	res, err := s.db.ExecContext(ctx, query, goalID, empresaID)

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

// dateTruncUnit traduce el período del dominio (en español) al unit que
// entiende date_trunc. La traducción vive en Go para que un período inválido
// falle antes de llegar a Postgres.
func dateTruncUnit(period model.GoalPeriod) (string, error) {
	switch period {
	case model.GoalPeriodDia:
		return "day", nil
	case model.GoalPeriodSemana:
		return "week", nil
	case model.GoalPeriodMes:
		return "month", nil
	default:
		return "", fmt.Errorf("período inválido: %q", period)
	}
}

// GetIncomesTotalByPeriod suma los ingresos de la empresa dentro del período
// actual (día, semana o mes). El rango es semiabierto [inicio, fin) sobre
// fecha, y date_trunc('week') usa la semana ISO (lunes).
func (s *StoreGoal) GetIncomesTotalByPeriod(ctx context.Context, empresaID uuid.UUID, period model.GoalPeriod) (int64, error) {
	unit, err := dateTruncUnit(period)
	if err != nil {
		return 0, err
	}

	query := `
		SELECT COALESCE(SUM(monto), 0)::bigInt
		FROM incomes
		WHERE empresa_id = $1
		  AND fecha >= date_trunc($2::text, CURRENT_DATE::timestamp)::date
		  AND fecha < (date_trunc($2::text, CURRENT_DATE::timestamp) + CASE $2::text
		      WHEN 'day' THEN interval '1 day'
		      WHEN 'week' THEN interval '1 week'
		      WHEN 'month' THEN interval '1 month'
		  END)::date
	`

	ctx, cancel := context.WithTimeout(ctx, util.QueryTimeoutDuration)
	defer cancel()

	var total int64

	if err := s.db.QueryRowContext(ctx, query, empresaID, unit).Scan(&total); err != nil {
		return 0, err
	}

	return total, nil
}

func (s *StoreGoal) GetGoal(ctx context.Context, goalID uuid.UUID, empresaID uuid.UUID) (model.Goal, error) {
	query := `
		SELECT id, empresa_id, nombre, periodo, monto_meta, created_At
		FROM goals
		WHERE empresa_id = $1 and id = $2
	`

	ctx, cancel := context.WithTimeout(ctx, util.QueryTimeoutDuration)
	defer cancel()

	var goal model.Goal

	err := s.db.QueryRowContext(ctx, query, empresaID, goalID).Scan(
		&goal.ID,
		&goal.EmpresaID,
		&goal.Nombre,
		&goal.Periodo,
		&goal.MontoMeta,
		&goal.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Goal{}, util.ErrNotFound
		}

		return model.Goal{}, err
	}

	return goal, nil
}

func (s *StoreGoal) GetGoals(ctx context.Context, empresaID uuid.UUID) ([]model.Goal, error) {
	query := `
		SELECT id, empresa_id, nombre, periodo, monto_meta, created_At
		FROM goals
		WHERE empresa_id = $1
	`
	ctx, cancel := context.WithTimeout(ctx, util.QueryTimeoutDuration)
	defer cancel()

	goals := make([]model.Goal, 0)

	rows, err := s.db.QueryContext(ctx, query, empresaID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var goal model.Goal

		if err := rows.Scan(
			&goal.ID,
			&goal.EmpresaID,
			&goal.Nombre,
			&goal.Periodo,
			&goal.MontoMeta,
			&goal.CreatedAt,
		); err != nil {
			return nil, err
		}

		goals = append(goals, goal)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return goals, nil
}
