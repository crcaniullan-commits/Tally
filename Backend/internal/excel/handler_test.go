package excel

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/crcaniullan-commits/Tally/internal/errores"
	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/internal/utilidades"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// newTestHandler arma el Handler con el service mockeado.
func newTestHandler(svc IServiceExcel) *Handler {
	return NuevoHandlerExcel(svc, errores.NuevaRespuestaErrores(zap.NewNop().Sugar()))
}

// requestConContexto mete en el contexto el usuario y la empresa, que es lo
// que inyectan los middlewares de token y de empresa.
func requestConContexto(r *http.Request) *http.Request {
	usuario := &model.User{ID: testUserID, Role: utilidades.RolUsuario}
	empresa := &model.Empresa{ID: testEmpresaID, UserID: testUserID}

	ctx := context.WithValue(r.Context(), utilidades.ContextoUsuario, usuario)
	ctx = context.WithValue(ctx, utilidades.ContextoEmpresa, empresa)

	return r.WithContext(ctx)
}

// requestConID agrega el URL param "id" al route context de chi.
func requestConID(r *http.Request, id string) *http.Request {
	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("id", id)

	ctx := context.WithValue(r.Context(), chi.RouteCtxKey, routeCtx)

	return r.WithContext(ctx)
}

// bodyMultipart arma el body de una subida de archivo.
func bodyMultipart(t *testing.T, nombre string, contenido []byte) (io.Reader, string) {
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

func TestExcelHandler_Upload(t *testing.T) {
	t.Run("responde 200 con el archivo creado", func(t *testing.T) {
		svc := new(IServiceExcelMock)

		esperado := newTestFile()

		svc.On("Save", mock.Anything, testUserID, testEmpresaID, "reporte.xlsx", mock.Anything, int64(len(xlsxFalso))).
			Return(&esperado, nil).Once()

		h := newTestHandler(svc)

		body, contentType := bodyMultipart(t, "reporte.xlsx", xlsxFalso)

		req := httptest.NewRequest(http.MethodPost, "/excel", body)
		req.Header.Set("Content-Type", contentType)

		w := httptest.NewRecorder()
		h.Upload(w, requestConContexto(req))

		require.Equal(t, http.StatusOK, w.Code, w.Body.String())

		var resp struct {
			Datos File `json:"datos"`
		}

		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, testEmpresaID, resp.Datos.EmpresaID)
		assert.Equal(t, "reporte.xlsx", resp.Datos.Nombre)
		assert.Equal(t, int64(len(xlsxFalso)), resp.Datos.SizeBytes)

		svc.AssertExpectations(t)
	})

	t.Run("responde 400 si no viene el archivo", func(t *testing.T) {
		svc := new(IServiceExcelMock)
		h := newTestHandler(svc)

		req := httptest.NewRequest(http.MethodPost, "/excel", strings.NewReader(""))
		req.Header.Set("Content-Type", "multipart/form-data")

		w := httptest.NewRecorder()
		h.Upload(w, requestConContexto(req))

		require.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())
		assert.JSONEq(t, `{"error":"archivo requerido o demasiado grande"}`, w.Body.String())

		svc.AssertNotCalled(t, "Save", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("responde 400 si el archivo no es un xlsx", func(t *testing.T) {
		svc := new(IServiceExcelMock)
		h := newTestHandler(svc)

		body, contentType := bodyMultipart(t, "reporte.txt", []byte("no soy un zip"))

		req := httptest.NewRequest(http.MethodPost, "/excel", body)
		req.Header.Set("Content-Type", contentType)

		w := httptest.NewRecorder()
		h.Upload(w, requestConContexto(req))

		require.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())
		assert.JSONEq(t, `{"error":"el archivo debe ser un .xlsx válido"}`, w.Body.String())

		svc.AssertNotCalled(t, "Save", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("responde 400 si el archivo está vacío", func(t *testing.T) {
		svc := new(IServiceExcelMock)
		h := newTestHandler(svc)

		body, contentType := bodyMultipart(t, "reporte.xlsx", []byte{})

		req := httptest.NewRequest(http.MethodPost, "/excel", body)
		req.Header.Set("Content-Type", contentType)

		w := httptest.NewRecorder()
		h.Upload(w, requestConContexto(req))

		require.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())
		assert.JSONEq(t, `{"error":"el archivo está vacío o incompleto"}`, w.Body.String())
	})

	t.Run("responde 500 si el service falla", func(t *testing.T) {
		svc := new(IServiceExcelMock)

		svc.On("Save", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(nil, errors.New("boom")).Once()

		h := newTestHandler(svc)

		body, contentType := bodyMultipart(t, "reporte.xlsx", xlsxFalso)

		req := httptest.NewRequest(http.MethodPost, "/excel", body)
		req.Header.Set("Content-Type", contentType)

		w := httptest.NewRecorder()
		h.Upload(w, requestConContexto(req))

		require.Equal(t, http.StatusInternalServerError, w.Code, w.Body.String())
		assert.JSONEq(t, `{"error":"el servidor encontró un problema"}`, w.Body.String())

		svc.AssertExpectations(t)
	})
}

func TestExcelHandler_List(t *testing.T) {
	t.Run("responde 200 con la lista de archivos", func(t *testing.T) {
		svc := new(IServiceExcelMock)

		esperado := newTestFile()

		svc.On("List", mock.Anything, testEmpresaID, 20, 0).
			Return([]File{esperado}, nil).Once()

		h := newTestHandler(svc)

		req := httptest.NewRequest(http.MethodGet, "/excel", nil)

		w := httptest.NewRecorder()
		h.List(w, requestConContexto(req))

		require.Equal(t, http.StatusOK, w.Code, w.Body.String())

		var resp struct {
			Datos []File `json:"datos"`
		}

		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		require.Len(t, resp.Datos, 1)
		assert.Equal(t, testFileID, resp.Datos[0].ID)

		svc.AssertExpectations(t)
	})

	t.Run("usa el limit y el offset de la query", func(t *testing.T) {
		svc := new(IServiceExcelMock)

		svc.On("List", mock.Anything, testEmpresaID, 10, 5).
			Return([]File{}, nil).Once()

		h := newTestHandler(svc)

		req := httptest.NewRequest(http.MethodGet, "/excel?limit=10&offset=5", nil)

		w := httptest.NewRecorder()
		h.List(w, requestConContexto(req))

		require.Equal(t, http.StatusOK, w.Code, w.Body.String())

		svc.AssertExpectations(t)
	})

	t.Run("responde 500 si el service falla", func(t *testing.T) {
		svc := new(IServiceExcelMock)

		svc.On("List", mock.Anything, testEmpresaID, mock.Anything, mock.Anything).
			Return(nil, errors.New("boom")).Once()

		h := newTestHandler(svc)

		req := httptest.NewRequest(http.MethodGet, "/excel", nil)

		w := httptest.NewRecorder()
		h.List(w, requestConContexto(req))

		require.Equal(t, http.StatusInternalServerError, w.Code, w.Body.String())
		assert.JSONEq(t, `{"error":"el servidor encontró un problema"}`, w.Body.String())

		svc.AssertExpectations(t)
	})
}

func TestExcelHandler_Download(t *testing.T) {
	t.Run("responde 200 con el binario del archivo", func(t *testing.T) {
		svc := new(IServiceExcelMock)

		esperado := newTestFile()

		svc.On("Open", mock.Anything, testEmpresaID, testFileID).
			Return(&esperado, io.NopCloser(bytes.NewReader(xlsxFalso)), nil).Once()

		h := newTestHandler(svc)

		req := httptest.NewRequest(http.MethodGet, "/excel/"+testFileID.String(), nil)

		w := httptest.NewRecorder()
		h.Download(w, requestConID(requestConContexto(req), testFileID.String()))

		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Equal(t, xlsxType, w.Header().Get("Content-Type"))
		assert.Contains(t, w.Header().Get("Content-Disposition"), "reporte.xlsx")
		assert.Equal(t, xlsxFalso, w.Body.Bytes())

		svc.AssertExpectations(t)
	})

	t.Run("responde 400 si el id no es un uuid", func(t *testing.T) {
		svc := new(IServiceExcelMock)
		h := newTestHandler(svc)

		req := httptest.NewRequest(http.MethodGet, "/excel/no-es-uuid", nil)

		w := httptest.NewRecorder()
		h.Download(w, requestConID(requestConContexto(req), "no-es-uuid"))

		require.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())

		svc.AssertNotCalled(t, "Open", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("responde 404 si el archivo no existe", func(t *testing.T) {
		svc := new(IServiceExcelMock)

		svc.On("Open", mock.Anything, testEmpresaID, testFileID).
			Return(nil, nil, ErrNotFound).Once()

		h := newTestHandler(svc)

		req := httptest.NewRequest(http.MethodGet, "/excel/"+testFileID.String(), nil)

		w := httptest.NewRecorder()
		h.Download(w, requestConID(requestConContexto(req), testFileID.String()))

		require.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
		assert.JSONEq(t, `{"error":"no encontrado"}`, w.Body.String())

		svc.AssertExpectations(t)
	})

	t.Run("responde 500 si el service falla", func(t *testing.T) {
		svc := new(IServiceExcelMock)

		svc.On("Open", mock.Anything, testEmpresaID, testFileID).
			Return(nil, nil, errors.New("boom")).Once()

		h := newTestHandler(svc)

		req := httptest.NewRequest(http.MethodGet, "/excel/"+testFileID.String(), nil)

		w := httptest.NewRecorder()
		h.Download(w, requestConID(requestConContexto(req), testFileID.String()))

		require.Equal(t, http.StatusInternalServerError, w.Code, w.Body.String())
		assert.JSONEq(t, `{"error":"el servidor encontró un problema"}`, w.Body.String())

		svc.AssertExpectations(t)
	})
}

func TestExcelHandler_Delete(t *testing.T) {
	t.Run("responde 204 sin cuerpo", func(t *testing.T) {
		svc := new(IServiceExcelMock)

		svc.On("Delete", mock.Anything, testEmpresaID, testFileID).
			Return(nil).Once()

		h := newTestHandler(svc)

		req := httptest.NewRequest(http.MethodDelete, "/excel/"+testFileID.String(), nil)

		w := httptest.NewRecorder()
		h.Delete(w, requestConID(requestConContexto(req), testFileID.String()))

		require.Equal(t, http.StatusNoContent, w.Code, w.Body.String())
		assert.Empty(t, w.Body.String())

		svc.AssertExpectations(t)
	})

	t.Run("responde 400 si el id no es un uuid", func(t *testing.T) {
		svc := new(IServiceExcelMock)
		h := newTestHandler(svc)

		req := httptest.NewRequest(http.MethodDelete, "/excel/no-es-uuid", nil)

		w := httptest.NewRecorder()
		h.Delete(w, requestConID(requestConContexto(req), "no-es-uuid"))

		require.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())

		svc.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("responde 404 si el archivo no existe", func(t *testing.T) {
		svc := new(IServiceExcelMock)

		svc.On("Delete", mock.Anything, testEmpresaID, testFileID).
			Return(ErrNotFound).Once()

		h := newTestHandler(svc)

		req := httptest.NewRequest(http.MethodDelete, "/excel/"+testFileID.String(), nil)

		w := httptest.NewRecorder()
		h.Delete(w, requestConID(requestConContexto(req), testFileID.String()))

		require.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
		assert.JSONEq(t, `{"error":"no encontrado"}`, w.Body.String())

		svc.AssertExpectations(t)
	})

	t.Run("responde 500 si el service falla", func(t *testing.T) {
		svc := new(IServiceExcelMock)

		svc.On("Delete", mock.Anything, testEmpresaID, testFileID).
			Return(errors.New("boom")).Once()

		h := newTestHandler(svc)

		req := httptest.NewRequest(http.MethodDelete, "/excel/"+testFileID.String(), nil)

		w := httptest.NewRecorder()
		h.Delete(w, requestConID(requestConContexto(req), testFileID.String()))

		require.Equal(t, http.StatusInternalServerError, w.Code, w.Body.String())
		assert.JSONEq(t, `{"error":"el servidor encontró un problema"}`, w.Body.String())

		svc.AssertExpectations(t)
	})
}

// TestExcelHandler_ServiceRealSobreStoreMockeado comprueba el cableado: el
// handler funciona con el Service real sobre un store mockeado.
func TestExcelHandler_ServiceRealSobreStoreMockeado(t *testing.T) {
	store := new(fileStoreMock)
	fs := new(fileStorageMock)

	store.On("Get", mock.Anything, testEmpresaID, testFileID).
		Return(newTestFile(), nil).Once()

	fs.On("Get", mock.Anything, "clave/prueba.xlsx").
		Return(io.NopCloser(bytes.NewReader(xlsxFalso)), nil).Once()

	h := newTestHandler(&Service{store: store, storage: fs})

	req := httptest.NewRequest(http.MethodGet, "/excel/"+testFileID.String(), nil)

	w := httptest.NewRecorder()
	h.Download(w, requestConID(requestConContexto(req), testFileID.String()))

	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	assert.Equal(t, xlsxFalso, w.Body.Bytes())

	store.AssertExpectations(t)
	fs.AssertExpectations(t)
}
