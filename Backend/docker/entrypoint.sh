#!/bin/sh
set -e

#las migraciones las aplica el admin por fuera (go tool task migration-up)-
#Solo se corren aquí si se defini MIGRATE_DB_ADDR. con un usuario que pueda hacer DDL.
if [ -n "$MIGRATE_DB_ADDR"]; then
    echo "Aplicando migraciones de la base de datos..."
    /app/migrate --path=/app/migrations -database="$MIGRATE_DB_ADDR" up
fi

echo "iniciando servidor..."
exec /app/server