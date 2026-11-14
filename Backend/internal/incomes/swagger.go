package incomes

// AgregarIngreso
//
//	@Summary		Registrar ingreso
//	@Description	Registra un nuevo ingreso asociado a la empresa autenticada. La respuesta viene envuelta en {"datos": ...} con el ingreso creado, incluyendo el id y las fechas que asigna la base. "monto" es un entero en pesos chilenos (CLP, sin decimales). "categoria_id" es opcional (la columna es nullable) y debe ser el UUID de una categoría global o de la empresa; si no existe o pertenece a otra empresa responde 400.
//	@Tags			Ingresos
//	@Accept			json
//	@Produce		json
//	@Param			empresaID	path		string				true	"ID de la empresa"
//	@Param			payload		body		PayloadIngreso		true	"payload"
//	@Success		201			{object}	model.Income		"ingreso creado"
//	@Failure		400			{object}	map[string]string	"payload del ingreso erróneo"
//	@Failure		500			{object}	map[string]string	"error interno del servidor"
//	@Security		ApiKeyAuth
//	@Router			/aplicacion/empresas/{empresaID}/ingresos	[post]
func AgregarIngreso() {}

// ActualizarIngreso
//
//	@Summary		Editar ingreso
//	@Description	Edita un ingreso de la empresa autenticada. Solo se modifican los campos que vienen en el payload: lo que se omite se queda como estaba. "fecha" (AAAA-MM-DD) permite corregir la fecha del movimiento; un formato inválido responde 400. Si el ingreso no existe o es de otra empresa responde 404. La respuesta viene envuelta en {"datos": ...} con el ingreso ya actualizado.
//	@Tags			Ingresos
//	@Accept			json
//	@Produce		json
//	@Param			empresaID	path		string						true	"ID de la empresa"
//	@Param			ingresoID	path		string						true	"ID del ingreso"
//	@Param			payload		body		PayloadActualizarIngreso	true	"campos a actualizar"
//	@Success		200			{object}	model.Income				"ingreso actualizado"
//	@Failure		400			{object}	map[string]string			"payload erróneo, fecha inválida o categoría ajena"
//	@Failure		404			{object}	map[string]string			"ingreso no encontrado"
//	@Failure		500			{object}	map[string]string			"error interno del servidor"
//	@Security		ApiKeyAuth
//	@Router			/aplicacion/empresas/{empresaID}/ingresos/{ingresoID}	[patch]
func Actualizar() {}

// EliminarIngreso
//
//	@Summary		Eliminar ingreso
//	@Description	Elimina un ingreso por su ID, siempre que pertenezca a la empresa autenticada. La respuesta viene envuelta en {"datos": ...} con el mensaje de confirmación.
//	@Tags			Ingresos
//	@Produce		json
//	@Param			empresaID	path		string				true	"ID de la empresa"
//	@Param			ingresoID	path		string				true	"ID del ingreso"
//	@Success		200			{object}	string				"ingreso eliminado"
//	@Failure		400			{object}	map[string]string	"ingresoID invalido"
//	@Failure		404			{object}	map[string]string	"ingreso no encontrado"
//	@Failure		500			{object}	map[string]string	"error interno del servidor"
//	@Security		ApiKeyAuth
//	@Router			/aplicacion/empresas/{empresaID}/ingresos/{ingresoID}	[delete]
func Eliminar() {}

// ObtenerIngresosDeEmpresa
//
//	@Summary		Listar ingresos de la empresa
//	@Description	Devuelve una página de los ingresos de la empresa autenticada, ordenados del más reciente al más antiguo. La respuesta viene envuelta en {"datos": [...]}. Paginá con limite (entre 1 y 30, por defecto 30) y desplazamiento; el rango se acota con desde y hasta en formato AAAA-MM-DD, y por categoría con categoria_id (el UUID de la categoría, no su nombre). Un parámetro mal formado responde 400. Los montos son enteros en pesos chilenos (CLP, sin decimales).
//	@Tags			Ingresos
//	@Produce		json
//	@Param			empresaID		path		string						true	"ID de la empresa"
//	@Param			limite			query		int							false	"cantidad maxima de ingresos a devolver (1-30, default 30)"
//	@Param			desplazamiento	query		int							false	"ingresos a saltar, para pedir la pagina siguiente"
//	@Param			desde			query		string						false	"fecha minima del rango (AAAA-MM-DD)"
//	@Param			hasta			query		string						false	"fecha maxima del rango (AAAA-MM-DD)"
//	@Param			categoria_id	query		string						false	"UUID de la categoria por la que filtrar"
//	@Success		200				{object}	map[string][]model.Income	"pagina de ingresos"
//	@Failure		400				{object}	map[string]string			"parametros de paginacion erroneos"
//	@Failure		500				{object}	map[string]string			"error interno del servidor"
//	@Security		ApiKeyAuth
//	@Router			/aplicacion/empresas/{empresaID}/ingresos	[get]
func ObtenerIngresosDeEmpresa() {}
