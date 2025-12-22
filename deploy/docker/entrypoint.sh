#!/bin/sh
set -e

echo "=== Landly Backend Starting ==="

# Опциональное ожидание PostgreSQL (для docker-compose)
if [ "${WAIT_FOR_DB:-0}" = "1" ]; then
    DB_HOST="${LANDLY_DATABASE_POSTGRES_HOST:-postgres}"
    DB_PORT="${LANDLY_DATABASE_POSTGRES_PORT:-5432}"
    
    if [ -z "$DB_HOST" ] || [ -z "$DB_PORT" ]; then
        echo "ERROR: WAIT_FOR_DB=1 requires LANDLY_DATABASE_POSTGRES_HOST and LANDLY_DATABASE_POSTGRES_PORT"
        exit 1
    fi
    
    echo "Waiting for PostgreSQL at $DB_HOST:$DB_PORT..."
    max_tries=30
    try=0
    while [ $try -lt $max_tries ]; do
        if nc -z "$DB_HOST" "$DB_PORT" 2>/dev/null; then
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
else
    echo "Skipping PostgreSQL wait (WAIT_FOR_DB not set)"
fi

# Опциональное применение миграций
if [ "${RUN_MIGRATIONS:-0}" = "1" ]; then
    DB_HOST="${LANDLY_DATABASE_POSTGRES_HOST:-postgres}"
    DB_PORT="${LANDLY_DATABASE_POSTGRES_PORT:-5432}"
    DB_USER="${LANDLY_DATABASE_POSTGRES_USER:-landly}"
    DB_PASSWORD="${LANDLY_DATABASE_POSTGRES_PASSWORD:-landly}"
    DB_NAME="${LANDLY_DATABASE_POSTGRES_DBNAME:-landly}"
    DB_SSLMODE="${LANDLY_DATABASE_POSTGRES_SSLMODE:-disable}"
    
    if [ -z "$DB_HOST" ] || [ -z "$DB_PORT" ] || [ -z "$DB_USER" ] || [ -z "$DB_PASSWORD" ] || [ -z "$DB_NAME" ]; then
        echo "ERROR: RUN_MIGRATIONS=1 requires database connection env vars"
        exit 1
    fi
    
    echo "Running database migrations..."
    cd /app
    DSN="postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSLMODE}"
    ./goose -dir ./migrations postgres "$DSN" up
    
    if [ $? -eq 0 ]; then
        echo "Migrations applied successfully!"
    else
        echo "ERROR: Failed to apply migrations"
        exit 1
    fi
else
    echo "Skipping migrations (RUN_MIGRATIONS not set)"
fi

# Запускаем приложение
echo "Starting application..."
exec "$@"

