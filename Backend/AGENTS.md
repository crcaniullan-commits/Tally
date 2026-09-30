# Tally — AGENTS.md

API financiera (ingresos, gastos, deudores, metas) para emprendedores informales. Go + chi + Postgres.

Este archivo cubre **solo el backend**. El frontend es un proyecto separado en `../Frontend/` con su propio `AGENTS.md`; no mezcles las convenciones de ambos.

## Comandos

Todos desde este directorio (`Backend/`). Las herramientas están declaradas en el bloque `tool` de `go.mod` (Go 1.24+), así que se invocan con `go tool <name>` — no hace falta instalarlas.

```bash
go tool task test        # vet + go test -race -cover ./...
go tool task test TEST_FLAGS='-run TestIncomeService -v'   # test suelto
go tool task test-cover   # coverage.out + desglose por función
go tool task clean-cover
go tool task gen-docs     # regenera docs/
go tool task migration NAME=algo   # crea 00000N_*.{up,down}.sql
go tool task migration-up         # necesita Postgres + DB_ADDR
GOPROXY=off go test ./internal/incomes/...   # cache de módulos ya está caliente
```

`make test`, `make test-cover` y `make clean-cover` funcionan igual. El resto de targets de `make` **no**:

- `make build` es un no-op silencioso. El catch-all `%:` del final del makefile captura cualquier target inexistente y sale con 0 sin hacer nada. No confíes en `exit 0`.
- `go tool task build` falla: hace `go build -o bin/server main.go` pero el entrypoint está en `cmd/`. Compilá con `go build -o bin/main ./cmd`. `task start` hereda el fallo porque depende de `build`.
- `make migration*` y `make gen-docs` invocan binarios `migrate`/`swag` sueltos que no están instalados. Usá el `go tool task` equivalente.

## Migraciones

`golang-migrate`, archivos `NNN_nombre.{up,down}.sql` en `migrations/`. Editar una migración ya aplicada no la reaplica: creá una nueva.

`.envrc` está en `.gitignore` (tiene credenciales reales) pero `makefile` lo incluye; `go tool task` lo lee vía `dotenv`. Sin él, `make test` igual corre.

## Montos de dinero: `int64`, nunca `float64`

CLP es una moneda de exponente 0 (no usa centavos), así que los montos son enteros de pesos: `int64` en Go, `BIGINT` en Postgres. Aplica a `incomes.monto`, `expenses.monto`, `debtors.monto`, `goals.monto_meta`.

La migración `000003_montos_to_bigint` pasó las columnas de `NUMERIC(12,2)` a `BIGINT`, y **no es opcional**: `lib/pq` devuelve los `NUMERIC` como `[]byte` y `database/sql` los convierte con `strconv.ParseInt`, que rechaza `"1500.00"`. Con la columna como `NUMERIC`, escanear a `int64` falla en toda fila, incluso las de monto `0.00`.

Consecuencias al escribir código de montos:
- Sin aritmética flotante ni `math.*` / `FormatFloat` / `%.2f` en el proyecto. No los reintroduzcas.
- En `validate`, `required` solo verifica `!= 0`, así que no rechaza negativos. Los montos llevan `required,gt=0`.
- El cast `::bigint` redondea (half away from zero): `1500.50` pasa a `1501`. El `.down.sql` no es reversible de verdad.
- JSON hacia `int64` rechaza `1500.0` y `1500.5`, no solo los no-enteros. Mandá enteros pelados.

## Tags `validate`: el espacio después de la coma entra en panic

```go
validate:"required, gt=0"                    // PANIC: Undefined validation function ' gt'
validate:"required,oneof=debito credito ..." // ok
validate:"omitempy,max=100"                  // PANIC: Undefined validation function 'omitempy'
```

Un tag mal escrito no devuelve 400: **el `Validate.Struct` entra en panic** y el `Recoverer` de chi lo convierte en un 500 sin cuerpo. Un tag roto tumba el endpoint entero. Sin espacios tras las comas.

## Tests

Solo `internal/incomes/` y `internal/users/` tienen tests. Convención, un archivo por capa dentro del paquete (tests en `package incomes`, no `incomes_test`):

| Archivo | Qué va |
|---|---|
| `mocks_test.go` | `StoreIncomesMock` (capa store) + `ServiceIncomesMock` (capa service) + fixtures `newTestIncome()` / `newTestPayload()` |
| `service_test.go` | lógica de `IncomeService` contra el store mockeado |
| `handler_test.go` | HTTP con `httptest` + `chi.NewRouteContext()` para los URL params; helpers `newTestHandler`, `requestWithUser`, `requestWithIncomeIDParam` |
| `storage_test.go` | `go-sqlmock` sobre el store real |

`testify` (`assert` / `require` / `mock`) y `go-sqlmock` son dependencias directas.

**sqlmock no parsea SQL**: solo compara strings y devuelve las filas que le digas. No detecta un `Scan` desalineado ni una columna inexistente. Por eso `incomes/storage_test.go` usa un `QueryMatcherFunc` permisivo que captura el SQL realmente ejecutado y lo audita a mano; `users/storage_test.go` alcanza con el matcher por defecto.

## Arquitectura

`handler.go` → `service.go` → `storage.go` por módulo, con interfaces inversas (`StoreIncomes`, `ServiceIncomes`) definidas en service/handler. `router.go` expone `InitModule(r, db, logger, middl)`.

Solo `auth`, `users` e `incomes` están cableados en `cmd/application/app.go` (rutas bajo `/v1/app`, middleware JWT). `debtors`, `expenses`, `goals`, `categories` y `accessKeys` tienen código pero **no están montados**: no son accesibles por HTTP todavía.

El userId del request se saca del contexto con `users.GetUserFromContext(r)`, no del body. `middl.CheckOwnership(rol, handler)` valida el **rol**, no la pertenencia: el acotado por usuario va en el `WHERE` del store.

Respuestas siempre envueltas en `{"data": ...}` vía `util.JsonResponse`. `util.ReadJSON` usa `DisallowUnknownFields`, así que un campo desconocido es 500, no 400.

## Swagger

Las anotaciones viven en un `swagger.go` por módulo (funciones vacías, no se llaman). `docs/docs.go`, `docs/swagger.json` y `docs/swagger.yaml` son **generados**: después de tocar anotaciones o tipos, `go tool task gen-docs` y commiteá el resultado. La UI queda en `/v1/swagger/`.

Ojo: `make gen-docs` apunta a `-g cmd/src/main.go`, ruta que no existe. El target de Taskfile (`-g main.go`) es el correcto.

## CI

`../.github/workflows/test.yml` corre solo en `Backend` (vet + `go test -race -covermode=atomic`) y sube a Codecov. El workflow vive en la raíz del repo, no acá, porque GitHub Actions solo lee `.github/` del root. Triggers: push a `main`, PRs a `main` y manual.

## Otros

- `../Frontend/` está vacío salvo `.gitkeep`: no hay consumidor del contrato JSON todavía, así que cambiar un tag no rompe nada visible.
- `../Fase 1/` y `../Fase 2/` son carpetas de evidencias del curso, sin código.
- `gofmt` no está enganchado a ningún hook: corré `gofmt -w` sobre lo que toques.
- `internal/auth/swagger.go` sale sin formatear de `gofmt` y es preexistente, no lo reformatees de paso.
