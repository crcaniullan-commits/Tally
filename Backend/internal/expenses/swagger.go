package expenses

// Create
//
//	@Summary		Registrar gasto
//	@Description	Registra un nuevo gasto asociado a la empresa autenticada. La respuesta viene envuelta en {"datos": ...} con el gasto creado, incluyendo el id y las fechas que asigna la base. "monto" es un entero en pesos chilenos (CLP, sin decimales). "payment_method" es obligatorio y es uno de debito, credito, transferencia o efectivo; un valor distinto responde 400. "categoria_id" es obligatorio y debe ser el UUID de una categoría global o de la empresa de tipo "gasto"; si no existe, pertenece a otra empresa o es de tipo "ingreso" responde 400. El payload no acepta "fecha": la asigna la base de datos con CURRENT_DATE.
//	@Tags			Gastos
//	@Accept			json
//	@Produce		json
//	@Param			empresaID	path		string						true	"ID de la empresa"
//	@Param			payload		body		PayloadGasto				true	"payload"
//	@Success		201			{object}	map[string]model.Expense	"gasto creado"
//	@Failure		400			{object}	map[string]string			"empresaID inválido o payload del gasto erróneo"
//	@Failure		401			{object}	map[string]string			"no autorizado"
//	@Failure		403			{object}	map[string]string			"prohibido (se requiere rol usuario)"
//	@Failure		404			{object}	map[string]string			"empresa no encontrada o ajena"
//	@Failure		500			{object}	map[string]string			"error interno del servidor"
//	@Security		ApiKeyAuth
//	@Router			/aplicacion/empresas/{empresaID}/gastos	[post]
func Create() {}

// Update
//
//	@Summary		Editar gasto
//	@Description	Edita un gasto de la empresa autenticada. Solo se modifican los campos que vienen en el payload (monto, payment_method, category_id, descripcion): lo que se omite se queda como estaba. "payment_method", si viene, debe ser uno de debito, credito, transferencia o efectivo. La categoría, si viene, debe ser de tipo "gasto". La fecha del movimiento NO se edita: mandarla en el payload es un 400 por campo desconocido. Si el gasto no existe o es de otra empresa responde 404. La respuesta viene envuelta en {"datos": ...} con el gasto ya actualizado.
//	@Tags			Gastos
//	@Accept			json
//	@Produce		json
//	@Param			empresaID	path		string						true	"ID de la empresa"
//	@Param			gastoID		path		string						true	"ID del gasto"
//	@Param			payload		body		PayloadActualizarGasto		true	"campos a actualizar"
//	@Success		200			{object}	map[string]model.Expense	"gasto actualizado"
//	@Failure		400			{object}	map[string]string			"empresaID o gastoID inválido, payload erróneo, sin campos o categoría ajena"
//	@Failure		401			{object}	map[string]string			"no autorizado"
//	@Failure		403			{object}	map[string]string			"prohibido (se requiere rol usuario)"
//	@Failure		404			{object}	map[string]string			"empresa ajena o gasto no encontrado"
//	@Failure		500			{object}	map[string]string			"error interno del servidor"
//	@Security		ApiKeyAuth
//	@Router			/aplicacion/empresas/{empresaID}/gastos/{gastoID}	[patch]
func Update() {}

// Delete
//
//	@Summary		Eliminar gasto
//	@Description	Elimina un gasto por su ID, siempre que pertenezca a la empresa autenticada. La respuesta viene envuelta en {"datos": ...} con el mensaje de confirmación.
//	@Tags			Gastos
//	@Produce		json
//	@Param			empresaID	path		string				true	"ID de la empresa"
//	@Param			gastoID		path		string				true	"ID del gasto"
//	@Success		200			{object}	map[string]string	"gasto eliminado"
//	@Failure		400			{object}	map[string]string	"empresaID o gastoID invalido"
//	@Failure		401			{object}	map[string]string	"no autorizado"
//	@Failure		403			{object}	map[string]string	"prohibido (se requiere rol usuario)"
//	@Failure		404			{object}	map[string]string	"empresa ajena o gasto no encontrado"
//	@Failure		500			{object}	map[string]string	"error interno del servidor"
//	@Security		ApiKeyAuth
//	@Router			/aplicacion/empresas/{empresaID}/gastos/{gastoID}	[delete]
func Delete() {}

// ObtenerGastosDeEmpresa
//
//	@Summary		Listar gastos de la empresa
//	@Description	Devuelve una página de los gastos de la empresa autenticada, ordenados del más reciente al más antiguo. La respuesta viene envuelta en {"datos": [...]}. Paginá con limite (entre 1 y 30, por defecto 30) y desplazamiento; el rango se acota con desde y hasta en formato AAAA-MM-DD, y por categoría con categoria_id (el UUID de la categoría, no su nombre). Un parámetro mal formado responde 400. Los montos son enteros en pesos chilenos (CLP, sin decimales).
//	@Tags			Gastos
//	@Produce		json
//	@Param			empresaID		path		string						true	"ID de la empresa"
//	@Param			limite			query		int							false	"cantidad maxima de gastos a devolver (1-30, default 30)"
//	@Param			desplazamiento	query		int							false	"gastos a saltar, para pedir la pagina siguiente"
//	@Param			desde			query		string						false	"fecha minima del rango (AAAA-MM-DD)"
//	@Param			hasta			query		string						false	"fecha maxima del rango (AAAA-MM-DD)"
//	@Param			categoria_id	query		string						false	"UUID de la categoria por la que filtrar"
//	@Success		200				{object}	map[string][]model.Expense	"pagina de gastos"
//	@Failure		400				{object}	map[string]string			"empresaID inválido o parametros de paginacion erroneos"
//	@Failure		401				{object}	map[string]string			"no autorizado"
//	@Failure		403				{object}	map[string]string			"prohibido (se requiere rol usuario)"
//	@Failure		404				{object}	map[string]string			"empresa no encontrada o ajena"
//	@Failure		500				{object}	map[string]string			"error interno del servidor"
//	@Security		ApiKeyAuth
//	@Router			/aplicacion/empresas/{empresaID}/gastos	[get]
func ObtenerGastosDeEmpresa() {}
