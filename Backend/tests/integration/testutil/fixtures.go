package testutil

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/crcaniullan-commits/Tally/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type User struct {
	ID       uuid.UUID
	Token    string
	Email    string
	Password string
	RUT      string
}

var (
	municipalToken string
	municipalRUT   string
	seq            atomic.Uint64
)

// SetupSuite levanta una base de test aislada, monta el router real y registra
// una sola vez el usuario municipal que usan los fixtures.
func SetupSuite(m *testing.M) int {
	env, cleanup, err := newEnv()
	if err != nil {
		fmt.Fprintf(os.Stderr, "testutil: %v\n", err)
		return 1
	}

	current = env
	defer cleanup()

	token, rut, err := registerMunicipal()
	if err != nil {
		fmt.Fprintf(os.Stderr, "testutil: %v\n", err)
		return 1
	}

	municipalToken = token
	municipalRUT = rut

	return m.Run()
}

func registerMunicipal() (string, string, error) {
	email := UniqueEmail("municipal")
	rut := NewRUT()

	res, err := doRequest(http.MethodPost, "/v1/auth/municipal", "", map[string]string{
		"email":    email,
		"password": "Municipal123",
		"nombre":   "Municipal Integration",
		"rut":      rut,
	})
	if err != nil {
		return "", "", err
	}

	if res.Status != http.StatusCreated {
		return "", "", fmt.Errorf("registro municipal: status %d, body %s", res.Status, res.Body)
	}

	var envelope struct {
		Data string `json:"data"`
	}
	if err := json.Unmarshal(res.Body, &envelope); err != nil {
		return "", "", fmt.Errorf("decodificando token municipal: %w", err)
	}

	return envelope.Data, rut, nil
}

func MunicipalToken(t *testing.T) string {
	t.Helper()
	require.NotEmpty(t, municipalToken, "el municipal de TestMain no fue registrado")
	return municipalToken
}

func MunicipalRUT(t *testing.T) string {
	t.Helper()
	require.NotEmpty(t, municipalRUT, "el municipal de TestMain no fue registrado")
	return municipalRUT
}

func UniqueEmail(prefix string) string {
	return fmt.Sprintf("%s_%d_%d@example.com", prefix, time.Now().UnixNano(), seq.Add(1))
}

func UniqueName(prefix string) string {
	return fmt.Sprintf("%s %d", prefix, seq.Add(1))
}

// NewUser emite una llave con el municipal, registra un usuario con ese código
// y devuelve sus credenciales. El registro consume la llave dentro de la
// transacción; si falla, la llave no queda canjeada.
func NewUser(t *testing.T) *User {
	t.Helper()

	email := UniqueEmail("user")
	password := "Password123"
	key := IssueKey(t, email)

	var last Result

	for attempt := 0; attempt < 5; attempt++ {
		rut := NewRUT()

		res := Request(t, http.MethodPost, "/v1/auth", "", map[string]string{
			"email":    email,
			"password": password,
			"nombre":   "Usuario Integration",
			"rut":      rut,
			"code":     key.Code,
		})

		if res.Status == http.StatusCreated {
			user := &User{
				ID:       userIDByEmail(t, email),
				Token:    DecodeData[string](t, res),
				Email:    email,
				Password: password,
				RUT:      rut,
			}

			t.Cleanup(func() {
				_, _ = GetEnv().DB.Exec(`DELETE FROM users WHERE id = $1`, user.ID)
			})

			return user
		}

		last = res

		if res.Status == http.StatusBadRequest && strings.Contains(DecodeError(t, res), "rut already exists") {
			continue
		}

		require.FailNowf(t, "no se pudo registrar el usuario fixture", "status %d, body %s", res.Status, res.Body)
	}

	require.FailNowf(t, "no se pudo registrar el usuario fixture tras varios intentos", "status %d, body %s", last.Status, last.Body)

	return nil
}

func IssueKey(t *testing.T, email string) model.AccessKey {
	t.Helper()

	res := Request(t, http.MethodPost, "/v1/app/key", MunicipalToken(t), map[string]string{
		"email": email,
	})

	RequireStatus(t, res, http.StatusCreated)

	return DecodeData[model.AccessKey](t, res)
}

func userIDByEmail(t *testing.T, email string) uuid.UUID {
	t.Helper()

	var id uuid.UUID

	require.NoError(t, GetEnv().DB.QueryRow(`SELECT id FROM users WHERE email = $1`, email).Scan(&id))

	return id
}

func EmpresaPath(empresaID uuid.UUID, suffix string) string {
	return "/v1/app/empresas/" + empresaID.String() + suffix
}

func CreateEmpresa(t *testing.T, user *User) model.Empresa {
	t.Helper()

	res := Request(t, http.MethodPost, "/v1/app/empresas", user.Token, map[string]string{
		"nombre_empresa": UniqueName("Empresa"),
	})

	RequireStatus(t, res, http.StatusCreated)

	return DecodeData[model.Empresa](t, res)
}

func CreateCategory(t *testing.T, user *User, empresaID uuid.UUID) model.Categorie {
	t.Helper()

	res := Request(t, http.MethodPost, EmpresaPath(empresaID, "/categories"), user.Token, map[string]string{
		"nombre": UniqueName("Cat"),
	})

	RequireStatus(t, res, http.StatusCreated)

	return DecodeData[model.Categorie](t, res)
}

func CreateIncome(t *testing.T, user *User, empresaID uuid.UUID, monto int64, categoryID *uuid.UUID) model.Income {
	t.Helper()

	payload := map[string]any{
		"monto":          monto,
		"payment_method": "efectivo",
	}
	if categoryID != nil {
		payload["category_id"] = categoryID.String()
	}

	res := Request(t, http.MethodPost, EmpresaPath(empresaID, "/incomes"), user.Token, payload)

	RequireStatus(t, res, http.StatusCreated)

	return DecodeData[model.Income](t, res)
}

func CreateExpense(t *testing.T, user *User, empresaID uuid.UUID, monto int64, categoryID uuid.UUID) model.Expense {
	t.Helper()

	res := Request(t, http.MethodPost, EmpresaPath(empresaID, "/expenses"), user.Token, map[string]any{
		"monto":       monto,
		"category_id": categoryID.String(),
	})

	RequireStatus(t, res, http.StatusCreated)

	return DecodeData[model.Expense](t, res)
}

func CreateDebtor(t *testing.T, user *User, empresaID uuid.UUID, monto int64) model.Debtor {
	t.Helper()

	res := Request(t, http.MethodPost, EmpresaPath(empresaID, "/debtors"), user.Token, map[string]any{
		"nombre":       UniqueName("Deudor"),
		"monto":        monto,
		"descripcion":  "deuda de prueba",
		"fecha_limite": time.Now().Add(24 * time.Hour).UTC().Format(time.RFC3339),
	})

	RequireStatus(t, res, http.StatusCreated)

	return DecodeData[model.Debtor](t, res)
}

func CreateGoal(t *testing.T, user *User, empresaID uuid.UUID, periodo string, montoMeta int64) model.Goal {
	t.Helper()

	res := Request(t, http.MethodPost, EmpresaPath(empresaID, "/goal"), user.Token, map[string]any{
		"nombre":     UniqueName("Meta"),
		"periodo":    periodo,
		"monto_meta": montoMeta,
	})

	RequireStatus(t, res, http.StatusCreated)

	return DecodeData[model.Goal](t, res)
}
