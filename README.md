# Landly

SaaS-платформа для генерации landing pages с помощью AI.

## Быстрый старт

**Требуется только Docker!**

```bash
# Клонировать репозиторий
git clone https://github.com/DimaGlobin/landly.git
cd landly

# Запустить все сервисы одной командой
make run
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
├── docs/             # Документация (см. ниже)
└── config.yml        # Конфигурация приложения
```

## Документация

| Документ | Описание |
|----------|----------|
| [docs/API_ENDPOINTS.md](docs/API_ENDPOINTS.md) | Описание всех API endpoints с примерами |
| [docs/openapi.yaml](docs/openapi.yaml) | OpenAPI 3.0 спецификация |
| [docs/ENV_VARIABLES.md](docs/ENV_VARIABLES.md) | Полный справочник переменных окружения |
| [docs/MONITORING.md](docs/MONITORING.md) | Логирование, метрики, Prometheus, Grafana |
| [docs/PUBLISHING_ARCHITECTURE.md](docs/PUBLISHING_ARCHITECTURE.md) | Архитектура публикации сайтов и CDN |
| [apps/backend/migrations/README.md](apps/backend/migrations/README.md) | Работа с миграциями БД |

## Конфигурация

### Локальная разработка

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

### Production / Serverless Deployment

В production образ **не содержит config.yml** (по соображениям безопасности).  
Вся конфигурация должна передаваться через environment variables.

**Обязательные переменные окружения:**

| Variable | Description | Example |
|----------|-------------|---------|
| `LANDLY_AUTH_JWT_SECRET` | JWT secret (min 32 chars) | `super-secret-key-at-least-32-characters` |
| `LANDLY_DATABASE_POSTGRES_HOST` | PostgreSQL host | `rc1a-xxx.mdb.yandexcloud.net` |
| `LANDLY_DATABASE_POSTGRES_PASSWORD` | PostgreSQL password | `your-db-password` |
| `LANDLY_STORAGE_S3_BUCKET` | S3 bucket name | `landly-sites` |
| `LANDLY_STORAGE_S3_ENDPOINT` | S3 endpoint | `storage.yandexcloud.net` |
| `LANDLY_STORAGE_S3_ACCESS_KEY` | S3 access key | `YCAxxxxx` |
| `LANDLY_STORAGE_S3_SECRET_KEY` | S3 secret key | `YCPxxxxx` |

**Опциональные переменные:**

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | HTTP listen port (set by Yandex Serverless) |
| `LANDLY_APP_ENV` | `production` | Environment name |
| `LANDLY_DATABASE_POSTGRES_PORT` | `5432` | PostgreSQL port |
| `LANDLY_DATABASE_POSTGRES_USER` | `landly` | PostgreSQL user |
| `LANDLY_DATABASE_POSTGRES_DBNAME` | `landly` | PostgreSQL database |
| `LANDLY_DATABASE_POSTGRES_SSLMODE` | `require` | SSL mode |
| `LANDLY_STORAGE_S3_USE_SSL` | `true` | Use HTTPS for S3 |
| `LANDLY_AI_PROVIDER` | `mock` | AI provider (mock/openai/anthropic) |
| `LANDLY_BOOTSTRAP_MODE` | `0` | Skip required validation (for health checks) |

**Порядок приоритета конфигурации** (от низшего к высшему):
1. Встроенные дефолты
2. `config.yml` (если есть)
3. `config.local.yml` (если есть)
4. Environment variables (`LANDLY_*`)
5. `PORT` env var (наивысший приоритет для listen address)

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
make run              # Запустить тесты и приложение в docker
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

## CI/CD

### GitHub Environments

Проект использует GitHub Environment **"staging"** для деплоя в Yandex Cloud Serverless Containers.

### Backend Workflow

При push в `main` (изменения в `apps/backend/**`):
1. **Unit tests** — запускаются тесты backend
2. **Integration tests** — тесты с PostgreSQL, Redis, MinIO
3. **Build** — сборка Docker образа и push в GHCR + Yandex Container Registry
4. **Deploy** — деплой новой revision в Serverless Container с immutable image digest
5. **Smoke test** — проверка `/health` и `/readyz` endpoints

### Frontend Workflow

При push в `main` (изменения в `apps/frontend/**`):
1. **Unit tests** — lint + Jest tests
2. **Build** — сборка Docker образа с `NEXT_PUBLIC_API_URL` и push в GHCR + YCR
3. **Deploy** — деплой новой revision в Serverless Container с immutable image digest
4. **Smoke test** — проверка homepage (HTTP 200)

### Required Secrets (GitHub Environment: staging)

