package accesskeys

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var errStoreBoom = errors.New("boom: fallo en la capa de persistencia")
var errCorreoBoom = errors.New("boom: fallo en el cliente de correo")

// newTestService arma un AccessKeyService con el store y el cliente de correo
// mockeados.
func newTestService(store AccessKeyStore, email SendEmail) *AccessKeyService {
	return NewAccessKeyService(store, email)
}

func TestAccessKeyService_IssueForUser(t *testing.T) {
	t.Run("crea la llave a nombre del municipal y la devuelve", func(t *testing.T) {
		store := new(AccessKeyStoreMock)
		email := newEmailQueAcepta(nil)
		service := newTestService(store, email)

		var guardado *model.AccessKey

		store.On("Create", mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) { guardado = args.Get(1).(*model.AccessKey) }).
			Return(nil).Once()

		accessKey, err := service.IssueForUser(context.Background(), testMunicipalID, testEmail)

		require.NoError(t, err)
		require.NotNil(t, guardado)
		assert.Equal(t, testMunicipalID, accessKey.CreatedBy, "created_by es el municipal emisor")
		assert.NotEmpty(t, accessKey.Code, "el código se genera en el service")
		assert.Equal(t, *guardado, accessKey, "la llave devuelta es la que se intentó guardar")
		store.AssertExpectations(t)
		email.esperar(t, 1)
	})

	t.Run("el código que se guarda es el mismo que va por correo", func(t *testing.T) {
		store := new(AccessKeyStoreMock)
		email := newEmailQueAcepta(nil)
		service := newTestService(store, email)

		var guardado *model.AccessKey

		store.On("Create", mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) { guardado = args.Get(1).(*model.AccessKey) }).
			Return(nil).Once()

		accessKey, err := service.IssueForUser(context.Background(), testMunicipalID, testEmail)

		require.NoError(t, err)

		enviados := email.esperar(t, 1)
		require.Len(t, enviados, 1)

		assert.Equal(t, testEmail, enviados[0].To, "el correo va al email del payload, no al del municipal")
		assert.Equal(t, accessKey.Code, enviados[0].Code, "por correo debe viajar el código guardado")
		require.NotNil(t, accessKey.ExpiresAt)
		assert.True(t, enviados[0].ExpiresAt.Equal(*accessKey.ExpiresAt))
		require.NotNil(t, guardado)
		assert.Equal(t, guardado.Code, enviados[0].Code)
		email.AssertExpectations(t)
	})

	t.Run("vence el código en un año", func(t *testing.T) {
		store := new(AccessKeyStoreMock)
		email := newEmailQueAcepta(nil)
		service := newTestService(store, email)

		store.On("Create", mock.Anything, mock.Anything).Return(nil).Once()

		accessKey, err := service.IssueForUser(context.Background(), testMunicipalID, testEmail)

		require.NoError(t, err)
		require.NotNil(t, accessKey.ExpiresAt)

		restante := accessKey.ExpiresAt.Sub(time.Now())
		assert.Greater(t, restante, 364*24*time.Hour, "no debe vencer antes de un año")
		assert.LessOrEqual(t, restante, 366*24*time.Hour, "no debe vencer despues de un año")
		email.esperar(t, 1)
	})

	t.Run("genera un código distinto en cada emisión", func(t *testing.T) {
		store := new(AccessKeyStoreMock)
		email := newEmailQueAcepta(nil)
		service := newTestService(store, email)

		store.On("Create", mock.Anything, mock.Anything).Return(nil).Twice()

		primera, err := service.IssueForUser(context.Background(), testMunicipalID, testEmail)
		require.NoError(t, err)

		segunda, err := service.IssueForUser(context.Background(), testMunicipalID, testEmail)
		require.NoError(t, err)

		assert.NotEqual(t, primera.Code, segunda.Code, "el código se sortea en cada emisión")

		enviados := email.esperar(t, 2)
		require.Len(t, enviados, 2)

		// Cada correo va en su propia goroutine, así que el orden en que llegan
		// no es el de emisión: lo que importa es que los dos códigos emitidos
		// llegaron por correo.
		assert.ElementsMatch(t,
			[]string{primera.Code, segunda.Code},
			[]string{enviados[0].Code, enviados[1].Code},
		)
		assert.Equal(t, testEmail, enviados[0].To)
		assert.Equal(t, testEmail, enviados[1].To)
	})

	t.Run("no manda correo ni devuelve llave si el store falla", func(t *testing.T) {
		store := new(AccessKeyStoreMock)
		email := newEmailQueAcepta(nil)
		service := newTestService(store, email)

		store.On("Create", mock.Anything, mock.Anything).Return(errStoreBoom).Once()

		accessKey, err := service.IssueForUser(context.Background(), testMunicipalID, testEmail)

		require.ErrorIs(t, err, errStoreBoom)
		assert.Equal(t, model.AccessKey{}, accessKey, "con error se devuelve la llave cero")
		email.sinCorreos(t)
		store.AssertExpectations(t)
	})

	t.Run("un fallo del correo no invalida la emisión", func(t *testing.T) {
		// El correo se manda en una goroutine: su error solo queda en el log,
		// la llave ya está en la base y el municipal ya recibió un 200.
		store := new(AccessKeyStoreMock)
		email := newEmailQueAcepta(errCorreoBoom)
		service := newTestService(store, email)

		store.On("Create", mock.Anything, mock.Anything).Return(nil).Once()

		accessKey, err := service.IssueForUser(context.Background(), testMunicipalID, testEmail)

		require.NoError(t, err)
		assert.NotEmpty(t, accessKey.Code)

		enviados := email.esperar(t, 1)
		require.Len(t, enviados, 1)
		assert.ErrorIs(t, enviados[0].Err, errCorreoBoom, "el error queda registrado, no propagado")
		store.AssertExpectations(t)
	})

	t.Run("propaga el error de la base tal cual", func(t *testing.T) {
		store := new(AccessKeyStoreMock)
		email := newEmailQueAcepta(nil)
		service := newTestService(store, email)

		store.On("Create", mock.Anything, mock.Anything).Return(util.ErrNotFound).Once()

		_, err := service.IssueForUser(context.Background(), testMunicipalID, testEmail)

		require.ErrorIs(t, err, util.ErrNotFound)
		email.sinCorreos(t)
	})
}

