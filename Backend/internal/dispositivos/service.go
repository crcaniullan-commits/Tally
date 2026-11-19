package dispositivos

import (
	"context"
	"errors"
	"strings"

	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/google/uuid"
)

var (
	ErrorTokenRequerido     = errors.New("el token es requerido")
	ErrorPlataformaInvalida = errors.New("la plataforma debe ser android, ios o web")
)

type RepositorioDispositivos interface {
	Upsert(ctx context.Context, userID uuid.UUID, token, platform string) error
	Eliminar(ctx context.Context, userID uuid.UUID, token string) error
	ObtenerTokensPorUsuario(ctx context.Context, userID uuid.UUID) ([]string, error)
	ListarPorUsuario(ctx context.Context, userID uuid.UUID) ([]model.DeviceToken, error)
	EliminarTokens(ctx context.Context, tokens []string) error
}

type ServicioDispositivos struct {
	store RepositorioDispositivos
}

func NuevoServicioDispositivos(store RepositorioDispositivos) *ServicioDispositivos {
	return &ServicioDispositivos{store: store}
}

func (s *ServicioDispositivos) Registrar(ctx context.Context, userID uuid.UUID, payload PayloadDispositivo) error {
	token := strings.TrimSpace(payload.Token)
	if token == "" {
		return ErrorTokenRequerido
	}

	plataforma := strings.ToLower(strings.TrimSpace(payload.Platform))
	if plataforma != "android" && plataforma != "ios" && plataforma != "web" {
		return ErrorPlataformaInvalida
	}

	return s.store.Upsert(ctx, userID, token, plataforma)
}

func (s *ServicioDispositivos) Desvincular(ctx context.Context, userID uuid.UUID, token string) error {
	token = strings.TrimSpace(token)
	if token == "" {
		return ErrorTokenRequerido
	}

	return s.store.Eliminar(ctx, userID, token)
}

func (s *ServicioDispositivos) Listar(ctx context.Context, userID uuid.UUID) ([]model.DeviceToken, error) {
	return s.store.ListarPorUsuario(ctx, userID)
}

func (s *ServicioDispositivos) ObtenerTokensPorUsuario(ctx context.Context, userID uuid.UUID) ([]string, error) {
	return s.store.ObtenerTokensPorUsuario(ctx, userID)
}

func (s *ServicioDispositivos) EliminarTokensInvalidos(ctx context.Context, tokens []string) error {
	return s.store.EliminarTokens(ctx, tokens)
}
