package categories

import "errors"

var (
	ErrorCategoriaDuplicada = errors.New("La categoría ya existe")
	// ErrorCategoriaConMovimientos: la categoría tiene ingresos o gastos
	// asociados y borrarla se los llevaría por delante (los gastos se borran
	// por ON DELETE CASCADE y los ingresos pierden la categoría).
	ErrorCategoriaConMovimientos = errors.New("la categoría tiene movimientos asociados y no se puede eliminar")
)
