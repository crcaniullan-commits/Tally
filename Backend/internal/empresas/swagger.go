package empresas

// CreateEmpresa
//
//	@Summary		Crear empresa
//	@Description	Crea una nueva empresa asociada al usuario autenticado. La respuesta viene envuelta en {"datos": ...} con la empresa creada.
//	@Tags			Empresas
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		PayloadEmpresa		true	"payload"
//	@Success		201		{object}	model.Empresa		"empresa creada"
//	@Failure		400		{object}	map[string]string	"payload de la empresa erroneo"
//	@Failure		500		{object}	map[string]string	"error interno del servidor"
//	@Security		ApiKeyAuth
//	@Router			/aplicacion/empresas	[post]
func CreateEmpresa() {}

// ObtenerEmpresas
//
//	@Summary		Listar empresas
//	@Description	Devuelve todas las empresas asociadas al usuario autenticado. La respuesta viene envuelta en {"datos": ...}.
//	@Tags			Empresas
//	@Produce		json
//	@Success		200	{object}	[]model.Empresa		"lista de empresas"
//	@Failure		404	{object}	map[string]string	"empresas no encontradas"
//	@Failure		500	{object}	map[string]string	"error interno del servidor"
//	@Security		ApiKeyAuth
//	@Router			/aplicacion/empresas	[get]
func ObtenerEmpresas() {}

// CambiarNombre
//
//	@Summary		Cambiar nombre de empresa
//	@Description	Cambia el nombre de una empresa, siempre que pertenezca al usuario autenticado. La respuesta viene envuelta en {"datos": ...} con el mensaje de confirmación.
//	@Tags			Empresas
//	@Produce		json
//	@Param			empresaID	path		string				true	"ID de la empresa"
//	@Param			nombre		path		string				true	"Nuevo nombre de la empresa"
//	@Success		200			{object}	string				"nombre actualizado"
//	@Failure		400			{object}	map[string]string	"ruta debe incluir el nombre de la empresa"
//	@Failure		404			{object}	map[string]string	"empresa no encontrada"
//	@Failure		500			{object}	map[string]string	"error interno del servidor"
//	@Security		ApiKeyAuth
//	@Router			/aplicacion/empresas/{empresaID}/{nombre}	[patch]
func CambiarNombre() {}

// BorrarEmpresa
//
//	@Summary		Eliminar empresa
//	@Description	Elimina una empresa, siempre que pertenezca al usuario autenticado. La respuesta viene envuelta en {"datos": ...} con el mensaje de confirmación.
//	@Tags			Empresas
//	@Produce		json
//	@Param			empresaID	path		string				true	"ID de la empresa"
//	@Success		200			{object}	string				"empresa borrada"
//	@Failure		404			{object}	map[string]string	"empresa no encontrada"
//	@Failure		500			{object}	map[string]string	"error interno del servidor"
//	@Security		ApiKeyAuth
//	@Router			/aplicacion/empresas/{empresaID}	[delete]
func BorrarEmpresa() {}
