package users

// UpdateUser
//
//	@Summary		Actualizar usuario
//	@Description	Actualiza un usuario en la base de datos. La respuesta viene envuelta en {"datos": ...} con el mensaje de confirmación.
//	@Tags			Usuarios
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		PayloadActualizarUsuario	true	"payload"
//	@Success		200		{object}	map[string]string			"usuario actualizado"
//	@Failure		400		{object}	map[string]string			"payload del usuario erróneo"
//	@Failure		401		{object}	map[string]string			"no autorizado"
//	@Failure		403		{object}	map[string]string			"prohibido (se requiere rol usuario)"
//	@Failure		404		{object}	map[string]string			"usuario no encontrado"
//	@Failure		500		{object}	map[string]string			"error interno del servidor"
//	@Security		ApiKeyAuth
//	@Router			/aplicacion/usuarios 	[patch]
func Update() {}

// DeleteUser
//
//	@Summary		Eliminar usuario
//	@Description	Elimina un usuario de la base de datos. La respuesta viene envuelta en {"datos": ...} con el mensaje de confirmación.
//	@Tags			Usuarios
//	@Success		200	{object}	map[string]string	"usuario eliminado"
//	@Failure		400	{object}	map[string]string	"userID invalido"
//	@Failure		401	{object}	map[string]string	"no autorizado"
//	@Failure		403	{object}	map[string]string	"prohibido (se requiere rol usuario)"
//	@Failure		404	{object}	map[string]string	"usuario no encontrado"
//	@Failure		500	{object}	map[string]string	"error interno del servidor"
//	@Security		ApiKeyAuth
//	@Router			/aplicacion/usuarios	[delete]
func Delete() {}

// GetUserByRut
//
//	@Summary		Buscar usuario por rut
//	@Description	Busca un usuario por su rut. La respuesta viene envuelta en {"datos": ...} con el usuario.
//	@Tags			Usuarios
//	@Produce		json
//	@Param			rut	path		string					true	"rut del usuario"
//	@Success		200	{object}	map[string]model.User	"usuario"
//	@Failure		400	{object}	map[string]string		"rut invalido"
//	@Failure		401	{object}	map[string]string		"no autorizado"
//	@Failure		403	{object}	map[string]string		"prohibido (se requiere rol municipal)"
//	@Failure		404	{object}	map[string]string		"usuario no encontrado"
//	@Failure		500	{object}	map[string]string		"error interno del servidor"
//	@Security		ApiKeyAuth
//	@Router			/aplicacion/usuarios/municipal/{rut}	[get]
func ObtenerPorRUT() {}
