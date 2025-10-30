# Landly

SaaS-платформа для генерации landing pages с помощью AI.

## Быстрый старт

**Требуется только Docker!**

```bash
# Клонировать репозиторий
git clone https://github.com/DimaGlobin/landly.git
cd landly

# Запустить все сервисы одной командой
make dev
```

Всё! Приложение запустится автоматически. Подождите ~30-60 секунд пока соберутся образы и применятся миграции.

После запуска доступно:
- **Frontend:** http://localhost:3000
- **Backend API:** http://localhost:8080
- **MinIO Console:** http://localhost:9001 (login: minioadmin / minioadmin)

Проверить статус: `make logs`  
Остановить: `make down`

## Требования

- Docker и Docker Compose
- Go 1.23+ (для локальной разработки backend)
- Node.js 18+ (для локальной разработки frontend)

## Структура проекта

```
landly/
├── apps/
│   ├── backend/      # Go API сервер
│   └── frontend/     # Next.js приложение
├── deploy/
│   ├── docker/       # Docker Compose конфигурация
│   └── ci/           # GitHub Actions workflows
├── docs/
│   ├── API_ENDPOINTS.md  # API документация
│   └── openapi.yaml      # OpenAPI спецификация
└── config.yml        # Конфигурация приложения
```

## Конфигурация

Базовая конфигурация в `config.yml` (готова для docker-compose).

Для локальной разработки создайте `config.local.yml`:

```bash
cp config.example.yml config.local.yml
# Отредактируйте config.local.yml (хосты на localhost)
```

Environment variables с префиксом `LANDLY_` переопределяют любые настройки:

```bash
export LANDLY_AUTH_JWT_SECRET="your-secret"
export LANDLY_DATABASE_POSTGRES_HOST="localhost"
```

## Разработка

### Backend

```bash
cd apps/backend

# Запустить локально
go run cmd/api/main.go

# Тесты
go test ./...
make test-integration  # с docker-compose
```

### Frontend

```bash
cd apps/frontend

# Установить зависимости
npm install

# Запустить dev server
npm run dev

# Тесты
npm test
npm run test:e2e
```

## API

Документация: [docs/API_ENDPOINTS.md](docs/API_ENDPOINTS.md)  
OpenAPI: [docs/openapi.yaml](docs/openapi.yaml)

Основные endpoints:

- `POST /api/auth/signup` - регистрация
- `POST /api/auth/signin` - вход
- `POST /api/projects/simple` - создать проект
- `GET /api/projects/:id` - получить проект
- `POST /api/projects/:id/publish` - опубликовать

## Makefile команды

```bash
make dev              # Запустить все сервисы
make down             # Остановить сервисы
make logs             # Показать логи

make test             # Все тесты
make test-backend     # Backend unit tests
make test-frontend    # Frontend unit tests
make test-integration # Backend integration tests
make test-e2e         # Frontend E2E tests

make build            # Собрать все
make lint             # Запустить линтеры
```

## Технологии

**Backend:**
- Go 1.22
- PostgreSQL
- Redis
- MinIO (S3)

**Frontend:**
- Next.js 14
- React 18
- TypeScript
- Tailwind CSS

## Лицензия

MIT
