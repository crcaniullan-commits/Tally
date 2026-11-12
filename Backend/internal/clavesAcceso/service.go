package clavesacceso

import (
	"context"
	"crypto/rand"
	"log"
	"time"

	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/internal/utilidades"
	"github.com/google/uuid"
)

type RepositorioClavesAcceso interface {
	Crear(context.Context, *model.AccessKey) error
	ObtenerPorID(context.Context, uuid.UUID) (model.AccessKey, error)
	Revocar(context.Context, DatosCodigoRevocacion) error
}

type SendEmail interface {
	SendActivationCode(ctx context.Context, to, code string, expiresAt time.Time) error
}

type ServicioClavesAcceso struct {
	store RepositorioClavesAcceso
	email SendEmail
}

func NuevoServicioClavesAcceso(store RepositorioClavesAcceso, email SendEmail) *ServicioClavesAcceso {
	return &ServicioClavesAcceso{store, email}
}

func (s *ServicioClavesAcceso) EmitirParaUsuario(ctx context.Context, municipalID uuid.UUID, email string) (model.AccessKey, error) {
	expire := time.Now().AddDate(1, 0, 0)
	code := rand.Text()[:6]

	accessKey := &model.AccessKey{
		Code:      code,
		CreatedBy: &municipalID,
		ExpiresAt: &expire,
	}

	if err := s.store.Crear(ctx, accessKey); err != nil {
		return model.AccessKey{}, err
	}

	s.sendEmailAsync(email, code, expire)

	return *accessKey, nil

}

func (s *ServicioClavesAcceso) Reenviar(ctx context.Context, accessKeyId uuid.UUID, email string) error {
	accessKey, err := s.store.ObtenerPorID(ctx, accessKeyId)

	if err != nil {
		return err
	}

	if accessKey.ExpiresAt == nil {
		return utilidades.ErrorNoEncontrado
	}

	s.sendEmailAsync(email, accessKey.Code, *accessKey.ExpiresAt)

	return nil
}

func (s *ServicioClavesAcceso) Revocar(ctx context.Context, municipalID uuid.UUID, code string) error {
	revoke := &DatosCodigoRevocacion{
		RevocadoPor: municipalID,
		code:        code,
	}

	if err := s.store.Revocar(ctx, *revoke); err != nil {
		return err
	}

	return nil
}

func (s *ServicioClavesAcceso) sendEmailAsync(to, code string, expire time.Time) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := s.email.SendActivationCode(ctx, to, code, expire); err != nil {
			log.Printf("error enviando correo de activación a %s: %v", to, err)
		}
	}()
}
