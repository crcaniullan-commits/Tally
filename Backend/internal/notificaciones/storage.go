package notificaciones

import (
	"context"
	"database/sql"

	"github.com/crcaniullan-commits/Tally/internal/bdtx"
	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/internal/utilidades"
	"github.com/google/uuid"
)

type AlmacenNotificaciones struct {
	db *sql.DB
}

func NuevoAlmacenNotificaciones(db *sql.DB) *AlmacenNotificaciones {
	return &AlmacenNotificaciones{db: db}
}

func (s *AlmacenNotificaciones) Claim(ctx context.Context, userID uuid.UUID, kind string, day string) (bool, error) {
	query := `
		INSERT INTO notification_log (user_id, kind, ref_date)
		VALUES ($1, $2, $3)
		ON CONFLICT DO NOTHING;
	`
	ctx, cancel := context.WithTimeout(ctx, utilidades.DuracionTimeoutConsulta)
	defer cancel()

	res, err := bdtx.FromContext(ctx, s.db).ExecContext(ctx, query, userID, kind, day)
	if err != nil {
		return false, err
	}
	filas, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return filas == 1, nil
}

func (s *AlmacenNotificaciones) Unclaim(ctx context.Context, userID uuid.UUID, kind string, day string) error {
	query := `
		DELETE FROM notification_log
		WHERE user_id = $1 AND kind = $2 AND ref_date = $3;
	`
	ctx, cancel := context.WithTimeout(ctx, utilidades.DuracionTimeoutConsulta)
	defer cancel()

	_, err := bdtx.FromContext(ctx, s.db).ExecContext(ctx, query, userID, kind, day)
	return err
}

func (s *AlmacenNotificaciones) ListarPorUsuario(ctx context.Context, userID uuid.UUID) ([]model.NotificationLog, error) {
	query := `
		SELECT user_id, kind, ref_date, sent_at
		FROM notification_log
		WHERE user_id = $1
		ORDER BY sent_at DESC;
	`
	ctx, cancel := context.WithTimeout(ctx, utilidades.DuracionTimeoutConsulta)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []model.NotificationLog
	for rows.Next() {
		var l model.NotificationLog
		if err := rows.Scan(&l.UserID, &l.Kind, &l.RefDate, &l.SentAt); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return logs, nil
}
