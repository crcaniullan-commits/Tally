package notificaciones

// ListarHistorial
//
//	@Summary		Listar historial de notificaciones enviadas
//	@Description	Devuelve el registro de notificaciones y recordatorios emitidos para el usuario autenticado (notification_log), ordenados del más reciente al más antiguo. La respuesta viene envuelta en {"datos": [...]}.
//	@Tags			Notificaciones
//	@Produce		json
//	@Success		200	{object}	map[string][]model.NotificationLog	"historial de notificaciones"
//	@Failure		401	{object}	map[string]string					"no autorizado"
//	@Failure		403	{object}	map[string]string					"prohibido (se requiere rol usuario)"
//	@Failure		500	{object}	map[string]string					"error interno del servidor"
//	@Security		ApiKeyAuth
//	@Router			/aplicacion/notificaciones/historial	[get]
func ListarHistorialSwagger() {}

// ProbarNotificacion
//
//	@Summary		Enviar notificación push de prueba
//	@Description	Despacha inmediatamente una notificación push de prueba a todos los dispositivos registrados del usuario autenticado. La respuesta viene envuelta en {"datos": ...}.
//	@Tags			Notificaciones
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		PayloadPruebaNotificacion	false	"título y cuerpo opcionales"
//	@Success		200		{object}	map[string]string			"notificación de prueba enviada"
//	@Failure		401		{object}	map[string]string			"no autorizado"
//	@Failure		403		{object}	map[string]string			"prohibido (se requiere rol usuario)"
//	@Failure		500		{object}	map[string]string			"error interno del servidor"
//	@Security		ApiKeyAuth
//	@Router			/aplicacion/notificaciones/probar	[post]
func ProbarNotificacionSwagger() {}
