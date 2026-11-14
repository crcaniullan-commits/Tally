package incomes

import "errors"

var (
	// ErrorFechaInvalida: la fecha no viene en el formato AAAA-MM-DD.
	ErrorFechaInvalida = errors.New("la fecha debe tener el formato AAAA-MM-DD")
	// ErrorSinCambios: el payload de actualización no trae ningún campo.
	ErrorSinCambios = errors.New("no se enviaron campos para actualizar")
	// ErrorMovimientoNoEncontrado: el ingreso no existe o no es de la empresa.
	ErrorMovimientoNoEncontrado = errors.New("el ingreso no fue encontrado")
)
