# Tally — AGENTS.md

API financiera (ingresos, gastos, deudores, metas) para emprendedores informales. Go 1.27 + chi + Postgres 16.

Este archivo cubre **solo el backend**. El frontend es un proyecto separado en `../Frontend/` con su propio `AGENTS.md`; no mezcles las convenciones de ambos.

## Norma estricta: preguntar antes de tocar código

**No edites ningún archivo sin autorización explícita del usuario en la conversación actual.** Antes de la primera modificación de una tarea, decí qué archivo vas a tocar, qué cambiás y por qué, y esperá el visto bueno. Autorización de una tarea anterior no habilita la siguiente.

Alcance:

- Aplica a todo archivo del repo: `.go`, tests, migraciones, `docs/` generado, `AGENTS.md`, Taskfile/makefile, swagger.
- Solo queda exceptuado lo que el usuario pidió explícitamente en el mensaje actual ("arreglalo", "agregá el test", "actualizá AGENTS.md"). Pedir algo no autoriza los cambios que yo crea necesarios alrededor.
- **Límite de módulos:** solo podés modificar los módulos y archivos que el usuario nombre explícitamente en la conversación actual. No toques otros módulos aunque el build o los tests fallen por ellos. Si un cambio pedido rompe otro módulo, reportá el conflicto y esperá instrucciones.
- **No** hace falta permiso para leer, buscar, correr `go tool task test`, `gofmt -l`, `git diff` ni para reportar un bug.
- Si encontrás un bug que no te pidieron arreglar, **reportalo y preguntá**. No lo fixes de pasada, por obvio que parezca.
- Si el arreglo que te pidieron contradice un test existente (o al revés), avisá el conflicto y preguntá cuál de los dos cambia. No elijas por tu cuenta.

## Comandos

Todos desde este directorio (`Backend/`). Las herramientas están declaradas en el bloque `tool` de `go.mod` (`task`, `migrate`, `swag`), así que se invocan con `go tool <name>` — no hace falta instalarlas.

```bash
go tool task test        # go vet ./... + go test -race -cover ./...
go tool task test TEST_FLAGS='-run TestServicioCategorias -v'   # test suelto
go tool task test-cover   # coverage.out + desglose por función
go tool task clean-cover
go tool task gen-docs     # regenera docs/
go tool task migration NAME=algo          # crea 00000N_*.{up,down}.sql
go tool task migration-up                 # necesita Postgres + DB_ADDR
go tool task migration-down STEPS=1
GOPROXY=off go test ./internal/categories/...   # cache de módulos ya está caliente
```

`make test`, `make test-cover` y `make clean-cover` funcionan igual (el makefile corre `go vet` + `go test`). El resto de targets de `make` **no**:

- `make build` es un no-op silencioso: no existe el target y el catch-all `%:` del final del makefile sale con 0. No confíes en `exit 0`.
- `go tool task build` falla: hace `go build -o bin/server main.go` pero el entrypoint está en `cmd/main.go`. Compilá con `go build -o bin/main ./cmd`. `task start` hereda el fallo porque depende de `build`.
- `make migration*` y `make gen-docs` invocan binarios `migrate`/`swag` sueltos que no están instalados. Usá el `go tool task` equivalente.

`Taskfile.yml` carga `.envrc` y `.env` vía `dotenv` (ambos opcionales para correr tests). El makefile incluye `.envrc` directamente.

## Migraciones

`golang-migrate`, archivos `NNN_nombre.{up,down}.sql` en `migrations/`. Editar una migración ya aplicada no la reaplica: creá una nueva. Última migración del módulo `auth`: `000010_add_password_reset_codes` (`000009_exel_files` es del módulo `excel`, en desarrollo).

`.envrc` está en `.gitignore` (tiene credenciales reales). Sin él, los tests igual corren; `migration-up` necesita `DB_ADDR`.

## Esquema de la base de datos

Estado tras `000010`. El número entre paréntesis es la migración que introdujo el elemento.

ENUMs (`000001`): `user_role` (`usuario`, `municipal`), `payment_method` (`debito`, `credito`, `transferencia`, `efectivo`), `goal_period` (`dia`, `semana`, `mes`).

