package debtors

// Create
//
//	@Summary		Crear deuda
//	@Description	Crea una deuda asociada a la empresa autenticada. "monto" es un entero en pesos chilenos (CLP, sin decimales) y debe ser mayor que cero. "fecha_limite" se envía como fecha-hora RFC3339. La respuesta viene envuelta en {"datos": ...} con la deuda creada.
//	@Tags			Deudas
//	@Accept			json
//	@Produce		json
//	@Param			empresaID	path		string				true	"ID de la empresa"
//	@Param			payload		body		PayloadDeudor		true	"payload"
//	@Success		201			{object}	model.Debtor		"deuda creada"
//	@Failure		400			{object}	map[string]string	"payload inválido"
//	@Failure		500			{object}	map[string]string	"error interno del servidor"
//	@Security		ApiKeyAuth
//	@Router			/aplicacion/empresas/{empresaID}/deudas	[post]
func Create() {}

// ObtenerDeudoresDeEmpresa
//
//	@Summary		Listar deudas
//	@Description	Devuelve todas las deudas de la empresa autenticada, ordenadas de la más nueva a la más antigua. La respuesta viene envuelta en {"datos": [...]}. Si no hay deudas devuelve una lista vacía.
//	@Tags			Deudas
//	@Produce		json
//	@Param			empresaID	path		string						true	"ID de la empresa"
//	@Success		200			{object}	map[string][]model.Debtor	"deudas de la empresa"
//	@Failure		500			{object}	map[string]string			"error interno del servidor"
//	@Security		ApiKeyAuth
//	@Router			/aplicacion/empresas/{empresaID}/deudas	[get]
func ObtenerDeudoresDeEmpresa() {}

// Update
//
//	@Summary		Actualizar deuda
//	@Description	Actualiza "nombre", "monto" y "fecha_limite" de una deuda de la empresa autenticada. La descripción no se modifica. "monto" es un entero en pesos chilenos (CLP, sin decimales) y debe ser mayor que cero. La respuesta viene envuelta en {"datos": ...}.
//	@Tags			Deudas
//	@Accept			json
//	@Produce		json
//	@Param			empresaID	path		string						true	"ID de la empresa"
//	@Param			deudorID	path		string						true	"ID de la deuda"
//	@Param			payload		body		PayloadActualizacionDeudor	true	"payload"
//	@Success		200			{object}	string						"deuda actualizada"
//	@Failure		400			{object}	map[string]string			"deudorID o payload inválido"
//	@Failure		404			{object}	map[string]string			"deuda no encontrada"
//	@Failure		500			{object}	map[string]string			"error interno del servidor"
//	@Security		ApiKeyAuth
//	@Router			/aplicacion/empresas/{empresaID}/deudas/{deudorID}	[patch]
func Update() {}

// MarcarDeudorPagado
//
//	@Summary		Marcar deuda como pagada
//	@Description	Marca como pagada una deuda de la empresa autenticada. La operación es idempotente. La respuesta viene envuelta en {"datos": ...} con el mensaje de confirmación.
//	@Tags			Deudas
//	@Produce		json
//	@Param			empresaID	path		string				true	"ID de la empresa"
//	@Param			deudorID	path		string				true	"ID de la deuda"
//	@Success		200			{object}	string				"deuda pagada"
//	@Failure		400			{object}	map[string]string	"deudorID invalido"
//	@Failure		404			{object}	map[string]string	"deuda no encontrada"
//	@Failure		500			{object}	map[string]string	"error interno del servidor"
//	@Security		ApiKeyAuth
//	@Router			/aplicacion/empresas/{empresaID}/deudas/{deudorID}/pagada	[patch]
func MarcarDeudorPagado() {}
