package incomes

import (
	"context"
	"database/sql"

	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/internal/paginacion"
	"github.com/crcaniullan-commits/Tally/internal/utilidades"
	"github.com/google/uuid"
)

type AlmacenIngresos struct {
	db *sql.DB
}

func NuevoAlmacenIngresos(db *sql.DB) *AlmacenIngresos {
	return &AlmacenIngresos{db}
}

func (s *AlmacenIngresos) AgregarIngreso(ctx context.Context, income *model.Income) error {
	query := `
		INSERT INTO incomes (empresa_id, monto, payment_method, descripcion, category_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, fecha, created_at
	`

	ctx, cancel := context.WithTimeout(ctx, utilidades.DuracionTimeoutConsulta)
	defer cancel()

	// income.CategoryID es un *uuid.UUID porque la columna es nullable: un nil
	// viaja a la base como NULL, no como un UUID cero que no matchearía ninguna
	// categoría.
	err := s.db.QueryRowContext(ctx, query,
		income.EmpresaID,
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

// CategoriaExiste verifica que la categoría exista y sea visible para la
// empresa: global (empresa_id IS NULL) o propia. Una categoría de otra empresa
// no pasa el filtro y el service la trata como inexistente.
func (s *AlmacenIngresos) CategoriaExiste(ctx context.Context, categoryID uuid.UUID, empresaID uuid.UUID) (bool, error) {
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

func (s *AlmacenIngresos) EliminarIngreso(ctx context.Context, ingresoID uuid.UUID, empresaID uuid.UUID) error {
	query := `
		DELETE FROM incomes
		WHERE id = $1 and empresa_id = $2
	`

	ctx, cancel := context.WithTimeout(ctx, utilidades.DuracionTimeoutConsulta)
	defer cancel()

	res, err := s.db.ExecContext(ctx, query, ingresoID, empresaID)

	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()

	if err != nil {
		return err
	}

	if rows == 0 {
		return utilidades.ErrorNoEncontrado
	}

	return nil
}

func (s *AlmacenIngresos) ObtenerIngresosDeEmpresa(ctx context.Context, empresaID uuid.UUID, fq paginacion.ConsultaPaginacionIngresos) ([]model.Income, error) {
	// El store es la última línea de defensa de la paginación: un
	// IncomePaginationQuery{} (valor cero) pondría LIMIT 0 y devolvería siempre
	// lista vacía, y un offset negativo hace fallar la query en Postgres.
	if fq.Limite <= 0 {
		fq.Limite = paginacion.LimitePorDefecto
	}
	if fq.Desplazamiento < 0 {
		fq.Desplazamiento = 0
	}

	// El ORDER BY no es cosmético: sin un orden determinista el LIMIT/OFFSET
	// puede repetir o saltear filas entre páginas. El id desempata los ingresos
	// que comparten fecha.
	//
	// category_id va último en la lista y en el Scan a propósito: mantenerlo al
	// final deja intactas las posiciones de las otras 7 columnas.
	query := `
		SELECT id, empresa_id, monto, payment_method, descripcion, fecha, created_at, category_id
		FROM incomes
		WHERE empresa_id = $1
			  AND ($4 = '' OR fecha >= $4::date)
			  AND ($5 = '' OR fecha <= $5::date)
			  AND ($6 = '' OR category_id = $6::uuid)
		ORDER BY fecha DESC, id DESC
		LIMIT $2 OFFSET $3
	`

	ctx, cancel := context.WithTimeout(ctx, utilidades.DuracionTimeoutConsulta)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query,
		empresaID,
		fq.Limite,
		fq.Desplazamiento,
		fq.Desde,
		fq.Hasta,
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
			&income.EmpresaID,
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
