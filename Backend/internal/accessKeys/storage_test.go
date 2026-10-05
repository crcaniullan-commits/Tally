package accesskeys

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/crcaniullan-commits/Tally/internal/dbtx"
	"github.com/crcaniullan-commits/Tally/internal/util"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errDBBoom = errors.New("boom: fallo de la base de datos")

// newMockStore levanta un StoreAccessKey sobre una conexión sqlmockeada.
func newMockStore(t *testing.T) (*StoreAccessKey, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	t.Cleanup(func() {
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("expectativas de sql sin cumplir: %v", err)
		}
		mock.ExpectClose()
		_ = db.Close()
	})

	return NewStorage(db), mock
}

// Fragments de las queries reales, para no duplicar el SQL en cada test.
var (
	regexInsert  = regexp.QuoteMeta("INSERT INTO access_keys")
	regexGetByID = regexp.QuoteMeta("SELECT id, code, created_by, redeemed_by")
	regexRedeem  = regexp.QuoteMeta("UPDATE access_keys")
	regexRevoke  = regexp.QuoteMeta("UPDATE access_keys")
)

// accessKeyColumns son las columnas de GetByID, en el mismo orden en que el
// Scan las lee: id, code, created_by, redeemed_by, redeemed_at, expires_at,
// revoked_at, revoked_by, created_at.
func accessKeyColumns() []string {
	return []string{
		"id", "code", "created_by", "redeemed_by",
		"redeemed_at", "expires_at", "revoked_at", "revoked_by", "created_at",
	}
}

// accessKeyRow devuelve una llave vigente y no canjeada. redeemed_by,
// redeemed_at, revoked_at y revoked_by van en NULL, que es el estado normal de
// una llave recién emitida.
func accessKeyRow(expiresAt time.Time) []driver.Value {
	return []driver.Value{
		testAccessKeyID,
		testAccessKeyCod,
		testMunicipalID,
		nil,
		nil,
		expiresAt,
		nil,
		nil,
		time.Now().Add(-time.Hour),
	}
}

func TestAccessKeyStore_Create(t *testing.T) {
	t.Run("inserta código, emisor y vencimiento, y devuelve el id generado", func(t *testing.T) {
		store, mock := newMockStore(t)

		accessKey := newTestAccessKey()

		mock.ExpectQuery(regexInsert).
			WithArgs(testAccessKeyCod, testMunicipalID, *accessKey.ExpiresAt).
			WillReturnRows(sqlmock.NewRows([]string{"code", "id"}).
				AddRow(testAccessKeyCod, testAccessKeyID))

		require.NoError(t, store.Create(context.Background(), &accessKey))

		assert.Equal(t, testAccessKeyID, accessKey.ID, "el id vuelve en el RETURNING")
		assert.Equal(t, testAccessKeyCod, accessKey.Code)
	})

	t.Run("persiste el código que le pasó el service, no uno nuevo", func(t *testing.T) {
		store, mock := newMockStore(t)

		accessKey := newTestAccessKey()
		accessKey.Code = "ZCL35ZHKHS5YDWKIPAWCIAB3JR"

		mock.ExpectQuery(regexInsert).
			WithArgs("ZCL35ZHKHS5YDWKIPAWCIAB3JR", testMunicipalID, *accessKey.ExpiresAt).
			WillReturnRows(sqlmock.NewRows([]string{"code", "id"}).
				AddRow("ZCL35ZHKHS5YDWKIPAWCIAB3JR", uuid.New()))

		require.NoError(t, store.Create(context.Background(), &accessKey))
		assert.Equal(t, "ZCL35ZHKHS5YDWKIPAWCIAB3JR", accessKey.Code, "el RETURNING no puede pisar el código")
	})

	t.Run("acepta una llave sin vencimiento", func(t *testing.T) {
		store, mock := newMockStore(t)

		accessKey := newTestAccessKey()
		accessKey.ExpiresAt = nil

		mock.ExpectQuery(regexInsert).
			WithArgs(testAccessKeyCod, testMunicipalID, nil).
			WillReturnRows(sqlmock.NewRows([]string{"code", "id"}).
				AddRow(testAccessKeyCod, testAccessKeyID))

		require.NoError(t, store.Create(context.Background(), &accessKey))
	})

	t.Run("propaga el error de la base", func(t *testing.T) {
		store, mock := newMockStore(t)

		accessKey := newTestAccessKey()

		mock.ExpectQuery(regexInsert).
			WithArgs(testAccessKeyCod, testMunicipalID, *accessKey.ExpiresAt).
			WillReturnError(errDBBoom)

		err := store.Create(context.Background(), &accessKey)

		require.ErrorIs(t, err, errDBBoom)
	})
}

