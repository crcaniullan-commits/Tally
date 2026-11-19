package expenses

import "errors"

var (
	// ErrorSinCambios: el payload de actualización no trae ningún campo.
	ErrorSinCambios = errors.New("no se enviaron campos para actualizar")
	// ErrorMovimientoNoEncontrado: el gasto no existe o no es de la empresa.
	ErrorMovimientoNoEncontrado = errors.New("el gasto no fue encontrado")
)
