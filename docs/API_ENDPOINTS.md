# API Эндпоинты Landly

**Base URL:** `http://localhost:8080`

## 🔓 Публичные эндпоинты (без авторизации)

### Health Checks

#### GET `/healthz`
Проверка здоровья сервиса

**Ответ:**
```json
{
  "status": "ok"
}
```

#### GET `/readyz`
Проверка готовности сервиса

**Ответ:**
```json
{
  "status": "ready"
}
```

---

## 👤 Аутентификация

### POST `/v1/auth/signup`
Регистрация нового пользователя

**Запрос:**
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**Ответ:**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2025-10-13T00:00:00Z"
}
```

**Ошибки:**
- `400` - Invalid input
- `400` - User already exists

---

### POST `/v1/auth/login`
Вход в систему

**Запрос:**
```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**Ответ:**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2025-10-13T00:00:00Z"
}
```

**Ошибки:**
- `401` - Invalid credentials

---

### POST `/v1/auth/refresh`
Обновление токена

**Запрос:**
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Ответ:**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2025-10-13T00:00:00Z"
}
```

---

## 🔐 Приватные эндпоинты (требуют JWT токен)

**Заголовок авторизации:**
```
Authorization: Bearer <access_token>
```

---

## 📁 Управление проектами

### POST `/v1/projects`
Создать новый проект

**Запрос:**
```json
{
  "name": "Мой проект",
  "niche": "Онлайн-образование"
}
```

**Ответ:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "user_id": "123e4567-e89b-12d3-a456-426614174000",
  "name": "Мой проект",
  "niche": "Онлайн-образование",
  "status": "draft",
  "created_at": "2025-10-12T00:00:00Z",
  "updated_at": "2025-10-12T00:00:00Z"
}
```

**Ошибки:**
- `400` - Invalid input
- `401` - Unauthorized

---

### GET `/v1/projects`
Получить список всех проектов пользователя

**Ответ:**
```json
{
  "projects": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "user_id": "123e4567-e89b-12d3-a456-426614174000",
      "name": "Мой проект",
      "niche": "Онлайн-образование",
      "status": "published",
      "created_at": "2025-10-12T00:00:00Z",
      "updated_at": "2025-10-12T00:00:00Z"
    }
  ],
  "total": 1
}
```

---

### GET `/v1/projects/:id`
Получить проект по ID

**Параметры:**
- `id` - UUID проекта

**Ответ:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "user_id": "123e4567-e89b-12d3-a456-426614174000",
  "name": "Мой проект",
  "niche": "Онлайн-образование",
  "status": "published",
  "created_at": "2025-10-12T00:00:00Z",
  "updated_at": "2025-10-12T00:00:00Z"
}
```

**Ошибки:**
- `404` - Project not found
- `403` - Access denied

---

### DELETE `/v1/projects/:id`
Удалить проект

**Параметры:**
- `id` - UUID проекта

**Ответ:**
```
204 No Content
```

**Ошибки:**
- `404` - Project not found
- `403` - Access denied

---

## 🤖 Генерация и публикация

### POST `/v1/projects/:id/generate`
Сгенерировать лендинг с помощью AI

**Параметры:**
- `id` - UUID проекта

**Запрос:**
```json
{
  "prompt": "Создай лендинг для онлайн-курса по программированию. Аудитория: новички. Оффер: научим программировать за 3 месяца. Преимущества: опытные преподаватели, практика, трудоустройство.",
  "payment_url": "https://pay.prodamus.ru/checkout-abc123"
}
```

