package excel

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/internal/utilidades"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// rutaExcel arma la URL de las rutas de excel tal como las declara
// excel.InitModule.
func rutaExcel(sufijo string) string {
	return "/v1/aplicacion/empresas/" + testEmpresaID.String() + "/excel" + sufijo
}

// contextoMiddleware inyecta en el contexto el usuario y la empresa, que es lo
// que hacen los middlewares de token y de empresa en la app real.
func contextoMiddleware(usuario *model.User, empresa *model.Empresa) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), utilidades.ContextoUsuario, usuario)
			ctx = context.WithValue(ctx, utilidades.ContextoEmpresa, empresa)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// newTestRouter monta el módulo de excel igual que la app: un router chi con
// las rutas declaradas por InitModule, el storage mockeado y la db con
// sqlmock.
func newTestRouter(t *testing.T, db *sql.DB, fs FileStorage, usuario *model.User, empresa *model.Empresa) http.Handler {
	t.Helper()

	r := chi.NewRouter()

	// la app usa los middlewares globales (Logger, Recoverer, etc.): el
	// Recoverer es el que convierte un panic en un 500. Acá va una versión
	// silenciosa para no ensuciar la salida del test.
	r.Use(recovererSilencioso)

	r.Route("/v1/aplicacion", func(r chi.Router) {
		// el módulo de excel va dentro del grupo de la empresa, igual que en
		// cmd/application/app.go
		r.Route("/empresas/{empresaID}", func(r chi.Router) {
			r.Use(contextoMiddleware(usuario, empresa))

			InitModule(r, db, fs, newTestLogger(), newTestMiddleware())
		})
	})

	return r
}

// recovererSilencioso hace lo mismo que el Recoverer de chi (panic → 500 sin
// cuerpo), pero sin imprimir el stack.
func recovererSilencioso(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// newTestUsuarioEmpresa arma el usuario y la empresa que inyecta el middleware.
func newTestUsuarioEmpresa() (*model.User, *model.Empresa) {
	return &model.User{ID: testUserID, Role: utilidades.RolUsuario},
		&model.Empresa{ID: testEmpresaID, UserID: testUserID}
}

// archivoMultipart arma el body de un POST multipart con el campo "file".
func archivoMultipart(t *testing.T, nombre string, contenido []byte) (io.Reader, string) {
	t.Helper()

	var buf bytes.Buffer

	w := multipart.NewWriter(&buf)

	parte, err := w.CreateFormFile("file", nombre)
	require.NoError(t, err)

	_, err = parte.Write(contenido)
	require.NoError(t, err)

	require.NoError(t, w.Close())

	return &buf, w.FormDataContentType()
}

// filasExcel arma las filas que devuelve el store para listar u obtener un
// archivo.
func filasExcel() *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "empresa_id", "nombre", "storage_key", "size_bytes", "created_at"}).
		AddRow(testFileID, testEmpresaID, "reporte.xlsx", "clave/prueba.xlsx", int64(len(xlsxFalso)), testCreatedAt)
}