- `users` — `email VARCHAR(255) UNIQUE NOT NULL`, `password_hash`, `nombre`, `role` (`user_role`, default `usuario`), `created_at`, `updated_at`, `rut` UNIQUE (`000002`), `plan_expires_at` nullable (`000004`).
- `empresas` (`000007`) — `nombre_empresa VARCHAR(150) NOT NULL`, `user_id NOT NULL → users` con `ON DELETE CASCADE`. Un usuario tiene N empresas; índice `idx_empresas_user_id`.
- `access_keys` — `code VARCHAR(50) UNIQUE NOT NULL`, `created_by → users` **nullable** con `ON DELETE SET NULL` (`000007`), `redeemed_by → users` (`SET NULL`), `redeemed_at`, `expires_at`, `revoked_at` y `revoked_by → users` (`000005`, `SET NULL`), `created_at`.
- `categories` — `nombre VARCHAR(100) NOT NULL`; `empresa_id → empresas` **nullable** (`000007`, `ON DELETE CASCADE`), donde NULL = categoría global; índice `idx_categories_empresa_id`. `000008` siembra las globales `Ventas`, `Servicios` y `Pedidos`.
- `incomes` — `empresa_id NOT NULL → empresas` (`000007`, `CASCADE`), `monto BIGINT NOT NULL` (`000003`), `payment_method`, `descripcion TEXT` nullable, `fecha DATE NOT NULL DEFAULT CURRENT_DATE`, `created_at`, `category_id → categories` nullable (`000006`, `ON DELETE SET NULL`).
- `expenses` — `empresa_id NOT NULL → empresas` (`000007`, `CASCADE`), `monto BIGINT NOT NULL`, `category_id NOT NULL → categories` (`ON DELETE CASCADE`), `descripcion`, `fecha DATE NOT NULL DEFAULT CURRENT_DATE`, `created_at`.
- `debtors` — `empresa_id NOT NULL → empresas` (`000007`, `CASCADE`), `nombre`, `monto BIGINT NOT NULL`, `descripcion`, `pagado BOOLEAN NOT NULL DEFAULT false`, `fecha_limite DATE` nullable, `created_at`.
- `goals` — `empresa_id NOT NULL → empresas` (`000007`, `CASCADE`), `nombre`, `periodo` (`goal_period`), `monto_meta BIGINT NOT NULL`, `created_at`.
- `password_reset_codes` (`000010`) — `user_id NOT NULL → users` (`CASCADE`), `codigo VARCHAR(64) NOT NULL` (hash SHA-256 del código de 6 dígitos, nunca el código en claro), `expira_en TIMESTAMPTZ NOT NULL`, `verificado_en` y `usado_en` TIMESTAMPTZ nullable; índice `idx_password_reset_codes_user_id`. Es la tabla del flujo de recuperación de contraseña: el código pasa por *emitido → verificado → usado*.

`000007` sacó `user_id` de `incomes`, `expenses`, `debtors`, `goals` y `categories`: la tenencia pasa a ser `empresa_id → empresas → users`. Crea los índices `idx_incomes_empresa_id`, `idx_expenses_empresa_id`, `idx_debtors_empresa_id`, `idx_goals_empresa_id` y `idx_categories_empresa_id`.

**Unicidad de categorías (solo aplicación):** una empresa no puede repetir nombre de categoría, pero el mismo nombre sí puede existir en otra empresa. La regla vive en `ServicioCategorias` (`normalizarNombre` + `validarNombreUnico`) y **no tiene índice ni constraint en la base**. La comparación ignora mayúsculas, acentos y espacios (`"Comida" == "comida" == "COMIDA"`, `"Áqui" == "aqui"`). Al ser un chequeo previo al `INSERT`, dos creates concurrentes de la misma empresa pueden colarse; cerrarlo requiere transacción + `pg_advisory_xact_lock` o un `UNIQUE (empresa_id, nombre)`.

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

Un tag roto no devuelve 400: **el `Validador.Struct` entra en panic** y el `Recoverer` de chi lo convierte en un 500 sin cuerpo. Un tag roto tumba el endpoint entero. Sin espacios tras las comas.

`utilidades.Validador` es un `*validator.Validate` global creado con `validator.WithRequiredStructEnabled()`; por eso `required` aplica también a structs anidados.

## Tests

