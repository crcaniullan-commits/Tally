package excel

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestStore arma el Store real sobre una conexión sqlmock.
func newTestStore(t *testing.T) (*Store, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, _ := newTestDB(t)

	return &Store{db: db}, mock
}
func TestExcelStore_Insert(t *testing.T) {
	t.Run("inserta el archivo y trae el created_at", func(t *testing.T) {
		store, mock := newTestStore(t)

		mock.ExpectQuery(".*").
			WillReturnRows(sqlmock.NewRows([]string{"created_at"}).AddRow(testCreatedAt))

		archivo := &File{
			ID:         testFileID,
			EmpresaID:  testEmpresaID,
			Nombre:     "reporte.xlsx",
			StorageKey: "clave/prueba.xlsx",
			SizeBytes:  int64(len(xlsxFalso)),
		}

		err := store.Insert(context.Background(), archivo)

		require.NoError(t, err)
		assert.Equal(t, testCreatedAt, archivo.CreatedAt.UTC())

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("propaga el error de la base", func(t *testing.T) {
		store, mock := newTestStore(t)

		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		archivo := &File{
			ID:         testFileID,
			EmpresaID:  testEmpresaID,
			Nombre:     "reporte.xlsx",
			StorageKey: "clave/prueba.xlsx",
			SizeBytes:  int64(len(xlsxFalso)),
		}

		err := store.Insert(context.Background(), archivo)

		require.Error(t, err)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestExcelStore_List(t *testing.T) {
	t.Run("devuelve los archivos de la empresa", func(t *testing.T) {
		store, mock := newTestStore(t)

		mock.ExpectQuery(".*").WillReturnRows(filasExcel())

		archivos, err := store.List(context.Background(), testEmpresaID, 20, 0)

		require.NoError(t, err)
		require.Len(t, archivos, 1)

		assert.Equal(t, testFileID, archivos[0].ID)
		assert.Equal(t, testEmpresaID, archivos[0].EmpresaID)
		assert.Equal(t, "reporte.xlsx", archivos[0].Nombre)
		assert.Equal(t, "clave/prueba.xlsx", archivos[0].StorageKey)
		assert.Equal(t, int64(len(xlsxFalso)), archivos[0].SizeBytes)
		assert.Equal(t, testCreatedAt, archivos[0].CreatedAt.UTC())

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("devuelve una lista vacía sin error", func(t *testing.T) {
		store, mock := newTestStore(t)

		mock.ExpectQuery(".*").
			WillReturnRows(sqlmock.NewRows([]string{"id", "empresa_id", "nombre", "storage_key", "size_bytes", "created_at"}))

		archivos, err := store.List(context.Background(), testEmpresaID, 20, 0)

		require.NoError(t, err)
		assert.Empty(t, archivos)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("propaga el error de escaneo de una fila", func(t *testing.T) {
		store, mock := newTestStore(t)

		// size_bytes con un valor que no entra en int64
		mock.ExpectQuery(".*").WillReturnRows(
			sqlmock.NewRows([]string{"id", "empresa_id", "nombre", "storage_key", "size_bytes", "created_at"}).
				AddRow(testFileID, testEmpresaID, "reporte.xlsx", "clave/prueba.xlsx", "no-es-un-numero", testCreatedAt),
		)

		archivos, err := store.List(context.Background(), testEmpresaID, 20, 0)

		require.Error(t, err)
		assert.Nil(t, archivos)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("propaga el error de la base", func(t *testing.T) {
		store, mock := newTestStore(t)

		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		archivos, err := store.List(context.Background(), testEmpresaID, 20, 0)

		require.Error(t, err)
		assert.Nil(t, archivos)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestExcelStore_Get(t *testing.T) {
	t.Run("devuelve el archivo de la empresa", func(t *testing.T) {
		store, mock := newTestStore(t)

		mock.ExpectQuery(".*").WillReturnRows(filasExcel())

		archivo, err := store.Get(context.Background(), testEmpresaID, testFileID)

		require.NoError(t, err)
		assert.Equal(t, testFileID, archivo.ID)
		assert.Equal(t, testEmpresaID, archivo.EmpresaID)
		assert.Equal(t, "clave/prueba.xlsx", archivo.StorageKey)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("devuelve ErrNotFound si no hay filas", func(t *testing.T) {
		store, mock := newTestStore(t)

		mock.ExpectQuery(".*").
			WillReturnRows(sqlmock.NewRows([]string{"id", "empresa_id", "nombre", "storage_key", "size_bytes", "created_at"}))

		archivo, err := store.Get(context.Background(), testEmpresaID, testFileID)

		require.ErrorIs(t, err, ErrNotFound)
		assert.Equal(t, File{}, archivo)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("propaga el error de la base", func(t *testing.T) {
		store, mock := newTestStore(t)

		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		archivo, err := store.Get(context.Background(), testEmpresaID, testFileID)

		require.Error(t, err)
		require.NotErrorIs(t, err, ErrNotFound)
		assert.Equal(t, File{}, archivo)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestExcelStore_Delete(t *testing.T) {
	t.Run("borra el archivo y devuelve la clave del objeto", func(t *testing.T) {
		store, mock := newTestStore(t)

		mock.ExpectQuery(".*").
			WillReturnRows(sqlmock.NewRows([]string{"storage_key"}).AddRow("clave/prueba.xlsx"))

		clave, err := store.Delete(context.Background(), testEmpresaID, testFileID)

		require.NoError(t, err)
		assert.Equal(t, "clave/prueba.xlsx", clave)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("devuelve ErrNotFound si no hay filas", func(t *testing.T) {
		store, mock := newTestStore(t)

		mock.ExpectQuery(".*").WillReturnRows(sqlmock.NewRows([]string{"storage_key"}))

		clave, err := store.Delete(context.Background(), testEmpresaID, testFileID)

		require.ErrorIs(t, err, ErrNotFound)
		assert.Empty(t, clave)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("propaga el error de la base", func(t *testing.T) {
		store, mock := newTestStore(t)

		mock.ExpectQuery(".*").WillReturnError(errors.New("boom"))

		clave, err := store.Delete(context.Background(), testEmpresaID, testFileID)

		require.Error(t, err)
		require.NotErrorIs(t, err, ErrNotFound)
		assert.Empty(t, clave)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

// TestExcelStore_UsaLaEmpresaEnElWhere audita el SQL real del store: el
// acotado por empresa va en el WHERE, no en la aplicación.
func TestExcelStore_UsaLaEmpresaEnElWhere(t *testing.T) {
	t.Run("el listado filtra por empresa_id", func(t *testing.T) {
		db, mock, captura := newTestDB(t)
		store := &Store{db: db}

		mock.ExpectQuery(".*").WillReturnRows(filasExcel())

		_, err := store.List(context.Background(), testEmpresaID, 20, 0)
		require.NoError(t, err)

		sql := captura.get()

		assert.Contains(t, sql, "FROM excel_files")
		assert.Contains(t, sql, "WHERE empresa_id = $1")
		assert.NotContains(t, sql, "emrpesa_id")

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("la obtención filtra por id y empresa", func(t *testing.T) {
		db, mock, captura := newTestDB(t)
		store := &Store{db: db}

		mock.ExpectQuery(".*").WillReturnRows(filasExcel())

		_, err := store.Get(context.Background(), testEmpresaID, testFileID)
		require.NoError(t, err)

		sql := captura.get()

		assert.Contains(t, sql, "FROM excel_files")
		assert.Contains(t, sql, "WHERE id = $1 AND empresa_id = $2")
		assert.NotContains(t, sql, "exel_files")

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("el borrado filtra por id y empresa", func(t *testing.T) {
		db, mock, captura := newTestDB(t)
		store := &Store{db: db}

		mock.ExpectQuery(".*").
			WillReturnRows(sqlmock.NewRows([]string{"storage_key"}).AddRow("clave/prueba.xlsx"))

		_, err := store.Delete(context.Background(), testEmpresaID, testFileID)
		require.NoError(t, err)

		sql := captura.get()

		assert.Contains(t, sql, "DELETE FROM excel_files")
		assert.Contains(t, sql, "WHERE id = $1 AND empresa_id = $2")

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("el insert guarda la empresa y la clave del objeto", func(t *testing.T) {
		db, mock, captura := newTestDB(t)
		store := &Store{db: db}

		mock.ExpectQuery(".*").
			WillReturnRows(sqlmock.NewRows([]string{"created_at"}).AddRow(testCreatedAt))

		archivo := &File{
			ID:         testFileID,
			EmpresaID:  testEmpresaID,
			Nombre:     "reporte.xlsx",
			StorageKey: "clave/prueba.xlsx",
			SizeBytes:  int64(len(xlsxFalso)),
		}

		require.NoError(t, store.Insert(context.Background(), archivo))

		sql := captura.get()

		assert.Contains(t, sql, "INSERT INTO excel_files")
		assert.Contains(t, sql, "empresa_id")

		require.NoError(t, mock.ExpectationsWereMet())
	})
}