func TestExcelRutas_Upload(t *testing.T) {
	t.Run("guarda el archivo y responde 200 con el registro", func(t *testing.T) {
		db, mock, _ := newTestDB(t)
		fs := newStorageMock()
		usuario, empresa := newTestUsuarioEmpresa()
		r := newTestRouter(t, db, fs, usuario, empresa)

		mock.ExpectQuery(".*").
			WillReturnRows(sqlmock.NewRows([]string{"created_at"}).AddRow(testCreatedAt))

		body, contentType := archivoMultipart(t, "reporte.xlsx", xlsxFalso)

		req := httptest.NewRequest(http.MethodPost, rutaExcel("/"), body)
		req.Header.Set("Content-Type", contentType)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code, w.Body.String())

		var resp struct {
			Datos File `json:"datos"`
		}

		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

		assert.Equal(t, testEmpresaID, resp.Datos.EmpresaID)
		assert.Equal(t, "reporte.xlsx", resp.Datos.Nombre)
		assert.Equal(t, int64(len(xlsxFalso)), resp.Datos.SizeBytes)
		assert.Equal(t, testCreatedAt, resp.Datos.CreatedAt.UTC())

		// el objeto quedó en el storage mock, con la clave userID/empresaID/uuid.xlsx
		require.NotEmpty(t, fs.ultimaClave)
		assert.True(t, strings.HasPrefix(fs.ultimaClave, testUserID.String()+"/"+testEmpresaID.String()+"/"),
			"la clave debería llevar usuario y empresa: %s", fs.ultimaClave)
		assert.True(t, strings.HasSuffix(fs.ultimaClave, ".xlsx"),
			"la clave debería terminar en .xlsx: %s", fs.ultimaClave)
		assert.True(t, fs.existe(fs.ultimaClave))
		assert.Equal(t, xlsxType, fs.tipos[fs.ultimaClave])

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("responde 400 si no viene el archivo", func(t *testing.T) {
		db, mock, _ := newTestDB(t)
		fs := newStorageMock()
		usuario, empresa := newTestUsuarioEmpresa()
		r := newTestRouter(t, db, fs, usuario, empresa)

		req := httptest.NewRequest(http.MethodPost, rutaExcel("/"), strings.NewReader(""))
		req.Header.Set("Content-Type", "multipart/form-data")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())
		assert.JSONEq(t, `{"error":"archivo requerido o demasiado grande"}`, w.Body.String())
		assert.Equal(t, 0, fs.puts)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("responde 400 si el archivo no es un xlsx", func(t *testing.T) {
		db, mock, _ := newTestDB(t)
		fs := newStorageMock()
		usuario, empresa := newTestUsuarioEmpresa()
		r := newTestRouter(t, db, fs, usuario, empresa)

		body, contentType := archivoMultipart(t, "reporte.txt", []byte("no soy un zip"))

		req := httptest.NewRequest(http.MethodPost, rutaExcel("/"), body)
		req.Header.Set("Content-Type", contentType)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())
		assert.JSONEq(t, `{"error":"el archivo debe ser un .xlsx válido"}`, w.Body.String())
		assert.Equal(t, 0, fs.puts)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("responde 400 si el archivo está vacío", func(t *testing.T) {
		db, mock, _ := newTestDB(t)
		fs := newStorageMock()
		usuario, empresa := newTestUsuarioEmpresa()
		r := newTestRouter(t, db, fs, usuario, empresa)

		body, contentType := archivoMultipart(t, "reporte.xlsx", []byte{})

		req := httptest.NewRequest(http.MethodPost, rutaExcel("/"), body)
		req.Header.Set("Content-Type", contentType)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())
		assert.JSONEq(t, `{"error":"el archivo está vacío o incompleto"}`, w.Body.String())
		assert.Equal(t, 0, fs.puts)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("si el registro no se puede guardar, borra el objeto del storage", func(t *testing.T) {
		db, mock, _ := newTestDB(t)
		fs := newStorageMock()
		usuario, empresa := newTestUsuarioEmpresa()
		r := newTestRouter(t, db, fs, usuario, empresa)

		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		body, contentType := archivoMultipart(t, "reporte.xlsx", xlsxFalso)

		req := httptest.NewRequest(http.MethodPost, rutaExcel("/"), body)
		req.Header.Set("Content-Type", contentType)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusInternalServerError, w.Code, w.Body.String())
		assert.Equal(t, 0, fs.objetosCount(), "el objeto debería quedar compensado")
		assert.Equal(t, 1, fs.deletes)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestExcelRutas_List(t *testing.T) {
	db, mock, _ := newTestDB(t)
	fs := newStorageMock()
	usuario, empresa := newTestUsuarioEmpresa()
	r := newTestRouter(t, db, fs, usuario, empresa)

	mock.ExpectQuery(".*").WillReturnRows(filasExcel())

	req := httptest.NewRequest(http.MethodGet, rutaExcel("/?limit=10&offset=0"), nil)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code, w.Body.String())

	// el listado usa el mismo envelope que el resto de la API: {"datos": [...]}
	var resp struct {
		Datos []File `json:"datos"`
	}

	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Len(t, resp.Datos, 1)

	assert.Equal(t, testFileID, resp.Datos[0].ID)
	assert.Equal(t, testEmpresaID, resp.Datos[0].EmpresaID)
	assert.Equal(t, "reporte.xlsx", resp.Datos[0].Nombre)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExcelRutas_Download(t *testing.T) {
	t.Run("devuelve el archivo con el content-type de xlsx", func(t *testing.T) {
		db, mock, _ := newTestDB(t)
		fs := newStorageMock()
		usuario, empresa := newTestUsuarioEmpresa()
		r := newTestRouter(t, db, fs, usuario, empresa)

		clave := "clave/prueba.xlsx"
		require.NoError(t, fs.Put(context.Background(), clave, bytes.NewReader(xlsxFalso), int64(len(xlsxFalso)), xlsxType))

		mock.ExpectQuery(".*").WillReturnRows(filasExcel())

		req := httptest.NewRequest(http.MethodGet, rutaExcel("/"+testFileID.String()), nil)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Equal(t, xlsxType, w.Header().Get("Content-Type"))
		assert.Contains(t, w.Header().Get("Content-Disposition"), "reporte.xlsx")
		assert.Equal(t, xlsxFalso, w.Body.Bytes())

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("responde 404 si el archivo no existe", func(t *testing.T) {
		db, mock, _ := newTestDB(t)
		fs := newStorageMock()
		usuario, empresa := newTestUsuarioEmpresa()
		r := newTestRouter(t, db, fs, usuario, empresa)

		mock.ExpectQuery(".*").
			WillReturnRows(sqlmock.NewRows([]string{"id", "empresa_id", "nombre", "storage_key", "size_bytes", "created_at"}))

		req := httptest.NewRequest(http.MethodGet, rutaExcel("/"+testFileID.String()), nil)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
		assert.JSONEq(t, `{"error":"no encontrado"}`, w.Body.String())

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("responde 400 si el id no es un uuid", func(t *testing.T) {
		db, mock, _ := newTestDB(t)
		fs := newStorageMock()
		usuario, empresa := newTestUsuarioEmpresa()
		r := newTestRouter(t, db, fs, usuario, empresa)

		req := httptest.NewRequest(http.MethodGet, rutaExcel("/no-es-uuid"), nil)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestExcelRutas_Delete(t *testing.T) {
	t.Run("borra el registro y el objeto", func(t *testing.T) {
		db, mock, _ := newTestDB(t)
		fs := newStorageMock()
		usuario, empresa := newTestUsuarioEmpresa()
		r := newTestRouter(t, db, fs, usuario, empresa)

		clave := "clave/prueba.xlsx"
		require.NoError(t, fs.Put(context.Background(), clave, bytes.NewReader(xlsxFalso), int64(len(xlsxFalso)), xlsxType))

		mock.ExpectQuery(".*").
			WillReturnRows(sqlmock.NewRows([]string{"storage_key"}).AddRow(clave))

		req := httptest.NewRequest(http.MethodDelete, rutaExcel("/"+testFileID.String()), nil)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusNoContent, w.Code, w.Body.String())
		assert.False(t, fs.existe(clave), "el objeto debería estar borrado")
		assert.Equal(t, 1, fs.deletes)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("responde 404 si el archivo no existe", func(t *testing.T) {
		db, mock, _ := newTestDB(t)
		fs := newStorageMock()
		usuario, empresa := newTestUsuarioEmpresa()
		r := newTestRouter(t, db, fs, usuario, empresa)

		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"storage_key"}))

		req := httptest.NewRequest(http.MethodDelete, rutaExcel("/"+testFileID.String()), nil)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
		assert.JSONEq(t, `{"error":"no encontrado"}`, w.Body.String())
		assert.Equal(t, 0, fs.deletes, "no debería tocar el storage si el registro no existe")

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

// TestExcelRutas_SQLDelStore audita a mano el SQL que el store manda a la
// base, porque sqlmock no parsea SQL: solo compara strings.
func TestExcelRutas_SQLDelStore(t *testing.T) {
	t.Run("el listado consulta la tabla excel_files y la columna empresa_id", func(t *testing.T) {
		db, mock, captura := newTestDB(t)
		fs := newStorageMock()
		usuario, empresa := newTestUsuarioEmpresa()
		r := newTestRouter(t, db, fs, usuario, empresa)

		mock.ExpectQuery(".*").WillReturnRows(filasExcel())

		req := httptest.NewRequest(http.MethodGet, rutaExcel("/"), nil)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code, w.Body.String())

		assert.Contains(t, captura.get(), "excel_files")
		assert.Contains(t, captura.get(), "empresa_id")
		assert.NotContains(t, captura.get(), "emrpesa_id")

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("la descarga consulta la tabla excel_files", func(t *testing.T) {
		db, mock, captura := newTestDB(t)
		fs := newStorageMock()
		usuario, empresa := newTestUsuarioEmpresa()
		r := newTestRouter(t, db, fs, usuario, empresa)

		mock.ExpectQuery(".*").WillReturnRows(filasExcel())

		require.NoError(t, fs.Put(context.Background(), "clave/prueba.xlsx",
			bytes.NewReader(xlsxFalso), int64(len(xlsxFalso)), xlsxType))

		req := httptest.NewRequest(http.MethodGet, rutaExcel("/"+testFileID.String()), nil)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code, w.Body.String())

		assert.Contains(t, captura.get(), "excel_files")
		assert.NotContains(t, captura.get(), "exel_files")

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestExcelRutas_Permisos(t *testing.T) {
	t.Run("responde 403 si el usuario es municipal", func(t *testing.T) {
		db, mock, _ := newTestDB(t)
		fs := newStorageMock()
		usuario, empresa := newTestUsuarioEmpresa()
		usuario.Role = utilidades.RolMunicipal

		r := newTestRouter(t, db, fs, usuario, empresa)

		req := httptest.NewRequest(http.MethodGet, rutaExcel("/"), nil)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusForbidden, w.Code, w.Body.String())
		assert.JSONEq(t, `{"error":"prohibido"}`, w.Body.String())

		require.NoError(t, mock.ExpectationsWereMet())
	})
}
