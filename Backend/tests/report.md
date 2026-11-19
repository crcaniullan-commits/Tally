# Reporte de tests de integración

Fecha: 2026-11-18

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

Última corrida completa: 14 flujos, 18 tests top-level y 423 subtests, todos en
verde (441 líneas `--- PASS`, ~27 s en frío).
`gofmt -l tests/integration/` y `go vet ./tests/integration/...` sin hallazgos.

| Flujo | Archivo | Resultado | Casos |
|---|---|---|---|
| Auth | `integration/auth_flow_test.go` | PASS | 36 |
| Recuperación de contraseña | `integration/password_recovery_flow_test.go` | PASS | 12 |
| Claves de acceso | `integration/accesskeys_flow_test.go` | PASS | 20 |
| Empresas | `integration/empresas_flow_test.go` | PASS | 27 |
| Categorías | `integration/categories_flow_test.go` | PASS | 35 |
| Ingresos | `integration/incomes_flow_test.go` | PASS | 54 |
| Gastos | `integration/expenses_flow_test.go` | PASS | 57 |
| Deudores | `integration/debtors_flow_test.go` | PASS | 45 |
| Metas | `integration/goals_flow_test.go` | PASS | 40 |
| Usuarios | `integration/users_flow_test.go` | PASS | 29 |
| Archivos Excel | `integration/excel_flow_test.go` | PASS | 20 |
| Cache | `integration/cache_flow_test.go` | PASS | 6 |
| Aislamiento multi-empresa | `integration/tenant_isolation_test.go` | PASS | 23 |
| Dispositivos, Notificaciones y Jobs | `integration/notificaciones_jobs_flow_test.go` | PASS | 37 |

Los flujos que ejercitan rutas protegidas verifican, salvo las excepciones
anotadas abajo, una ruta sin token y con token inválido (`401 no autorizado`).
`password_recovery_flow_test.go` es público por diseño; `cache_flow_test.go` y
`tenant_isolation_test.go` usan tokens válidos y validan el acceso por empresa,
pero no repiten el chequeo de `401`.

## Estructura

```text
tests/
├── report.md
└── integration/
    ├── main_test.go              # TestMain -> testutil.SetupSuite
    ├── helpers_test.go           # requireProtected, dbDate
    ├── <flujo>_flow_test.go      # un archivo por flujo
    ├── password_recovery_flow_test.go
    ├── excel_flow_test.go
    ├── cache_flow_test.go
    ├── tenant_isolation_test.go
    └── testutil/
        ├── env.go                # base efímera, migraciones, router real
        ├── cache.go              # cacheMemoria: Cache en memoria para la suite
        ├── client.go             # Request, RequestRaw, DecodeData, DecodeError, Require*
        ├── fixtures.go           # NewUser, CreateEmpresa, CreateCategory, ...
        └── rut.go                # generador de RUT módulo 11
```

## Cobertura por flujo

- **Auth:** registro municipal, registro canjeando llave, `plan_expires_at`,
  login, llave reutilizada/inexistente, email/RUT duplicado, RUT inválido,
  validaciones de payload y acceso a ruta protegida.
- **Recuperación de contraseña:** envío del código de 6 dígitos por correo,
  email no registrado (`400`), código incorrecto, expirado o ya usado, verificación
  previa al cambio de contraseña, login con la contraseña nueva y con la vieja.
- **Access keys:** emisión, reenvío por ID, revocación por código, canje de llave
  revocada, permisos de rol y validaciones.
- **Empresas:** creación, listado, renombrado (con tope de 30 caracteres),
  borrado, doble borrado, listado vacío, validaciones, rol municipal y
  aislamiento entre usuarios.
- **Categorías:** creación con `tipo` (`ingreso`/`gasto`), listado propio + globales
  (las globales de ventas son de tipo `ingreso` y las de gastos, de tipo `gasto`),
  renombrado, duplicados ignorando mayúsculas/acentos, mismo nombre permitido con
  otro tipo, mismo nombre en otra empresa, categorías globales no modificables
  desde la empresa, borrado, borrado bloqueado (`409`) cuando la categoría tiene
  ingresos o gastos, PATCH con `tipo` rechazado (`400` por campo desconocido) y
  validaciones de payload.
