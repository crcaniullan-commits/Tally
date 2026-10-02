package incomes

// AddIncome
//
//	@Summary		Registrar ingreso
//	@Description	Registra un nuevo ingreso asociado al usuario autenticado. La respuesta viene envuelta en {"data": ...} con el ingreso creado, incluyendo el id y las fechas que asigna la base. "monto" es un entero en pesos chilenos (CLP, sin decimales).
//	@Tags			Incomes
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		IncomePayload		true	"payload"
//	@Success		201		{object}	IncomeStorage		"ingreso creado"
//	@Failure		400		{object}	map[string]string	"payload del ingreso erroneo"
//	@Failure		500		{object}	map[string]string	"error interno del servidor"
//	@Security		ApiKeyAuth
//	@Router			/app/incomes	[post]
func AddIncome() {}

// DeleteIncome
//
//	@Summary		Eliminar ingreso
//	@Description	Elimina un ingreso por su ID, siempre que pertenezca al usuario autenticado. La respuesta viene envuelta en {"data": ...} con el mensaje de confirmación.
//	@Tags			Incomes
//	@Produce		json
//	@Param			incomeID	path		string				true	"ID del ingreso"
//	@Success		200			{object}	string				"ingreso eliminado"
//	@Failure		400			{object}	map[string]string	"incomeID invalido"
//	@Failure		404			{object}	map[string]string	"ingreso no encontrado"
//	@Failure		500			{object}	map[string]string	"error interno del servidor"
//	@Security		ApiKeyAuth
//	@Router			/app/incomes/{incomeID}	[delete]
func Delete() {}

// GetIncomesOfUser
//
//	@Summary		Listar ingresos del usuario
//	@Description	Devuelve una página de los ingresos del usuario autenticado, ordenados del más reciente al más antiguo. La respuesta viene envuelta en {"data": [...]}. Paginá con limit (entre 1 y 30, por defecto 30) y offset; el rango se acota con since y until en formato AAAA-MM-DD. Un parámetro mal formado responde 400. Los montos son enteros en pesos chilenos (CLP, sin decimales).
//	@Tags			Incomes
//	@Produce		json
//	@Param			limit	query		int							false	"cantidad maxima de ingresos a devolver (1-30, default 30)"
//	@Param			offset	query		int							false	"ingresos a saltar, para pedir la pagina siguiente"
//	@Param			since	query		string						false	"fecha minima del rango (AAAA-MM-DD)"
//	@Param			until	query		string						false	"fecha maxima del rango (AAAA-MM-DD)"
//	@Success		200		{object}	map[string][]IncomeStorage	"pagina de ingresos"
//	@Failure		400		{object}	map[string]string			"parametros de paginacion erroneos"
//	@Failure		500		{object}	map[string]string			"error interno del servidor"
//	@Security		ApiKeyAuth
//	@Router			/app/incomes	[get]
func GetIncomesOfUser() {}
