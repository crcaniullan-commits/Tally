package debtors

// Create
//
//	@Summary		Crear deuda
//	@Description	Crea una deuda asociada a la empresa autenticada. "monto" es un entero en pesos chilenos (CLP, sin decimales) y debe ser mayor que cero. "fecha_limite" se envía como fecha-hora RFC3339. La respuesta viene envuelta en {"data": ...} con la deuda creada.
//	@Tags			Debtors
//	@Accept			json
//	@Produce		json
//	@Param			empresaID	path		string				true	"ID de la empresa"
//	@Param			payload		body		DebtorPayload		true	"payload"
//	@Success		201			{object}	model.Debtor		"deuda creada"
//	@Failure		400			{object}	map[string]string	"payload invalido"
//	@Failure		500			{object}	map[string]string	"error interno del servidor"
//	@Security		ApiKeyAuth
//	@Router			/app/empresas/{empresaID}/debtors	[post]
func Create() {}

// GetDebtorsOfEmpresa
//
//	@Summary		Listar deudas
//	@Description	Devuelve todas las deudas de la empresa autenticada, ordenadas de la más nueva a la más antigua. La respuesta viene envuelta en {"data": [...]}. Si no hay deudas devuelve una lista vacía.
//	@Tags			Debtors
//	@Produce		json
//	@Param			empresaID	path		string						true	"ID de la empresa"
//	@Success		200			{object}	map[string][]model.Debtor	"deudas de la empresa"
//	@Failure		500			{object}	map[string]string			"error interno del servidor"
//	@Security		ApiKeyAuth
//	@Router			/app/empresas/{empresaID}/debtors	[get]
func GetDebtorsOfEmpresa() {}

// Update
//
//	@Summary		Actualizar deuda
//	@Description	Actualiza "nombre", "monto" y "fecha_limite" de una deuda de la empresa autenticada. La descripción no se modifica. "monto" es un entero en pesos chilenos (CLP, sin decimales) y debe ser mayor que cero. La respuesta viene envuelta en {"data": ...}.
//	@Tags			Debtors
//	@Accept			json
//	@Produce		json
//	@Param			empresaID	path		string				true	"ID de la empresa"
//	@Param			debtorID	path		string				true	"ID de la deuda"
//	@Param			payload		body		UpdatePayload		true	"payload"
//	@Success		200			{object}	string				"deuda actualizada"
//	@Failure		400			{object}	map[string]string	"debtorID o payload invalido"
//	@Failure		404			{object}	map[string]string	"deuda no encontrada"
//	@Failure		500			{object}	map[string]string	"error interno del servidor"
//	@Security		ApiKeyAuth
//	@Router			/app/empresas/{empresaID}/debtors/{debtorID}	[patch]
func Update() {}

// PaidDebtor
//
//	@Summary		Marcar deuda como pagada
//	@Description	Marca como pagada una deuda de la empresa autenticada. La operación es idempotente. La respuesta viene envuelta en {"data": ...} con el mensaje de confirmación.
//	@Tags			Debtors
//	@Produce		json
//	@Param			empresaID	path		string				true	"ID de la empresa"
//	@Param			debtorID	path		string				true	"ID de la deuda"
//	@Success		200			{object}	string				"deuda pagada"
//	@Failure		400			{object}	map[string]string	"debtorID invalido"
//	@Failure		404			{object}	map[string]string	"deuda no encontrada"
//	@Failure		500			{object}	map[string]string	"error interno del servidor"
//	@Security		ApiKeyAuth
//	@Router			/app/empresas/{empresaID}/debtors/{debtorID}/paid	[patch]
func PaidDebtor() {}