func TestAccessKeyService_Resend(t *testing.T) {
	t.Run("reenvía el mismo código de la llave, sin crear otra", func(t *testing.T) {
		store := new(AccessKeyStoreMock)
		email := newEmailQueAcepta(nil)
		service := newTestService(store, email)

		accessKey := newTestAccessKey()

		store.On("GetByID", mock.Anything, testAccessKeyID).Return(accessKey, nil).Once()

		require.NoError(t, service.Resend(context.Background(), testAccessKeyID, testEmail))

		enviados := email.esperar(t, 1)
		require.Len(t, enviados, 1)

		assert.Equal(t, testEmail, enviados[0].To)
		assert.Equal(t, accessKey.Code, enviados[0].Code, "el reenvío reutiliza el código vigente")
		require.NotNil(t, accessKey.ExpiresAt)
		assert.True(t, enviados[0].ExpiresAt.Equal(*accessKey.ExpiresAt),
			"no se renueva el vencimiento: reenviar no extiende la vida del código")

		store.AssertExpectations(t)
		store.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	})

	t.Run("no manda correo si la llave no existe", func(t *testing.T) {
		store := new(AccessKeyStoreMock)
		email := newEmailQueAcepta(nil)
		service := newTestService(store, email)

		// Así lo devuelve el store real: GetByID no traduce sql.ErrNoRows.
		store.On("GetByID", mock.Anything, testAccessKeyID).
			Return(model.AccessKey{}, sql.ErrNoRows).Once()

		err := service.Resend(context.Background(), testAccessKeyID, testEmail)

		require.ErrorIs(t, err, sql.ErrNoRows)
		email.sinCorreos(t)
		store.AssertExpectations(t)
	})

	t.Run("propaga otros errores del store", func(t *testing.T) {
		store := new(AccessKeyStoreMock)
		email := newEmailQueAcepta(nil)
		service := newTestService(store, email)

		store.On("GetByID", mock.Anything, testAccessKeyID).
			Return(model.AccessKey{}, errStoreBoom).Once()

		err := service.Resend(context.Background(), testAccessKeyID, testEmail)

		require.ErrorIs(t, err, errStoreBoom)
		email.sinCorreos(t)
	})

	t.Run("BUG: entra en panic si la llave no tiene vencimiento", func(t *testing.T) {
		// access_keys.expires_at es nullable (migrations/000001) y Resend lo
		// desreferencia a ciegas en service.go:63, asi que una llave emitida
		// sin vencimiento tumba el request (el Recoverer de chi lo convierte en
		// un 500 sin cuerpo). Este test fija el comportamiento ACTUAL.
		store := new(AccessKeyStoreMock)
		email := newEmailQueAcepta(nil)
		service := newTestService(store, email)

		accessKey := newTestAccessKey()
		accessKey.ExpiresAt = nil

		store.On("GetByID", mock.Anything, testAccessKeyID).Return(accessKey, nil).Once()

		assert.Panics(t, func() {
			_ = service.Resend(context.Background(), testAccessKeyID, testEmail)
		})
		email.sinCorreos(t)
	})
}

