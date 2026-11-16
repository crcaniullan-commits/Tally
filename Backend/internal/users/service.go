package users

import (
	"context"

	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/internal/utilidades"
	"github.com/google/uuid"
)

type RepositorioUsuarios interface {
	Actualizar(context.Context, CambiosUsuario, uuid.UUID) error
	Eliminar(context.Context, uuid.UUID) error
	ObtenerPorRUT(context.Context, string) (*model.User, error)
}

// CambiosUsuario son los campos a modificar de un usuario. Un puntero nil
// significa "no tocar este campo".
//
// El store NO reescribe la fila completa: aplica solo lo que viene acá. Eso
// importa porque el usuario del contexto puede venir del cache (ver
// internal/cache), y en ese caso model.User.PasswordHash llega vacío (el
// cache serializa a JSON y ese campo tiene json:"-"). Si el store
// reescribiera la fila entera con ese objeto, un PATCH que solo cambia el
// nombre escribiría un password_hash vacío en una columna NOT NULL.
type CambiosUsuario struct {
	Nombre         *string
	PasswordHash   []byte
	CambiaPassword bool
}

// Vacio indica que no hay nada que actualizar.
func (c CambiosUsuario) Vacio() bool {
	return c.Nombre == nil && !c.CambiaPassword
}

type invalidarUsuario interface {
	Invalidar(ctx context.Context, id uuid.UUID)
}

type ServicioUsuarios struct {
	store       RepositorioUsuarios
	invalidador invalidarUsuario
}

func NuevoServicioUsuarios(store RepositorioUsuarios, inv invalidarUsuario) *ServicioUsuarios {
	return &ServicioUsuarios{store: store, invalidador: inv}
}

// hashPassword hashea una contraseña en claro con la misma política que usa
// el modelo (bcrypt, mismo costo). Se usa un model.User descartable porque
// el tipo del hash es privado de model.
func hashPassword(texto string) ([]byte, error) {
	var auxiliar model.User

	if err := auxiliar.PasswordHash.Set(texto); err != nil {
		return nil, err
	}

	return auxiliar.PasswordHash.GetHash(), nil
}

// Actualizar aplica los cambios del payload. El usuario del contexto solo se
// usa para saber el ID: los valores nuevos salen del payload, no del objeto
// (que puede venir cacheado y desactualizado).
func (s *ServicioUsuarios) Actualizar(ctx context.Context, user *model.User, payload *PayloadActualizarUsuario) error {
	cambios := CambiosUsuario{}

	if payload.Password != "" {
		hash, err := hashPassword(payload.Password)
		if err != nil {
			return err
		}

		cambios.PasswordHash = hash
		cambios.CambiaPassword = true
	}

	if payload.Nombre != "" {
		cambios.Nombre = &payload.Nombre
	}

	if cambios.Vacio() {
		// Payload vacío: no hay nada que escribir. La fila no se toca.
		return nil
	}

	if err := s.store.Actualizar(ctx, cambios, user.ID); err != nil {
		return err
	}

	// El usuario cambió: la entrada cacheada quedó vieja.
	s.invalidador.Invalidar(ctx, user.ID)

	return nil
}

func (s *ServicioUsuarios) Eliminar(ctx context.Context, userID uuid.UUID) error {
	if err := s.store.Eliminar(ctx, userID); err != nil {
		return err
	}

	s.invalidador.Invalidar(ctx, userID)

	return nil
}

func (s *ServicioUsuarios) ObtenerPorRUT(ctx context.Context, rut utilidades.RUT) (*model.User, error) {
	user, err := s.store.ObtenerPorRUT(ctx, rut.String())

	if err != nil {
		return nil, err
	}
	return user, nil
}
