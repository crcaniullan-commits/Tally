# Notificaciones push en Tally: cómo funcionan y cómo implementarlas

*Guía para entender el tema desde cero y dejar listos los dos jobs: recordatorio diario (solo si el usuario no registró movimientos ese día) y aviso de cobros vencidos (todos los días).*

## 1. Cómo funciona una notificación push

Tu backend no puede "llamar" a un teléfono: el teléfono está detrás de la red móvil y no tiene una dirección a la que conectarte. Lo que sí tiene es **una única conexión permanente con el servicio de push de su sistema operativo**: Google (FCM) en Android y Apple (APNs) en iOS. Tu backend le entrega el mensaje a ese servicio, y el servicio lo hace llegar al teléfono aunque la app esté cerrada.

```
(A) Registro: una vez por dispositivo
    App ---- token ------------------------> Backend (lo guarda en la BD)

(B) Envío: cada vez que hay algo que avisar
    Job del backend ---- mensaje + token --> Servicio de push (FCM)
    Servicio de push ---- entrega ---------> Teléfono (app abierta o cerrada)
```

### Conceptos que aparecen

- **Token de dispositivo:** un identificador que el servicio de push le da a esa instalación de la app. Es la "dirección" a la que apuntas al enviar. Puede cambiar (reinstalación, restauración), por eso la app debe reenviarlo al backend cada vez que cambie.
- **Servicio de push (FCM/APNs):** el intermediario. Solo él puede hablar con el teléfono. Tu backend le hace una petición HTTP con el token y el contenido.
- **Mensaje:** un título y un cuerpo que el sistema operativo muestra como notificación, más datos opcionales (por ejemplo, qué pantalla abrir al tocarla).
- **Job o tarea programada:** código que se ejecuta solo en un horario, sin que nadie lo llame. Aquí, un planificador dentro del backend.
- **Expresión cron:** la forma estándar de escribir un horario con cinco campos: `minuto hora día-del-mes mes día-de-la-semana`. Por ejemplo, `0 20 * * *` significa "minuto 0 de la hora 20, todos los días", es decir, las 20:00.
- **Idempotencia:** una operación es idempotente si ejecutarla dos veces deja el mismo resultado que ejecutarla una. Importa porque un job puede dispararse dos veces (reinicio del backend, dos instancias) y no queremos mandar dos notificaciones iguales.

## 2. Qué vamos a construir

| Pieza | Para qué sirve |
| --- | --- |
| Tabla `device_tokens` y endpoints `PUT`/`DELETE /v1/app/devices` | Saber a qué dispositivos de cada usuario enviar |
| Paquete `notify` con la interfaz `Notifier` | Enviar sin depender de Firebase todavía (mock por consola, como el correo) |
| Tabla `notification_log` | Evitar notificaciones duplicadas |
| Paquete `jobs` con dos jobs | Decidir a quién avisar y qué decirle |
| Planificador (cron) en `app.go` | Disparar los jobs a su hora |

## 3. Paso a paso

### Paso 1: migraciones

```sql
CREATE TABLE device_tokens (
    id           UUID PRIMARY KEY,
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token        TEXT NOT NULL UNIQUE,
    platform     TEXT NOT NULL,
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_device_tokens_user ON device_tokens (user_id);

CREATE TABLE notification_log (
    user_id  UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    kind     TEXT NOT NULL,
    ref_date DATE NOT NULL,
    sent_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, kind, ref_date)
);
```

Down:

```sql
DROP TABLE IF EXISTS notification_log;
DROP TABLE IF EXISTS device_tokens;
```

**Por qué así:**

- `token` es `UNIQUE`: un mismo token solo puede pertenecer a un usuario. Si alguien cierra sesión en un teléfono y otro inicia sesión en él, el token cambia de dueño.
- `last_seen_at` guarda cuándo la app lo envió por última vez. Firebase lo recomienda: <https://firebase.google.com/docs/cloud-messaging/manage-tokens> explica que conviene guardar una marca de tiempo en cada subida para detectar dispositivos inactivos.
- La clave primaria de `notification_log` es `(usuario, tipo, fecha)`. Eso significa que solo puede existir **una** fila por usuario, tipo de aviso y día. La usaremos como candado (ver paso 4).

### Paso 2: el paquete `notify`

La interfaz que usarán los jobs. Detrás puede haber un mock o Firebase; los jobs no se enteran.

