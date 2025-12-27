# Environment Variables Reference

Канонический список переменных окружения для Landly backend.

## Источники конфигурации

Приоритет (от низшего к высшему):
1. Встроенные дефолты (`setDefaults()` в `config/config.go`)
2. `config.yml` (если есть)
3. `config.local.yml` (если есть)
4. Environment variables (`LANDLY_*`)
5. `PORT` env var (наивысший приоритет для listen address)

---

## Required Variables

Без этих переменных приложение не запустится (если `LANDLY_BOOTSTRAP_MODE != 1`):

| Переменная | Описание | Пример |
|------------|----------|--------|
| `LANDLY_AUTH_JWT_SECRET` | JWT секрет (мин. 32 символа) | `your-super-secret-key-min-32-chars` |
| `LANDLY_DATABASE_POSTGRES_HOST` | Хост PostgreSQL | `rc1a-xxx.mdb.yandexcloud.net` |
| `LANDLY_DATABASE_POSTGRES_PASSWORD` | Пароль PostgreSQL | `your-db-password` |
| `LANDLY_DATABASE_POSTGRES_USER` | Пользователь PostgreSQL (default: `landly`) | `landly` |
| `LANDLY_DATABASE_POSTGRES_DBNAME` | Имя базы данных (default: `landly`) | `landly` |
| `LANDLY_STORAGE_S3_ENDPOINT` | S3 endpoint | `storage.yandexcloud.net` |
| `LANDLY_STORAGE_S3_ACCESS_KEY` | S3 access key | `YCAJExxxxxxxx` |
| `LANDLY_STORAGE_S3_SECRET_KEY` | S3 secret key | `YCPxxxxxxxx` |
| `LANDLY_STORAGE_S3_BUCKET` | S3 bucket name | `landly-sites` |

> **Note:** `LANDLY_DATABASE_POSTGRES_USER` и `LANDLY_DATABASE_POSTGRES_DBNAME` имеют defaults в коде, но для production их следует указывать явно.

---

## Optional Variables

### Application

| Переменная | Default | Описание |
|------------|---------|----------|
| `LANDLY_APP_ENV` | `production` | Окружение (`development` / `production`) |
| `LANDLY_APP_NAME` | `landly` | Имя приложения |
| `LANDLY_APP_VERSION` | `1.0.0` | Версия |
| `LANDLY_APP_BASE_URL` | `""` | Базовый URL для публичных ссылок |

### Server

| Переменная | Default | Описание |
|------------|---------|----------|
| `LANDLY_SERVER_HTTP_ADDR` | `:8080` | Адрес HTTP сервера |
| `LANDLY_SERVER_HTTP_READ_TIMEOUT` | `30s` | Таймаут чтения |
| `LANDLY_SERVER_HTTP_WRITE_TIMEOUT` | `30s` | Таймаут записи |
| `LANDLY_SERVER_CORS_ALLOWED_ORIGINS` | `*` | Разрешённые CORS origins |
| `LANDLY_SERVER_CORS_ALLOWED_METHODS` | `GET,POST,PUT,DELETE,OPTIONS` | Разрешённые методы |
| `LANDLY_SERVER_CORS_ALLOWED_HEADERS` | `Authorization,Content-Type` | Разрешённые заголовки |

### Auth / JWT

| Переменная | Default | Описание |
|------------|---------|----------|
| `LANDLY_AUTH_JWT_ACCESS_TOKEN_TTL` | `15m` | TTL access token |
| `LANDLY_AUTH_JWT_REFRESH_TOKEN_TTL` | `168h` | TTL refresh token (7 дней) |

### Database (PostgreSQL)

| Переменная | Default | Описание |
|------------|---------|----------|
| `LANDLY_DATABASE_POSTGRES_PORT` | `5432` | Порт |
| `LANDLY_DATABASE_POSTGRES_USER` | `landly` | Пользователь |
| `LANDLY_DATABASE_POSTGRES_DBNAME` | `landly` | Имя базы данных |
| `LANDLY_DATABASE_POSTGRES_SSLMODE` | `require` | SSL режим |
| `LANDLY_DATABASE_POSTGRES_MAX_OPEN_CONNS` | `25` | Макс. открытых соединений |
| `LANDLY_DATABASE_POSTGRES_MAX_IDLE_CONNS` | `5` | Макс. idle соединений |
| `LANDLY_DATABASE_POSTGRES_CONN_MAX_LIFETIME` | `5m` | Макс. время жизни соединения |

### Database (Redis)

