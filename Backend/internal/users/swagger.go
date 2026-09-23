package users

// UpdateUser
//
//	@Summary		Actualizar usuario
//	@Description	Actualiza un usuario en la base de datos
//	@Tags			Users
//	@Accept			json
//	@Produce		json
//	@Param			userID	path		string				true	"ID del usuario"
//	@Param			payload	body		UpdateUserPayload	true	"payload"
//	@Success		200		{object}	string				"usuario actualizado"
//	@Failure		400		{object}	error				"payload del usuario erroneo"
//	@Failure		404		{object}	error				"usuario no encontrado"
//	@Failure		500		{object}	error				"error interno del servidor"
//	@Security		ApiKeyAuth
//	@Router			/app/users/{userID}	[patch]
func Update() {}

// DeleteUser
//
//	@Summary		Eliminar usuario
//	@Description	Elimina un usuario de la base de datos
//	@Tags			Users
//	@Param			userID	path		string	true	"ID del usuario"
//	@Success		200		{object}	string	"usuario eliminado"
//	@Failure		400		{object}	error	"userID invalido"
//	@Failure		404		{object}	error	"usuario no encontrado"
//	@Failure		500		{object}	error	"error interno del servidor"
//	@Security		ApiKeyAuth
//	@Router			/app/users/{userID}	[delete]
func Delete() {}

// GetUserByRut
//
//	@Summary		Buscar usuario por rut
//	@Description	Busca un usuario por su rut
//	@Tags			Users
//	@Produce		json
//	@Param			rut	path		string	true	"rut del usuario"
//	@Success		200	{object}	Users	"usuario"
//	@Failure		400	{object}	error	"rut invalido"
//	@Failure		404	{object}	error	"usuario no encontrado"
//	@Failure		500	{object}	error	"error interno del servidor"
//	@Security		ApiKeyAuth
//	@Router			/app/users/municipal/{rut}	[get]
func GetByRut() {}
