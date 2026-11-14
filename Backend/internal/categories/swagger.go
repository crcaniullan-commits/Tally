package categories

// ObtenerCategorias
//
//	@Summary		Listar categorías
//	@Description	Devuelve las categorías de la empresa autenticada y las globales. Las globales se distinguen porque "empresa_id" viene null. La respuesta viene envuelta en {"datos": [...]} con las categorías propias primero, ordenadas por nombre.
//	@Tags			Categorias
//	@Produce		json
//	@Param			empresaID	path		string				true	"ID de la empresa"
//	@Success		200			{array}		model.Categorie		"categorías de la empresa y globales"
//	@Failure		500			{object}	map[string]string	"error interno del servidor"
//	@Security		ApiKeyAuth
//	@Router			/aplicacion/empresas/{empresaID}/categorias	[get]
func ObtenerCategorias() {}

// Crear
//
//	@Summary		Crear categoría
//	@Description	Crea una nueva categoría asociada a la empresa autenticada. La respuesta viene envuelta en {"datos": ...} con la categoría creada.
//	@Tags			Categorias
//	@Accept			json
//	@Produce		json
//	@Param			empresaID	path		string				true	"ID de la empresa"
//	@Param			payload		body		PayloadCategoria	true	"payload"
//	@Success		201			{object}	model.Categorie		"categoría creada"
//	@Failure		400			{object}	map[string]string	"payload de la categoría erróneo o la categoría ya existe"
//	@Failure		500			{object}	map[string]string	"error interno del servidor"
//	@Security		ApiKeyAuth
//	@Router			/aplicacion/empresas/{empresaID}/categorias	[post]
func Crear() {}

// CambiarNombre
//
//	@Summary		Cambiar nombre de categoría
//	@Description	Cambia el nombre de una categoría, siempre que pertenezca a la empresa autenticada. La respuesta viene envuelta en {"datos": ...} con el mensaje de confirmación.
//	@Tags			Categorias
//	@Produce		json
//	@Param			empresaID	path		string				true	"ID de la empresa"
//	@Param			categoriaID	path		string				true	"ID de la categoría"
//	@Param			payload		body		PayloadCategoria	true	"payload"
//	@Success		200			{object}	string				"nombre cambiado"
//	@Failure		400			{object}	map[string]string	"categoriaID inválido o payload erróneo"
//	@Failure		404			{object}	map[string]string	"categoría no encontrada"
//	@Failure		500			{object}	map[string]string	"error interno del servidor"
//	@Security		ApiKeyAuth
//	@Router			/aplicacion/empresas/{empresaID}/categorias/{categoriaID}	[patch]
func CambiarNombre() {}

// BorrarCategoria
//
//	@Summary		Eliminar categoría
//	@Description	Elimina una categoría, siempre que pertenezca a la empresa autenticada. Si la categoría tiene ingresos o gastos asociados no se puede borrar: responde 409 y la categoría sigue existiendo. La respuesta viene envuelta en {"datos": ...} con el mensaje de confirmación.
//	@Tags			Categorias
//	@Produce		json
//	@Param			empresaID	path		string				true	"ID de la empresa"
//	@Param			categoria	path		string				true	"ID de la categoría"
//	@Success		200			{object}	string				"categoría eliminada"
//	@Failure		400			{object}	map[string]string	"categoria invalido"
//	@Failure		404			{object}	map[string]string	"categoría no encontrada"
//	@Failure		409			{object}	map[string]string	"la categoría tiene movimientos asociados"
//	@Failure		500			{object}	map[string]string	"error interno del servidor"
//	@Security		ApiKeyAuth
//	@Router			/aplicacion/empresas/{empresaID}/categorias/{categoria}	[delete]
func BorrarCategoria() {}
