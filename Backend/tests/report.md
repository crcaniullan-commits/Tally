# Reporte de tests de integración

Fecha: 2026-11-14

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

Última corrida completa: **OK**, 12 flujos y 393 subtests directos (~23 s).
`gofmt -l tests/integration/` y `go vet ./tests/integration/...` sin hallazgos.

| Flujo | Archivo | Resultado | Subtests |
|---|---|---|---|
| Auth | `integration/auth_flow_test.go` | PASS | 36 |
| Recuperación de contraseña | `integration/password_recovery_flow_test.go` | PASS | 12 |
| Claves de acceso | `integration/accesskeys_flow_test.go` | PASS | 20 |
| Empresas | `integration/empresas_flow_test.go` | PASS | 27 |
| Categorías | `integration/categories_flow_test.go` | PASS | 35 |
| Ingresos | `integration/incomes_flow_test.go` | PASS | 54 |
| Gastos | `integration/expenses_flow_test.go` | PASS | 56 |
| Deudores | `integration/debtors_flow_test.go` | PASS | 45 |
| Metas | `integration/goals_flow_test.go` | PASS | 40 |
| Usuarios | `integration/users_flow_test.go` | PASS | 28 |
| Archivos Excel | `integration/excel_flow_test.go` | PASS | 20 |
| Cache | `integration/cache_flow_test.go` | PASS | 5 |
| Aislamiento multi-empresa | `integration/tenant_isolation_test.go` | PASS | 23 |

Todos los flujos verifican al menos una ruta protegida sin token y con token
inválido (`401 no autorizado`).

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
- **Empresas:** creación, listado, renombrado, borrado, doble borrado, listado
  vacío, validaciones, rol municipal y aislamiento entre usuarios.
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
  RUT, validaciones, borrado de cuenta e invalidación del token.
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

Los tests pasan assertando el comportamiento actual. No se corrigió código de
producción en los flujos que no son parte de un cambio pedido. Cada caso está
marcado con `BUG:` o `DISCREPANCIA:` en el test.

| # | Ruta | Esperado | Actual | Causa |
|---|---|---|---|---|
| 1 | `GET /v1/aplicacion/usuarios/municipal/{rut}` | `created_at` y `updated_at` reales | `0001-01-01T00:00:00Z` | `ObtenerPorRUT` no selecciona esas columnas. |

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

## Diferencias entre swagger y código

| # | Punto | Swagger | Código |
|---|---|---|---|
| 1 | `POST /aplicacion/empresas` | `200` | `201` |
| 2 | `GET /aplicacion/empresas` sin empresas | `404` | `200 {"datos":null}` |
| 3 | `PATCH /aplicacion/empresas/{empresaID}/{nombre}` | Documenta `400`, sin límite explícito | Acepta nombres de 31+ caracteres; la creación valida `max=30` |
| 4 | `POST /aplicacion/usuarios/exchange/{code}` | Ya no documentado | No montado (`404`); el canje real es `POST /v1/autenticacion` |
| 5 | `model.RUT` | `{"cuerpo":..., "dv":"..."}` | Serializa `{"Cuerpo":..., "DV":"..."}` por falta de tags JSON |
| 6 | `GET` de listados (`categorias`, `metas`, `deudas`, etc.) | Solo documentan `200/500` | El middleware puede devolver `400` (UUID de empresa inválido), `401` y `403` |

## Casos omitidos

- Expiración/refresh de JWT: no hay endpoint de refresh y esperar la expiración
  haría el test lento; solo se prueba token inválido.
- Envío real de correo SMTP: los tests usan un remitente que solo guarda el
  código enviado (`testutil.UltimoCodigoDeRecuperacion`).
- Vencimiento del código de recuperación (15 minutos): no se prueba esperando el
  reloj; se cubre la unidad de tiempo del código en el service.
- `informe`: es un esqueleto sin rutas ni lógica.
- Storage de archivos en S3: el suite usa `storage.NewDisk` sobre un directorio temporal; el cliente S3 no existe todavía.
- Subida de un archivo de más de 10 MB: el tope lo aplica `http.MaxBytesReader` y no se prueba con un cuerpo de 10 MB.
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
