package notificaciones

import (
	"context"

	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type BuscadorDispositivos interface {
	ObtenerTokensPorUsuario(ctx context.Context, userID uuid.UUID) ([]string, error)
	EliminarTokensInvalidos(ctx context.Context, tokens []string) error
}

type RepositorioNotificaciones interface {
	Claim(ctx context.Context, userID uuid.UUID, kind string, day string) (bool, error)
	Unclaim(ctx context.Context, userID uuid.UUID, kind string, day string) error
	ListarPorUsuario(ctx context.Context, userID uuid.UUID) ([]model.NotificationLog, error)
}

type Notificador interface {
	Send(ctx context.Context, userID uuid.UUID, n model.Notification) error
	Claim(ctx context.Context, userID uuid.UUID, kind string, day string) (bool, error)
	Unclaim(ctx context.Context, userID uuid.UUID, kind string, day string) error
	Listar(ctx context.Context, userID uuid.UUID) ([]model.NotificationLog, error)
}

type ServicioNotificaciones struct {
	store        RepositorioNotificaciones
	transporte   EnviadorPush
	dispositivos BuscadorDispositivos
	logger       *zap.SugaredLogger
}

func NuevoServicioNotificaciones(
	store RepositorioNotificaciones,
	transporte EnviadorPush,
	dispositivos BuscadorDispositivos,
	logger *zap.SugaredLogger,
) *ServicioNotificaciones {
	return &ServicioNotificaciones{
		store:        store,
		transporte:   transporte,
		dispositivos: dispositivos,
		logger:       logger,
	}
}

func (s *ServicioNotificaciones) Send(ctx context.Context, userID uuid.UUID, n model.Notification) error {
	if s.dispositivos == nil || s.transporte == nil {
		return nil
	}

	tokens, err := s.dispositivos.ObtenerTokensPorUsuario(ctx, userID)
	if err != nil {
		if s.logger != nil {
			s.logger.Errorw("error obteniendo tokens de usuario para push", "user_id", userID, "error", err)
		}
		return err
	}

	if len(tokens) == 0 {
		return nil
	}

	tokensInvalidos, err := s.transporte.EnviarPush(ctx, tokens, n)
	if err != nil {
		if s.logger != nil {
			s.logger.Errorw("error enviando notificación push", "user_id", userID, "error", err)
		}
		return err
	}

	if len(tokensInvalidos) > 0 {
		_ = s.dispositivos.EliminarTokensInvalidos(ctx, tokensInvalidos)
	}

	return nil
}

func (s *ServicioNotificaciones) Claim(ctx context.Context, userID uuid.UUID, kind string, day string) (bool, error) {
	return s.store.Claim(ctx, userID, kind, day)
}

func (s *ServicioNotificaciones) Unclaim(ctx context.Context, userID uuid.UUID, kind string, day string) error {
	return s.store.Unclaim(ctx, userID, kind, day)
}

func (s *ServicioNotificaciones) Listar(ctx context.Context, userID uuid.UUID) ([]model.NotificationLog, error) {
	return s.store.ListarPorUsuario(ctx, userID)
}
