package expenses

import "errors"

var (
	// ErrorFechaInvalida: la fecha no viene en el formato AAAA-MM-DD.
	ErrorFechaInvalida = errors.New("la fecha debe tener el formato AAAA-MM-DD")
	// ErrorSinCambios: el payload de actualización no trae ningún campo.
	ErrorSinCambios = errors.New("no se enviaron campos para actualizar")
	// ErrorMovimientoNoEncontrado: el gasto no existe o no es de la empresa.
	ErrorMovimientoNoEncontrado = errors.New("el gasto no fue encontrado")
)
