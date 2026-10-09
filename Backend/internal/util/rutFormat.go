package util

import "github.com/crcaniullan-commits/Tally/internal/model"

// RUT y su lógica viven en model; acá solo se re-exportan para no cambiar los
// call sites.
type RUT = model.RUT

var (
	ErrFormatoInvalido = model.ErrFormatoInvalido
	ErrDVInvalido      = model.ErrDVInvalido
)

// ParseRUT convierte un string a un RUT validando SOLO el formato básico.
func ParseRUT(rutRaw string) (RUT, error) {
	return model.ParseRUT(rutRaw)
}