func TestAccessKeyService_RevokePremature(t *testing.T) {
	t.Run("revoca el código a nombre del municipal", func(t *testing.T) {
		store := new(AccessKeyStoreMock)
		email := newEmailQueAcepta(nil)
		service := newTestService(store, email)

		var revocada RevokeCodeData

		store.On("Revoke", mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) { revocada = args.Get(1).(RevokeCodeData) }).
			Return(nil).Once()

		require.NoError(t, service.RevokePremature(context.Background(), testMunicipalID, testAccessKeyCod))

		assert.Equal(t, testMunicipalID, revocada.RevokeBy)
		assert.Equal(t, testAccessKeyCod, revocada.code)
		store.AssertExpectations(t)
	})

	t.Run("propaga util.ErrNotFound si el código no existe", func(t *testing.T) {
		// Es lo que devuelve el store cuando el UPDATE no matchea filas, y lo
		// que el handler traduce a 404.
		store := new(AccessKeyStoreMock)
		email := newEmailQueAcepta(nil)
		service := newTestService(store, email)

		store.On("Revoke", mock.Anything, mock.Anything).Return(util.ErrNotFound).Once()

		err := service.RevokePremature(context.Background(), testMunicipalID, "no-existe")

		require.ErrorIs(t, err, util.ErrNotFound)
		store.AssertExpectations(t)
	})

	t.Run("propaga otros errores del store", func(t *testing.T) {
		store := new(AccessKeyStoreMock)
		email := newEmailQueAcepta(nil)
		service := newTestService(store, email)

		store.On("Revoke", mock.Anything, mock.Anything).Return(errStoreBoom).Once()

		err := service.RevokePremature(context.Background(), testMunicipalID, testAccessKeyCod)

		require.ErrorIs(t, err, errStoreBoom)
		store.AssertExpectations(t)
	})

	t.Run("revocar no manda correo", func(t *testing.T) {
		store := new(AccessKeyStoreMock)
		email := newEmailQueAcepta(nil)
		service := newTestService(store, email)

		store.On("Revoke", mock.Anything, mock.Anything).Return(nil).Once()

		require.NoError(t, service.RevokePremature(context.Background(), testMunicipalID, testAccessKeyCod))
		email.sinCorreos(t)
	})
}
