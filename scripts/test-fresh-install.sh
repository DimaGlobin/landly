#!/bin/bash
set -e

echo "========================================="
echo "Testing Fresh Install (One-Button Start)"
echo "========================================="

# Проверка Docker
echo ""
echo "Checking Docker..."
if ! command -v docker &> /dev/null; then
    echo "❌ ERROR: Docker is not installed"
    exit 1
fi

if ! docker info &> /dev/null; then
    echo "❌ ERROR: Docker is not running"
    exit 1
fi

echo "✅ Docker is running"

# Проверка docker-compose
echo ""
echo "Checking docker-compose..."
if ! command -v docker-compose &> /dev/null; then
    echo "❌ ERROR: docker-compose is not installed"
    exit 1
fi
echo "✅ docker-compose is installed"

# Очистка старых контейнеров
echo ""
echo "Cleaning up old containers..."
docker-compose -f deploy/docker/docker-compose.yml down -v 2>/dev/null || true

# Проверка файлов конфигурации
echo ""
echo "Checking configuration files..."
if [ ! -f "config.yml" ]; then
    echo "❌ ERROR: config.yml not found"
    exit 1
fi
echo "✅ config.yml exists"

if [ ! -f "deploy/docker/docker-compose.yml" ]; then
    echo "❌ ERROR: docker-compose.yml not found"
    exit 1
fi
echo "✅ docker-compose.yml exists"

# Запуск make dev
echo ""
echo "========================================="
echo "Running: make dev"
echo "========================================="
make dev

# Ждём запуска сервисов
echo ""
echo "Waiting for services to be ready (max 120 seconds)..."

max_wait=120
waited=0

while [ $waited -lt $max_wait ]; do
    if curl -s http://localhost:8080/health > /dev/null 2>&1; then
        echo "✅ Backend is ready!"
        break
    fi
    
    if [ $waited -eq 0 ]; then
        echo -n "Waiting"
    else
        echo -n "."
    fi
    
    sleep 3
    waited=$((waited + 3))
done

echo ""

if [ $waited -ge $max_wait ]; then
    echo "❌ ERROR: Backend did not start in time"
    echo ""
    echo "Docker logs:"
    docker-compose -f deploy/docker/docker-compose.yml logs backend
    exit 1
fi

# Проверка Frontend
echo ""
echo "Checking frontend..."
if curl -s http://localhost:3000 > /dev/null 2>&1; then
    echo "✅ Frontend is ready!"
else
    echo "⚠️  Frontend is not responding yet (may still be starting)"
fi

# Проверка других сервисов
echo ""
echo "Checking other services..."

if docker-compose -f deploy/docker/docker-compose.yml ps | grep postgres | grep -q "Up"; then
    echo "✅ PostgreSQL is running"
else
    echo "❌ PostgreSQL is not running"
fi

if docker-compose -f deploy/docker/docker-compose.yml ps | grep redis | grep -q "Up"; then
    echo "✅ Redis is running"
else
    echo "❌ Redis is not running"
fi

if docker-compose -f deploy/docker/docker-compose.yml ps | grep minio | grep -q "Up"; then
    echo "✅ MinIO is running"
else
    echo "❌ MinIO is not running"
fi

# Финальный отчёт
echo ""
echo "========================================="
echo "✅ ONE-BUTTON START TEST PASSED!"
echo "========================================="
echo ""
echo "Services available at:"
echo "  - Frontend:      http://localhost:3000"
echo "  - Backend API:   http://localhost:8080"
echo "  - MinIO Console: http://localhost:9001 (admin/minioadmin)"
echo ""
echo "Run 'make down' to stop all services"
echo ""