func TestAccessKeyStore_GetByID(t *testing.T) {
	t.Run("devuelve la llave con las 9 columnas de access_keys", func(t *testing.T) {
		store, mock := newMockStore(t)

		expire := time.Now().AddDate(1, 0, 0)

		mock.ExpectQuery(regexGetByID).
			WithArgs(testAccessKeyID).
			WillReturnRows(sqlmock.NewRows(accessKeyColumns()).AddRow(accessKeyRow(expire)...))

		accessKey, err := store.GetByID(context.Background(), testAccessKeyID)

		require.NoError(t, err)
		assert.Equal(t, testAccessKeyID, accessKey.ID)
		assert.Equal(t, testAccessKeyCod, accessKey.Code)
		assert.Equal(t, testMunicipalID, accessKey.CreatedBy)
		assert.Nil(t, accessKey.RedeemedBy)
		assert.Nil(t, accessKey.RedeemedAt)
		require.NotNil(t, accessKey.ExpiresAt)
		assert.True(t, expire.Equal(*accessKey.ExpiresAt))
		assert.Nil(t, accessKey.RevokedAt)
		assert.Nil(t, accessKey.RevokedBy)
	})

	t.Run("devuelve una llave ya canjeada y revocada con sus auditores", func(t *testing.T) {
		store, mock := newMockStore(t)

		expire := time.Now().AddDate(1, 0, 0)
		canjeado := time.Now().Add(-2 * time.Hour)
		revocado := time.Now().Add(-time.Hour)

		row := accessKeyRow(expire)
		row[3] = testEmprendedor
		row[4] = canjeado
		row[6] = revocado
		row[7] = testMunicipalID

		mock.ExpectQuery(regexGetByID).
			WithArgs(testAccessKeyID).
			WillReturnRows(sqlmock.NewRows(accessKeyColumns()).AddRow(row...))

		accessKey, err := store.GetByID(context.Background(), testAccessKeyID)

		require.NoError(t, err)
		require.NotNil(t, accessKey.RedeemedBy)
		assert.Equal(t, testEmprendedor, *accessKey.RedeemedBy)
		require.NotNil(t, accessKey.RedeemedAt)
		assert.True(t, canjeado.Equal(*accessKey.RedeemedAt))
		require.NotNil(t, accessKey.RevokedBy)
		assert.Equal(t, testMunicipalID, *accessKey.RevokedBy)
		require.NotNil(t, accessKey.RevokedAt)
		assert.True(t, revocado.Equal(*accessKey.RevokedAt))
	})

	t.Run("tolera expires_at NULL", func(t *testing.T) {
		store, mock := newMockStore(t)

		row := accessKeyRow(time.Now())
		row[5] = nil

		mock.ExpectQuery(regexGetByID).
			WithArgs(testAccessKeyID).
			WillReturnRows(sqlmock.NewRows(accessKeyColumns()).AddRow(row...))

		accessKey, err := store.GetByID(context.Background(), testAccessKeyID)

		require.NoError(t, err)
		assert.Nil(t, accessKey.ExpiresAt)
	})

	t.Run("traduce sql.ErrNoRows a util.ErrNotFound", func(t *testing.T) {
		// Es lo que permite que el handler de Resend distinga "no existe" (404)
		// de "se rompió algo" (500), como ya hacían Revoke y los stores de
		// users.
		store, mock := newMockStore(t)

		mock.ExpectQuery(regexGetByID).
			WithArgs(testAccessKeyID).
			WillReturnRows(sqlmock.NewRows(accessKeyColumns()))

		accessKey, err := store.GetByID(context.Background(), testAccessKeyID)

		require.ErrorIs(t, err, util.ErrNotFound)
		assert.NotErrorIs(t, err, sql.ErrNoRows, "el error de sql no debe filtrarse al handler")
		assert.Equal(t, uuid.Nil, accessKey.ID)
	})

	t.Run("propaga otros errores de la base", func(t *testing.T) {
		store, mock := newMockStore(t)

		mock.ExpectQuery(regexGetByID).
			WithArgs(testAccessKeyID).
			WillReturnError(errDBBoom)

		accessKey, err := store.GetByID(context.Background(), testAccessKeyID)

		require.ErrorIs(t, err, errDBBoom)
		assert.Equal(t, uuid.Nil, accessKey.ID)
	})
}

