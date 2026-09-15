#!/bin/sh
set -e

echo "Aplicando migraciones de la base de datos..."
/app/migrate -path=/app/migrations -database="$DB_ADDR" up

echo "Migraciones aplicadas. Iniciando servidor..."
exec /app/server