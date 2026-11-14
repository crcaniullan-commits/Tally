package clavesacceso

// EmitirParaUsuario
//
//	@Summary		Emitir un código de acceso
//	@Description	Genera un código de acceso, lo guarda asociado al municipal autenticado (created_by) con un año de vencimiento y lo envía por correo al destinatario del payload. La respuesta viene envuelta en {"datos": ...} con la llave creada; el código también viaja en el correo. Emitir es independiente de canjear: el emprendedor lo canjea después en POST /aplicacion/autenticacion.
//	@Tags			ClavesAcceso
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		PayloadCorreo		true	"correo del destinatario"
//	@Success		201		{object}	model.AccessKey		"código creado"
//	@Failure		400		{object}	map[string]string	"payload inválido o email mal formado"
//	@Failure		500		{object}	map[string]string	"error interno del servidor"
//	@Security		ApiKeyAuth
//	@Router			/aplicacion/claves	[post]
func EmitirParaUsuario() {}

// Reenviar
//
//	@Summary		Reenviar el código de acceso
//	@Description	Vuelve a mandar por correo el código de una llave existente, sin crear una llave nueva ni mover su vencimiento: si el correo se perdió, el código sigue siendo el mismo. La respuesta viene envuelta en {"datos": ...} con el mensaje de confirmación. Un claveID que no sea un UUID responde 400 y una llave inexistente, 404.
//	@Tags			ClavesAcceso
//	@Accept			json
//	@Produce		json
//	@Param			claveID	path		string				true	"ID de la llave de acceso"
//	@Param			payload	body		PayloadCorreo		true	"correo del destinatario"
//	@Success		200		{object}	string				"correo reenviado"
//	@Failure		400		{object}	map[string]string	"claveID inválido o payload inválido"
//	@Failure		404		{object}	map[string]string	"llave de acceso inexistente"
//	@Failure		500		{object}	map[string]string	"error interno del servidor"
//	@Security		ApiKeyAuth
//	@Router			/aplicacion/claves/IDAcceso/{claveID}	[post]
func Reenviar() {}

// Revocar
//
//	@Summary		Revocar un código de acceso
//	@Description	Revoca una llave por su código, dejando registro de quién la revocó (revoked_by) y cuándo (revoked_at). Revocar es ortogonal a canjear y a vencer: deja la llave inútil aunque siga vigente. La respuesta viene envuelta en {"datos": ...} con el mensaje de confirmación. Un código que no existe responde 404.
//	@Tags			ClavesAcceso
//	@Produce		json
//	@Param			code	path		string				true	"código de acceso a revocar"
//	@Success		200		{object}	string				"código revocado"
//	@Failure		400		{object}	map[string]string	"código vacío"
//	@Failure		404		{object}	map[string]string	"código inexistente"
//	@Failure		500		{object}	map[string]string	"error interno del servidor"
//	@Security		ApiKeyAuth
//	@Router			/aplicacion/claves/revocar/{codigo}	[post]
func Revocar() {}
