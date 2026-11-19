package jobs

import (
	"context"
	"database/sql"
	"time"

	"github.com/crcaniullan-commits/Tally/internal/utilidades"
	"github.com/google/uuid"
)

type ResumenCobrosVencidos struct {
	UserID   uuid.UUID
	Cantidad int
	Total    int64
}

type AlmacenJobs struct {
	db *sql.DB
}

func NuevoAlmacenJobs(db *sql.DB) *AlmacenJobs {
	return &AlmacenJobs{db: db}
}

func (s *AlmacenJobs) UsuariosSinMovimientos(ctx context.Context, start, end time.Time) ([]uuid.UUID, error) {
	query := `
		SELECT DISTINCT dt.user_id
		FROM device_tokens dt
		WHERE NOT EXISTS (
			SELECT 1 FROM incomes i
			JOIN empresas e ON e.id = i.empresa_id
			WHERE e.user_id = dt.user_id
			  AND i.created_at >= $1 AND i.created_at < $2
		)
		AND NOT EXISTS (
			SELECT 1 FROM expenses x
			JOIN empresas e ON e.id = x.empresa_id
			WHERE e.user_id = dt.user_id
			  AND x.created_at >= $1 AND x.created_at < $2
		);
	`
	ctx, cancel := context.WithTimeout(ctx, utilidades.DuracionTimeoutConsulta)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userIDs []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		userIDs = append(userIDs, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return userIDs, nil
}

func (s *AlmacenJobs) CobrosVencidosPorUsuario(ctx context.Context, day string) ([]ResumenCobrosVencidos, error) {
	query := `
		SELECT e.user_id, COUNT(*) AS cantidad, SUM(d.monto) AS total
		FROM debtors d
		JOIN empresas e ON e.id = d.empresa_id
		WHERE d.fecha_limite < $1::date
		  AND d.pagado = false
		GROUP BY e.user_id;
	`
	ctx, cancel := context.WithTimeout(ctx, utilidades.DuracionTimeoutConsulta)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query, day)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var resumenes []ResumenCobrosVencidos
	for rows.Next() {
		var r ResumenCobrosVencidos
		if err := rows.Scan(&r.UserID, &r.Cantidad, &r.Total); err != nil {
			return nil, err
		}
		resumenes = append(resumenes, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return resumenes, nil
}
