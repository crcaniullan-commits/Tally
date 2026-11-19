package goals

// Create
//
//	@Summary		Crear meta
//	@Description	Crea un meta asociada a la empresa autenticada. "periodo" acepta dia, semana o mes. "monto_meta" es un entero en pesos chilenos (CLP, sin decimales) y debe ser mayor que cero. La respuesta viene envuelta en {"datos": ...} con la meta creada.
//	@Tags			Metas
//	@Accept			json
//	@Produce		json
//	@Param			empresaID	path		string					true	"ID de la empresa"
//	@Param			payload		body		PayloadCrearMeta		true	"payload"
//	@Success		201			{object}	map[string]model.Goal	"objetivo creado"
//	@Failure		400			{object}	map[string]string		"empresaID inválido o payload inválido"
//	@Failure		401			{object}	map[string]string		"no autorizado"
//	@Failure		403			{object}	map[string]string		"prohibido (se requiere rol usuario)"
//	@Failure		404			{object}	map[string]string		"empresa no encontrada o ajena"
//	@Failure		500			{object}	map[string]string		"error interno del servidor"
//	@Security		ApiKeyAuth
//	@Router			/aplicacion/empresas/{empresaID}/metas	[post]
func Create() {}

// ObtenerMetas
//
//	@Summary		Listar metas
//	@Description	Devuelve todas las metas de la empresa autenticada. La respuesta viene envuelta en {"datos": [...]}. Una meta no tiene estado: su progreso se calcula contra los ingresos del periodo actual y se reinicia con cada dia, semana o mes.
//	@Tags			Metas
//	@Produce		json
//	@Param			empresaID	path		string					true	"ID de la empresa"
//	@Success		200			{object}	map[string][]model.Goal	"metas de la empresa"
//	@Failure		400			{object}	map[string]string		"empresaID inválido"
//	@Failure		401			{object}	map[string]string		"no autorizado"
//	@Failure		403			{object}	map[string]string		"prohibido (se requiere rol usuario)"
//	@Failure		404			{object}	map[string]string		"empresa no encontrada o ajena"
//	@Failure		500			{object}	map[string]string		"error interno del servidor"
//	@Security		ApiKeyAuth
//	@Router			/aplicacion/empresas/{empresaID}/metas	[get]
func ObtenerMetas() {}

// ObtenerMeta
//
//	@Summary		Obtener meta
//	@Description	Devuelve una meta de la empresa autenticada por su ID. La respuesta viene envuelta en {"datos": ...}.
//	@Tags			Metas
//	@Produce		json
//	@Param			empresaID	path		string					true	"ID de la empresa"
//	@Param			metaID		path		string					true	"ID de la meta"
//	@Success		200			{object}	map[string]model.Goal	"meta encontrada"
//	@Failure		400			{object}	map[string]string		"empresaID o metaID invalido"
//	@Failure		401			{object}	map[string]string		"no autorizado"
//	@Failure		403			{object}	map[string]string		"prohibido (se requiere rol usuario)"
//	@Failure		404			{object}	map[string]string		"empresa ajena o meta no encontrada"
//	@Failure		500			{object}	map[string]string		"error interno del servidor"
//	@Security		ApiKeyAuth
//	@Router			/aplicacion/empresas/{empresaID}/metas/{metaID}	[get]
func ObtenerMeta() {}

// EliminarMeta
//
//	@Summary		Eliminar meta
//	@Description	Elimina una meta de la empresa autenticada por su ID. La respuesta viene envuelta en {"datos": ...} con el mensaje de confirmacion.
//	@Tags			Metas
//	@Produce		json
//	@Param			empresaID	path		string				true	"ID de la empresa"
//	@Param			metaID		path		string				true	"ID de la meta"
//	@Success		200			{object}	map[string]string	"meta eliminada"
//	@Failure		400			{object}	map[string]string	"empresaID o metaID invalido"
//	@Failure		401			{object}	map[string]string	"no autorizado"
//	@Failure		403			{object}	map[string]string	"prohibido (se requiere rol usuario)"
//	@Failure		404			{object}	map[string]string	"empresa ajena o meta no encontrada"
//	@Failure		500			{object}	map[string]string	"error interno del servidor"
//	@Security		ApiKeyAuth
//	@Router			/aplicacion/empresas/{empresaID}/metas/{metaID}	[delete]
func EliminarMeta() {}

// ObtenerEstadoMeta
//
//	@Summary		Estado de la meta
//	@Description	Suma los ingresos de la empresa en el periodo actual (dia, semana desde el lunes, o mes) y calcula el avance contra "monto_meta". La respuesta viene envuelta en {"datos": ...} con monto_actual, porcentaje_actual, monto_pendiente y porcentaje_faltante. Los montos son enteros en pesos chilenos (CLP, sin decimales).
//	@Tags			Metas
//	@Produce		json
//	@Param			empresaID	path		string					true	"ID de la empresa"
//	@Param			metaID		path		string					true	"ID de la meta"
//	@Success		200			{object}	map[string]EstadoMeta	"estado del objetivo"
//	@Failure		400			{object}	map[string]string		"empresaID o metaID invalido"
//	@Failure		401			{object}	map[string]string		"no autorizado"
//	@Failure		403			{object}	map[string]string		"prohibido (se requiere rol usuario)"
//	@Failure		404			{object}	map[string]string		"empresa ajena o meta no encontrada"
//	@Failure		500			{object}	map[string]string		"error interno del servidor"
//	@Security		ApiKeyAuth
//	@Router			/aplicacion/empresas/{empresaID}/metas/{metaID}/estado	[get]
func ObtenerEstadoMeta() {}
