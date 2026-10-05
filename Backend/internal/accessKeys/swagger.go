package accesskeys

// IssueForUser
//
//	@Summary		Emitir un código de acceso
//	@Description	Genera un código de acceso, lo guarda asociado al municipal autenticado (created_by) con un año de vencimiento y lo envía por correo al destinatario del payload. La respuesta viene envuelta en {"data": ...} con la llave creada; el código también viaja en el correo. Emitir es independiente de canjear: el emprendedor lo canjea después en POST /app/users/exchange/{code}.
//	@Tags			AccessKeys
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		EmailPayload		true	"correo del destinatario"
//	@Success		200		{object}	model.AccessKey		"código creado"
//	@Failure		400		{object}	map[string]string	"payload inválido o email mal formado"
//	@Failure		500		{object}	map[string]string	"error interno del servidor"
//	@Security		ApiKeyAuth
//	@Router			/app/key	[post]
func IssueForUser() {}

// Resend
//
//	@Summary		Reenviar el código de acceso
//	@Description	Vuelve a mandar por correo el código de una llave existente, sin crear una llave nueva ni mover su vencimiento: si el correo se perdió, el código sigue siendo el mismo. La respuesta viene envuelta en {"data": ...} con el mensaje de confirmación. Un keyID que no sea un UUID responde 400.
//	@Tags			AccessKeys
//	@Accept			json
//	@Produce		json
//	@Param			keyID	path		string				true	"ID de la llave de acceso"
//	@Param			payload	body		EmailPayload		true	"correo del destinatario"
//	@Success		200		{object}	string				"correo reenviado"
//	@Failure		400		{object}	map[string]string	"keyID inválido o payload inválido"
//	@Failure		500		{object}	map[string]string	"error interno del servidor o llave inexistente"
//	@Security		ApiKeyAuth
//	@Router			/app/key/AccessID/{keyID}	[post]
func Resend() {}

// Revoke
//
//	@Summary		Revocar un código de acceso
//	@Description	Revoca una llave por su código, dejando registro de quién la revocó (revoked_by) y cuándo (revoked_at). Revocar es ortogonal a canjear y a vencer: deja la llave inútil aunque siga vigente. La respuesta viene envuelta en {"data": ...} con el mensaje de confirmación. Un código que no existe responde 404.
//	@Tags			AccessKeys
//	@Produce		json
//	@Param			code	path		string				true	"código de acceso a revocar"
//	@Success		200		{object}	string				"código revocado"
//	@Failure		400		{object}	map[string]string	"código vacío"
//	@Failure		404		{object}	map[string]string	"código inexistente"
//	@Failure		500		{object}	map[string]string	"error interno del servidor"
//	@Security		ApiKeyAuth
//	@Router			/app/key/revoke/{code}	[post]
func Revoke() {}
