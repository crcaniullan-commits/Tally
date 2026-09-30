package incomes

// AddIncome
//
//	@Summary		Registrar ingreso
//	@Description	Registra un nuevo ingreso asociado al usuario autenticado. La respuesta viene envuelta en {"data": ...} con el ingreso creado, incluyendo el id y las fechas que asigna la base. "monto" es un entero en pesos chilenos (CLP, sin decimales).
//	@Tags			Incomes
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		IncomePayload	true	"payload"
//	@Success		201		{object}	IncomeStorage	"ingreso creado"
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
//	@Description	Devuelve todos los ingresos del usuario autenticado. La respuesta viene envuelta en {"data": [...]}. Si el usuario no tiene ingresos, "data" es una lista vacia. Los montos son enteros en pesos chilenos (CLP, sin decimales).
//	@Tags			Incomes
//	@Produce		json
//	@Success		200	{object}	map[string][]IncomeStorage	"lista de ingresos"
//	@Failure		500	{object}	map[string]string			"error interno del servidor"
//	@Security		ApiKeyAuth
//	@Router			/app/incomes	[get]
func GetIncomesOfUser() {}
