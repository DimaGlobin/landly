#!/bin/bash

# Скрипт быстрой проверки проекта Landly
# Использование: ./scripts/verify.sh

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

echo "╔══════════════════════════════════════════════════════════════╗"
echo "║         Проверка проекта Landly                             ║"
echo "╚══════════════════════════════════════════════════════════════╝"
echo ""

# Backend
echo "🔍 Проверка Backend..."
cd "$PROJECT_DIR/apps/backend"

echo "  ✓ Go modules..."
go mod verify > /dev/null 2>&1 && echo "    ✅ go mod verify" || echo "    ❌ go mod verify"

echo "  ✓ Go vet..."
go vet ./... > /dev/null 2>&1 && echo "    ✅ go vet" || echo "    ❌ go vet"

echo "  ✓ Компиляция API..."
go build -o /tmp/landly-api ./cmd/api/main.go > /dev/null 2>&1 && echo "    ✅ API компилируется" || echo "    ❌ API не компилируется"

echo "  ✓ Компиляция Worker..."
go build -o /tmp/landly-worker ./cmd/worker/main.go > /dev/null 2>&1 && echo "    ✅ Worker компилируется" || echo "    ❌ Worker не компилируется"

# Frontend
echo ""
echo "🔍 Проверка Frontend..."
cd "$PROJECT_DIR/apps/frontend"

echo "  ✓ package.json..."
cat package.json | python3 -m json.tool > /dev/null 2>&1 && echo "    ✅ package.json валиден" || echo "    ❌ package.json невалиден"

echo "  ✓ tsconfig.json..."
cat tsconfig.json | python3 -m json.tool > /dev/null 2>&1 && echo "    ✅ tsconfig.json валиден" || echo "    ❌ tsconfig.json невалиден"

# Docker
echo ""
echo "🔍 Проверка Docker..."
cd "$PROJECT_DIR"

echo "  ✓ docker-compose..."
docker compose -f deploy/docker/docker-compose.yml config > /dev/null 2>&1 && echo "    ✅ docker-compose.yml валиден" || echo "    ❌ docker-compose.yml невалиден"

# Структура
echo ""
echo "🔍 Проверка структуры..."

REQUIRED_DIRS=(
  "apps/backend/internal/domain"
  "apps/backend/internal/usecase"
  "apps/backend/internal/interface"
  "apps/backend/internal/infrastructure"
  "apps/backend/cmd/api"
  "apps/backend/cmd/worker"
  "apps/frontend/app"
  "apps/frontend/components"
  "apps/frontend/lib"
  "deploy/docker"
  "docs"
)

for dir in "${REQUIRED_DIRS[@]}"; do
  if [ -d "$PROJECT_DIR/$dir" ]; then
    echo "  ✅ $dir"
  else
    echo "  ❌ $dir - отсутствует"
  fi
done

echo ""
echo "╔══════════════════════════════════════════════════════════════╗"
echo "║         ✅ Проверка завершена                               ║"
echo "╚══════════════════════════════════════════════════════════════╝"
echo ""
echo "Для запуска проекта используйте:"
echo "  make dev      # Запустить все сервисы"
echo "  make migrate  # Применить миграции"
echo "  make logs     # Посмотреть логи"
echo ""