- **Ingresos:** creación con/sin categoría, categoría global, listado, filtros
  por categoría y fechas, paginación, edición (sin tocar la fecha), borrado,
  categoría inexistente/ajena, categoría de tipo `gasto` rechazada en el alta y
  en la edición, y validaciones de payload/query.
- **Gastos:** creación con categoría y descripción, rechazo de las globales de tipo
  `ingreso` y uso de las globales de tipo `gasto`, listado, filtros, paginación,
  edición (sin tocar la fecha), borrado, categoría inexistente/ajena, categoría de
  tipo `ingreso` rechazada en el alta y en la edición, método de pago obligatorio
  al crear y editable en el `PATCH`, y validaciones.
- **Deudores:** creación, listado, actualización (nombre, monto y fecha límite),
  actualización del comentario (`descripcion`) que se queda si no se manda, marcado
  como pagado, borrado (`200` con `Deuda eliminada`, también de una deuda ya
  pagada), borrado de una deuda de otra empresa (`404`), campos desconocidos y
  validaciones.
- **Metas:** creación por período, listado, detalle, cálculo de estado con
  ingresos del día/semana/mes, pendiente no negativo, borrado y validaciones.
- **Usuarios:** actualización de nombre+contraseña, login con nueva contraseña,
  actualización solo de contraseña, actualización solo de nombre, payload vacío,
  que un PATCH posterior no revierte el nombre anterior, consulta municipal por
  RUT (con el objeto `rut` serializado como `{"cuerpo":..., "dv":"..."}`),
  validaciones, borrado de cuenta e invalidación del token.
- **Archivos Excel:** subida de un `.xlsx` (multipart, campo `file`), listado con
  el envelope `{"datos": [...]}`, descarga con el binario y el content-type de
  hoja de cálculo, borrado (`204`), archivo que no es un zip válido (`400`),
  subida sin archivo (`400`), descarga y borrado de un archivo inexistente
  (`404`), aislamiento entre empresas (listado vacío y `404` por id), municipal
  sin acceso a la empresa (`404`), registro en la tabla `excel_files` y borrado
  en cascada al eliminar la empresa.
- **Cache:** el router de la suite se arma con los mismos buscadores cacheados
  que usa el backend real (`users.BuscadorUsuarioConCache` y
  `empresas.BuscadorEmpresaConCache`) sobre un `cache.Cache` en memoria
  (`testutil.cacheMemoria`), así que los flujos ejercitan el camino cacheado
  sin depender de un Redis. Cubre: el middleware resuelve la empresa con el
  cache, el cache no deja pasar a una empresa ajena (la clave incluye el
  usuario), el rename y el borrado de una empresa invalidan su entrada
  cacheada, y los gastos siguen siendo accesibles con el cache activo.
- **Aislamiento multi-empresa:** acceso a empresa ajena (`404` en middleware),
  recursos ajenos por ID desde la empresa propia (`404` en el `WHERE`), listados
  que excluyen datos ajenos, categoría ajena rechazada y rol municipal sin
  acceso a empresas.

## Bugs del backend documentados

Los tests no corrigen código de producción: documentan el comportamiento actual
con `BUG:` o `DISCREPANCIA:` y el arreglo se hace en un cambio aparte. Los bugs ya
corregidos figuran en la tabla siguiente.

| # | Ruta | Esperado | Actual | Causa |
|---|---|---|---|---|
| 1 | `GET /v1/aplicacion/usuarios/municipal/{rut}` | `created_at` y `updated_at` reales | `0001-01-01T00:00:00Z` | `ObtenerPorRUT` no selecciona esas columnas. |

Las discrepancias entre Swagger y el código que quedaban pendientes están
corregidas: van a la tabla de [Bugs ya corregidos](#bugs-ya-corregidos) (14 a
20). No queda ninguna marcada con `DISCREPANCIA:` en los tests.

