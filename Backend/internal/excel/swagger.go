package excel

// Upload
//
//	@Summary		Subir archivo xlsx
//	@Description	Sube un archivo .xlsx a la empresa autenticada. El archivo va en el campo multipart "file" y no puede pesar más de 10 MB. El nombre original se guarda y el binario queda en el storage de la empresa; la respuesta viene envuelta en {"datos": ...} con el registro del archivo.
//	@Tags			Excel
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			empresaID	path		string				true	"ID de la empresa"
//	@Param			file		formData	file				true	"archivo .xlsx"
//	@Success		200			{object}	File				"archivo subido"
//	@Failure		400			{object}	map[string]string	"falta el archivo, no es un .xlsx válido o excede el tamaño máximo"
//	@Failure		500			{object}	map[string]string	"error interno del servidor"
//	@Security		ApiKeyAuth
//	@Router			/aplicacion/empresas/{empresaID}/excel	[post]
func Upload() {}

// List
//
//	@Summary		Listar archivos
//	@Description	Devuelve los archivos .xlsx de la empresa autenticada, del más nuevo al más viejo. La respuesta viene envuelta en {"datos": [...]}. Admite "limit" (por defecto 20, máximo 100) y "offset".
//	@Tags			Excel
//	@Produce		json
//	@Param			empresaID	path		string				true	"ID de la empresa"
//	@Param			limit		query		int					false	"cantidad de archivos a devolver"
//	@Param			offset		query		int					false	"archivos a saltar"
//	@Success		200			{object}	map[string][]File	"archivos de la empresa"
//	@Failure		500			{object}	map[string]string	"error interno del servidor"
//	@Security		ApiKeyAuth
//	@Router			/aplicacion/empresas/{empresaID}/excel	[get]
func List() {}

// Download
//
//	@Summary		Descargar archivo
//	@Description	Descarga el contenido de un archivo .xlsx de la empresa autenticada. La respuesta es el binario del archivo, con el nombre original en el header Content-Disposition.
//	@Tags			Excel
//	@Produce		application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
//	@Param			empresaID	path		string				true	"ID de la empresa"
//	@Param			id			path		string				true	"ID del archivo"
//	@Success		200			{file}		binary				"archivo"
//	@Failure		400			{object}	map[string]string	"id invalido"
//	@Failure		404			{object}	map[string]string	"archivo no encontrado"
//	@Failure		500			{object}	map[string]string	"error interno del servidor"
//	@Security		ApiKeyAuth
//	@Router			/aplicacion/empresas/{empresaID}/excel/{id}	[get]
func Download() {}

// Delete
//
//	@Summary		Eliminar archivo
//	@Description	Elimina un archivo .xlsx de la empresa autenticada: se borra el registro y el binario asociado. Responde 204 sin cuerpo.
//	@Tags			Excel
//	@Produce		json
//	@Param			empresaID	path		string				true	"ID de la empresa"
//	@Param			id			path		string				true	"ID del archivo"
//	@Success		204			{string}	string				"archivo eliminado"
//	@Failure		400			{object}	map[string]string	"id invalido"
//	@Failure		404			{object}	map[string]string	"archivo no encontrado"
//	@Failure		500			{object}	map[string]string	"error interno del servidor"
//	@Security		ApiKeyAuth
//	@Router			/aplicacion/empresas/{empresaID}/excel/{id}	[delete]
func Delete() {}