| Переменная | Default | Описание |
|------------|---------|----------|
| `LANDLY_DATABASE_REDIS_ADDR` | `""` | Адрес Redis |
| `LANDLY_DATABASE_REDIS_PASSWORD` | `""` | Пароль |
| `LANDLY_DATABASE_REDIS_DB` | `0` | Номер базы |
| `LANDLY_DATABASE_REDIS_POOL_SIZE` | `10` | Размер пула |

### Storage (S3)

| Переменная | Default | Описание |
|------------|---------|----------|
| `LANDLY_STORAGE_S3_USE_SSL` | `true` | Использовать HTTPS |
| `LANDLY_STORAGE_S3_REGION` | `ru-central1` | Регион |

### Storage (CDN)

| Переменная | Default | Описание |
|------------|---------|----------|
| `LANDLY_STORAGE_CDN_BASE_URL` | `""` | Базовый URL CDN |
| `LANDLY_STORAGE_CDN_ENABLED` | `false` | Включить CDN |

### AI

| Переменная | Default | Описание |
|------------|---------|----------|
| `LANDLY_AI_PROVIDER` | `mock` | Провайдер (`mock` / `openai` / `anthropic`) |
| `LANDLY_AI_OPENAI_API_KEY` | `""` | OpenAI API key |
| `LANDLY_AI_OPENAI_MODEL` | `gpt-4` | Модель OpenAI |
| `LANDLY_AI_OPENAI_MAX_TOKENS` | `4000` | Макс. токенов |
| `LANDLY_AI_OPENAI_TEMPERATURE` | `0.7` | Temperature |
| `LANDLY_AI_ANTHROPIC_API_KEY` | `""` | Anthropic API key |
| `LANDLY_AI_ANTHROPIC_MODEL` | `claude-3-opus-20240229` | Модель Anthropic |
| `LANDLY_AI_ANTHROPIC_MAX_TOKENS` | `4000` | Макс. токенов |

### Render

| Переменная | Default | Описание |
|------------|---------|----------|
| `LANDLY_RENDER_TMP_DIR` | `/tmp/landly` | Временная директория |
| `LANDLY_RENDER_CLEANUP_AFTER` | `1h` | Время очистки |

### Logging

| Переменная | Default | Описание |
|------------|---------|----------|
| `LANDLY_LOGGING_LEVEL` | `info` | Уровень логирования |
| `LANDLY_LOGGING_FORMAT` | `json` | Формат (`json` / `console`) |
| `LANDLY_LOGGING_OUTPUT` | `stdout` | Вывод |

### Observability

| Переменная | Default | Описание |
|------------|---------|----------|
| `LANDLY_OBSERVABILITY_METRICS_ENABLED` | `false` | Включить метрики |
| `LANDLY_OBSERVABILITY_METRICS_PORT` | `9090` | Порт метрик |
| `LANDLY_OBSERVABILITY_TRACING_ENABLED` | `false` | Включить трейсинг |
| `LANDLY_OBSERVABILITY_TRACING_ENDPOINT` | `""` | Endpoint трейсинга |
| `LANDLY_OBSERVABILITY_TRACING_SAMPLE_RATE` | `0.1` | Sample rate |

---

## Frontend Variables

Переменные окружения для Next.js frontend (должны начинаться с `NEXT_PUBLIC_` для доступа в браузере):

| Переменная | Default | Описание |
|------------|---------|----------|
| `NEXT_PUBLIC_SITE_BASE_URL` | `https://landlify.ru` | Базовый URL для публичных ссылок на опубликованные сайты |

**Пример использования:**

```bash
# Production
NEXT_PUBLIC_SITE_BASE_URL=https://landlify.ru

# Staging
NEXT_PUBLIC_SITE_BASE_URL=https://staging.landlify.ru

# Local development
NEXT_PUBLIC_SITE_BASE_URL=http://localhost:8080
```

---

## Special Variables

Читаются напрямую через `os.Getenv()`, не через Viper:

| Переменная | Default | Описание |
|------------|---------|----------|
| `PORT` | — | Перезаписывает `LANDLY_SERVER_HTTP_ADDR`. Устанавливается Yandex Serverless автоматически. |
| `LANDLY_BOOTSTRAP_MODE` | `0` | Bootstrap режим (`1` / `true` / `yes`). Позволяет запуститься без required переменных для health checks. |

---

## GitHub Secrets (Environment: staging)

### Yandex Cloud Infrastructure (Shared)

