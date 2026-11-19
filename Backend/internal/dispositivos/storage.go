package dispositivos

import (
	"context"
	"database/sql"

	"github.com/crcaniullan-commits/Tally/internal/bdtx"
	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/internal/utilidades"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type AlmacenDispositivos struct {
	db *sql.DB
}

func NuevoAlmacenDispositivos(db *sql.DB) *AlmacenDispositivos {
	return &AlmacenDispositivos{db: db}
}

func (s *AlmacenDispositivos) Upsert(ctx context.Context, userID uuid.UUID, token, platform string) error {
	query := `
		INSERT INTO device_tokens (id, user_id, token, platform, last_seen_at)
		VALUES ($1, $2, $3, $4, now())
		ON CONFLICT (token) DO UPDATE
		SET user_id = EXCLUDED.user_id,
		    platform = EXCLUDED.platform,
		    last_seen_at = now();
	`
	ctx, cancel := context.WithTimeout(ctx, utilidades.DuracionTimeoutConsulta)
	defer cancel()

	_, err := bdtx.FromContext(ctx, s.db).ExecContext(ctx, query, uuid.New(), userID, token, platform)
	return err
}

func (s *AlmacenDispositivos) Eliminar(ctx context.Context, userID uuid.UUID, token string) error {
	query := `
		DELETE FROM device_tokens
		WHERE token = $1 AND user_id = $2;
	`
	ctx, cancel := context.WithTimeout(ctx, utilidades.DuracionTimeoutConsulta)
	defer cancel()

	res, err := bdtx.FromContext(ctx, s.db).ExecContext(ctx, query, token, userID)
	if err != nil {
		return err
	}
	filas, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if filas == 0 {
		return utilidades.ErrorNoEncontrado
	}
	return nil
}

func (s *AlmacenDispositivos) ObtenerTokensPorUsuario(ctx context.Context, userID uuid.UUID) ([]string, error) {
	query := `
		SELECT token
		FROM device_tokens
		WHERE user_id = $1
		ORDER BY last_seen_at DESC;
	`
	ctx, cancel := context.WithTimeout(ctx, utilidades.DuracionTimeoutConsulta)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokens []string
	for rows.Next() {
		var token string
		if err := rows.Scan(&token); err != nil {
			return nil, err
		}
		tokens = append(tokens, token)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return tokens, nil
}

func (s *AlmacenDispositivos) ListarPorUsuario(ctx context.Context, userID uuid.UUID) ([]model.DeviceToken, error) {
	query := `
		SELECT id, user_id, token, platform, last_seen_at
		FROM device_tokens
		WHERE user_id = $1
		ORDER BY last_seen_at DESC;
	`
	ctx, cancel := context.WithTimeout(ctx, utilidades.DuracionTimeoutConsulta)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dispositivos []model.DeviceToken
	for rows.Next() {
		var d model.DeviceToken
		if err := rows.Scan(&d.ID, &d.UserID, &d.Token, &d.Platform, &d.LastSeenAt); err != nil {
			return nil, err
		}
		dispositivos = append(dispositivos, d)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return dispositivos, nil
}

func (s *AlmacenDispositivos) EliminarTokens(ctx context.Context, tokens []string) error {
	if len(tokens) == 0 {
		return nil
	}
	query := `
		DELETE FROM device_tokens
		WHERE token = ANY($1);
	`
	ctx, cancel := context.WithTimeout(ctx, utilidades.DuracionTimeoutConsulta)
	defer cancel()

	_, err := bdtx.FromContext(ctx, s.db).ExecContext(ctx, query, pq.Array(tokens))
	return err
}