```go
package notify

import (
    "context"

    "github.com/google/uuid"
    "go.uber.org/zap"
)

type Notification struct {
    Title string
    Body  string
    Data  map[string]string // para que la app sepa qué pantalla abrir
}

type Notifier interface {
    Send(ctx context.Context, userID uuid.UUID, n Notification) error
}

// LogNotifier es el mock: imprime en vez de enviar.
type LogNotifier struct{ Log *zap.SugaredLogger }

func (l *LogNotifier) Send(ctx context.Context, userID uuid.UUID, n Notification) error {
    l.Log.Infow("push (mock)", "user", userID, "title", n.Title, "body", n.Body)
    return nil
}
```

### Paso 3: registrar dispositivos

Store:

```go
func (s *Store) Upsert(ctx context.Context, userID uuid.UUID, token, platform string) error {
    _, err := dbtx.FromContext(ctx, s.db).ExecContext(ctx,
        `INSERT INTO device_tokens (id, user_id, token, platform, last_seen_at)
         VALUES ($1, $2, $3, $4, now())
         ON CONFLICT (token) DO UPDATE
         SET user_id = EXCLUDED.user_id,
             platform = EXCLUDED.platform,
             last_seen_at = now()`,
        uuid.New(), userID, token, platform)
    return err
}

func (s *Store) Delete(ctx context.Context, userID uuid.UUID, token string) error {
    _, err := dbtx.FromContext(ctx, s.db).ExecContext(ctx,
        `DELETE FROM device_tokens WHERE token = $1 AND user_id = $2`, token, userID)
    return err
}
```

Handler (adapta la lectura del JSON y las respuestas a tus helpers):

```go
type registerReq struct {
    Token    string `json:"token"    validate:"required"`
    Platform string `json:"platform" validate:"required,oneof=android ios"`
}
// PUT    /v1/app/devices  -> Upsert(ctx, ctxUserID(ctx), req.Token, req.Platform)  -> 204
// DELETE /v1/app/devices  -> Delete(ctx, ctxUserID(ctx), req.Token)                -> 204
```

Van en el grupo que pasa por el middleware de autenticación, pero **no** por el de empresa: un dispositivo pertenece al usuario, no a una empresa.

`ON CONFLICT (token) DO UPDATE` hace que la misma ruta sirva para registrar un token nuevo y para refrescar uno existente. Eso es lo que permite que la app lo llame en cada arranque.

### Paso 4: las consultas y los jobs

**Quién recibe el recordatorio diario:** usuarios con al menos un dispositivo que hoy no hayan registrado ningún ingreso ni egreso. Supuesto: "registrar" significa *cuándo se creó el registro* (`created_at`), no la fecha del movimiento, porque alguien puede cargar hoy una venta de ayer.

```sql
SELECT DISTINCT dt.user_id
FROM device_tokens dt
WHERE NOT EXISTS (
    SELECT 1 FROM incomes i
    JOIN empresas e ON e.id = i.empresa_id
    WHERE e.user_id = dt.user_id
      AND i.created_at >= $1 AND i.created_at < $2
)
AND NOT EXISTS (
    SELECT 1 FROM expenses x
    JOIN empresas e ON e.id = x.empresa_id
    WHERE e.user_id = dt.user_id
      AND x.created_at >= $1 AND x.created_at < $2
);
```

`$1` y `$2` son el inicio de hoy y el inicio de mañana **en hora de Chile**, calculados en Go. Se compara con un rango y no con `created_at::date` para que la base pueda usar un índice. Si quieres notificar solo a usuarios con plan vigente, agrega `AND EXISTS (SELECT 1 FROM users u WHERE u.id = dt.user_id AND u.plan_expires_at > now())`.

**Quién recibe el aviso de cobros vencidos:** usuarios con deudores vencidos y sin pagar, resumidos en una fila por usuario. Los nombres de columna (`vence_el`, `pagado_at`, `monto`) son ejemplos; usa los de tu tabla `debtors`.

```sql
SELECT e.user_id, COUNT(*) AS cantidad, SUM(d.monto) AS total
FROM debtors d
JOIN empresas e ON e.id = d.empresa_id
WHERE d.vence_el < $1::date
  AND d.pagado_at IS NULL
GROUP BY e.user_id;
```

`$1` es la fecha de hoy en Chile como texto (`2026-11-15`). Se manda **un solo** aviso por usuario con el resumen, y no uno por deudor, para no llenar el teléfono.

**El candado anti-duplicados.** Antes de enviar, el job intenta insertar una fila en `notification_log`. Si la inserta, es el primero y envía; si ya existía, otro ya lo envió y se salta.

