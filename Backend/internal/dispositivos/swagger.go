package dispositivos

// Registrar
//
//	@Summary		Registrar o refrescar token de dispositivo
//	@Description	Registra un nuevo token de notificación push o actualiza la última conexión (last_seen_at) de un dispositivo para el usuario autenticado. Admite plataformas android, ios y web. Responde 204 sin contenido.
//	@Tags			Dispositivos
//	@Accept			json
//	@Param			payload	body	PayloadDispositivo	true	"token y plataforma del dispositivo"
//	@Success		204		"dispositivo registrado"
//	@Failure		400		{object}	map[string]string	"payload inválido o plataforma no soportada"
//	@Failure		401		{object}	map[string]string	"no autorizado"
//	@Failure		500		{object}	map[string]string	"error interno del servidor"
//	@Security		ApiKeyAuth
//	@Router			/aplicacion/dispositivos	[put]
func RegistrarSwagger() {}

// Listar
//
//	@Summary		Listar dispositivos registrados
//	@Description	Devuelve la lista de dispositivos y tokens activos asociados al usuario autenticado. La respuesta viene envuelta en {"datos": [...]}.
//	@Tags			Dispositivos
//	@Produce		json
//	@Success		200	{object}	map[string][]model.DeviceToken	"lista de dispositivos"
//	@Failure		401	{object}	map[string]string				"no autorizado"
//	@Failure		500	{object}	map[string]string				"error interno del servidor"
//	@Security		ApiKeyAuth
//	@Router			/aplicacion/dispositivos	[get]
func ListarSwagger() {}

// Desvincular
//
//	@Summary		Desvincular token de dispositivo
//	@Description	Elimina el token de dispositivo del usuario autenticado (por ejemplo, al cerrar sesión en la app móvil). Responde 204 sin contenido.
//	@Tags			Dispositivos
//	@Param			token	path	string	true	"token del dispositivo"
//	@Success		204		"dispositivo desvinculado"
//	@Failure		400		{object}	map[string]string	"token inválido"
//	@Failure		401		{object}	map[string]string	"no autorizado"
//	@Failure		404		{object}	map[string]string	"dispositivo no encontrado"
//	@Failure		500		{object}	map[string]string	"error interno del servidor"
//	@Security		ApiKeyAuth
//	@Router			/aplicacion/dispositivos/{token}	[delete]
func DesvincularSwagger() {}
