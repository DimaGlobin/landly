#!/bin/sh
set -e

echo "=== Landly Backend Starting ==="

# Ждём готовности PostgreSQL
echo "Waiting for PostgreSQL..."
max_tries=30
try=0
while [ $try -lt $max_tries ]; do
    if nc -z postgres 5432 2>/dev/null; then
        echo "PostgreSQL is ready!"
        break
    fi
    try=$((try + 1))
    echo "Attempt $try/$max_tries: PostgreSQL not ready yet, waiting..."
    sleep 2
done

if [ $try -eq $max_tries ]; then
    echo "ERROR: PostgreSQL did not become ready in time"
    exit 1
fi

# Применяем миграции
echo "Running database migrations..."
cd /app
./goose -dir ./migrations postgres "postgres://landly:landly@postgres:5432/landly?sslmode=disable" up

if [ $? -eq 0 ]; then
    echo "Migrations applied successfully!"
else
    echo "ERROR: Failed to apply migrations"
    exit 1
fi

# Запускаем приложение
echo "Starting application..."
exec "$@"