```go
func (s *Store) Claim(ctx context.Context, userID uuid.UUID, kind, day string) (bool, error) {
    res, err := s.db.ExecContext(ctx,
        `INSERT INTO notification_log (user_id, kind, ref_date)
         VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`, userID, kind, day)
    if err != nil {
        return false, err
    }
    n, err := res.RowsAffected()
    return n == 1, err
}

func (s *Store) Unclaim(ctx context.Context, userID uuid.UUID, kind, day string) error {
    _, err := s.db.ExecContext(ctx,
        `DELETE FROM notification_log WHERE user_id = $1 AND kind = $2 AND ref_date = $3`,
        userID, kind, day)
    return err
}
```

El orden es **reclamar, enviar, y si el envío falla, devolver el reclamo** para que se reintente. Así, si el backend se reinicia a las 20:00 y el job corre dos veces, el usuario recibe un solo aviso. Se pasa `day` como texto (`2026-11-15`) para evitar que el cambio de zona horaria mueva la fecha.

Los jobs:

```go
package jobs

type Jobs struct {
    store    *Store
    notifier notify.Notifier
    loc      *time.Location
    log      *zap.SugaredLogger
}

func (j *Jobs) DailyReminder() {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
    defer cancel()

    now := time.Now().In(j.loc)
    start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, j.loc)
    end := start.AddDate(0, 0, 1)
    day := start.Format("2006-01-02")

    users, err := j.store.UsersWithoutMovements(ctx, start, end)
    if err != nil {
        j.log.Errorw("daily reminder: consulta", "err", err)
        return
    }
    for _, id := range users {
        ok, err := j.store.Claim(ctx, id, "daily", day)
        if err != nil || !ok {
            continue
        }
        err = j.notifier.Send(ctx, id, notify.Notification{
            Title: "¿Cómo va tu día?",
            Body:  "Aún no registras movimientos hoy. Anótalos para no olvidarlos.",
        })
        if err != nil {
            _ = j.store.Unclaim(ctx, id, "daily", day)
            j.log.Errorw("daily reminder: envío", "user", id, "err", err)
        }
    }
}

func (j *Jobs) OverdueDebts() {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
    defer cancel()

    day := time.Now().In(j.loc).Format("2006-01-02")
    rows, err := j.store.OverdueByUser(ctx, day) // []Overdue{UserID, Cantidad, Total}
    if err != nil {
        j.log.Errorw("overdue: consulta", "err", err)
        return
    }
    for _, r := range rows {
        ok, err := j.store.Claim(ctx, r.UserID, "overdue", day)
        if err != nil || !ok {
            continue
        }
        err = j.notifier.Send(ctx, r.UserID, notify.Notification{
            Title: "Cobros vencidos",
            Body:  fmt.Sprintf("Tienes %d cobros vencidos por %s.", r.Cantidad, clp(r.Total)),
        })
        if err != nil {
            _ = j.store.Unclaim(ctx, r.UserID, "overdue", day)
            j.log.Errorw("overdue: envío", "user", r.UserID, "err", err)
        }
    }
}

// clp formatea pesos chilenos: 45000 -> $45.000
func clp(n int64) string {
    s := strconv.FormatInt(n, 10)
    for i := len(s) - 3; i > 0; i -= 3 {
        s = s[:i] + "." + s[i:]
    }
    return "$" + s
}
```

Si prefieres que el aviso de cobros sea semanal en vez de diario, cambia `day` por la fecha del lunes de esa semana: el candado hará el resto.

### Paso 5: el planificador

En `go.mod`: `go get github.com/robfig/cron/v3`. En `app.go`, después de registrar los módulos:

```go
loc, err := time.LoadLocation("America/Santiago")
if err != nil {
    return err
}
j := jobs.New(jobsStore, notifier, loc, logger)

c := cron.New(cron.WithLocation(loc), cron.WithChain(cron.Recover(cron.DefaultLogger)))
if _, err := c.AddFunc(env.DailyCron, j.DailyReminder); err != nil { // por defecto "0 20 * * *"
    return err
}
if _, err := c.AddFunc(env.OverdueCron, j.OverdueDebts); err != nil { // por defecto "0 9 * * *"
    return err
}
c.Start()
// al apagar el servidor: <-c.Stop().Done()
```

Las horas (20:00 y 9:00) son ejemplos; déjalas como variables de entorno.

Tres detalles que se escapan:

- **Zona horaria y Alpine.** Tu imagen `runtime` solo instala `ca-certificates`, y `LoadLocation` falla si no encuentra la base de zonas horarias. Agrega `import _ "time/tzdata"` en `main.go`, que la incrusta en el binario.
- **`cron.Recover`:** sin él, un error grave dentro de un job tumbaría todo el backend.
- **Hora de la base de datos:** la base trabaja en UTC. Por eso el job calcula el inicio y el fin del día en hora de Chile y se los pasa a las consultas, en lugar de usar `CURRENT_DATE`.

### Paso 6: probar sin app ni Firebase

1. Inserta a mano un token de prueba para un usuario real: `INSERT INTO device_tokens (id, user_id, token, platform) VALUES (gen_random_uuid(), '<userID>', 'token-de-prueba', 'android');`
2. Levanta el backend con `DAILY_CRON="* * * * *"` (cada minuto) y mira los logs: debe aparecer `push (mock)`.
3. Espera otro minuto: no debe aparecer de nuevo, porque el candado lo impide.
4. Registra un ingreso hoy, borra la fila de `notification_log` y espera: el usuario ya no debe recibir el aviso.
5. Para los tests unitarios, mockea el store y el notifier, como en tus otros módulos.

### Paso 7: cambiar al envío real (cuando elijan framework)

El `Notifier` real busca los tokens del usuario, envía uno por uno y limpia los inválidos. Con Firebase Cloud Messaging:

1. Crear un proyecto en Firebase y descargar la cuenta de servicio (un JSON con credenciales). Va fuera del repositorio y se pasa por variable de entorno.
2. Usar el SDK de administrador de Firebase para Go (`firebase.google.com/go/v4`, módulo `messaging`). No verifiqué ahora sus nombres de funciones; revisa su documentación.
3. Si FCM responde que un token es inválido, borrarlo de la tabla. Según su documentación, <https://firebase.google.com/docs/cloud-messaging/manage-tokens>, los errores `UNREGISTERED` (404) e `INVALID_ARGUMENT` (400) indican tokens inválidos o expirados.
4. Un job mensual que borre los tokens sin actividad: FCM considera obsoleto un token cuyo dispositivo no se conecta hace un mes, y en Android caduca a los 270 días. El umbral lo defines tú; por ejemplo, `DELETE FROM device_tokens WHERE last_seen_at < now() - interval '60 days'`.
5. En iOS, según recuerdo, hay que subir una clave de APNs al proyecto de Firebase. Confírmalo en su documentación.

En el cambio solo se reemplaza `LogNotifier` por el nuevo en `app.go`; los jobs no se tocan.

### Lo que debe hacer el frontend

Se lo puedes pasar al equipo:

- Pedir permiso de notificaciones (en Android 13 o superior es un permiso en tiempo de ejecución).
- Obtener el token y mandarlo a `PUT /v1/app/devices` en cada arranque y cada vez que cambie.
- Llamar a `DELETE /v1/app/devices` al cerrar sesión.
- Decidir qué pantalla abrir según `Data` cuando el usuario toque la notificación.

## 4. Cosas que conviene entender

- **Por qué el candado y no solo el horario.** Un planificador dentro del backend dispara el job una vez por proceso. Si hay dos instancias, o el backend se reinicia justo a esa hora, correría dos veces. La clave primaria de `notification_log` hace que el resultado sea el mismo.
- **Una notificación por usuario, no por deudor.** Más de una al día suele terminar con el usuario desactivando todas las notificaciones de la app.
- **Qué pasa si falla el envío.** Se devuelve el candado, pero el siguiente intento es el día siguiente, porque el cron corre una vez al día. Si quieres reintentos el mismo día, hay que programar el job más veces; el candado evita que lleguen duplicados.
- **Render.** Un planificador interno solo corre mientras el servicio está despierto, y en el plan gratuito se apaga tras 15 minutos sin tráfico. Ahí tendrías que disparar los jobs desde un servicio externo que llame a un endpoint protegido. En tu docker-compose local no hay problema.

## 5. Supuestos y límites

- Nada de esto está compilado ni probado.
- Asumí que existen `users`, `empresas`, `incomes`, `expenses` y `debtors` con las columnas descritas; `ctxUserID`, `dbtx` y los helpers de respuesta son los tuyos.
- Las consultas dependen de las columnas reales de `debtors`; ajusta los nombres.
- Verificado en fuentes: el comportamiento de tokens y errores de FCM (documentación de Firebase). El resto es diseño propio y conocimiento general de Go, Postgres y cron.
