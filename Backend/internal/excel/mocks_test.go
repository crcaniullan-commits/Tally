package excel

import (
	"bytes"
	"context"
	"database/sql"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/crcaniullan-commits/Tally/internal/auth"
	"github.com/crcaniullan-commits/Tally/internal/errores"
	"github.com/crcaniullan-commits/Tally/internal/intermediario"
	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/crcaniullan-commits/Tally/internal/storage"
	"github.com/crcaniullan-commits/Tally/internal/utilidades"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

var (
	testUserID    = uuid.MustParse("6f1a1b3c-2d4e-4f60-8a9b-0c1d2e3f4a5b")
	testEmpresaID = uuid.MustParse("1a2b3c4d-5e6f-4a7b-8c9d-0e1f2a3b4c5d")
	testFileID    = uuid.MustParse("d4e5f6a7-8b9c-4d0e-9f1a-2b3c4d5e6f7a")
)

// xlsxFalso es un archivo con el magic de un zip ("PK"), que es lo que el
// handler exige para tratarlo como .xlsx. No es un xlsx válido: para las rutas
// probadas solo importa que pase la validación de los primeros dos bytes.
var xlsxFalso = []byte("PK\x03\x04contenido-de-prueba")

var testCreatedAt = time.Date(2026, 11, 14, 12, 0, 0, 0, time.UTC)

// storageMock es un mock de FileStorage que guarda los objetos en memoria.
// Sirve para probar las rutas de excel sin tocar disco ni S3.
type storageMock struct {
	mu      sync.Mutex
	objetos map[string][]byte
	tipos   map[string]string

	puts    int
	gets    int
	deletes int

	ultimaClave string
	putErr      error
	getErr      error
	deleteErr   error
}

func newStorageMock() *storageMock {
	return &storageMock{
		objetos: make(map[string][]byte),
		tipos:   make(map[string]string),
	}
}

func (m *storageMock) Put(ctx context.Context, key string, body io.ReadSeeker, size int64, contentType string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.puts++
	m.ultimaClave = key

	if m.putErr != nil {
		return m.putErr
	}

	data, err := io.ReadAll(body)
	if err != nil {
		return err
	}

	m.objetos[key] = data
	m.tipos[key] = contentType

	return nil
}

func (m *storageMock) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.gets++

	if m.getErr != nil {
		return nil, m.getErr
	}

	data, ok := m.objetos[key]
	if !ok {
		return nil, storage.ErrNotFound
	}

	return io.NopCloser(bytes.NewReader(data)), nil
}

func (m *storageMock) Delete(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.deletes++

	if m.deleteErr != nil {
		return m.deleteErr
	}

	delete(m.objetos, key)

	return nil
}

// existe informa si el objeto está guardado en el mock.
func (m *storageMock) existe(key string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	_, ok := m.objetos[key]

	return ok
}

// objetosCount devuelve cuántos objetos hay guardados.
func (m *storageMock) objetosCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()

	return len(m.objetos)
}

// fileStoreMock es un mock de la interfaz fileStore (la que usa el service).
type fileStoreMock struct {
	mock.Mock
}

func (m *fileStoreMock) Insert(ctx context.Context, f *File) error {
	args := m.Called(ctx, f)

	return args.Error(0)
}

func (m *fileStoreMock) List(ctx context.Context, empresaID uuid.UUID, limit, offset int) ([]File, error) {
	args := m.Called(ctx, empresaID, limit, offset)

	files, _ := args.Get(0).([]File)

	return files, args.Error(1)
}

func (m *fileStoreMock) Get(ctx context.Context, empresaID, id uuid.UUID) (File, error) {
	args := m.Called(ctx, empresaID, id)

	f, _ := args.Get(0).(File)

	return f, args.Error(1)
}

func (m *fileStoreMock) Delete(ctx context.Context, empresaID, id uuid.UUID) (string, error) {
	args := m.Called(ctx, empresaID, id)

	key, _ := args.Get(0).(string)

	return key, args.Error(1)
}

// fileStorageMock es un mock de la interfaz FileStorage (la que usa el
// service). A diferencia de storageMock, este no guarda nada: solo registra
// las llamadas.
type fileStorageMock struct {
	mock.Mock
}

func (m *fileStorageMock) Put(ctx context.Context, key string, body io.ReadSeeker, size int64, contentType string) error {
	args := m.Called(ctx, key, body, size, contentType)

	return args.Error(0)
}