#### Yandex Cloud Infrastructure (Shared)

| Secret | Required | Description | Example |
|--------|:--------:|-------------|---------|
| `YC_FOLDER_ID` | ✅ | Yandex Cloud folder ID | `b1gxxxxxxxxx` |
| `YC_REGISTRY_ID` | ✅ | Container Registry ID | `crpxxxxxxxxx` |
| `YC_SERVICE_ACCOUNT_KEY_JSON` | ✅ | Service account key (JSON) | `{"id": "...", ...}` |
| `YC_SERVICE_ACCOUNT_ID` | ✅ | Service account ID | `ajexxxxxxxxx` |

#### Backend Secrets

| Secret | Required | Description | Example |
|--------|:--------:|-------------|---------|
| `YCR_REPOSITORY` | ✅ | Repository name in YCR | `landly-backend` |
| `YC_CONTAINER_ID` | ✅ | Serverless Container ID | `bbsxxxxxxxxx` |
| `STAGING_URL` | ✅ | Public URL of staging backend | `https://bbsxxxxxxxxx.containers.yandexcloud.net` |

#### Frontend Secrets

| Secret | Required | Description | Example |
|--------|:--------:|-------------|---------|
| `YCR_FRONTEND_REPOSITORY` | ✅ | Repository name in YCR | `landly-frontend` |
| `YC_FRONTEND_CONTAINER_ID` | ✅ | Serverless Container ID | `bbsyyyyyyyyy` |
| `STAGING_FRONTEND_URL` | ✅ | Public URL of staging frontend | `https://bbsyyyyyyyyy.containers.yandexcloud.net` |
| `NEXT_PUBLIC_API_URL` | ✅ | Backend API URL (build-time) | `https://api.staging.landly.ru` |

#### Backend Application Configuration (LANDLY_*)

| Secret | Description | Example |
|--------|-------------|---------|
| `LANDLY_AUTH_JWT_SECRET` | JWT secret (min 32 chars) | `your-super-secret-production-key` |
| `LANDLY_DATABASE_POSTGRES_HOST` | PostgreSQL host | `rc1a-xxx.mdb.yandexcloud.net` |
| `LANDLY_DATABASE_POSTGRES_PORT` | PostgreSQL port | `6432` |
| `LANDLY_DATABASE_POSTGRES_USER` | PostgreSQL user | `landly` |
| `LANDLY_DATABASE_POSTGRES_PASSWORD` | PostgreSQL password | `your-db-password` |
| `LANDLY_DATABASE_POSTGRES_DBNAME` | PostgreSQL database name | `landly` |
| `LANDLY_DATABASE_POSTGRES_SSLMODE` | PostgreSQL SSL mode | `require` |
| `LANDLY_STORAGE_S3_ENDPOINT` | S3 endpoint | `storage.yandexcloud.net` |
| `LANDLY_STORAGE_S3_BUCKET` | S3 bucket name | `landly-sites` |
| `LANDLY_STORAGE_S3_ACCESS_KEY` | S3 access key | `YCAJExxxxxxxx` |
| `LANDLY_STORAGE_S3_SECRET_KEY` | S3 secret key | `YCPxxxxxxxx` |
| `LANDLY_SERVER_CORS_ALLOWED_ORIGINS` | CORS allowed origins | `https://app.landly.ru` |

### How Secrets are Used

1. **Build job** — uses YC_* secrets to push image to Yandex Container Registry
2. **Deploy job** — passes all LANDLY_* secrets as environment variables to Serverless Container revision via `--environment` flags
3. **Smoke test** — calls STAGING_URL to verify deployment

### Adding a New Secret

1. Go to GitHub repo → Settings → Environments → staging
2. Add the secret with exact name from table above
3. Secrets are automatically picked up on next deploy

### Manual Deploy

If you need to trigger deploy manually:
```bash
gh workflow run backend-go.yml --ref main
```

## Архитектура

### Логирование и обработка ошибок

- **Структурированные логи**: JSON формат через Zap с trace_id для корреляции
- **Безопасные ответы**: клиенту отправляются только доменные ошибки, внутренние детали логируются
- **Централизованные helper-функции**: `RespondError()`, `RespondInternalError()`, `LogSQLError()`

### Публикация сайтов

- Статические сайты рендерятся и загружаются в S3
- Поддержка CDN для production (опционально)
- Атомарное переключение версий через release ID
- Инкрементальная загрузка с manifest-файлами

Подробнее: [docs/PUBLISHING_ARCHITECTURE.md](docs/PUBLISHING_ARCHITECTURE.md)

## Технологии

**Backend:**
- Go 1.23
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