| Secret | Required | Описание | Пример |
|--------|:--------:|----------|--------|
| `YC_FOLDER_ID` | ✅ | Yandex Cloud folder ID | `b1gxxxxxxxxx` |
| `YC_REGISTRY_ID` | ✅ | Container Registry ID | `crpxxxxxxxxx` |
| `YC_SERVICE_ACCOUNT_KEY_JSON` | ✅ | Service account key (JSON) | `{"id":"...","private_key":"..."}` |
| `YC_SERVICE_ACCOUNT_ID` | ✅ | Service account ID для контейнеров | `ajexxxxxxxxx` |

### Backend Secrets

| Secret | Required | Описание | Пример |
|--------|:--------:|----------|--------|
| `YCR_REPOSITORY` | ✅ | Имя репозитория backend в YCR | `landly-backend` |
| `YC_CONTAINER_ID` | ✅ | Serverless Container ID backend | `bbsxxxxxxxxx` |
| `STAGING_URL` | ✅ | Публичный URL staging backend | `https://bbsxxxxxxxxx.containers.yandexcloud.net` |

### Frontend Secrets

| Secret | Required | Описание | Пример |
|--------|:--------:|----------|--------|
| `YCR_FRONTEND_REPOSITORY` | ✅ | Имя репозитория frontend в YCR | `landly-frontend` |
| `YC_FRONTEND_CONTAINER_ID` | ✅ | Serverless Container ID frontend | `bbsyyyyyyyyy` |
| `STAGING_FRONTEND_URL` | ✅ | Публичный URL staging frontend | `https://bbsyyyyyyyyy.containers.yandexcloud.net` |
| `NEXT_PUBLIC_API_URL` | ✅ | URL backend API (передаётся при build) | `https://api.staging.landly.ru` |
| `NEXT_PUBLIC_SITE_BASE_URL` | ❌ | Базовый URL для публичных сайтов (default: `https://landlify.ru`) | `https://landlify.ru` |

### Backend Application Config (передаются в container revision)

| Secret | Описание |
|--------|----------|
| `LANDLY_AUTH_JWT_SECRET` | JWT секрет |
| `LANDLY_DATABASE_POSTGRES_HOST` | Хост PostgreSQL |
| `LANDLY_DATABASE_POSTGRES_PORT` | Порт PostgreSQL |
| `LANDLY_DATABASE_POSTGRES_USER` | Пользователь PostgreSQL |
| `LANDLY_DATABASE_POSTGRES_PASSWORD` | Пароль PostgreSQL |
| `LANDLY_DATABASE_POSTGRES_DBNAME` | Имя базы данных |
| `LANDLY_DATABASE_POSTGRES_SSLMODE` | SSL режим |
| `LANDLY_STORAGE_S3_ENDPOINT` | S3 endpoint |
| `LANDLY_STORAGE_S3_BUCKET` | S3 bucket |
| `LANDLY_STORAGE_S3_ACCESS_KEY` | S3 access key |
| `LANDLY_STORAGE_S3_SECRET_KEY` | S3 secret key |
| `LANDLY_SERVER_CORS_ALLOWED_ORIGINS` | CORS origins |

---

## Frontend Environment Variables

Frontend (Next.js) использует `NEXT_PUBLIC_*` переменные, которые встраиваются в bundle **при сборке**.

### Build-time Variables

| Переменная | Required | Описание | Пример |
|------------|:--------:|----------|--------|
| `NEXT_PUBLIC_API_URL` | ✅ | URL backend API | `https://api.landly.ru` |
| `NEXT_PUBLIC_SITE_BASE_URL` | ❌ | Базовый URL для публичных сайтов | `https://landlify.ru` |

> **Важно:** `NEXT_PUBLIC_*` переменные читаются **только при `npm run build`**, не при runtime. Для production нужно передавать их как `--build-arg` в Docker или как env при сборке в CI.

### Runtime Variables

| Переменная | Default | Описание |
|------------|---------|----------|
| `PORT` | `3000` | Порт Next.js сервера (устанавливается Yandex Serverless автоматически) |
| `NODE_ENV` | `production` | Режим Node.js |

---

## Сравнение источников