func TestAccessKeyStore_Redeem(t *testing.T) {
	t.Run("marca la llave como canjeada y devuelve su vencimiento", func(t *testing.T) {
		store, mock := newMockStore(t)

		expire := time.Now().AddDate(1, 0, 0)

		mock.ExpectQuery(regexRedeem).
			WithArgs(testEmprendedor, testAccessKeyCod).
			WillReturnRows(sqlmock.NewRows([]string{"id", "expires_at"}).
				AddRow(testAccessKeyID, expire))

		accessKey, err := store.Redeem(context.Background(), testAccessKeyCod, testEmprendedor)

		require.NoError(t, err)
		require.NotNil(t, accessKey)
		assert.Equal(t, testAccessKeyID, accessKey.ID)
		require.NotNil(t, accessKey.ExpiresAt)
		assert.True(t, expire.Equal(*accessKey.ExpiresAt))
	})

	t.Run("traduce sql.ErrNoRows a ErrNotRedeemable", func(t *testing.T) {
		// El WHERE filtra redeemed_by IS NULL AND expires_at > now() AND
		// revoked_at IS NULL, así que cero filas = canjeado, vencido o revocado.
		store, mock := newMockStore(t)

		mock.ExpectQuery(regexRedeem).
			WithArgs(testEmprendedor, testAccessKeyCod).
			WillReturnRows(sqlmock.NewRows([]string{"id", "expires_at"}))

		accessKey, err := store.Redeem(context.Background(), testAccessKeyCod, testEmprendedor)

		require.ErrorIs(t, err, ErrNotRedeemable)
		assert.NotErrorIs(t, err, sql.ErrNoRows, "el error de sql no debe filtrarse al handler")
		assert.Nil(t, accessKey)
	})

	t.Run("envuelve otros errores de la base", func(t *testing.T) {
		store, mock := newMockStore(t)

		mock.ExpectQuery(regexRedeem).
			WithArgs(testEmprendedor, testAccessKeyCod).
			WillReturnError(errDBBoom)

		accessKey, err := store.Redeem(context.Background(), testAccessKeyCod, testEmprendedor)

		require.ErrorIs(t, err, errDBBoom)
		assert.Contains(t, err.Error(), "canjeando access key")
		assert.Nil(t, accessKey)
	})

	t.Run("canjea por la transacción del contexto si hay una abierta", func(t *testing.T) {
		// dbtx.FromContext: el canje y la extensión del plan en users viajan en
		// la misma transacción, así que esto no puede salir por la conexión
		// suelta o el canje queda confirmado sin plan.
		db, sqlm, err := sqlmock.New()
		require.NoError(t, err)

		t.Cleanup(func() {
			if err := sqlm.ExpectationsWereMet(); err != nil {
				t.Errorf("expectativas de sql sin cumplir: %v", err)
			}
			_ = db.Close()
		})

		expire := time.Now().AddDate(1, 0, 0)

		sqlm.ExpectBegin()
		sqlm.ExpectQuery(regexRedeem).
			WithArgs(testEmprendedor, testAccessKeyCod).
			WillReturnRows(sqlmock.NewRows([]string{"id", "expires_at"}).
				AddRow(testAccessKeyID, expire))
		sqlm.ExpectCommit()

		tx, err := db.BeginTx(context.Background(), nil)
		require.NoError(t, err)

		accessKey, err := NewStorage(db).
			Redeem(dbtx.WithTx(context.Background(), tx), testAccessKeyCod, testEmprendedor)

		require.NoError(t, err)
		require.NotNil(t, accessKey)
		require.NoError(t, tx.Commit())
	})
}

