package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/crcaniullan-commits/Tally/internal/bdtx"
	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/internal/utilidades"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	// longitudCodigoRecuperacion son los dígitos del código de recuperación.
	// El usuario lo tipea, así que 6 es el máximo cómodo de tipear sin errar.
	longitudCodigoRecuperacion = 6
	// ventanaCodigoRecuperacion es cuánto vive el código desde que se emite.
	ventanaCodigoRecuperacion = 15 * time.Minute
)

type RepositorioAuth interface {
	Create(context.Context, *model.User) error
	ObtenerPorEmail(context.Context, string) (*model.User, error)
	establecerVencimiento(context.Context, time.Time, uuid.UUID) error
	CrearCodigoRecuperacion(context.Context, *model.CodigoRecuperacion) error
	ObtenerCodigoRecuperacion(context.Context, uuid.UUID) (*model.CodigoRecuperacion, error)
	MarcarCodigoVerificado(context.Context, uuid.UUID) error
	MarcarCodigoUsado(context.Context, uuid.UUID) error
	ActualizarPassword(context.Context, uuid.UUID, []byte) error
}

type Canjeador interface {
	Canjear(ctx context.Context, code string, userID uuid.UUID) (*model.AccessKey, error)
}

// Remitente es lo que el módulo necesita para mandar el código de recuperación
// por correo. El módulo no depende de internal/correo ni de internal/clavesAcceso:
// solo de esta interfaz, así el remitente se puede mockear en los tests.
type Remitente interface {
	SendPasswordResetCode(ctx context.Context, to, code string, expiresAt time.Time) error
}

type ServicioAuth struct {
	store         RepositorioAuth
	redeemer      Canjeador
	authenticator Authenticator
	transactor    *bdtx.Transactor
	email         Remitente
}

func NuevoServicioAuth(store RepositorioAuth, auth Authenticator, t *bdtx.Transactor, r Canjeador, email Remitente) *ServicioAuth {
	return &ServicioAuth{store, r, auth, t, email}
}

func (s *ServicioAuth) RegistrarUsuario(ctx context.Context, payload *PayloadCrearUsuario) (string, error) {
	token, err := s.transactor.WithinTxWithResult(ctx, func(ctx context.Context) (string, error) {
		user := &model.User{
			Nombre: payload.Nombre,
			Email:  payload.Email,
			Role:   utilidades.RolUsuario,
		}

		if err := user.PasswordHash.Set(payload.Password); err != nil {
			return "", err
		}

		rut, err := utilidades.ParseRUT(payload.Rut)

		if err != nil {
			return "", err
		}

		user.Rut = rut

		err = s.store.Create(ctx, user)

		if err != nil {
			return "", err
		}

		token, err := s.crearToken(user)

		if err != nil {
			return "", err
		}

		accesskey, err := s.redeemer.Canjear(ctx, payload.Code, user.ID)

		if err != nil {
			return "", err
		}

		if err := s.store.establecerVencimiento(ctx, *accesskey.ExpiresAt, user.ID); err != nil {
			return "", err
		}

		return token, nil
	})

	if err != nil {
		return "", err
	}

	if token == "" {
		return "", errors.New("error al generar el token")
	}

	return token, nil

}

func (s *ServicioAuth) Login(ctx context.Context, payload *PayloadInicioSesion) (string, error) {
	user, err := s.store.ObtenerPorEmail(ctx, payload.Email)

	if err != nil {
		return "", err
	}

	if err = user.PasswordHash.Compare(payload.Password); err != nil {
		return "", err
	}

	token, err := s.crearToken(user)

	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *ServicioAuth) crearToken(user *model.User) (string, error) {
	claims := jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(cExp).Unix(),
		"iat": time.Now().Unix(),
		"nbf": time.Now().Unix(),
		"iss": cIss,
		"aud": cIss,
	}

	token, err := s.authenticator.GenerarToken(claims)
	if err != nil {
		return "", err
	}

	return token, err

}

func (s *ServicioAuth) CrearMunicipal(ctx context.Context, payload *PayloadCrearMunicipal) (string, error) {
	user := &model.User{
		Nombre: payload.Nombre,
		Email:  payload.Email,
		Role:   utilidades.RolMunicipal,
	}

	if err := user.PasswordHash.Set(payload.Password); err != nil {
		return "", err
	}

	rut, err := utilidades.ParseRUT(payload.Rut)

	if err != nil {
		return "", err
	}

	user.Rut = rut

	err = s.store.Create(ctx, user)

	if err != nil {
		return "", err
	}

	token, err := s.crearToken(user)

	if err != nil {
		return "", err
	}

	return token, nil
}

