package goals

// Create
//
//	@Summary		Crear objetivo
//	@Description	Crea un objetivo asociado a la empresa autenticada. "periodo" acepta dia, semana o mes. "monto_meta" es un entero en pesos chilenos (CLP, sin decimales) y debe ser mayor que cero. La respuesta viene envuelta en {"data": ...} con el objetivo creado.
//	@Tags			Goals
//	@Accept			json
//	@Produce		json
//	@Param			empresaID	path		string				true	"ID de la empresa"
//	@Param			payload		body		CreateGoalPayload	true	"payload"
//	@Success		201			{object}	model.Goal			"objetivo creado"
//	@Failure		400			{object}	map[string]string	"payload invalido o falta la empresa en el contexto"
//	@Failure		500			{object}	map[string]string	"error interno del servidor"
//	@Security		ApiKeyAuth
//	@Router			/app/empresas/{empresaID}/goal	[post]
func Create() {}

// GetGoals
//
//	@Summary		Listar objetivos
//	@Description	Devuelve todos los objetivos de la empresa autenticada. La respuesta viene envuelta en {"data": [...]}. Un objetivo no tiene estado: su progreso se calcula contra los ingresos del periodo actual y se reinicia con cada dia, semana o mes.
//	@Tags			Goals
//	@Produce		json
//	@Param			empresaID	path		string					true	"ID de la empresa"
//	@Success		200			{object}	map[string][]model.Goal	"objetivos de la empresa"
//	@Failure		500			{object}	map[string]string		"error interno del servidor"
//	@Security		ApiKeyAuth
//	@Router			/app/empresas/{empresaID}/goal	[get]
func GetGoals() {}

// GetGoal
//
//	@Summary		Obtener objetivo
//	@Description	Devuelve un objetivo de la empresa autenticada por su ID. La respuesta viene envuelta en {"data": ...}.
//	@Tags			Goals
//	@Produce		json
//	@Param			empresaID	path		string				true	"ID de la empresa"
//	@Param			goalID		path		string				true	"ID del objetivo"
//	@Success		200			{object}	model.Goal			"objetivo encontrado"
//	@Failure		400			{object}	map[string]string	"goalID invalido"
//	@Failure		404			{object}	map[string]string	"objetivo no encontrado"
//	@Failure		500			{object}	map[string]string	"error interno del servidor"
//	@Security		ApiKeyAuth
//	@Router			/app/empresas/{empresaID}/goal/{goalID}	[get]
func GetGoal() {}

// DeleteGoal
//
//	@Summary		Eliminar objetivo
//	@Description	Elimina un objetivo de la empresa autenticada por su ID. La respuesta viene envuelta en {"data": ...} con el mensaje de confirmacion.
//	@Tags			Goals
//	@Produce		json
//	@Param			empresaID	path		string				true	"ID de la empresa"
//	@Param			goalID		path		string				true	"ID del objetivo"
//	@Success		200			{object}	string				"objetivo eliminado"
//	@Failure		400			{object}	map[string]string	"goalID invalido"
//	@Failure		404			{object}	map[string]string	"objetivo no encontrado"
//	@Failure		500			{object}	map[string]string	"error interno del servidor"
//	@Security		ApiKeyAuth
//	@Router			/app/empresas/{empresaID}/goal/{goalID}	[delete]
func DeleteGoal() {}

// GetGoalState
//
//	@Summary		Estado del objetivo
//	@Description	Suma los ingresos de la empresa en el periodo actual (dia, semana desde el lunes, o mes) y calcula el avance contra "monto_meta". La respuesta viene envuelta en {"data": ...} con monto_actual, porcentaje_actual, monto_pendiente y porcentaje_faltante. Los montos son enteros en pesos chilenos (CLP, sin decimales).
//	@Tags			Goals
//	@Produce		json
//	@Param			empresaID	path		string				true	"ID de la empresa"
//	@Param			goalID		path		string				true	"ID del objetivo"
//	@Success		200			{object}	GoalState			"estado del objetivo"
//	@Failure		400			{object}	map[string]string	"goalID invalido"
//	@Failure		404			{object}	map[string]string	"objetivo no encontrado"
//	@Failure		500			{object}	map[string]string	"error interno del servidor"
//	@Security		ApiKeyAuth
//	@Router			/app/empresas/{empresaID}/goal/{goalID}/state	[get]
func GetGoalState() {}
