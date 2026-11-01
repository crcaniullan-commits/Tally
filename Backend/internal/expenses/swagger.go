package expenses

// Create
//
//	@Summary		Registrar gasto
//	@Description	Registra un nuevo gasto asociado a la empresa autenticada. La respuesta viene envuelta en {"data": ...} con el gasto creado, incluyendo el id y las fechas que asigna la base. "monto" es un entero en pesos chilenos (CLP, sin decimales). "category_id" es obligatorio y debe ser el UUID de una categoría global o de la empresa; si no existe o pertenece a otra empresa responde 400. El payload no acepta "fecha": la asigna la base de datos con CURRENT_DATE.
//	@Tags			Expenses
//	@Accept			json
//	@Produce		json
//	@Param			empresaID	path		string				true	"ID de la empresa"
//	@Param			payload		body		ExpensePayload		true	"payload"
//	@Success		201			{object}	model.Expense		"gasto creado"
//	@Failure		400			{object}	map[string]string	"payload del gasto erroneo"
//	@Failure		500			{object}	map[string]string	"error interno del servidor"
//	@Security		ApiKeyAuth
//	@Router			/app/empresas/{empresaID}/expenses	[post]
func Create() {}

// Delete
//
//	@Summary		Eliminar gasto
//	@Description	Elimina un gasto por su ID, siempre que pertenezca a la empresa autenticada. La respuesta viene envuelta en {"data": ...} con el mensaje de confirmación.
//	@Tags			Expenses
//	@Produce		json
//	@Param			empresaID	path		string				true	"ID de la empresa"
//	@Param			expenseID	path		string				true	"ID del gasto"
//	@Success		200			{object}	string				"gasto eliminado"
//	@Failure		400			{object}	map[string]string	"expenseID invalido"
//	@Failure		404			{object}	map[string]string	"gasto no encontrado"
//	@Failure		500			{object}	map[string]string	"error interno del servidor"
//	@Security		ApiKeyAuth
//	@Router			/app/empresas/{empresaID}/expenses/{expenseID}	[delete]
func Delete() {}

// GetExpensesOfEmpresa
//
//	@Summary		Listar gastos de la empresa
//	@Description	Devuelve una página de los gastos de la empresa autenticada, ordenados del más reciente al más antiguo. La respuesta viene envuelta en {"data": [...]}. Paginá con limit (entre 1 y 30, por defecto 30) y offset; el rango se acota con since y until en formato AAAA-MM-DD, y por categoría con category_id (el UUID de la categoría, no su nombre). Un parámetro mal formado responde 400. Los montos son enteros en pesos chilenos (CLP, sin decimales).
//	@Tags			Expenses
//	@Produce		json
//	@Param			empresaID	path		string						true	"ID de la empresa"
//	@Param			limit		query		int							false	"cantidad maxima de gastos a devolver (1-30, default 30)"
//	@Param			offset		query		int							false	"gastos a saltar, para pedir la pagina siguiente"
//	@Param			since		query		string						false	"fecha minima del rango (AAAA-MM-DD)"
//	@Param			until		query		string						false	"fecha maxima del rango (AAAA-MM-DD)"
//	@Param			category_id	query		string						false	"UUID de la categoria por la que filtrar"
//	@Success		200			{object}	map[string][]model.Expense	"pagina de gastos"
//	@Failure		400			{object}	map[string]string			"parametros de paginacion erroneos"
//	@Failure		500			{object}	map[string]string			"error interno del servidor"
//	@Security		ApiKeyAuth
//	@Router			/app/empresas/{empresaID}/expenses	[get]
func GetExpensesOfEmpresa() {}