// SolicitarRecuperacion emite un código de un solo uso para el email dado y lo
// manda por correo.
//
// Si el email no está registrado devuelve utilidades.ErrorNoEncontrado: el
// handler lo traduce a 400, no a un 500.
func (s *ServicioAuth) SolicitarRecuperacion(ctx context.Context, email string) error {
	user, err := s.store.ObtenerPorEmail(ctx, email)

	if err != nil {
		return err
	}

	codigo, err := generarCodigo(longitudCodigoRecuperacion)

	if err != nil {
		return err
	}

	expira := time.Now().Add(ventanaCodigoRecuperacion)

	registro := &model.CodigoRecuperacion{
		UserID:   user.ID,
		Codigo:   hashearCodigo(codigo),
		ExpiraEn: expira,
	}

	if err := s.store.CrearCodigoRecuperacion(ctx, registro); err != nil {
		return err
	}

	if err := s.email.SendPasswordResetCode(ctx, user.Email, codigo, expira); err != nil {
		return err
	}

	return nil
}

// VerificarCodigoRecuperacion comprueba el código que recibió el usuario. Deja el
// código marcado como verificado, que es lo que habilita el cambio de contraseña.
func (s *ServicioAuth) VerificarCodigoRecuperacion(ctx context.Context, email, codigo string) error {
	registro, err := s.codigoVigente(ctx, email, codigo)

	if err != nil {
		return err
	}

	if registro.VerificadoEn != nil {
		return nil
	}

	return s.store.MarcarCodigoVerificado(ctx, registro.ID)
}

// ConfirmarRecuperacion cambia la contraseña del usuario. Exige un código
// verificado y vigente: es el segundo paso del flujo de recuperación.
func (s *ServicioAuth) ConfirmarRecuperacion(ctx context.Context, email, codigo, password string) error {
	registro, err := s.codigoVigente(ctx, email, codigo)

	if err != nil {
		return err
	}

	if registro.VerificadoEn == nil {
		return ErrorCodigoRecuperacionNoVerificado
	}

	usuario := &model.User{}

	if err := usuario.PasswordHash.Set(password); err != nil {
		return err
	}

	if err := s.store.ActualizarPassword(ctx, registro.UserID, usuario.PasswordHash.GetHash()); err != nil {
		return err
	}

	return s.store.MarcarCodigoUsado(ctx, registro.ID)
}

// codigoVigente devuelve el último código emitido para el email, chequeando que
// siga siendo utilizable: que no se haya usado, que no haya vencido y que el
// código enviado coincida con el guardado.
func (s *ServicioAuth) codigoVigente(ctx context.Context, email, codigo string) (*model.CodigoRecuperacion, error) {
	user, err := s.store.ObtenerPorEmail(ctx, email)

	if err != nil {
		return nil, err
	}

	registro, err := s.store.ObtenerCodigoRecuperacion(ctx, user.ID)

	if err != nil {
		return nil, err
	}

	if registro.UsadoEn != nil {
		return nil, ErrorCodigoRecuperacionUsado
	}

	if time.Now().After(registro.ExpiraEn) {
		return nil, ErrorCodigoRecuperacionVencido
	}

	if registro.Codigo != hashearCodigo(codigo) {
		return nil, ErrorCodigoRecuperacionInvalido
	}

	return registro, nil
}

// generarCodigo devuelve un código numérico aleatorio de los dígitos pedidos,
// con ceros a la izquierda para que siempre tenga la misma longitud.
func generarCodigo(digitos int) (string, error) {
	maximo := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(digitos)), nil)

	numero, err := rand.Int(rand.Reader, maximo)

	if err != nil {
		return "", fmt.Errorf("generando el código de recuperación: %w", err)
	}

	return fmt.Sprintf("%0*d", digitos, numero.Int64()), nil
}

// hashearCodigo devuelve el SHA-256 del código en hexadecimal. Es lo que se
// guarda en la base: el código que el usuario recibió nunca se persiste.
func hashearCodigo(codigo string) string {
	suma := sha256.Sum256([]byte(codigo))
	return hex.EncodeToString(suma[:])
}
