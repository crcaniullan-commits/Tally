package integration

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"testing"

	"github.com/crcaniullan-commits/Tally/tests/integration/testutil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// xlsxContenido es un archivo con el magic de un zip ("PK"). No es un xlsx
// válido: el módulo solo verifica los primeros dos bytes.
var xlsxContenido = []byte("PK\x03\x04contenido-del-archivo-de-prueba")

// multipartArchivo arma el body multipart de la subida de archivos.
func multipartArchivo(t *testing.T, nombre string, contenido []byte) ([]byte, string) {
	t.Helper()

	var buf bytes.Buffer

	w := multipart.NewWriter(&buf)

	parte, err := w.CreateFormFile("file", nombre)
	require.NoError(t, err)

	_, err = parte.Write(contenido)
	require.NoError(t, err)

	require.NoError(t, w.Close())

	return buf.Bytes(), w.FormDataContentType()
}

// archivoExcel es el registro que devuelve la API.
type archivoExcel struct {
	ID        uuid.UUID `json:"id"`
	EmpresaID uuid.UUID `json:"empresa_id"`
	Nombre    string    `json:"nombre"`
	SizeBytes int64     `json:"size_bytes"`
}

func TestExcelFlow(t *testing.T) {
	usuario := testutil.NewUser(t)
	empresa := testutil.CreateEmpresa(t, usuario)

	rutas := testutil.EmpresaPath(empresa.ID, "/excel")

	t.Run("las rutas están detrás del token", func(t *testing.T) {
		requireProtected(t, http.MethodGet, rutas)
		requireProtected(t, http.MethodPost, rutas)
		requireProtected(t, http.MethodDelete, rutas+"/"+uuid.NewString())
	})

	var archivoID uuid.UUID

	t.Run("sube un archivo xlsx", func(t *testing.T) {
		body, contentType := multipartArchivo(t, "reporte.xlsx", xlsxContenido)

		res := testutil.RequestRaw(t, http.MethodPost, rutas, usuario.Token, contentType, body)

		testutil.RequireStatus(t, res, http.StatusOK)

		archivo := testutil.DecodeData[archivoExcel](t, res)

		assert.Equal(t, empresa.ID, archivo.EmpresaID)
		assert.Equal(t, "reporte.xlsx", archivo.Nombre)
		assert.Equal(t, int64(len(xlsxContenido)), archivo.SizeBytes)

		archivoID = archivo.ID
	})

	t.Run("lista los archivos de la empresa", func(t *testing.T) {
		res := testutil.Request(t, http.MethodGet, rutas, usuario.Token, nil)

		testutil.RequireStatus(t, res, http.StatusOK)

		archivos := testutil.DecodeData[[]archivoExcel](t, res)

		require.Len(t, archivos, 1)
		assert.Equal(t, archivoID, archivos[0].ID)
		assert.Equal(t, "reporte.xlsx", archivos[0].Nombre)
	})

	t.Run("descarga el archivo", func(t *testing.T) {
		res := testutil.Request(t, http.MethodGet, rutas+"/"+archivoID.String(), usuario.Token, nil)

		testutil.RequireStatus(t, res, http.StatusOK)
		assert.Equal(t, xlsxContenido, res.Body)
		assert.Contains(t, res.Header.Get("Content-Type"), "spreadsheetml")
	})

	t.Run("rechaza un archivo que no es xlsx", func(t *testing.T) {
		body, contentType := multipartArchivo(t, "reporte.txt", []byte("no soy un zip"))

		res := testutil.RequestRaw(t, http.MethodPost, rutas, usuario.Token, contentType, body)

		testutil.RequireError(t, res, http.StatusBadRequest, "el archivo debe ser un .xlsx válido")
	})

	t.Run("rechaza la subida sin archivo", func(t *testing.T) {
		res := testutil.RequestRaw(t, http.MethodPost, rutas, usuario.Token, "multipart/form-data", nil)

		testutil.RequireError(t, res, http.StatusBadRequest, "archivo requerido o demasiado grande")
	})

	t.Run("otra empresa no ve el archivo", func(t *testing.T) {
		otro := testutil.NewUser(t)
		otraEmpresa := testutil.CreateEmpresa(t, otro)

		res := testutil.Request(t, http.MethodGet, testutil.EmpresaPath(otraEmpresa.ID, "/excel"), otro.Token, nil)

		testutil.RequireStatus(t, res, http.StatusOK)
		assert.Empty(t, testutil.DecodeData[[]archivoExcel](t, res))

		// y tampoco puede descargarlo por id
		res = testutil.Request(t, http.MethodGet,
			testutil.EmpresaPath(otraEmpresa.ID, "/excel")+"/"+archivoID.String(), otro.Token, nil)

		testutil.RequireError(t, res, http.StatusNotFound, "no encontrado")
	})

	t.Run("borra el archivo", func(t *testing.T) {
		res := testutil.Request(t, http.MethodDelete, rutas+"/"+archivoID.String(), usuario.Token, nil)

		testutil.RequireStatus(t, res, http.StatusNoContent)
	})

	t.Run("el archivo borrado ya no se puede descargar ni borrar", func(t *testing.T) {
		res := testutil.Request(t, http.MethodGet, rutas+"/"+archivoID.String(), usuario.Token, nil)

		testutil.RequireError(t, res, http.StatusNotFound, "no encontrado")

		res = testutil.Request(t, http.MethodDelete, rutas+"/"+archivoID.String(), usuario.Token, nil)

		testutil.RequireError(t, res, http.StatusNotFound, "no encontrado")
	})

	t.Run("el listado queda vacío", func(t *testing.T) {
		res := testutil.Request(t, http.MethodGet, rutas, usuario.Token, nil)

		testutil.RequireStatus(t, res, http.StatusOK)
		assert.Empty(t, testutil.DecodeData[[]archivoExcel](t, res))
	})
}

