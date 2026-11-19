package model

// CategoryType diferencia si una categoría aplica a los ingresos o a los gastos.
// Una categoría es de uno u otro tipo, nunca de los dos.
type CategoryType string

const (
	CategoryTypeIngreso CategoryType = "ingreso"
	CategoryTypeGasto   CategoryType = "gasto"
)
