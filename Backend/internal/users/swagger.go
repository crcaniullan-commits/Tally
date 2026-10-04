package users

// UpdateUser
//
//	@Summary		Actualizar usuario
//	@Description	Actualiza un usuario en la base de datos
//	@Tags			Users
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		UpdateUserPayload	true	"payload"
//	@Success		200		{object}	string				"usuario actualizado"
//	@Failure		400		{object}	error				"payload del usuario erroneo"
//	@Failure		404		{object}	error				"usuario no encontrado"
//	@Failure		500		{object}	error				"error interno del servidor"
//	@Security		ApiKeyAuth
//	@Router			/app/users 	[patch]
func Update() {}

// DeleteUser
//
//	@Summary		Eliminar usuario
//	@Description	Elimina un usuario de la base de datos
//	@Tags			Users
//	@Success		200	{object}	string	"usuario eliminado"
//	@Failure		400	{object}	error	"userID invalido"
//	@Failure		404	{object}	error	"usuario no encontrado"
//	@Failure		500	{object}	error	"error interno del servidor"
//	@Security		ApiKeyAuth
//	@Router			/app/users	[delete]
func Delete() {}

// GetUserByRut
//
//	@Summary		Buscar usuario por rut
//	@Description	Busca un usuario por su rut
//	@Tags			Users
//	@Produce		json
//	@Param			rut	path		string		true	"rut del usuario"
//	@Success		200	{object}	model.User	"usuario"
//	@Failure		400	{object}	error		"rut invalido"
//	@Failure		404	{object}	error		"usuario no encontrado"
//	@Failure		500	{object}	error		"error interno del servidor"
//	@Security		ApiKeyAuth
//	@Router			/app/users/municipal/{rut}	[get]
func GetByRut() {}

// Exchange
//
//	@Summary		Canjear un código de acceso
//	@Description	Canjea un código de acceso y extiende el plan del usuario autenticado hasta el vencimiento de la llave. Un código ya canjeado, vencido o revocado se responde 400; uno inexistente, 404.
//	@Tags			Users
//	@Produce		json
//	@Param			code	path		string				true	"codigo de acceso"
//	@Success		202		{object}	string				"codigo canjeado"
//	@Failure		400		{object}	map[string]string	"codigo invalido, vencido, revocado o ya canjeado"
//	@Failure		404		{object}	map[string]string	"codigo inexistente"
//	@Failure		500		{object}	map[string]string	"error interno del servidor"
//	@Security		ApiKeyAuth
//	@Router			/app/users/exchange/{code}	[post]
func Exchange() {}