### Bugs ya corregidos

| # | Ruta | Antes | Ahora | Causa |
|---|---|---|---|---|
| 2 | `POST /v1/autenticacion` y `POST /v1/autenticacion/municipal` con email/RUT duplicado | `500 {"error":"el servidor encontró un problema"}` | `400` con `ya existe un usuario con ese correo` / `ya existe un usuario con ese RUT` | `internal/auth/storage.go` comparaba `err.Error()` contra el texto exacto de pq, pero lib/pq agrega ` (23505)`. Ahora se mira el error tipado `*pq.Error` (código `23505` + constraint) con un fallback por texto. |
| 3 | `POST /v1/autenticacion` y `POST /v1/autenticacion/municipal` con RUT mal formado | `500 {"error":"el servidor encontró un problema"}` | `400` con `el formato del RUT no es válido` | El `ParseRUT` vive en el service y el handler de auth no mapeaba `utilidades.ErrFormatoInvalido`. |
| 4 | `PATCH /v1/aplicacion/usuarios` con solo `nombre` o payload vacío | `500 {"error":"el servidor encontró un problema"}` | `200` y el nombre queda actualizado | `MiddlewareTokenAutenticacion` usa `ObtenerPorID`, que no seleccionaba `password_hash`; el UPDATE escribía un hash vacío en una columna `NOT NULL`. Ahora `ObtenerPorID` trae el hash. |
| 5 | `DELETE /v1/aplicacion/empresas/{empresaID}/categorias/{categoria}` con movimientos | `200` y los gastos de la categoría borrados en cascada | `409` y la categoría intacta | La categoría se borraba sin mirar si tenía ingresos o gastos asociados. |
| 6 | `GET /v1/aplicacion/empresas/{empresaID}/excel` y `GET .../excel/{id}` | `200` | `500` (columna `emrpesa_id` y tabla `exel_files` inexistentes) | Typos en el SQL de `internal/excel/store.go`. |
| 7 | `POST /v1/aplicacion/empresas/{empresaID}/excel` con un archivo que no es `.xlsx` | `400` | `500` (panic por `ErrorInterno` con error `nil`) | `internal/excel/handler.go` mezclaba el error de lectura con el de magic bytes y pasaba un `nil` a `ErrorInterno`. |
| 8 | `DELETE /v1/aplicacion/empresas/{empresaID}/excel/{id}` con un id inexistente | `404` | `204` | `Service.Delete` se comía el error del store (`return nil`) y el `ErrNotFound` nunca llegaba al handler. |
| 9 | `GET /v1/aplicacion/empresas/{empresaID}/excel` | `{"datos":[...]}` | `[...]` (array pelado) | El handler usaba `utilidades.EscribirJSON` en vez de `utilidades.RespuestaJSON`. |
| 10 | `POST /v1/aplicacion/empresas/{empresaID}/excel` (ruta completa) | `404` (no existía) | `404` en una ruta duplicada (`/empresas/{id}/empresas/{id}/excel`) | `excel.InitModule` registraba `/empresas/{empresaID}/excel` pero se montaba dentro del grupo `/empresas/{empresaID}`. Ahora registra `/excel`. |
| 11 | `migrations/000009_excel_files.up.sql` | creaba la tabla `excel_files` | la dropeaba (el `CREATE TABLE` estaba en el `.down.sql`) | Los dos archivos de la migración estaban invertidos: la tabla nunca existía en la base. |
| 12 | `PATCH /v1/aplicacion/usuarios` con solo `nombre` o payload vacío, con el cache activo | `200` y el nombre queda actualizado | `500 {"error":"el servidor encontró un problema"}` | `users.BuscadorUsuarioConCache` cachea el `model.User` serializado a JSON, y `model.User.PasswordHash` tiene `json:"-"`, así que el usuario servido desde el cache llega **sin hash**. `AlmacenUsuarios.Actualizar` reescribía la fila completa con ese objeto, así que mandaba `password_hash = NULL` y Postgres lo rechazaba (`23502`). **Corregido**: el store ahora aplica solo los campos del payload (`CASE WHEN`), con el patrón de `CambiosGasto` de `expenses`, y `Actualizar` invalida la entrada cacheada. |
| 4 (regresión) | `PATCH /v1/aplicacion/usuarios` con solo `nombre` o payload vacío | `200` | `500` con el cache activo | El bug 4 se había corregido haciendo que `ObtenerPorID` traiga el `password_hash`. Al agregar el cache, el usuario vuelve a perderlo en el round-trip por JSON (`PasswordHash` tiene `json:"-"`). Ver el bug 12. |
| 13 | `cache.Version` / `cache.Bump` | `Bump` invalida la versión que lee `Version` | `Bump` incrementaba `var:<scope>` y `Version` leía `ver:<scope>`: son claves distintas, así que un `Bump` no cambiaba lo que devuelve `Version` | Prefijos desalineados en `internal/cache/cache.go`. **Corregido**: los dos usan ahora `claveVersion(scope)`. |
| 14 | `GET /v1/aplicacion/empresas` sin empresas | `404` (lo que decía el Swagger) | `200 {"datos":null}` | El handler nunca devolvía `404` en el listado: el `404` de la anotación era un error de documentación. **Corregido**: se eliminó el `@Failure 404` de la anotación y la descripción aclara que un usuario sin empresas recibe `200` con `{"datos":null}`. |
| 15 | `PATCH /v1/aplicacion/empresas/{empresaID}/{nombre}` con un nombre de más de 30 caracteres | `400` | `200` y el nombre se guardaba | El nombre viene del path, no del body, así que no pasaba por el `validate:"max=30"` del alta. **Corregido**: el handler valida el largo (30 caracteres) y responde `400` con `el nombre de la empresa no puede superar los 30 caracteres`. |
| 16 | `GET /v1/aplicacion/usuarios/municipal/{rut}`, campo `rut` | `{"cuerpo":..., "dv":"..."}` | `{"Cuerpo":..., "DV":"..."}` | `model.RUT` no tenía tags JSON, así que serializaba con los nombres de los campos de Go. **Corregido**: `internal/model/rutFormat.go` declara `json:"cuerpo"` y `json:"dv"`. |
| 17 | Todos los endpoints bajo `/v1/aplicacion` | Solo `200`/`500` en los listados | El middleware puede devolver `400`, `401`, `403` y `404` | Las anotaciones no documentaban las respuestas de los middlewares (`MiddlewareTokenAutenticacion`, `VerificarRol` y `EmpresaDelUsuario`). **Corregido**: los endpoints de `/aplicacion` documentan `401`, más `403` cuando exigen rol y `400`/`404` cuando pasan por `EmpresaDelUsuario`. |
| 18 | Esquemas de respuesta de todo el API | Mitad con el envelope `{"datos": ...}`, mitad con el objeto pelado | Siempre `{"datos": ...}` | Varias anotaciones declaraban `{object} model.X` o `{object} []model.X` cuando la respuesta real viene envuelta. **Corregido**: los esquemas de éxito son `map[string]X` o `map[string][]X`, salvo el `204` de excel/dispositivos (sin cuerpo) y el `200` binario de la descarga. |
| 19 | `POST /aplicacion/dispositivos` y `DELETE /aplicacion/dispositivos` | No documentados | Ambos están montados en `internal/dispositivos/router.go` | Faltaban las anotaciones de las rutas que operan sobre la colección. **Corregido**: se documentaron `POST /` (registrar) y `DELETE /` (desvincular por token en el body). |
| 20 | `POST /aplicacion/claves` (descripción) | El canje se hacía en `POST /aplicacion/autenticacion` | El canje real es `POST /autenticacion` | La ruta de auth se monta en `/v1/autenticacion`, no bajo `/aplicacion`. **Corregido** en la descripción de `internal/clavesAcceso/swagger.go`. |