Paquetes con tests: `clavesAcceso`, `auth`, `categories`, `debtors`, `empresas`, `expenses`, `goals`, `incomes`, `paginacion`, `users` y `utilidades`. Sin tests: `bd`, `bdtx`, `excel`, `informe`, `correo`, `entorno`, `errores`, `intermediario` y `model`.

Convención, un archivo por capa dentro del paquete (tests en `package categories`, no `categories_test`):

| Archivo | Qué va |
|---|---|
| `mocks_test.go` | mocks de store y service + fixtures (`newTestX`, IDs compartidos) |
| `service_test.go` | lógica del service contra el store mockeado |
| `handler_test.go` | HTTP con `httptest` + `chi.NewRouteContext()` para URL params; contexto con `utilidades.ContextoUsuario` / `utilidades.ContextoEmpresa` |
| `storage_test.go` | `go-sqlmock` sobre el store real |

`testify` (`assert` / `require` / `mock`) y `go-sqlmock` son dependencias directas.

**sqlmock no parsea SQL**: solo compara strings y devuelve las filas que le digas. No detecta un `Scan` desalineado ni una columna inexistente. Por eso la mayoría de los `storage_test.go` (`clavesAcceso`, `auth`, `categories`, `debtors`, `empresas`, `expenses`, `goals`, `incomes`, y un caso de `users`) usan un `QueryMatcherFunc` permisivo que captura el SQL realmente ejecutado y lo audita a mano. Los tests de integración contra Postgres real viven en `tests/integration/` (ver sección siguiente).

`categories` tiene tests específicos de la regla de negocio: `TestNormalizarNombre`, duplicados por mayúsculas/acentos en `Crear` y en `CambiarNombre`, y `ObtenerCategoriasPorEmpresa` / `ObtenerCategoriasVisibles` en el store.

`debtors` y `goals` tienen tests de service, handler y storage. `debtors` cubre CRUD, listado, `pagado` y mapeo de `ErrorNoEncontrado`; `goals` cubre el cálculo de estado y las rutas de detalle/estado.

### Tests de integración (`tests/integration/`)

Prueban flujos HTTP completos contra el router real y una base PostgreSQL real.
Complementan los tests unitarios; no los reemplazan. El resumen de cobertura,
bugs y discrepancias vive en `tests/report.md` y se actualiza con cada flujo.

**Infraestructura**

- `TestMain` está en `tests/integration/main_test.go` y delega en `testutil.SetupSuite`.
- `testutil/env.go` crea una base efímera `tally_it_<pid>` desde `DB_ADDR` (default `postgres://postgres:tally_database_env@localhost/tally_database_env?sslmode=disable`), aplica `migrations/*.up.sql` en orden y la dropea al terminar. `TEST_DB_ADDR` permite reutilizar una base existente sin crearla ni dropearla.
- El router se replica en `testutil/env.go` desde `cmd/application/app.go`. **No** se modifica el archivo de producción para hacer pasar un test.
- Se omite el middleware `Logger` para no ensuciar `go test -v`; se mantienen `RequestID`, `ClientIPFromHeader("X-Real-IP")` y `Recoverer`. El remitente de correo se reemplaza por uno que **guarda los códigos enviados** (`testutil.UltimoCodigoDeRecuperacion`), para poder completar el flujo de recuperación de contraseña sin SMTP real.
- `TestMain` registra un único usuario municipal; los tests lo obtienen con `testutil.MunicipalToken(t)` / `testutil.MunicipalRUT(t)`.

**Archivos y nombres**

- Un archivo por flujo: `tests/integration/<flujo>_flow_test.go` (`auth_flow_test.go`, `password_recovery_flow_test.go`, `incomes_flow_test.go`, etc.).
- El aislamiento multi-empresa vive en `tenant_isolation_test.go`.
- Helpers compartidos en `helpers_test.go` (`requireProtected`, `dbDate`); entorno, cliente HTTP y fixtures en `testutil/`.
- Los tests van en `package integration` (no `integration_test`).

**Reglas de escritura**

