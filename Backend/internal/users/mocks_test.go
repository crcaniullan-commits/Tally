package users

import (
	"context"

	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/internal/utilidades"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// RepositorioUsuariosMock es un mock de la capa de persistencia (RepositorioUsuarios) usado para
// probar ServicioUsuarios sin tocar la base de datos.
type RepositorioUsuariosMock struct {
	mock.Mock
}

func (m *RepositorioUsuariosMock) Actualizar(ctx context.Context, cambios CambiosUsuario, userID uuid.UUID) error {
	args := m.Called(ctx, cambios, userID)
	return args.Error(0)
}

func (m *RepositorioUsuariosMock) Eliminar(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *RepositorioUsuariosMock) ObtenerPorRUT(ctx context.Context, rut string) (*model.User, error) {
	args := m.Called(ctx, rut)

	user, _ := args.Get(0).(*model.User)

	return user, args.Error(1)
}

// InvalidadorUsuarioMock es un mock del invalidador de cache
// (invalidarUsuario) usado para verificar que el service invalida la entrada
// cacheada del usuario cuando se borra la cuenta.
type InvalidadorUsuarioMock struct {
	mock.Mock
}

func (m *InvalidadorUsuarioMock) Invalidar(ctx context.Context, id uuid.UUID) {
	m.Called(ctx, id)
}

// newInvalidadorUsuarioMock devuelve un invalidador que acepta cualquier
// llamada: los tests que no miran el cache no tienen que configurar nada, y
// los que sí lo verifican usan AssertCalled/AssertNotCalled.
func newInvalidadorUsuarioMock() *InvalidadorUsuarioMock {
	m := new(InvalidadorUsuarioMock)
	m.On("Invalidar", mock.Anything, mock.Anything).Return().Maybe()

	return m
}

// IServiceUsuariosMock es un mock de la capa de servicio (IServiceUsuarios) usado para
// probar HandlerUsuarios de forma aislada.
type IServiceUsuariosMock struct {
	mock.Mock
}

func (m *IServiceUsuariosMock) Actualizar(ctx context.Context, user *model.User, payload *PayloadActualizarUsuario) error {
	args := m.Called(ctx, user, payload)
	return args.Error(0)
}

func (m *IServiceUsuariosMock) Eliminar(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *IServiceUsuariosMock) ObtenerPorRUT(ctx context.Context, rut utilidades.RUT) (*model.User, error) {
	args := m.Called(ctx, rut)

	user, _ := args.Get(0).(*model.User)

	return user, args.Error(1)
}

var testUserID = uuid.MustParse("6f1a1b3c-2d4e-4f60-8a9b-0c1d2e3f4a5b")

// newTestUser construye un usuario válido y reutilizable en los tests.
func newTestUser() *model.User {
	return &model.User{
		ID:     testUserID,
		Email:  "emprendedor@correo.cl",
		Nombre: "Nombre Original",
		Role:   utilidades.RolUsuario,
		Rut:    utilidades.RUT{Cuerpo: 19234567, DV: "K"},
	}
}
