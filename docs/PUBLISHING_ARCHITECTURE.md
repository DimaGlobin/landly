# Архитектура публикации сайтов

## Обзор

Когда пользователь публикует сайт, он сохраняется в S3-совместимое хранилище (например, MinIO, AWS S3, или другой провайдер). Затем пользователи могут получить доступ к опубликованному сайту через CDN или через наш сервер.

## Процесс публикации

1. **Рендеринг**: Схема проекта (JSON) преобразуется в статические HTML/CSS/JS файлы
2. **Загрузка в S3**: Файлы загружаются в S3 по пути `sites/{subdomain}/...`
3. **Сохранение метаданных**: В базе данных создается/обновляется запись `PublishTarget` с subdomain

## Способы доступа к опубликованным сайтам

### Вариант 1: Через CDN (рекомендуется для продакшена)

**Как работает:**
- Если в конфигурации указан `storage.cdn.base_url`, система возвращает прямую ссылку на CDN
- CDN проксирует запросы к S3 и кеширует контент
- Пользователи получают контент напрямую с CDN, минуя наш сервер

**Преимущества:**
- ✅ Нет нагрузки на наш сервер для статического контента
- ✅ Быстрая загрузка благодаря кешированию и географическому распределению
- ✅ Масштабируемость - CDN автоматически обрабатывает большой трафик

**Настройка:**
```yaml
storage:
  cdn:
    base_url: "https://cdn.example.com"  # URL вашего CDN
    enabled: true
```

**Пример URL:** `https://cdn.example.com/sites/my-project-abc123/index.html`

### Вариант 2: Через наш сервер (fallback для dev)

**Как работает:**
- Если CDN не настроен, система возвращает URL вида `{base_url}/sites/{subdomain}`
- Запросы обрабатываются через наш Go сервер
- Сервер получает файлы из S3 через `GetObject()` и возвращает их пользователю

**Преимущества:**
- ✅ Простая настройка для разработки
- ✅ Можно добавить дополнительную логику (аналитика, авторизация и т.д.)

**Недостатки:**
- ❌ Создает нагрузку на сервер
- ❌ Медленнее, чем CDN
- ❌ Не масштабируется для большого трафика

**Пример URL:** `http://localhost:8080/sites/my-project-abc123`

## Роутинг запросов

### Роуты для опубликованных сайтов

```
GET /sites/:slug          - Главная страница сайта
GET /sites/:slug/*path     - Любые другие файлы (CSS, JS, изображения и т.д.)
GET /:slug                - Legacy роут (для обратной совместимости)
```

### Логика поиска файлов

При запросе `/sites/{subdomain}/path/to/file` система ищет файл в следующем порядке:

1. `sites/{subdomain}/path/to/file`
2. `sites/{actual_subdomain}/path/to/file` (если subdomain изменился)
3. `sites/{project_id}/path/to/file` (fallback по project ID)

## Интеграция с CDN провайдерами

### AWS CloudFront + S3

1. Создайте S3 bucket для статических сайтов
2. Настройте CloudFront distribution, указывающий на S3 bucket
3. Укажите CloudFront URL в конфигурации:
   ```yaml
   storage:
     s3:
       endpoint: "s3.amazonaws.com"
       bucket: "landly-sites"
     cdn:
       base_url: "https://d1234567890.cloudfront.net"
   ```

### Cloudflare R2 + CDN

1. Создайте R2 bucket
2. Настройте Cloudflare CDN для R2
3. Укажите CDN URL:
   ```yaml
   storage:
     s3:
       endpoint: "https://your-account-id.r2.cloudflarestorage.com"
       bucket: "landly-sites"
     cdn:
       base_url: "https://cdn.yourdomain.com"
   ```

### Другие провайдеры

Любой S3-совместимый провайдер (DigitalOcean Spaces, Backblaze B2, MinIO и т.д.) + любой CDN (Cloudflare, Fastly, KeyCDN и т.д.)

## Безопасность

### Публичный доступ к S3

S3 bucket должен иметь публичную политику для чтения:
```json
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": {"AWS": ["*"]},
    "Action": ["s3:GetObject"],
    "Resource": ["arn:aws:s3:::bucket-name/*"]
  }]
}
```

### Защита через CDN

CDN может добавить дополнительный уровень защиты:
- Rate limiting
- DDoS защита
- Географические ограничения
- WAF (Web Application Firewall)

## Мониторинг и аналитика

### Метрики для отслеживания

1. **Количество публикаций** - сколько сайтов опубликовано
2. **Трафик через сервер** - если используется fallback через сервер
3. **Трафик через CDN** - через CDN провайдера
4. **Ошибки доступа** - 404, 403 и т.д.

### Интеграция аналитики

Можно добавить аналитику на уровне:
- CDN (CloudFront access logs, Cloudflare Analytics)
- S3 (access logs)
- Наш сервер (если используется fallback)

## Рекомендации для продакшена

1. **Всегда используйте CDN** - это критично для производительности
2. **Настройте кеширование** - CDN должен кешировать статические файлы
3. **Используйте HTTPS** - обязательно для безопасности
4. **Мониторьте использование** - отслеживайте трафик и стоимость
5. **Настройте резервное копирование** - регулярно бэкапьте S3 bucket

## Пример конфигурации для продакшена

```yaml
app:
  base_url: "https://api.landly.com"

storage:
  s3:
    endpoint: "s3.amazonaws.com"
    access_key: "${AWS_ACCESS_KEY_ID}"
    secret_key: "${AWS_SECRET_ACCESS_KEY}"
    bucket: "landly-sites-prod"
    use_ssl: true
    region: "us-east-1"
  
  cdn:
    base_url: "https://cdn.landly.com"
    enabled: true
```

В этом случае:
- Публикация: файлы загружаются в `s3://landly-sites-prod/sites/{subdomain}/`
- Доступ: пользователи получают URL `https://cdn.landly.com/sites/{subdomain}/index.html`
- CDN проксирует запросы к S3 и кеширует контент



