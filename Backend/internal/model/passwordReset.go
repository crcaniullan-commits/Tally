package model

import (
	"time"

	"github.com/google/uuid"
)

// CodigoRecuperacion es el código de un solo uso que se manda por correo para
// recuperar la contraseña.
//
// Codigo guarda el hash SHA-256 del código real: el código que el usuario
// recibió nunca se escribe en la base. VerificadoEn marca que el usuario ya
// lo introdujo correctamente; UsadoEn que se consumió (al cambiar la contraseña
// o al pedir uno nuevo). Los dos nil son el estado "código vigente sin usar".
type CodigoRecuperacion struct {
	ID           uuid.UUID  `json:"id"`
	UserID       uuid.UUID  `json:"user_id"`
	Codigo       string     `json:"-"`
	ExpiraEn     time.Time  `json:"expira_en"`
	VerificadoEn *time.Time `json:"verificado_en,omitempty"`
	UsadoEn      *time.Time `json:"usado_en,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}
