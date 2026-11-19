package jobs

// RecordatorioDiario
//
//	@Summary		Ejecutar recordatorio diario de inactividad
//	@Description	Ejecuta bajo demanda el job que detecta usuarios con dispositivos registrados que hoy no han cargado ningún ingreso ni egreso (en hora de Chile), emitiendo la notificación push correspondiente. Responde 200 con mensaje de confirmación.
//	@Tags			Jobs
//	@Produce		json
//	@Success		200	{object}	map[string]string	"recordatorio diario ejecutado"
//	@Failure		401	{object}	map[string]string	"no autorizado"
//	@Failure		500	{object}	map[string]string	"error interno del servidor"
//	@Security		ApiKeyAuth
//	@Router			/aplicacion/jobs/recordatorio-diario	[post]
func RecordatorioDiarioSwagger() {}

// CobrosVencidos
//
//	@Summary		Ejecutar aviso de cobros vencidos
//	@Description	Ejecuta bajo demanda el job que detecta deudores vencidos e impagos agrupados por usuario, emitiendo una notificación consolidada en pesos chilenos. Responde 200 con mensaje de confirmación.
//	@Tags			Jobs
//	@Produce		json
//	@Success		200	{object}	map[string]string	"aviso de cobros vencidos ejecutado"
//	@Failure		401	{object}	map[string]string	"no autorizado"
//	@Failure		500	{object}	map[string]string	"error interno del servidor"
//	@Security		ApiKeyAuth
//	@Router			/aplicacion/jobs/cobros-vencidos	[post]
func CobrosVencidosSwagger() {}

// EjecutarTodos
//
//	@Summary		Ejecutar todas las tareas programadas
//	@Description	Ejecuta secuencialmente el recordatorio diario por inactividad y el aviso de cobros vencidos. Responde 200 con mensaje de confirmación.
//	@Tags			Jobs
//	@Produce		json
//	@Success		200	{object}	map[string]string	"todos los jobs ejecutados"
//	@Failure		401	{object}	map[string]string	"no autorizado"
//	@Failure		500	{object}	map[string]string	"error interno del servidor"
//	@Security		ApiKeyAuth
//	@Router			/aplicacion/jobs/ejecutar-todos	[post]
func EjecutarTodosSwagger() {}