func TestAccessKeyStore_Revoke(t *testing.T) {
	t.Run("marca la llave como revocada con su autor", func(t *testing.T) {
		store, mock := newMockStore(t)

		mock.ExpectExec(regexRevoke).
			WithArgs(testMunicipalID, testAccessKeyCod).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := store.Revoke(context.Background(), RevokeCodeData{
			RevokeBy: testMunicipalID,
			code:     testAccessKeyCod,
		})

		require.NoError(t, err)
	})

	t.Run("devuelve util.ErrNotFound si no se revocó ninguna fila", func(t *testing.T) {
		store, mock := newMockStore(t)

		mock.ExpectExec(regexRevoke).
			WithArgs(testMunicipalID, "no-existe").
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := store.Revoke(context.Background(), RevokeCodeData{
			RevokeBy: testMunicipalID,
			code:     "no-existe",
		})

		require.ErrorIs(t, err, util.ErrNotFound)
	})

	t.Run("revocar dos veces la misma llave no falla, la segunda pisa la auditoría", func(t *testing.T) {
		// El WHERE es solo por code, sin filtro de revoked_at IS NULL: el
		// segundo UPDATE matchea la misma fila y sobrescribe revoked_at y
		// revoked_by en lugar de reportar que ya estaba revocada.
		store, mock := newMockStore(t)

		mock.ExpectExec(regexRevoke).
			WithArgs(testMunicipalID, testAccessKeyCod).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectExec(regexRevoke).
			WithArgs(testEmprendedor, testAccessKeyCod).
			WillReturnResult(sqlmock.NewResult(0, 1))

		require.NoError(t, store.Revoke(context.Background(), RevokeCodeData{
			RevokeBy: testMunicipalID,
			code:     testAccessKeyCod,
		}))

		require.NoError(t, store.Revoke(context.Background(), RevokeCodeData{
			RevokeBy: testEmprendedor,
			code:     testAccessKeyCod,
		}))
	})

	t.Run("propaga el error de la base sin convertirlo", func(t *testing.T) {
		store, mock := newMockStore(t)

		mock.ExpectExec(regexRevoke).
			WithArgs(testMunicipalID, testAccessKeyCod).
			WillReturnError(errDBBoom)

		err := store.Revoke(context.Background(), RevokeCodeData{
			RevokeBy: testMunicipalID,
			code:     testAccessKeyCod,
		})

		require.ErrorIs(t, err, errDBBoom)
		assert.NotErrorIs(t, err, util.ErrNotFound)
	})

	t.Run("propaga el error de RowsAffected", func(t *testing.T) {
		store, mock := newMockStore(t)

		mock.ExpectExec(regexRevoke).
			WithArgs(testMunicipalID, testAccessKeyCod).
			WillReturnResult(sqlmock.NewErrorResult(errDBBoom))

		err := store.Revoke(context.Background(), RevokeCodeData{
			RevokeBy: testMunicipalID,
			code:     testAccessKeyCod,
		})

		require.ErrorIs(t, err, errDBBoom)
	})
}

func TestAccessKeyStore_ContextCancelado(t *testing.T) {
	t.Run("no ejecuta nada si el contexto ya está cancelado", func(t *testing.T) {
		db, _, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		store := NewStorage(db)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		accessKey := newTestAccessKey()

		require.ErrorIs(t, store.Create(ctx, &accessKey), context.Canceled)

		_, err = store.GetByID(ctx, testAccessKeyID)
		require.ErrorIs(t, err, context.Canceled)

		_, err = store.Redeem(ctx, testAccessKeyCod, testEmprendedor)
		require.ErrorIs(t, err, context.Canceled)

		require.ErrorIs(t, store.Revoke(ctx, RevokeCodeData{
			RevokeBy: testMunicipalID,
			code:     testAccessKeyCod,
		}), context.Canceled)
	})
}