**Ответ:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "user_id": "123e4567-e89b-12d3-a456-426614174000",
  "name": "Мой проект",
  "niche": "Онлайн-образование",
  "status": "generated",
  "created_at": "2025-10-12T00:00:00Z",
  "updated_at": "2025-10-12T00:00:00Z"
}
```

**Ошибки:**
- `404` - Project not found
- `403` - Access denied
- `500` - Generation failed

---

### GET `/v1/projects/:id/preview`
Получить JSON-схему лендинга для предпросмотра

**Параметры:**
- `id` - UUID проекта

**Ответ:**
```json
{
  "schema": {
    "version": "1.0",
    "pages": [
      {
        "path": "/",
        "title": "Онлайн-курс по программированию",
        "description": "Научитесь программировать за 3 месяца",
        "blocks": [
          {
            "type": "hero",
            "order": 0,
            "props": {
              "headline": "Научитесь программировать за 3 месяца",
              "subheadline": "Практический курс для начинающих",
              "ctaText": "Записаться на курс",
              "image": "https://example.com/hero.jpg"
            }
          },
          {
            "type": "features",
            "order": 1,
            "props": {
              "title": "Наши преимущества",
              "items": [
                {
                  "icon": "⚡",
                  "title": "Практические навыки",
                  "description": "Реальные проекты в портфолио"
                }
              ]
            }
          }
        ]
      }
    ],
    "theme": {
      "palette": {
        "primary": "#3B82F6",
        "secondary": "#8B5CF6"
      },
      "font": "inter",
      "borderRadius": "lg"
    },
    "payment": {
      "url": "https://pay.prodamus.ru/checkout-abc123",
      "buttonText": "Оплатить"
    }
  }
}
```

**Ошибки:**
- `404` - Project not found or no schema generated yet
- `403` - Access denied

---

### POST `/v1/projects/:id/publish`
Опубликовать лендинг в S3

**Параметры:**
- `id` - UUID проекта

**Ответ:**
```json
{
  "subdomain": "moj-proekt-550e8400",
  "public_url": "http://localhost:9000/landly-sites/sites/550e8400-e29b-41d4-a716-446655440000",
  "published_at": "2025-10-12T00:00:00Z"
}
```

**Ошибки:**
- `404` - Project not found
- `403` - Access denied
- `400` - Project has no generated schema
- `500` - Publishing failed

---

## 📊 Аналитика

### POST `/v1/analytics/:id/event`
Отправить событие аналитики (публичный эндпойнт)

**Параметры:**
- `id` - UUID проекта

**Запрос:**
```json
{
  "event_type": "pageview",
  "path": "/",
  "referrer": "https://google.com"
}
```

**Типы событий:**
- `pageview` - просмотр страницы
- `cta_click` - клик по CTA кнопке
- `pay_click` - клик по кнопке оплаты
- `form_submit` - отправка формы

**Ответ:**
```
204 No Content
```

**Примечание:** Этот эндпойнт публичный и используется для трекинга с опубликованных лендингов.

---

### GET `/v1/analytics/:id/stats`
Получить статистику по проекту (приватный эндпойнт)

**Параметры:**
- `id` - UUID проекта

**Ответ:**
```json
{
  "project_id": "550e8400-e29b-41d4-a716-446655440000",
  "total_pageviews": 1543,
  "total_cta_clicks": 87,
  "total_pay_clicks": 23,
  "unique_visitors": 892
}
```

**Ошибки:**
- `404` - Project not found
- `403` - Access denied

---

## 📝 Примеры использования

### Полный flow создания лендинга

1. **Регистрация:**
```bash
curl -X POST http://localhost:8080/v1/auth/signup \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}'
```

2. **Создание проекта:**
```bash
curl -X POST http://localhost:8080/v1/projects \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"name":"Мой курс","niche":"Образование"}'
```

3. **Генерация лендинга:**
```bash
curl -X POST http://localhost:8080/v1/projects/<project_id>/generate \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "prompt":"Создай лендинг для онлайн-курса",
    "payment_url":"https://pay.prodamus.ru/abc"
  }'
```

4. **Получение preview:**
```bash
curl -X GET http://localhost:8080/v1/projects/<project_id>/preview \
  -H "Authorization: Bearer <token>"
```

5. **Публикация:**
```bash
curl -X POST http://localhost:8080/v1/projects/<project_id>/publish \
  -H "Authorization: Bearer <token>"
```

6. **Получение статистики:**
```bash
curl -X GET http://localhost:8080/v1/analytics/<project_id>/stats \
  -H "Authorization: Bearer <token>"
```

---

## 🔍 Коды ошибок

| Код | Описание |
|-----|----------|
| 200 | OK - Успешный запрос |
| 201 | Created - Ресурс создан |
| 204 | No Content - Успешно, нет содержимого |
| 400 | Bad Request - Невалидные данные |
| 401 | Unauthorized - Не авторизован |
| 403 | Forbidden - Доступ запрещён |
| 404 | Not Found - Ресурс не найден |
| 500 | Internal Server Error - Ошибка сервера |

---

## 📖 Дополнительная документация

- Полная OpenAPI спецификация: [docs/openapi.yaml](openapi.yaml)
- Главная документация: [README.md](../README.md)
- Быстрый старт: [QUICKSTART.md](../QUICKSTART.md)