- HTTP en memoria con `httptest` a través de `testutil.Request`; no levantar un servidor real ni usar URLs externas.
- Usar `testify` (`require` / `assert`). No agregar dependencias nuevas.
- Flujo feliz secuencial reutilizando IDs generados; negativos en tablas con `t.Run`.
- Validar **status exacto** y **mensaje exacto** (`testutil.RequireStatus` / `testutil.RequireError`). No inventar rutas, campos ni mensajes: el código es la fuente de verdad; swagger solo orienta.
- Cada test debe ser independiente: crear sus propios usuarios/empresas con los fixtures (`NewUser`, `CreateEmpresa`, `CreateCategory`, `CreateIncome`, `CreateExpense`, `CreateDebtor`, `CreateGoal`) y limpiar con `t.Cleanup`. No depender del orden entre tests.
- Todo flujo debe verificar al menos una ruta protegida sin token y con token inválido usando `requireProtected` (`401 unauthorized`).
- Fechas: usar `dbDate(t, offset)` para alinear con `CURRENT_DATE` de Postgres y no con el reloj del host.
- Cubrir explícitamente el aislamiento por empresa: acceso a la empresa ajena y a recursos ajenos por ID, además de listados que excluyan datos de otra empresa.

**Bugs y discrepancias**

- No corregir bugs de producción desde un test. Documentarlos en el test con `BUG:` (comportamiento incorrecto) o `DISCREPANCIA:` (swagger/código o código/código) y assertar el comportamiento actual.
- Mantener `tests/report.md` actualizado con el resultado por flujo, bugs, discrepancias y casos omitidos.

**Ejecución**

```bash
go tool task test                                   # vet + go test ./... (incluye integración si hay Postgres)
GOPROXY=off go test -v ./tests/integration/...      # integración completa
GOPROXY=off go test -v ./tests/integration/ -run TestIncomesFlow
```

Requiere PostgreSQL. Sin `TEST_DB_ADDR`, la suite crea y dropea su propia base.

**CI:** el workflow levanta un service container `postgres:16` y exporta `DB_ADDR`; la suite crea su base efímera y aplica las migraciones sola. No hace falta un step de migraciones.

## Arquitectura

`handler.go` → `service.go` → `storage.go` por módulo, con interfaces inversas (`RepositorioIngresos`, `IServiceIngresos`, etc.) definidas en service/handler. Cada módulo implementado expone un `router.go` con `InitModule`, pero la firma varía: la mayoría `(r, db, logger, middl)`; `auth` recibe además `*bdtx.Transactor` y la config JWT; `clavesAcceso` recibe el `SendEmail`; `empresas.InitModule` no recibe router y devuelve `Routes` + `intermediario.BuscadorEmpresa`.

Los módulos `excel` y `informe` son esqueletos: tienen `handler.go`, `service.go`, `storage.go` y `router.go` con TODOs, sin lógica ni rutas. `excel` generará un archivo Excel y lo guardará en S3 por usuario y empresa; `informe` obtendrá una sola vez toda la información de un usuario con todas sus empresas o de una sola empresa.

Cableado actual en `cmd/application/app.go`:

- `/v1/swagger/*` — UI de Swagger.
- `/v1/autenticacion` — `POST /` (registro con código de acceso), `POST /inicio-sesion`, `POST /municipal`, más el flujo de recuperación de contraseña (`POST /recuperar-password` emite el código de 6 dígitos y lo manda por correo, `POST /recuperar-password/verificar` lo valida, `POST /recuperar-password/confirmar` cambia la contraseña). Sin JWT.
- `/v1/aplicacion` — todo lo demás, detrás de `MiddlewareTokenAutenticacion`:
  - `/usuarios` — `PATCH /`, `DELETE /`, `GET /municipal/{rut}` (este último rol municipal).
  - `/claves` — `POST /` (emitir), `POST /IDAcceso/{claveID}` (reenviar), `POST /revocar/{codigo}` (revocar). Solo rol municipal.
  - `/empresas` — `POST /` y `GET /`.
  - `/empresas/{empresaID}` — detrás de `MiddlewareEmpresa.EmpresaDelUsuario`: `PATCH /{nombre}` y `DELETE /` de la empresa, más `ingresos` (`POST /`, `GET /`, `PATCH /{ingresoID}`, `DELETE /{ingresoID}`), `gastos` (`POST /`, `GET /`, `PATCH /{gastoID}`, `DELETE /{gastoID}`), `categorias` (`GET /` lista las propias y las globales, `POST /`, `PATCH /{categoriaID}`, `DELETE /{categoria}`), `metas` (prefijo `/metas`: `POST /`, `GET /`, `GET /{metaID}`, `DELETE /{metaID}`, `GET /{metaID}/estado`) y `deudas` (prefijo `/deudas`: `POST /`, `GET /`, `PATCH /{deudorID}`, `PATCH /{deudorID}/pagada`).