// newCapturingMockStore levanta el store con un matcher permisivo que guarda el
// SQL realmente ejecutado. sqlmock no parsea SQL (solo compara strings), así que
// sin esto una coma de menos o una columna cambiada de lugar pasarían los tests
// en verde y reventarían recién contra Postgres.
func newCapturingMockStore(t *testing.T) (*StoreAccessKey, sqlmock.Sqlmock, *string) {
	t.Helper()

	var capturada string

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(
		sqlmock.QueryMatcherFunc(func(_, actualSQL string) error {
			capturada = actualSQL
			return nil
		}),
	))
	require.NoError(t, err)

	t.Cleanup(func() {
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("expectativas de sql sin cumplir: %v", err)
		}
		mock.ExpectClose()
		_ = db.Close()
	})

	return NewStorage(db), mock, &capturada
}

// columnasDe extrae la lista de proyección de un SELECT y la parte por coma.
// Sirve para comparar la query contra el orden en que el Scan escanea.
func columnasDe(query string) []string {
	query = strings.TrimSpace(query)
	query = query[strings.Index(query, "SELECT")+len("SELECT"):]
	query = query[:strings.Index(query, "FROM")]

	proyeccion := strings.Split(query, ",")
	for i, col := range proyeccion {
		proyeccion[i] = strings.TrimSpace(col)
	}

	return proyeccion
}

func TestAccessKeyStore_Create_query(t *testing.T) {
	t.Run("el RETURNING está en el orden que lee el Scan", func(t *testing.T) {
		// RETURNING id, code con Scan(&Code, &ID) escribiría el UUID en el
		// string del código: una llave con el código corrupto que después no
		// se puede canjear ni revocar.
		store, mock, sql := newCapturingMockStore(t)

		accessKey := newTestAccessKey()

		mock.ExpectQuery("").
			WithArgs(testAccessKeyCod, testMunicipalID, *accessKey.ExpiresAt).
			WillReturnRows(sqlmock.NewRows([]string{"code", "id"}).
				AddRow(testAccessKeyCod, testAccessKeyID))

		require.NoError(t, store.Create(context.Background(), &accessKey))

		assert.Equal(t, testAccessKeyCod, accessKey.Code)
		assert.Equal(t, testAccessKeyID, accessKey.ID)
		assert.Contains(t, *sql, "RETURNING code, id")
	})

	t.Run("escribe en access_keys con las tres columnas de la llave", func(t *testing.T) {
		store, mock, sql := newCapturingMockStore(t)

		accessKey := newTestAccessKey()

		mock.ExpectQuery("").
			WillReturnRows(sqlmock.NewRows([]string{"code", "id"}).
				AddRow(testAccessKeyCod, testAccessKeyID))

		require.NoError(t, store.Create(context.Background(), &accessKey))

		assert.Contains(t, *sql, "INSERT INTO access_keys (code, created_by, expires_at)")
	})
}

func TestAccessKeyStore_GetByID_query(t *testing.T) {
	t.Run("la proyección está en el mismo orden que el Scan", func(t *testing.T) {
		store, mock, sql := newCapturingMockStore(t)

		mock.ExpectQuery("").
			WillReturnRows(sqlmock.NewRows(accessKeyColumns()).
				AddRow(accessKeyRow(time.Now().AddDate(1, 0, 0))...))

		_, err := store.GetByID(context.Background(), testAccessKeyID)

		require.NoError(t, err)
		assert.Equal(t, accessKeyColumns(), columnasDe(*sql))
	})

	t.Run("created_at es una columna propia, no un alias de revoked_by", func(t *testing.T) {
		// "revoked_by created_at" es SQL válido en Postgres (alias implícito),
		// por eso el typo no se ve al leer la query: devuelve revoked_by dos
		// veces, created_at queda NULL y el Scan lee la columna equivocada.
		store, mock, sql := newCapturingMockStore(t)

		mock.ExpectQuery("").
			WillReturnRows(sqlmock.NewRows(accessKeyColumns()).
				AddRow(accessKeyRow(time.Now().AddDate(1, 0, 0))...))

		_, err := store.GetByID(context.Background(), testAccessKeyID)

		require.NoError(t, err)
		assert.NotContains(t, *sql, "revoked_by created_at")
		assert.Contains(t, *sql, "revoked_by, created_at")
	})
}

