package accesskeys

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/internal/util"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// AccessKeyStoreMock es un mock de la capa de persistencia (AccessKeyStore)
// usado para probar AccessKeyService sin tocar la base de datos.
type AccessKeyStoreMock struct {
	mock.Mock
}

// Create recibe el puntero que el service armó: el test puede inspeccionar el
// AccessKey que quedó guardado, y el service espera que Create le rellene el id
// que devuelve la base.
func (m *AccessKeyStoreMock) Create(ctx context.Context, accessKey *model.AccessKey) error {
	args := m.Called(ctx, accessKey)
	return args.Error(0)
}

func (m *AccessKeyStoreMock) GetByID(ctx context.Context, accessKeyID uuid.UUID) (model.AccessKey, error) {
	args := m.Called(ctx, accessKeyID)

	accessKey, _ := args.Get(0).(model.AccessKey)

	return accessKey, args.Error(1)
}

func (m *AccessKeyStoreMock) Revoke(ctx context.Context, revokeData RevokeCodeData) error {
	args := m.Called(ctx, revokeData)
	return args.Error(0)
}

// correoEnviado es una llamada a SendActivationCode ya terminada.
type correoEnviado struct {
	To        string
	Code      string
	ExpiresAt time.Time
	Err       error
}

// SendEmailMock mockea el cliente de correo (internal/email). El service manda
// el correo en una goroutine, asi que el mock publica cada envio en un canal:
// los tests esperan esa señal en vez de dormir y no dependen del scheduler.
type SendEmailMock struct {
	mock.Mock

	mu        sync.Mutex
	recibidos []correoEnviado

	avisos chan struct{}
}

func newSendEmailMock() *SendEmailMock {
	return &SendEmailMock{avisos: make(chan struct{}, 16)}
}

// newEmailQueAcepta devuelve un cliente de correo que acepta cualquier envio y
// devuelve err. La expectativa va con Maybe() a proposito: a estos tests les
// interesa el contenido del correo (eso lo asserta esperar) y no que se mande
// exactamente una vez. Igual hay que llamar a esperar/sinCorreos en todo test
// que dispare un envio, porque la goroutine del service no se puede deixar
// correr despues de que el test termino.
func newEmailQueAcepta(err error) *SendEmailMock {
	email := newSendEmailMock()
	email.On("SendActivationCode", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(err).Maybe()

	return email
}

// El aviso del canal se publica despues de que testify anotó la llamada: al
// recibirlo, el test ya puede usar AssertExpectations sin racing con la
// goroutine del service.
func (m *SendEmailMock) SendActivationCode(ctx context.Context, to, code string, expiresAt time.Time) error {
	err := m.Called(ctx, to, code, expiresAt).Error(0)

	m.mu.Lock()
	m.recibidos = append(m.recibidos, correoEnviado{
		To:        to,
		Code:      code,
		ExpiresAt: expiresAt,
		Err:       err,
	})
	m.mu.Unlock()

	m.avisos <- struct{}{}

	return err
}

// esperar bloquea hasta que el service haya mandado n correos y devuelve los
// CorreosEnviados registrados. Si no llegan dentro del timeout, falla el test
// con un mensaje util en vez de colgarse.
func (m *SendEmailMock) esperar(t *testing.T, n int) []correoEnviado {
	t.Helper()

	for i := 0; i < n; i++ {
		select {
		case <-m.avisos:
		case <-time.After(2 * time.Second):
			t.Fatalf("se esperaban %d correos y solo llegaron %d", n, i)
		}
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	return append([]correoEnviado(nil), m.recibidos...)
}

// sinCorreos verifica que no se mandó ningún correo, esperando un rato lo
// suficiente para que una goroutine que ya salió del service se manifieste.
func (m *SendEmailMock) sinCorreos(t *testing.T) {
	t.Helper()

	select {
	case <-m.avisos:
		t.Fatal("se mandó un correo y no debía")
	case <-time.After(100 * time.Millisecond):
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.recibidos) != 0 {
		t.Fatalf("no debía mandarse ningún correo, llegaron %d", len(m.recibidos))
	}
}

// ServiceAccessKeyMock es un mock de la capa de servicio (ServiceAccessKey)
// usado para probar AccessKeysHandler de forma aislada.
type ServiceAccessKeyMock struct {
	mock.Mock
}

func (m *ServiceAccessKeyMock) IssueForUser(ctx context.Context, municipalID uuid.UUID, email string) (model.AccessKey, error) {
	args := m.Called(ctx, municipalID, email)

	accessKey, _ := args.Get(0).(model.AccessKey)

	return accessKey, args.Error(1)
}

func (m *ServiceAccessKeyMock) Resend(ctx context.Context, accessKeyID uuid.UUID, email string) error {
	args := m.Called(ctx, accessKeyID, email)
	return args.Error(0)
}

func (m *ServiceAccessKeyMock) RevokePremature(ctx context.Context, municipalID uuid.UUID, code string) error {
	args := m.Called(ctx, municipalID, code)
	return args.Error(0)
}

var (
	testMunicipalID  = uuid.MustParse("1c2d3e4f-5a6b-7c8d-9e0f-1a2b3c4d5e6f")
	testEmprendedor  = uuid.MustParse("6f1a1b3c-2d4e-4f60-8a9b-0c1d2e3f4a5b")
	testAccessKeyID  = uuid.MustParse("7a8b9c0d-1e2f-4a3b-8c9d-0e1f2a3b4c5d")
	testAccessKeyCod = "CODIGO-DE-PRUEBA"
	testEmail        = "emprendedor@correo.cl"
)

// newTestAccessKey construye una llave vigente, no canjeada y no revocada: el
// estado en el que un codigo se puede canjear. Los tests que necesitan otro
// estado parten de aca y cambian el campo correspondiente.
func newTestAccessKey() model.AccessKey {
	expire := time.Now().AddDate(0, 0, 365)

	return model.AccessKey{
		ID:        testAccessKeyID,
		Code:      testAccessKeyCod,
		CreatedBy: testMunicipalID,
		ExpiresAt: &expire,
		CreatedAt: time.Now(),
	}
}

// newTestEmailPayload arma el payload de envío de código con un correo válido.
func newTestEmailPayload() EmailPayload {
	return EmailPayload{Email: testEmail}
}

// newTestMunicipalUser es el usuario del contexto en estos endpoints: las tres
// rutas de accessKeys están detrás de CheckOwnership(UserRoleMunicipal).
func newTestMunicipalUser() *model.User {
	return &model.User{
		ID:     testMunicipalID,
		Email:  "municipal@correo.cl",
		Nombre: "Municipal de Prueba",
		Role:   util.UserRoleMunicipal,
	}
}

// Guardas de firma: si estas interfaces cambian, esto deja de compilar.
var (
	_ AccessKeyStore   = new(AccessKeyStoreMock)
	_ SendEmail        = new(SendEmailMock)
	_ ServiceAccessKey = new(ServiceAccessKeyMock)
)