## Diferencias entre swagger y código

Todas las diferencias conocidas entre la documentación y el código están
corregidas. El detalle de cada corrección está en la tabla de
[Bugs ya corregidos](#bugs-ya-corregidos).

| # | Punto | Swagger | Código |
|---|---|---|---|
| 1 | `POST /aplicacion/empresas` | `201` | `201` |
| 2 | `GET /aplicacion/empresas` sin empresas | `200 {"datos":null}` | `200 {"datos":null}` |
| 3 | `PATCH /aplicacion/empresas/{empresaID}/{nombre}` | `400` con nombre vacío o de más de 30 caracteres | `400` con nombre vacío o de más de 30 caracteres |
| 4 | `POST /aplicacion/usuarios/exchange/{code}` | No documentado | No montado (`404`); el canje real es `POST /v1/autenticacion` |
| 5 | `model.RUT` | `{"cuerpo":..., "dv":"..."}` | `{"cuerpo":..., "dv":"..."}` |
| 6 | Códigos de los listados | `400`/`401`/`403`/`404` además de `200`/`500` | Los mismos |
| 7 | `POST`/`DELETE` de `/aplicacion/dispositivos` | Documentados | Montados en el router |
| 8 | Canje de llaves en la descripción de `/aplicacion/claves` | `POST /autenticacion` | `POST /v1/autenticacion` |

## Casos omitidos

- Expiración/refresh de JWT: no hay endpoint de refresh y esperar la expiración
  haría el test lento; solo se prueba token inválido.
- Envío real de correo SMTP: los tests usan un remitente que solo guarda el
  código enviado (`testutil.UltimoCodigoDeRecuperacion`).
- Vencimiento del código de recuperación (15 minutos): no se prueba esperando el
  reloj; se cubre la unidad de tiempo del código en el service.
- `informe`: es un esqueleto sin rutas ni lógica.
- Storage de archivos en S3: el cliente existe (`internal/storage/s3.go`) y
  `cmd/application/app.go` lo usa cuando hay configuración `S3_*`; el suite usa
  `storage.NewDisk` sobre un directorio temporal, así que la integración con S3
  no se ejercita.
- Envío real a Expo Push: el suite usa `LogNotifier` (loguea y devuelve `nil`).
  El `ClienteExpoPush` real no se ejercita en integración; está cubierto en el test unitario de `notificaciones`.
- Scheduler cron real: `PlanificadorJobs` inicia en background desde
  `cmd/application/routes.go`, pero no en el router del suite.
  Los jobs se ejercitan vía HTTP (`POST /jobs/recordatorio-diario`, etc.).
- Historial de notificaciones post-job: el test verifica que el endpoint `GET /historial` responde 200.
  La verificación de que `notification_log` contiene una fila específica requeriría datos que activen el filtro SQL
  del job (usuario sin movimientos hoy / con cobros vencidos), lo que haría el test frágil al horario; se cubre
  en el test unitario del service.
- Subida de un archivo de más de 10 MB: el tope lo aplica `http.MaxBytesReader` y no se prueba con un cuerpo de 10 MB.
- Carrera de unicidad de categorías: deuda técnica conocida y no determinista.
- Borrado en cascada de empresas/ingresos/gastos al borrar un usuario: solo se
  verifica que el token deja de servir (`401`).
- UI de Swagger (`/v1/swagger/*`).
- Métodos HTTP no soportados (`405`) y payloads gigantes.
- Enforcement de `plan_expires_at`: el código no lo valida.
- Rate limiting: `errores.ErrorLimiteExcedido` existe pero ningún middleware lo llama.
- Revocación de llaves entre municipales distintos: solo hay un municipal en la
  suite.

## Mantenimiento

- Al agregar un flujo, crear `integration/<flujo>_flow_test.go`, correrlo solo y
  después la suite completa.
- Si se descubre un bug, no arreglarlo en el mismo cambio: documentarlo con
  `BUG:`/`DISCREPANCIA:` y actualizar este reporte.
- Si cambia el contrato (ruta, campo, status o mensaje), actualizar el test y
  este reporte.