func TestAccessKeyStore_Redeem_query(t *testing.T) {
	t.Run("solo canjea llaves sin canjear, sin vencer y no revocadas", func(t *testing.T) {
		// Si falta cualquiera de los tres filtros se puede canjear dos veces la
		// misma llave, o extender el plan con una llave ya vencida o revocada.
		store, mock, sql := newCapturingMockStore(t)

		mock.ExpectQuery("").
			WillReturnRows(sqlmock.NewRows([]string{"id", "expires_at"}).
				AddRow(testAccessKeyID, time.Now().AddDate(1, 0, 0)))

		_, err := store.Redeem(context.Background(), testAccessKeyCod, testEmprendedor)

		require.NoError(t, err)
		assert.Contains(t, *sql, "redeemed_by IS NULL")
		assert.Contains(t, *sql, "expires_at > now()")
		assert.Contains(t, *sql, "revoked_at IS NULL")
	})

	t.Run("el UPDATE es atómico: canjea con el filtro, no con un SELECT previo", func(t *testing.T) {
		store, mock, sql := newCapturingMockStore(t)

		mock.ExpectQuery("").
			WillReturnRows(sqlmock.NewRows([]string{"id", "expires_at"}).
				AddRow(testAccessKeyID, time.Now().AddDate(1, 0, 0)))

		_, err := store.Redeem(context.Background(), testAccessKeyCod, testEmprendedor)

		require.NoError(t, err)
		assert.Contains(t, *sql, "SET redeemed_by = $1, redeemed_at = now()")
		assert.Contains(t, *sql, "RETURNING id, expires_at")
		assert.NotContains(t, *sql, "SELECT", "un SELECT previo abre la carrera de doble canje")
	})

	t.Run("canjea por código, no por id", func(t *testing.T) {
		store, mock, sql := newCapturingMockStore(t)

		mock.ExpectQuery("").
			WillReturnRows(sqlmock.NewRows([]string{"id", "expires_at"}).
				AddRow(testAccessKeyID, time.Now().AddDate(1, 0, 0)))

		_, err := store.Redeem(context.Background(), testAccessKeyCod, testEmprendedor)

		require.NoError(t, err)
		assert.Contains(t, *sql, "WHERE code = $2")
	})
}

func TestAccessKeyStore_Revoke_query(t *testing.T) {
	t.Run("guarda quién revocó y cuándo", func(t *testing.T) {
		store, mock, sql := newCapturingMockStore(t)

		mock.ExpectExec("").
			WillReturnResult(sqlmock.NewResult(0, 1))

		require.NoError(t, store.Revoke(context.Background(), RevokeCodeData{
			RevokeBy: testMunicipalID,
			code:     testAccessKeyCod,
		}))

		assert.Contains(t, *sql, "SET revoked_by = $1, revoked_at = now()")
		assert.Contains(t, *sql, "WHERE code = $2")
	})

	t.Run("no toca redeemed_at: revocar no es canjear", func(t *testing.T) {
		store, mock, sql := newCapturingMockStore(t)

		mock.ExpectExec("").
			WillReturnResult(sqlmock.NewResult(0, 1))

		require.NoError(t, store.Revoke(context.Background(), RevokeCodeData{
			RevokeBy: testMunicipalID,
			code:     testAccessKeyCod,
		}))

		assert.NotContains(t, *sql, "redeemed_at")
		assert.NotContains(t, *sql, "redeemed_by =")
	})
}