**No montados:** `excel` y `informe` (esqueletos sin lógica ni rutas). El resto de los módulos con lógica (`clavesAcceso`, `auth`, `categories`, `debtors`, `empresas`, `expenses`, `goals`, `incomes`, `users`) está montado.

Middlewares:
- `intermediario.MiddlewaresGlobales()`: `RequestID`, `ClientIPFromHeader("X-Real-IP")`, `Logger`, `Recoverer`.
- `MiddlewareAutenticacion.MiddlewareTokenAutenticacion`: valida `Authorization: Bearer <jwt>`, saca `sub` del claim, carga el usuario y lo inyecta en `utilidades.ContextoUsuario`.
- `MiddlewareAutenticacion.VerificarRol(rol, handler)`: valida **rol**, no pertenencia. El acotado por usuario/empresa va en el `WHERE` del store.
- `MiddlewareEmpresa.EmpresaDelUsuario`: parsea `{empresaID}` y busca la empresa por `(empresaID, userID)`, inyectándola en `utilidades.ContextoEmpresa`.

El usuario y la empresa del request se sacan del contexto con `utilidades.ObtenerUsuarioDelContexto(r)` y `utilidades.ObtenerEmpresaDelContexto(r)` (no hay helpers en `users`). Devuelven `nil` si el middleware no los inyectó; usarlos sin el middleware correspondiente produce panic.

`users.AlmacenUsuarios.ObtenerPorID` (el que usa `MiddlewareTokenAutenticacion`) trae `password_hash` a propósito: `PATCH /v1/aplicacion/usuarios` reescribe la fila con ese usuario, así que si el hash no viniera el UPDATE escribiría un hash vacío en una columna `NOT NULL` y el endpoint devolvía 500. `ObtenerPorRUT` no lo trae: es solo de lectura.

`bdtx.Transactor` (`internal/bdtx`) envuelve `BeginTx`/`Commit`/`Rollback`. Los stores que participan de una transacción deben usar `bdtx.FromContext(ctx, s.db)` en vez de `s.db`; hoy lo hacen `auth` y `clavesAcceso`. `auth.RegistrarUsuario` crea el usuario, canjea la llave y setea `plan_expires_at` en una sola transacción.

Respuestas siempre envueltas en `{"datos": ...}` vía `utilidades.RespuestaJSON`; errores como `{"error": "..."}` vía `utilidades.EscribirErrorJSON`. `utilidades.LeerJSON` usa `DisallowUnknownFields` y el handler traduce el fallo a **400**. `utilidades.ErrorNoEncontrado` / `utilidades.ErrorConflicto` son los errores de dominio compartidos; `internal/errores` mapea a 400/401/403/404/409/500.

Paginación: `paginacion.ConsultaPaginacionIngresos` (compartida por `incomes` y `expenses`): `limite` default y máximo 30, `desplazamiento`, `desde`/`hasta` en `AAAA-MM-DD` y `categoria_id` UUID. Un parámetro inválido devuelve error → 400.

Categorías al crear ingresos/gastos: llegan por **`category_id`** (UUID). El service valida con el store `CategoriaExiste` que la categoría sea global (`empresa_id IS NULL`) o de la empresa; si no existe o es de otra empresa, el handler responde **400**. En incomes es opcional porque `incomes.category_id` es nullable; en expenses es obligatoria porque la columna es `NOT NULL`. El payload de expenses no acepta `fecha`: la asigna la base con `DEFAULT CURRENT_DATE`.

Edición de movimientos (`PATCH .../ingresos/{ingresoID}` y `PATCH .../gastos/{gastoID}`): el payload trae **punteros** y lo que no se manda se queda como está. `fecha` es opcional y va en `AAAA-MM-DD` (formato inválido → **400** `la fecha debe tener el formato AAAA-MM-DD`). Un payload sin ningún campo → **400** `no se enviaron campos para actualizar`. Un movimiento que no existe o es de otra empresa → **404**. La fecha viaja a la base como texto (`$n::date`) para no depender de la zona horaria de la sesión.

