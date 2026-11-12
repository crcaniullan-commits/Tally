package auth

// Register
//
//	@Summary		Registrar usuario
//	@Description	Registra un nuevo usuario y devuelve un token de autenticación
//	@Tags			Autenticacion
//	@Accept			json
//	@Produce		json
//	@Param			payload			body		PayloadCrearUsuario	true	"payload"
//	@Success		201				{object}	string				"token de autenticacion"
//	@Failure		400				{object}	map[string]string	"payload erróneo o datos duplicados"
//	@Failure		500				{object}	map[string]string	"error interno del servidor"
//	@Router			/autenticacion	[post]
func Register() {}

// Login
//
//	@Summary		Iniciar sesión
//	@Description	Inicia sesión con email y contraseña y devuelve un token de autenticación
//	@Tags			Autenticacion
//	@Accept			json
//	@Produce		json
//	@Param			payload							body		PayloadInicioSesion	true	"payload"
//	@Success		202								{object}	string				"token de autenticacion"
//	@Failure		400								{object}	map[string]string	"credenciales invalidas"
//	@Failure		500								{object}	map[string]string	"error interno del servidor"
//	@Router			/autenticacion/inicio-sesion	[post]
func Login() {}

// CrearMunicipal
//
//	@Summary		Registrar usuario municipal
//	@Description	Registra un usuario con rol municipal y devuelve un token de autenticación
//	@Tags			Autenticacion
//	@Accept			json
//	@Produce		json
//	@Param			payload						body		PayloadCrearMunicipal	true	"payload"
//	@Success		201							{object}	string					"token de autenticacion"
//	@Failure		400							{object}	map[string]string		"payload erróneo o datos duplicados"
//	@Failure		500							{object}	map[string]string		"error interno del servidor"
//	@Router			/autenticacion/municipal	[post]
func CrearMunicipal() {}
