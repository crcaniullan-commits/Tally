package auth

// Register
//
//	@Summary		Registrar usuario
//	@Description	Registra un nuevo usuario y devuelve un token de autenticación
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			payload	body	CreateUserPayload	true	"payload"
//	@Success		200		{object}	string			"token de autenticacion"
//	@Failure		400		{object}	error			"payload erroneo o datos duplicados"
//	@Failure		500		{object}	error			"error interno del servidor"
//	@Router			/auth	[post]
func Register() {}

// Login
//
//	@Summary		Iniciar sesión
//	@Description	Inicia sesión con email y contraseña y devuelve un token de autenticación
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			payload	body	LoginUserPayload	true	"payload"
//	@Success		202		{object}	string			"token de autenticacion"
//	@Failure		400		{object}	error			"credenciales invalidas"
//	@Failure		500		{object}	error			"error interno del servidor"
//	@Router			/auth/login	[post]
func Login() {}