Recuperación de contraseña (`internal/auth`): el código de 6 dígitos se genera con `crypto/rand` y **se guarda hasheado (SHA-256)**, nunca en claro. El estado vive en `password_reset_codes`: *emitido → verificado → usado*. Pedir un código nuevo invalida el anterior. Los errores del dominio (`ErrorCodigoRecuperacionInvalido`, `...Vencido`, `...Usado`, `...NoVerificado`) y el email no registrado (`utilidades.ErrorNoEncontrado`) responden **400**. El remitente es la interfaz `auth.Remitente` (`SendPasswordResetCode`), que implementan `correo.Client`, `correo.ClienteGmail` y el mock de `cmd/application/app.go`.

Borrado de categorías: una categoría con ingresos o gastos asociados **no se puede borrar** (`409` `la categoría tiene movimientos asociados y no se puede eliminar`), porque `expenses.category_id` es `NOT NULL` con `ON DELETE CASCADE` y `incomes.category_id` tiene `ON DELETE SET NULL`. El chequeo (`TieneMovimientos`) filtra por la empresa de la categoría: una categoría ajena sigue respondiendo **404**, no 409.

## Swagger

Las anotaciones viven en un `swagger.go` por módulo (funciones vacías, no se llaman): `clavesAcceso`, `auth`, `categories`, `debtors`, `empresas`, `expenses`, `goals`, `incomes` y `users`. `excel` e `informe` no tienen endpoints todavía.

`docs/docs.go`, `docs/swagger.json` y `docs/swagger.yaml` son **generados**: después de tocar anotaciones o tipos, `go tool task gen-docs` y commiteá el resultado. La UI queda en `/v1/swagger/`. El root de swagger se anota en `cmd/main.go` (`@BasePath /v1`, `@securityDefinitions.apiKey`).

## CI

`../.github/workflows/test.yml` corre solo el job `backend`: `go mod download`, `go vet ./...`, `go test -race -covermode=atomic -coverprofile=coverage.out ./...`, resumen con `go tool cover` y upload a Codecov (`fail_ci_if_error: false`). Usa `go-version-file: Backend/go.mod` y `working-directory: Backend`. El workflow vive en la raíz del repo porque GitHub Actions solo lee `.github/` del root. Triggers: push a `main`, PRs a `main` y manual.

El job levanta un service container `postgres:16` (usuario `postgres`, contraseña `tally_database_env`, base `tally_database_env`) con health check `pg_isready`, y exporta `DB_ADDR` apuntando a `localhost:5432`. Ese Postgres es efímero y vive solo durante el job. No hay step de migraciones: `testutil` crea `tally_it_<pid>` dentro de esa instancia y aplica `migrations/*.up.sql` ahí mismo.

## Deuda técnica conocida (verificada, sin arreglar)

- La unicidad de categorías es solo a nivel aplicación: hay ventana de carrera entre el chequeo y el `INSERT` (ver esquema). Cerrarla requiere transacción + `pg_advisory_xact_lock` o un `UNIQUE (empresa_id, nombre)`.
- `excel` e `informe` son solo esqueletos: no tienen handlers, services, stores, rutas, tests ni cliente S3 implementados.
- En la edición de movimientos no se puede **borrar** un campo: mandar `"descripcion": null` o `"category_id": null` equivale a "no tocarlo". Para quitar la categoría de un ingreso hay que hacerlo por otra vía.
- `GET /v1/aplicacion/usuarios/municipal/{rut}` sigue devolviendo `created_at` y `updated_at` en cero: `ObtenerPorRUT` no selecciona esas columnas (bug conocido, documentado en `tests/report.md`).

## Otros

- `../Frontend/` está vacío salvo `.gitkeep`: no hay consumidor del contrato JSON todavía, así que cambiar un tag no rompe nada visible.
- `../Fase 1/` y `../Fase 2/` son carpetas de evidencias del curso, sin código.
- `docker/` tiene `Dockerfile`, `docker-compose.yml` y `entrypoint.sh` para desplegar la API; el `docker-compose.yml` de la raíz solo levanta Postgres.
- `.air.toml` configura live reload con `go build -o ./bin/main ./cmd` (la ruta correcta del entrypoint).
- `gofmt` no está enganchado a ningún hook: corré `gofmt -w` sobre lo que toques.
- Todos los archivos, incluido `internal/auth/swagger.go`, están formateados con `gofmt`.
