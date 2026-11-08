# Reporte de tests de integración

Fecha: 2026-11-07

Suite HTTP de integración del backend de Tally. Ejercita el router real en memoria
contra una base PostgreSQL efímera, sin mocks de base de datos.

## Cómo ejecutar

```bash
cd Backend

# Suite completa
GOPROXY=off go test -v ./tests/integration/...

# Un flujo puntual
GOPROXY=off go test -v ./tests/integration/ -run TestIncomesFlow
```

Requisitos:

- PostgreSQL accesible. Por defecto se usa
  `postgres://postgres:tally_database_env@localhost/tally_database_env?sslmode=disable`,
  o la URL de `DB_ADDR`.
- `TEST_DB_ADDR` es opcional: si está definido, la suite reutiliza esa base sin
  crearla ni dropearla (solo asegura el esquema).
- Sin `TEST_DB_ADDR`, `TestMain` crea `tally_it_<pid>`, aplica
  `migrations/*.up.sql` en orden, y la elimina al terminar. No toca la base de
  desarrollo.

> **CI:** el workflow levanta un service container `postgres:16` y exporta
> `DB_ADDR`; la suite crea su base efímera y aplica las migraciones sola. No hay
> un step de migraciones separado.

## Resultado actual

Última corrida completa: **OK**, 10 flujos y 189 subtests directos (~3 s).
`gofmt -l tests/integration/` y `go vet ./tests/integration/...` sin hallazgos.

| Flujo | Archivo | Resultado | Subtests |
|---|---|---|---|
| Auth | `integration/auth_flow_test.go` | PASS | 21 |
| Access keys | `integration/accesskeys_flow_test.go` | PASS | 19 |
| Empresas | `integration/empresas_flow_test.go` | PASS | 18 |
| Categorías | `integration/categories_flow_test.go` | PASS | 20 |
| Ingresos | `integration/incomes_flow_test.go` | PASS | 21 |
| Gastos | `integration/expenses_flow_test.go` | PASS | 20 |
| Deudores | `integration/debtors_flow_test.go` | PASS | 19 |
| Metas | `integration/goals_flow_test.go` | PASS | 23 |
| Usuarios | `integration/users_flow_test.go` | PASS | 22 |
| Aislamiento multi-empresa | `integration/tenant_isolation_test.go` | PASS | 6 |

Todos los flujos verifican al menos una ruta protegida sin token y con token
inválido (`401 unauthorized`).

## Estructura

```text
tests/
├── report.md
└── integration/
    ├── main_test.go              # TestMain -> testutil.SetupSuite
    ├── helpers_test.go           # requireProtected, dbDate
    ├── <flujo>_flow_test.go      # un archivo por flujo
    ├── tenant_isolation_test.go
    └── testutil/
        ├── env.go                # base efímera, migraciones, router real
        ├── client.go             # Request, DecodeData, DecodeError, Require*
        ├── fixtures.go           # NewUser, CreateEmpresa, CreateCategory, ...
        └── rut.go                # generador de RUT módulo 11
```

## Cobertura por flujo

- **Auth:** registro municipal, registro canjeando llave, `plan_expires_at`,
  login, llave reutilizada/inexistente, email/RUT duplicado, RUT inválido,
  validaciones de payload y acceso a ruta protegida.
- **Access keys:** emisión, reenvío por ID, revocación por código, canje de llave
  revocada, permisos de rol y validaciones.
- **Empresas:** creación, listado, renombrado, borrado, doble borrado, listado
  vacío, validaciones, rol municipal y aislamiento entre usuarios.
- **Categorías:** creación, listado propio + globales, renombrado, duplicados
  ignorando mayúsculas/acentos, mismo nombre en otra empresa, categorías
  globales no modificables desde la empresa, borrado y validaciones.
- **Ingresos:** creación con/sin categoría, categoría global, listado, filtros
  por categoría y fechas, paginación, borrado, categoría inexistente/ajena y
  validaciones de payload/query.
- **Gastos:** creación con categoría y descripción, categoría global, listado,
  filtros, paginación, borrado, categoría inexistente/ajena y validaciones.
- **Deudores:** creación, listado, actualización, marcado como pagado, campos
  desconocidos, validaciones y errores de ruta.