func (m *fileStorageMock) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	args := m.Called(ctx, key)

	rc, _ := args.Get(0).(io.ReadCloser)

	return rc, args.Error(1)
}

func (m *fileStorageMock) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)

	return args.Error(0)
}

// newTestFile arma un File válido y reutilizable en los tests.
func newTestFile() File {
	return File{
		ID:         testFileID,
		EmpresaID:  testEmpresaID,
		Nombre:     "reporte.xlsx",
		SizeBytes:  int64(len(xlsxFalso)),
		CreatedAt:  testCreatedAt,
		StorageKey: "clave/prueba.xlsx",
	}
}

// IServiceExcelMock es un mock de IServiceExcel para probar el Handler de
// forma aislada, sin store ni storage.
type IServiceExcelMock struct {
	mock.Mock
}

func (m *IServiceExcelMock) Save(ctx context.Context, userID, empresaID uuid.UUID, nombre string, body io.ReadSeeker, size int64) (*File, error) {
	args := m.Called(ctx, userID, empresaID, nombre, body, size)

	f, _ := args.Get(0).(*File)

	return f, args.Error(1)
}

func (m *IServiceExcelMock) List(ctx context.Context, empresaID uuid.UUID, limit, offset int) ([]File, error) {
	args := m.Called(ctx, empresaID, limit, offset)

	files, _ := args.Get(0).([]File)

	return files, args.Error(1)
}

func (m *IServiceExcelMock) Open(ctx context.Context, empresaID, id uuid.UUID) (*File, io.ReadCloser, error) {
	args := m.Called(ctx, empresaID, id)

	f, _ := args.Get(0).(*File)
	rc, _ := args.Get(1).(io.ReadCloser)

	return f, rc, args.Error(2)
}

func (m *IServiceExcelMock) Delete(ctx context.Context, empresaID, id uuid.UUID) error {
	args := m.Called(ctx, empresaID, id)

	return args.Error(0)
}

// sqlCaptura guarda el último SQL que ejecutó el store. sqlmock no parsea SQL
// (solo compara strings), así que el SQL real se audita a mano.
type sqlCaptura struct {
	mu  sync.Mutex
	sql string
}

func (c *sqlCaptura) set(q string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.sql = q
}

func (c *sqlCaptura) get() string {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.sql
}

// newTestDB levanta una conexión sqlmock con un matcher permisivo: las
// expectativas se definen por orden de ejecución y el SQL capturado se audita
// a mano.
func newTestDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *sqlCaptura) {
	t.Helper()

	captura := &sqlCaptura{}

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(
		sqlmock.QueryMatcherFunc(func(expectedSQL, actualSQL string) error {
			captura.set(actualSQL)

			return nil
		}),
	))
	if err != nil {
		t.Fatalf("no se pudo crear el sqlmock: %v", err)
	}

	t.Cleanup(func() {
		mock.ExpectClose()

		if err := db.Close(); err != nil {
			t.Logf("error cerrando la db del mock: %v", err)
		}
	})

	return db, mock, captura
}

// stubAutenticador implementa auth.Authenticator. Las rutas de excel no
// validan tokens (solo el rol), así que no necesita hacer nada real.
type stubAutenticador struct{}

var _ auth.Authenticator = stubAutenticador{}

func (stubAutenticador) GenerarToken(claims jwt.Claims) (string, error) { return "", nil }

func (stubAutenticador) ValidarToken(token string) (*jwt.Token, error) { return nil, nil }

// stubBuscadorUsuario implementa intermediario.BuscadorUsuario.
type stubBuscadorUsuario struct{}

func (stubBuscadorUsuario) ObtenerPorID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	return &model.User{ID: id, Role: utilidades.RolUsuario}, nil
}

// newTestLogger devuelve un logger que no ensucia la salida del test.
func newTestLogger() *zap.SugaredLogger {
	return zap.NewNop().Sugar()
}

// newTestMiddleware arma el middleware de autenticación real (el que aplica
// VerificarRol) con stubs para lo que las rutas de excel no usan.
func newTestMiddleware() *intermediario.MiddlewareAutenticacion {
	return intermediario.NuevoMiddlewareAutenticacion(
		errores.NuevaRespuestaErrores(newTestLogger()),
		stubAutenticador{},
		stubBuscadorUsuario{},
	)
}
