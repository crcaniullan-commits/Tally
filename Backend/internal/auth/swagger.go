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

// SolicitarRecuperacion
//
//	@Summary		Pedir código de recuperación de contraseña
//	@Description	Genera un código numérico de 6 dígitos y lo manda por correo al email indicado. El código sirve para verificar al usuario y después cambiar su contraseña. Si el email no está registrado responde 400.
//	@Tags			Autenticacion
//	@Accept			json
//	@Produce		json
//	@Param			payload								body		PayloadSolicitarRecuperacion	true	"payload"
//	@Success		202									{object}	string							"código enviado"
//	@Failure		400									{object}	map[string]string				"payload erróneo o email no registrado"
//	@Failure		500									{object}	map[string]string				"error interno del servidor"
//	@Router			/autenticacion/recuperar-password	[post]
func SolicitarRecuperacion() {}

// VerificarCodigoRecuperacion
//
//	@Summary		Verificar código de recuperación
//	@Description	Verifica el código de 6 dígitos recibido por correo. Un código verificado habilita el cambio de contraseña. Un código incorrecto, vencido o ya usado responde 400.
//	@Tags			Autenticacion
//	@Accept			json
//	@Produce		json
//	@Param			payload										body		PayloadVerificarRecuperacion	true	"payload"
//	@Success		200											{object}	string							"código verificado"
//	@Failure		400											{object}	map[string]string				"payload erróneo o código inválido"
//	@Failure		500											{object}	map[string]string				"error interno del servidor"
//	@Router			/autenticacion/recuperar-password/verificar	[post]
func VerificarCodigoRecuperacion() {}

// ConfirmarRecuperacion
//
//	@Summary		Confirmar recuperación de contraseña
//	@Description	Cambia la contraseña del usuario con el código de recuperación ya verificado. El código queda consumido y no se puede reusar.
//	@Tags			Autenticacion
//	@Accept			json
//	@Produce		json
//	@Param			payload										body		PayloadConfirmarRecuperacion	true	"payload"
//	@Success		200											{object}	string							"contraseña actualizada"
//	@Failure		400											{object}	map[string]string				"payload erróneo o código inválido"
//	@Failure		500											{object}	map[string]string				"error interno del servidor"
//	@Router			/autenticacion/recuperar-password/confirmar	[post]
func ConfirmarRecuperacion() {}
