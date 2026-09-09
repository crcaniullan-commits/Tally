package users

// CreateUser
//
// @Summary	Crear usuario
// @Description	Crea un usuario en la base de datos
// @Tags		Users
// @Accept		json
// @Produce	json
// @Param		payload	body	CreateUserPayload	true	"payload"
// @Success	201	{object}	string	"usuario creado"
// @Failure	400	{object}	error	"payload del usuario erroneo"
// @Failure	500	{object}	error	"error interno del servidor"
// @Security	ApiKeyAuth
// @Router		/users	[post]
func Create() {}

// UpdateUser
//
// @Summary		Actualizar usuario
// @Description	Actualiza un usuario en la base de datos
// @Tags			Users
// @Accept		json
// @Produce		json
// @Param			userID	path	string	true	"ID del usuario"
// @Param			payload	body	UpdateUserPayload	true	"payload"
// @Success		200	{object}	string	"usuario actualizado"
// @Failure		400	{object}	error	"payload del usuario erroneo"
// @Failure		404	{object}	error	"usuario no encontrado"
// @Failure		500	{object}	error	"error interno del servidor"
// @Security		ApiKeyAuth
// @Router			/users/{userID}	[patch]
func Update() {}

// DeleteUser
//
// @Summary		Eliminar usuario
// @Description	Elimina un usuario de la base de datos
// @Tags			Users
// @Param			userID	path	string	true	"ID del usuario"
// @Success		200	{object}	string	"usuario eliminado"
// @Failure		400	{object}	error	"userID invalido"
// @Failure		404	{object}	error	"usuario no encontrado"
// @Failure		500	{object}	error	"error interno del servidor"
// @Security		ApiKeyAuth
// @Router			/users/{userID}	[delete]
func Delete() {}

// GetUserByRut
//
// @Summary		Buscar usuario por rut
// @Description	Busca un usuario por su rut
// @Tags			Users
// @Produce		json
// @Param			rut	path	string	true	"rut del usuario"
// @Success		200	{object}	Users	"usuario"
// @Failure		400	{object}	error	"rut invalido"
// @Failure		404	{object}	error	"usuario no encontrado"
// @Failure		500	{object}	error	"error interno del servidor"
// @Security		ApiKeyAuth
// @Router			/users/municipal/{rut}	[get]
func GetByRut() {}

// GetUserByEmail
//
// @Summary		Buscar usuario por email
// @Description	Busca un usuario por su email
// @Tags			Users
// @Produce		json
// @Param			email	path	string	true	"email del usuario"
// @Success		200	{object}	Users	"usuario"
// @Failure		404	{object}	error	"usuario no encontrado"
// @Failure		500	{object}	error	"error interno del servidor"
// @Security		ApiKeyAuth
// @Router			/users/{email}	[get]
func GetByEmail() {}
