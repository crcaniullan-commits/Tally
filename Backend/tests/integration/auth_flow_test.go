package integration

import (
	"database/sql"
	"net/http"
	"testing"
	"time"

	"github.com/crcaniullan-commits/Tally/tests/integration/testutil"
	"github.com/stretchr/testify/require"
)

func TestAuthFlow(t *testing.T) {
	municipalEmail := testutil.UniqueEmail("auth_municipal")
	municipalRUT := testutil.NewRUT()
	const municipalPassword = "Municipal123"

	t.Run("registra un usuario municipal", func(t *testing.T) {
		res := testutil.Request(t, http.MethodPost, "/v1/autenticacion/municipal", "", map[string]string{
			"email":    municipalEmail,
			"password": municipalPassword,
			"nombre":   "Municipal Auth",
			"rut":      municipalRUT,
		})

		testutil.RequireStatus(t, res, http.StatusCreated)
		require.NotEmpty(t, testutil.DecodeData[string](t, res))
	})

	t.Run("el municipal puede iniciar sesión", func(t *testing.T) {
		res := testutil.Request(t, http.MethodPost, "/v1/autenticacion/inicio-sesion", "", map[string]string{
			"email":    municipalEmail,
			"password": municipalPassword,
		})

		testutil.RequireStatus(t, res, http.StatusAccepted)
		require.NotEmpty(t, testutil.DecodeData[string](t, res))
	})

	email := testutil.UniqueEmail("auth_user")
	rut := testutil.NewRUT()
	const password = "Password123"
	key := testutil.IssueKey(t, email)

	var token string

	t.Run("registra un usuario canjeando una llave", func(t *testing.T) {
		res := testutil.Request(t, http.MethodPost, "/v1/autenticacion", "", map[string]string{
			"email":    email,
			"password": password,
			"nombre":   "Usuario Auth",
			"rut":      rut,
			"code":     key.Code,
		})

		testutil.RequireStatus(t, res, http.StatusCreated)
		token = testutil.DecodeData[string](t, res)
		require.NotEmpty(t, token)
	})

	t.Run("el registro setea plan_expires_at con el vencimiento de la llave", func(t *testing.T) {
		var planExpires sql.NullTime
		err := testutil.GetEnv().DB.QueryRow(`SELECT plan_expires_at FROM users WHERE email = $1`, email).Scan(&planExpires)
		require.NoError(t, err)
		require.True(t, planExpires.Valid, "plan_expires_at debería quedar seteado")
		require.NotNil(t, key.ExpiresAt)
		require.WithinDuration(t, *key.ExpiresAt, planExpires.Time, time.Second)
	})

	t.Run("el usuario registrado puede iniciar sesión", func(t *testing.T) {
		res := testutil.Request(t, http.MethodPost, "/v1/autenticacion/inicio-sesion", "", map[string]string{
			"email":    email,
			"password": password,
		})

		testutil.RequireStatus(t, res, http.StatusAccepted)
		require.NotEmpty(t, testutil.DecodeData[string](t, res))
	})

	t.Run("la misma llave no se puede canjear dos veces", func(t *testing.T) {
		res := testutil.Request(t, http.MethodPost, "/v1/autenticacion", "", map[string]string{
			"email":    testutil.UniqueEmail("auth_second"),
			"password": password,
			"nombre":   "Segundo Usuario",
			"rut":      testutil.NewRUT(),
			"code":     key.Code,
		})

		testutil.RequireError(t, res, http.StatusBadRequest, "El código no se puede canjear")
	})

	t.Run("una llave inexistente no se puede canjear", func(t *testing.T) {
		res := testutil.Request(t, http.MethodPost, "/v1/autenticacion", "", map[string]string{
			"email":    testutil.UniqueEmail("auth_badkey"),
			"password": password,
			"nombre":   "Sin Llave",
			"rut":      testutil.NewRUT(),
			"code":     "NO-EXISTE",
		})

		testutil.RequireError(t, res, http.StatusBadRequest, "El código no se puede canjear")
	})

	t.Run("rechaza un email ya registrado", func(t *testing.T) {
		otherKey := testutil.IssueKey(t, testutil.UniqueEmail("auth_dup_email"))
		res := testutil.Request(t, http.MethodPost, "/v1/autenticacion", "", map[string]string{
			"email":    email,
			"password": password,
			"nombre":   "Duplicado",
			"rut":      testutil.NewRUT(),
			"code":     otherKey.Code,
		})

		testutil.RequireError(t, res, http.StatusBadRequest, "ya existe un usuario con ese correo")
	})

	t.Run("rechaza un RUT ya registrado", func(t *testing.T) {
		otherKey := testutil.IssueKey(t, testutil.UniqueEmail("auth_dup_rut"))
		res := testutil.Request(t, http.MethodPost, "/v1/autenticacion", "", map[string]string{
			"email":    testutil.UniqueEmail("auth_dup_rut_user"),
			"password": password,
			"nombre":   "Duplicado",
			"rut":      rut,
			"code":     otherKey.Code,
		})

		testutil.RequireError(t, res, http.StatusBadRequest, "ya existe un usuario con ese RUT")
	})

	t.Run("rechaza un RUT con formato inválido", func(t *testing.T) {
		res := testutil.Request(t, http.MethodPost, "/v1/autenticacion", "", map[string]string{
			"email":    testutil.UniqueEmail("auth_bad_rut"),
			"password": password,
			"nombre":   "RUT Inválido",
			"rut":      "abc",
			"code":     "cualquiera",
		})

		testutil.RequireError(t, res, http.StatusBadRequest, "el formato del RUT no es válido")
	})

	t.Run("rechaza un RUT con formato inválido en el municipal", func(t *testing.T) {
		// "abcdefghi" pasa la validación de largo (min=9) pero no el formato del
		// RUT: el regex espera dígitos.
		res := testutil.Request(t, http.MethodPost, "/v1/autenticacion/municipal", "", map[string]string{
			"email":    testutil.UniqueEmail("municipal_bad_rut"),
			"password": municipalPassword,
			"nombre":   "Municipal RUT Inválido",
			"rut":      "abcdefghi",
		})

		testutil.RequireError(t, res, http.StatusBadRequest, "el formato del RUT no es válido")
	})

	t.Run("login rechaza una contraseña incorrecta", func(t *testing.T) {
		res := testutil.Request(t, http.MethodPost, "/v1/autenticacion/inicio-sesion", "", map[string]string{
			"email":    email,
			"password": "WrongPassword123",
		})

		testutil.RequireError(t, res, http.StatusBadRequest, "crypto/bcrypt: hashedPassword is not the hash of the given password")
	})

	t.Run("login rechaza un email inexistente", func(t *testing.T) {
		res := testutil.Request(t, http.MethodPost, "/v1/autenticacion/inicio-sesion", "", map[string]string{
			"email":    testutil.UniqueEmail("auth_nobody"),
			"password": password,
		})

		testutil.RequireError(t, res, http.StatusBadRequest, "recurso no encontrado")
	})

	t.Run("login valida el payload", func(t *testing.T) {
		cases := []struct {
			name    string
			body    any
			message string
		}{
			{
				name:    "email inválido",
				body:    map[string]string{"email": "no-es-email", "password": password},
				message: "Key: 'PayloadInicioSesion.Email' Error:Field validation for 'Email' failed on the 'email' tag",
			},
			{
				name:    "contraseña corta",
				body:    map[string]string{"email": email, "password": "short"},
				message: "Key: 'PayloadInicioSesion.Password' Error:Field validation for 'Password' failed on the 'min' tag",
			},
			{
				name:    "campos faltantes",
				body:    map[string]any{},
				message: "Key: 'PayloadInicioSesion.Email' Error:Field validation for 'Email' failed on the 'required' tag\nKey: 'PayloadInicioSesion.Password' Error:Field validation for 'Password' failed on the 'required' tag",
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				res := testutil.Request(t, http.MethodPost, "/v1/autenticacion/inicio-sesion", "", tc.body)
				testutil.RequireError(t, res, http.StatusBadRequest, tc.message)
			})
		}
	})

	t.Run("registro valida el payload", func(t *testing.T) {
		cases := []struct {
			name    string
			body    any
			message string
		}{
			{
				name:    "email inválido",
				body:    map[string]string{"email": "no-es-email", "password": password, "nombre": "X", "rut": testutil.NewRUT(), "code": "x"},
				message: "Key: 'PayloadCrearUsuario.Email' Error:Field validation for 'Email' failed on the 'email' tag",
			},
			{
				name:    "contraseña corta",
				body:    map[string]string{"email": "x@example.com", "password": "short", "nombre": "X", "rut": testutil.NewRUT(), "code": "x"},
				message: "Key: 'PayloadCrearUsuario.Password' Error:Field validation for 'Password' failed on the 'min' tag",
			},
			{
				name:    "contraseña larga",
				body:    map[string]string{"email": "x@example.com", "password": string(make([]byte, 73)), "nombre": "X", "rut": testutil.NewRUT(), "code": "x"},
				message: "Key: 'PayloadCrearUsuario.Password' Error:Field validation for 'Password' failed on the 'max' tag",
			},
			{
				name:    "campos faltantes",
				body:    map[string]any{},
				message: "Key: 'PayloadCrearUsuario.Email' Error:Field validation for 'Email' failed on the 'required' tag\nKey: 'PayloadCrearUsuario.Password' Error:Field validation for 'Password' failed on the 'required' tag\nKey: 'PayloadCrearUsuario.Nombre' Error:Field validation for 'Nombre' failed on the 'required' tag\nKey: 'PayloadCrearUsuario.Rut' Error:Field validation for 'Rut' failed on the 'required' tag\nKey: 'PayloadCrearUsuario.Code' Error:Field validation for 'Code' failed on the 'required' tag",
			},
			{
				name:    "JSON malformado",
				body:    `{"email":`,
				message: "unexpected EOF",
			},
			{
				name:    "tipo incorrecto",
				body:    `{"email":1,"password":true}`,
				message: "json: cannot unmarshal number into Go struct field PayloadCrearUsuario.email of type string",
			},
			{
				name:    "campo desconocido",
				body:    `{"email":"x@example.com","password":"Password123","nombre":"X","rut":"1234567-8","code":"x","extra":1}`,
				message: `json: unknown field "extra"`,
			},
			{
				name:    "nombre demasiado largo",
				body:    map[string]string{"email": "x@example.com", "password": password, "nombre": string(make([]byte, 101)), "rut": testutil.NewRUT(), "code": "x"},
				message: "Key: 'PayloadCrearUsuario.Nombre' Error:Field validation for 'Nombre' failed on the 'max' tag",
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				res := testutil.Request(t, http.MethodPost, "/v1/autenticacion", "", tc.body)
				testutil.RequireError(t, res, http.StatusBadRequest, tc.message)
			})
		}
	})

	t.Run("municipal valida el payload", func(t *testing.T) {
		cases := []struct {
			name    string
			body    any
			message string
		}{
			{
				name:    "RUT corto",
				body:    map[string]string{"email": testutil.UniqueEmail("muni_short"), "password": municipalPassword, "nombre": "X", "rut": "1234-5"},
				message: "Key: 'PayloadCrearMunicipal.Rut' Error:Field validation for 'Rut' failed on the 'min' tag",
			},
			{
				name:    "RUT largo",
				body:    map[string]string{"email": testutil.UniqueEmail("muni_long"), "password": municipalPassword, "nombre": "X", "rut": "1234567890123-4"},
				message: "Key: 'PayloadCrearMunicipal.Rut' Error:Field validation for 'Rut' failed on the 'max' tag",
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				res := testutil.Request(t, http.MethodPost, "/v1/autenticacion/municipal", "", tc.body)
				testutil.RequireError(t, res, http.StatusBadRequest, tc.message)
			})
		}
	})

	t.Run("rechaza un email municipal duplicado", func(t *testing.T) {
		res := testutil.Request(t, http.MethodPost, "/v1/autenticacion/municipal", "", map[string]string{
			"email":    municipalEmail,
			"password": municipalPassword,
			"nombre":   "Municipal Duplicado",
			"rut":      testutil.NewRUT(),
		})

		testutil.RequireError(t, res, http.StatusBadRequest, "ya existe un usuario con ese correo")
	})

	t.Run("el token permite acceder a una ruta protegida", func(t *testing.T) {
		res := testutil.Request(t, http.MethodGet, "/v1/aplicacion/empresas", token, nil)
		testutil.RequireStatus(t, res, http.StatusOK)
	})

	requireProtected(t, http.MethodGet, "/v1/aplicacion/empresas")
	requireProtected(t, http.MethodPatch, "/v1/aplicacion/usuarios")
}