- **Metas:** creación por período, listado, detalle, cálculo de estado con
  ingresos del día/semana/mes, pendiente no negativo, borrado y validaciones.
- **Usuarios:** actualización de nombre+contraseña, login con nueva contraseña,
  actualización solo de contraseña, consulta municipal por RUT, validaciones,
  borrado de cuenta e invalidación del token.
- **Aislamiento multi-empresa:** acceso a empresa ajena (`404` en middleware),
  recursos ajenos por ID desde la empresa propia (`404` en el `WHERE`), listados
  que excluyen datos ajenos, categoría ajena rechazada y rol municipal sin
  acceso a empresas.

## Bugs del backend documentados

Los tests pasan assertando el comportamiento actual. No se corrigió código de
producción. Cada caso está marcado con `BUG:` o `DISCREPANCIA:` en el test.

| # | Ruta | Esperado | Actual | Causa |
|---|---|---|---|---|
| 1 | `POST /v1/auth` y `POST /v1/auth/municipal` con email/RUT duplicado | `400` con `a user with that email already exists` / `a user with that rut already exists` | `500 {"error":"the server encountered a problem"}` | `internal/auth/storage.go` compara `err.Error()` contra el texto exacto de pq, pero lib/pq agrega ` (23505)`. |
| 2 | `POST /v1/auth` y `POST /v1/auth/municipal` con RUT mal formado | `400` con `el formato del RUT no es válido` | `500 {"error":"the server encountered a problem"}` | El handler de auth no mapea `model.ErrFormatoInvalido`. |
| 3 | `PATCH /v1/app/users` con solo `nombre` o payload vacío | `200` actualizando el nombre / no-op | `500 {"error":"the server encountered a problem"}` | `AuthTokenMiddleware` usa `GetByID`, que no selecciona `password_hash`; `Update` escribe `NULL` en una columna `NOT NULL`. |
| 4 | `GET /v1/app/users/municipal/{rut}` | `created_at` y `updated_at` reales | `0001-01-01T00:00:00Z` | `GetByRut` no selecciona esas columnas. |

## Diferencias entre swagger y código

| # | Punto | Swagger | Código |
|---|---|---|---|
| 1 | `POST /app/empresas` | `200` | `201` |
| 2 | `GET /app/empresas` sin empresas | `404` | `200 {"data":null}` |
| 3 | `PATCH /app/empresas/{empresaID}/{nombre}` | Documenta `400`, sin límite explícito | Acepta nombres de 31+ caracteres; la creación valida `max=30` |
| 4 | `POST /app/users/exchange/{code}` | Documentado con `202` | No montado (`404`); el canje real es `POST /v1/auth` |
| 5 | `model.RUT` | `{"cuerpo":..., "dv":"..."}` | Serializa `{"Cuerpo":..., "DV":"..."}` por falta de tags JSON |
| 6 | `GET` de listados (`categories`, `goal`, `debtors`, etc.) | Solo documentan `200/500` | El middleware puede devolver `400` (UUID de empresa inválido), `401` y `403` |

## Casos omitidos

- Expiración/refresh de JWT: no hay endpoint de refresh y esperar la expiración
  haría el test lento; solo se prueba token inválido.
- Envío real de correo SMTP: los tests usan un `SendEmail` no-op.
- `excel` y `feed`: son esqueletos sin rutas ni lógica.
- Carrera de unicidad de categorías: deuda técnica conocida y no determinista.
- Borrado en cascada de empresas/ingresos/gastos al borrar un usuario: solo se
  verifica que el token deja de servir (`401`).
- UI de Swagger (`/v1/swagger/*`).
- Métodos HTTP no soportados (`405`) y payloads gigantes.
- Enforcement de `plan_expires_at`: el código no lo valida.
- Rate limiting: el helper existe pero no está montado.
- Revocación de llaves entre municipales distintos: solo hay un municipal en la
  suite.

## Mantenimiento

- Al agregar un flujo, crear `integration/<flujo>_flow_test.go`, correrlo solo y
  después la suite completa.
- Si se descubre un bug, no arreglarlo en el mismo cambio: documentarlo con
  `BUG:`/`DISCREPANCIA:` y actualizar este reporte.
- Si cambia el contrato (ruta, campo, status o mensaje), actualizar el test y
  este reporte.