// TestExcelFlow_ModeloDejaElRegistroEnLaBase comprueba que el registro queda
// en la tabla excel_files con la empresa correcta.
func TestExcelFlow_RegistroEnLaBase(t *testing.T) {
	usuario := testutil.NewUser(t)
	empresa := testutil.CreateEmpresa(t, usuario)

	body, contentType := multipartArchivo(t, "reporte.xlsx", xlsxContenido)

	res := testutil.RequestRaw(t, http.MethodPost,
		testutil.EmpresaPath(empresa.ID, "/excel"), usuario.Token, contentType, body)

	testutil.RequireStatus(t, res, http.StatusOK)

	archivo := testutil.DecodeData[archivoExcel](t, res)

	var (
		empresaID  uuid.UUID
		nombre     string
		storageKey string
	)

	require.NoError(t, testutil.GetEnv().DB.QueryRow(
		`SELECT empresa_id, nombre, storage_key FROM excel_files WHERE id = $1`, archivo.ID,
	).Scan(&empresaID, &nombre, &storageKey))

	assert.Equal(t, empresa.ID, empresaID)
	assert.Equal(t, "reporte.xlsx", nombre)
	assert.Contains(t, storageKey, empresa.ID.String())

	// el archivo se guarda con la clave usuario/empresa/uuid.xlsx
	assert.Regexp(t, `^.+/`+empresa.ID.String()+`/.+\.xlsx$`, storageKey)
}

// TestExcelFlow_EmpresaBorradaBorraArchivos comprueba la cascada: al borrar la
// empresa se borran sus registros de excel_files.
func TestExcelFlow_EmpresaBorradaBorraArchivos(t *testing.T) {
	usuario := testutil.NewUser(t)
	empresa := testutil.CreateEmpresa(t, usuario)

	body, contentType := multipartArchivo(t, "reporte.xlsx", xlsxContenido)

	res := testutil.RequestRaw(t, http.MethodPost,
		testutil.EmpresaPath(empresa.ID, "/excel"), usuario.Token, contentType, body)

	testutil.RequireStatus(t, res, http.StatusOK)

	archivo := testutil.DecodeData[archivoExcel](t, res)

	_, err := testutil.GetEnv().DB.Exec(`DELETE FROM empresas WHERE id = $1`, empresa.ID)
	require.NoError(t, err)

	var queda bool

	require.NoError(t, testutil.GetEnv().DB.QueryRow(
		`SELECT EXISTS (SELECT 1 FROM excel_files WHERE id = $1)`, archivo.ID,
	).Scan(&queda))

	assert.False(t, queda, "el registro debería borrarse en cascada con la empresa")
}

// TestExcelFlow_UsuarioMunicipalNoPuedeVerLosArchivos verifica que el control
// de empresa bloquea a quien no es dueño de la empresa.
func TestExcelFlow_UsuarioMunicipalNoPuedeVerLosArchivos(t *testing.T) {
	usuario := testutil.NewUser(t)
	empresa := testutil.CreateEmpresa(t, usuario)

	// el municipal no es dueño de la empresa: el middleware de empresa lo
	// rechaza antes que el de rol.
	res := testutil.Request(t, http.MethodGet,
		testutil.EmpresaPath(empresa.ID, "/excel"), testutil.MunicipalToken(t), nil)

	testutil.RequireError(t, res, http.StatusNotFound, "no encontrado")
}
