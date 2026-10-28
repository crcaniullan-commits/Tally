package categories

// Create
//
//	@Summary		Crear categoría
//	@Description	Crea una nueva categoría asociada a la empresa autenticada. La respuesta viene envuelta en {"data": ...} con la categoría creada.
//	@Tags			Categories
//	@Accept			json
//	@Produce		json
//	@Param			empresaID	path		string				true	"ID de la empresa"
//	@Param			payload		body		CategoriaPayload	true	"payload"
//	@Success		201			{object}	model.Categorie		"categoría creada"
//	@Failure		400			{object}	map[string]string	"payload de la categoría erroneo o la categoría ya existe"
//	@Failure		500			{object}	map[string]string	"error interno del servidor"
//	@Security		ApiKeyAuth
//	@Router			/app/empresas/{empresaID}/categories	[post]
func Create() {}

// CambiarNombre
//
//	@Summary		Cambiar nombre de categoría
//	@Description	Cambia el nombre de una categoría, siempre que pertenezca a la empresa autenticada. La respuesta viene envuelta en {"data": ...} con el mensaje de confirmación.
//	@Tags			Categories
//	@Produce		json
//	@Param			empresaID	path		string				true	"ID de la empresa"
//	@Param			categoriaID	path		string				true	"ID de la categoría"
//	@Param			payload		body		CategoriaPayload	true	"payload"
//	@Success		200			{object}	string				"nombre cambiado"
//	@Failure		400			{object}	map[string]string	"categoriaID invalido o payload erroneo"
//	@Failure		404			{object}	map[string]string	"categoría no encontrada"
//	@Failure		500			{object}	map[string]string	"error interno del servidor"
//	@Security		ApiKeyAuth
//	@Router			/app/empresas/{empresaID}/categories/{categoriaID}	[patch]
func CambiarNombre() {}

// BorrarCategoria
//
//	@Summary		Eliminar categoría
//	@Description	Elimina una categoría, siempre que pertenezca a la empresa autenticada. La respuesta viene envuelta en {"data": ...} con el mensaje de confirmación.
//	@Tags			Categories
//	@Produce		json
//	@Param			empresaID	path		string				true	"ID de la empresa"
//	@Param			categoria	path		string				true	"ID de la categoría"
//	@Success		200			{object}	string				"categoría eliminada"
//	@Failure		400			{object}	map[string]string	"categoria invalido"
//	@Failure		404			{object}	map[string]string	"categoría no encontrada"
//	@Failure		500			{object}	map[string]string	"error interno del servidor"
//	@Security		ApiKeyAuth
//	@Router			/app/empresas/{empresaID}/categories/{categoria}	[delete]
func BorrarCategoria() {}
