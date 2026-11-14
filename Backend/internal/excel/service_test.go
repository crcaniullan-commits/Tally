package excel

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/crcaniullan-commits/Tally/internal/storage"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// newTestService arma el Service con el store y el storage mockeados.
func newTestService(store fileStore, fs FileStorage) *Service {
	return &Service{store: store, storage: fs}
}

func TestExcelService_Save(t *testing.T) {
	t.Run("guarda el objeto y después crea el registro", func(t *testing.T) {
		store := new(fileStoreMock)
		fs := new(fileStorageMock)
		svc := newTestService(store, fs)

		var clave string

		fs.On("Put", mock.Anything, mock.Anything, mock.Anything, int64(len(xlsxFalso)), xlsxType).
			Run(func(args mock.Arguments) { clave = args.String(1) }).
			Return(nil).Once()

		store.On("Insert", mock.Anything, mock.Anything).
			Return(nil).Once()

		archivo, err := svc.Save(context.Background(), testUserID, testEmpresaID,
			"reporte.xlsx", bytes.NewReader(xlsxFalso), int64(len(xlsxFalso)))

		require.NoError(t, err)
		require.NotNil(t, archivo)

		assert.Equal(t, testEmpresaID, archivo.EmpresaID)
		assert.Equal(t, "reporte.xlsx", archivo.Nombre)
		assert.Equal(t, int64(len(xlsxFalso)), archivo.SizeBytes)
		assert.NotEqual(t, uuid.Nil, archivo.ID)

		// la clave del objeto lleva usuario, empresa y el id del archivo
		assert.Equal(t, testUserID.String()+"/"+testEmpresaID.String()+"/"+archivo.ID.String()+".xlsx", archivo.StorageKey)
		assert.Equal(t, archivo.StorageKey, clave)

		fs.AssertExpectations(t)
		store.AssertExpectations(t)
	})

	t.Run("si el storage falla no crea el registro", func(t *testing.T) {
		store := new(fileStoreMock)
		fs := new(fileStorageMock)
		svc := newTestService(store, fs)

		fs.On("Put", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(errors.New("boom")).Once()

		archivo, err := svc.Save(context.Background(), testUserID, testEmpresaID,
			"reporte.xlsx", bytes.NewReader(xlsxFalso), int64(len(xlsxFalso)))

		require.Error(t, err)
		assert.Contains(t, err.Error(), "guardando archivo")
		assert.Nil(t, archivo)

		store.AssertNotCalled(t, "Insert", mock.Anything, mock.Anything)
		fs.AssertExpectations(t)
	})

	t.Run("si el registro falla borra el objeto para compensar", func(t *testing.T) {
		store := new(fileStoreMock)
		fs := new(fileStorageMock)
		svc := newTestService(store, fs)

		fs.On("Put", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(nil).Once()

		store.On("Insert", mock.Anything, mock.Anything).
			Return(errors.New("boom")).Once()

		var claveBorrada string

		fs.On("Delete", mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) { claveBorrada = args.String(1) }).
			Return(nil).Once()

		archivo, err := svc.Save(context.Background(), testUserID, testEmpresaID,
			"reporte.xlsx", bytes.NewReader(xlsxFalso), int64(len(xlsxFalso)))

		require.Error(t, err)
		assert.Nil(t, archivo)
		assert.Contains(t, claveBorrada, testEmpresaID.String())

		fs.AssertExpectations(t)
		store.AssertExpectations(t)
	})
}

func TestExcelService_List(t *testing.T) {
	t.Run("devuelve los archivos del store", func(t *testing.T) {
		store := new(fileStoreMock)
		fs := new(fileStorageMock)
		svc := newTestService(store, fs)

		esperados := []File{newTestFile()}

		store.On("List", mock.Anything, testEmpresaID, 20, 0).
			Return(esperados, nil).Once()

		archivos, err := svc.List(context.Background(), testEmpresaID, 20, 0)

		require.NoError(t, err)
		require.Len(t, archivos, 1)
		assert.Equal(t, testFileID, archivos[0].ID)

		store.AssertExpectations(t)
	})

	t.Run("propaga el error del store", func(t *testing.T) {
		store := new(fileStoreMock)
		fs := new(fileStorageMock)
		svc := newTestService(store, fs)

		store.On("List", mock.Anything, testEmpresaID, mock.Anything, mock.Anything).
			Return(nil, errors.New("boom")).Once()

		archivos, err := svc.List(context.Background(), testEmpresaID, 20, 0)

		require.Error(t, err)
		assert.Nil(t, archivos)

		store.AssertExpectations(t)
	})
}

func TestExcelService_Open(t *testing.T) {
	t.Run("devuelve el archivo y su contenido", func(t *testing.T) {
		store := new(fileStoreMock)
		fs := new(fileStorageMock)
		svc := newTestService(store, fs)

		esperado := newTestFile()

		store.On("Get", mock.Anything, testEmpresaID, testFileID).
			Return(esperado, nil).Once()

		fs.On("Get", mock.Anything, esperado.StorageKey).
			Return(io.NopCloser(bytes.NewReader(xlsxFalso)), nil).Once()

		archivo, rc, err := svc.Open(context.Background(), testEmpresaID, testFileID)

		require.NoError(t, err)
		require.NotNil(t, archivo)
		require.NotNil(t, rc)

		defer rc.Close()

		assert.Equal(t, testFileID, archivo.ID)

		contenido, err := io.ReadAll(rc)
		require.NoError(t, err)
		assert.Equal(t, xlsxFalso, contenido)

		store.AssertExpectations(t)
		fs.AssertExpectations(t)
	})

	t.Run("propaga el error del store", func(t *testing.T) {
		store := new(fileStoreMock)
		fs := new(fileStorageMock)
		svc := newTestService(store, fs)

		store.On("Get", mock.Anything, testEmpresaID, testFileID).
			Return(File{}, ErrNotFound).Once()

		archivo, rc, err := svc.Open(context.Background(), testEmpresaID, testFileID)

		require.ErrorIs(t, err, ErrNotFound)
		assert.Nil(t, archivo)
		assert.Nil(t, rc)

		fs.AssertNotCalled(t, "Get", mock.Anything, mock.Anything)
		store.AssertExpectations(t)
	})

	t.Run("traduce el objeto inexistente del storage a ErrNotFound", func(t *testing.T) {
		store := new(fileStoreMock)
		fs := new(fileStorageMock)
		svc := newTestService(store, fs)

		esperado := newTestFile()

		store.On("Get", mock.Anything, testEmpresaID, testFileID).
			Return(esperado, nil).Once()

		fs.On("Get", mock.Anything, esperado.StorageKey).
			Return(nil, storage.ErrNotFound).Once()

		archivo, rc, err := svc.Open(context.Background(), testEmpresaID, testFileID)

		require.ErrorIs(t, err, ErrNotFound)
		assert.Nil(t, archivo)
		assert.Nil(t, rc)

		store.AssertExpectations(t)
		fs.AssertExpectations(t)
	})

	t.Run("propaga otros errores del storage", func(t *testing.T) {
		store := new(fileStoreMock)
		fs := new(fileStorageMock)
		svc := newTestService(store, fs)

		esperado := newTestFile()

		store.On("Get", mock.Anything, testEmpresaID, testFileID).
			Return(esperado, nil).Once()

		fs.On("Get", mock.Anything, esperado.StorageKey).
			Return(nil, errors.New("boom")).Once()

		archivo, rc, err := svc.Open(context.Background(), testEmpresaID, testFileID)

		require.Error(t, err)
		assert.NotErrorIs(t, err, ErrNotFound)
		assert.Nil(t, archivo)
		assert.Nil(t, rc)

		store.AssertExpectations(t)
		fs.AssertExpectations(t)
	})
}

func TestExcelService_Delete(t *testing.T) {
	t.Run("borra el registro y el objeto", func(t *testing.T) {
		store := new(fileStoreMock)
		fs := new(fileStorageMock)
		svc := newTestService(store, fs)

		store.On("Delete", mock.Anything, testEmpresaID, testFileID).
			Return("clave/prueba.xlsx", nil).Once()

		fs.On("Delete", mock.Anything, "clave/prueba.xlsx").
			Return(nil).Once()

		err := svc.Delete(context.Background(), testEmpresaID, testFileID)

		require.NoError(t, err)

		store.AssertExpectations(t)
		fs.AssertExpectations(t)
	})

	t.Run("propaga el ErrNotFound del store y no toca el storage", func(t *testing.T) {
		store := new(fileStoreMock)
		fs := new(fileStorageMock)
		svc := newTestService(store, fs)

		store.On("Delete", mock.Anything, testEmpresaID, testFileID).
			Return("", ErrNotFound).Once()

		err := svc.Delete(context.Background(), testEmpresaID, testFileID)

		require.ErrorIs(t, err, ErrNotFound)

		fs.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything)
		store.AssertExpectations(t)
	})

	t.Run("ignora el error al borrar el objeto", func(t *testing.T) {
		store := new(fileStoreMock)
		fs := new(fileStorageMock)
		svc := newTestService(store, fs)

		store.On("Delete", mock.Anything, testEmpresaID, testFileID).
			Return("clave/prueba.xlsx", nil).Once()

		fs.On("Delete", mock.Anything, "clave/prueba.xlsx").
			Return(errors.New("boom")).Once()

		err := svc.Delete(context.Background(), testEmpresaID, testFileID)

		require.NoError(t, err, "el registro ya se borró: el objeto huérfano no es error")

		store.AssertExpectations(t)
		fs.AssertExpectations(t)
	})
}

// testStorageKeyFormato documenta el formato de la clave del objeto.
func TestExcelService_StorageKey(t *testing.T) {
	store := new(fileStoreMock)
	fs := new(fileStorageMock)
	svc := newTestService(store, fs)

	fs.On("Put", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil).Once()

	store.On("Insert", mock.Anything, mock.Anything).Return(nil).Once()

	archivo, err := svc.Save(context.Background(), testUserID, testEmpresaID,
		"reporte.xlsx", strings.NewReader("contenido"), 8)

	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(archivo.StorageKey, testUserID.String()+"/"))
	assert.True(t, strings.Contains(archivo.StorageKey, testEmpresaID.String()+"/"))
	assert.True(t, strings.HasSuffix(archivo.StorageKey, ".xlsx"))
}