| Переменная | config.go | .env.example | .env.prod | docker-compose | workflow |
|------------|:---------:|:------------:|:---------:|:--------------:|:--------:|
| `LANDLY_APP_ENV` | ✅ | ✅ | ✅ | ✅ | hardcoded |
| `LANDLY_APP_NAME` | ✅ | ✅ | ❌ | ❌ | ❌ |
| `LANDLY_APP_VERSION` | ✅ | ✅ | ❌ | ❌ | ❌ |
| `LANDLY_APP_BASE_URL` | ✅ | ✅ | ✅ | ✅ | ✅ |
| `LANDLY_AUTH_JWT_SECRET` | ✅ | ✅ | ✅ | ✅ | ✅ |
| `LANDLY_DATABASE_POSTGRES_*` | ✅ | ✅ | ✅ | ✅ | ✅ |
| `LANDLY_STORAGE_S3_*` | ✅ | ✅ | ✅ | ✅ | ✅ |
| `LANDLY_STORAGE_S3_REGION` | ✅ | ✅ | ✅ | ❌ | ❌ |
| `LANDLY_SERVER_CORS_ALLOWED_ORIGINS` | ✅ | ✅ | ❌ | ❌ | ✅ |
| `LANDLY_BOOTSTRAP_MODE` | ✅ | ✅ | ✅ | ❌ | hardcoded |

---

## Примеры конфигурации

### Local Development (docker-compose)

```bash
LANDLY_APP_ENV=development
LANDLY_AUTH_JWT_SECRET=dev-secret-change-in-production-please
LANDLY_DATABASE_POSTGRES_HOST=postgres
LANDLY_DATABASE_POSTGRES_PORT=5432
LANDLY_DATABASE_POSTGRES_USER=landly
LANDLY_DATABASE_POSTGRES_PASSWORD=landly
LANDLY_DATABASE_POSTGRES_DBNAME=landly
LANDLY_DATABASE_POSTGRES_SSLMODE=disable
LANDLY_STORAGE_S3_ENDPOINT=minio:9000
LANDLY_STORAGE_S3_BUCKET=landly-sites
LANDLY_STORAGE_S3_ACCESS_KEY=minioadmin
LANDLY_STORAGE_S3_SECRET_KEY=minioadmin
LANDLY_STORAGE_S3_USE_SSL=false
```

### Production (Managed PostgreSQL)

Для Yandex Managed PostgreSQL с SSL:

```bash
LANDLY_APP_ENV=production
LANDLY_BOOTSTRAP_MODE=0
LANDLY_AUTH_JWT_SECRET=your-super-secret-production-key-min-32-chars
LANDLY_DATABASE_POSTGRES_HOST=rc1a-xxx.mdb.yandexcloud.net
LANDLY_DATABASE_POSTGRES_PORT=6432
LANDLY_DATABASE_POSTGRES_USER=landly
LANDLY_DATABASE_POSTGRES_PASSWORD=your-secure-password
LANDLY_DATABASE_POSTGRES_DBNAME=landly
LANDLY_DATABASE_POSTGRES_SSLMODE=require
LANDLY_STORAGE_S3_ENDPOINT=storage.yandexcloud.net
LANDLY_STORAGE_S3_BUCKET=landly-sites
LANDLY_STORAGE_S3_ACCESS_KEY=YCAJExxxxxxxx
LANDLY_STORAGE_S3_SECRET_KEY=YCPxxxxxxxx
LANDLY_STORAGE_S3_USE_SSL=true
LANDLY_SERVER_CORS_ALLOWED_ORIGINS=https://app.landly.ru
LANDLY_APP_BASE_URL=https://api.landly.ru
```

### Production (PostgreSQL on VM / Docker)

Для PostgreSQL на собственной VM или в Docker без SSL:

```bash
LANDLY_APP_ENV=production
LANDLY_BOOTSTRAP_MODE=0
LANDLY_AUTH_JWT_SECRET=your-super-secret-production-key-min-32-chars
LANDLY_DATABASE_POSTGRES_HOST=192.168.1.100
LANDLY_DATABASE_POSTGRES_PORT=5432
LANDLY_DATABASE_POSTGRES_USER=landly
LANDLY_DATABASE_POSTGRES_PASSWORD=your-secure-password
LANDLY_DATABASE_POSTGRES_DBNAME=landly
LANDLY_DATABASE_POSTGRES_SSLMODE=disable
LANDLY_STORAGE_S3_ENDPOINT=s3.your-domain.com
LANDLY_STORAGE_S3_BUCKET=landly-sites
LANDLY_STORAGE_S3_ACCESS_KEY=your-access-key
LANDLY_STORAGE_S3_SECRET_KEY=your-secret-key
LANDLY_STORAGE_S3_USE_SSL=true
LANDLY_SERVER_CORS_ALLOWED_ORIGINS=https://app.landly.ru
LANDLY_APP_BASE_URL=https://api.landly.ru
```

