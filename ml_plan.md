# Landly — План реализации AI-генерации лендингов (MLE)

## 0) Цели и результат
- Заменить mock-генерацию на реальную AI-интеграцию с гарантией строгого JSON (PageSchema) и публикации без ручных правок.
- Обеспечить диалоговые правки через патчи (JSON Patch/Merge Patch) с валидацией и авто-фиксом.
- Сделать генерацию клиенториентированной: учитывать бренд/продукт, цены, ссылки, цитаты, предпочтения дизайна.
- Метрики качества/стоимости и управляемые ретраи.

Definition of Done (MVP):
- POST /v1/projects/:id/generate возвращает валидный PageSchema и публикуется end-to-end.
- Валидатор JSON Schema + бизнес-правила проходят ≥95% без ретрая; остальное — 1 ретрай/авто‑фикс.
- Чат-правки применяются патчем к текущей схеме и проходят валидацию.
- Поддержка минимум 1 провайдера (OpenAI или OpenRouter с моделью, поддерживающей structured outputs). Мок — как fallback.
- Базовые UI‑поля: goal, target_audience, tone, language, paymentURL.
- Метрики: latency, tokens, success, validation_errors, autofix_count.

---

## 1) Архитектура (высокоуровнево)
- AI Provider Layer: `internal/storage/ai` — реализация клиентов (openai, openrouter), выбор по cfg.AI.Provider.
- Prompt Layer: константы System/Dev/User для SGR + Structured Outputs; few-shot примеры.
- Validation Layer: JSON Schema PageSchema + бизнес‑правила; автопочинки и ретраи.
- CCE (Client Context Engine): сбор контекста бренда/продукта/сниппетов; валидация включения обязательных элементов.
- Chat Patch Engine: SGR-режим правок (RFC6902/RFC7386) + optimistic locking проекта.
- Observability: логи/метрики (latency, tokens, success, валидность, autofикс).

---

## 2) Фазы и дорожная карта

### Фаза P0 — Реальная генерация вместо mock (1–2 дня)
1. Конфиг
   - Добавить опции в `config.yml`: `ai.provider`, `ai.model`, `ai.response_format`, `ai.max_tokens`, `ai.temperature`, `ai.timeout_ms`.
   - Поддержать env LANDLY_AI_* (в `config.Load()`).
2. AI-клиент
   - `apps/backend/internal/storage/ai/openai_client.go` (или `openrouter_client.go`): Chat Completions с structured outputs/JSON mode.
   - Интерфейс `Client{ GenerateLandingSchema(ctx, prompt, paymentURL string) (string, error) }` уже есть — расширить внутри реализацию.
3. Промпты
   - Добавить `SYSTEM_PROMPT`, `DEV_PROMPT_TEMPLATE`, `USER_PROMPT_TEMPLATE` для генерации PageSchema.
   - Вставлять goal/audience/tone/language/paymentURL при сборке.
4. Валидация
   - Добавить JSON Schema (файл `docs/page_schema.json`).
   - Пакет валидации: `internal/validation/pageschema` с `Validate([]byte) error` + бизнес‑правила (≥3 features, CTA обязателен, лимиты длин).
5. Постпроцессинг
   - Авто‑фиксы: обрезка длины, заполнение отсутствующих обязательных полей безопасными значениями; `notes.validation_autofix`.
6. Ретраи/деградации
   - Один ретрай при невалидном JSON; затем fallback на mock с логом причины.
7. Интеграция
   - В `cmd/api/main.go` включить провайдер по конфигу (mock/openai/openrouter) без `log.Fatal` для не‑mock.
8. Документация
   - README: переменные окружения, включение реального провайдера, ограничения.

Выход: генерация работает end‑to‑end, публикация успешна.

### Фаза P1 — Client Context Engine (CCE) — минимум (2–4 дня)
1. Доменная модель и миграции (Postgres)
   - Таблицы: BrandProfile, ProductProfile, ContentSnippets (см. раздел 6 ниже).
   - Миграции в `apps/backend/migrations`.
2. CRUD API (минимум)
   - `PUT /v1/projects/:id/brand`, `PUT /v1/projects/:id/product`, `POST /v1/projects/:id/snippets` (за авторизацией).
3. Сервис и сборка контекста
   - `internal/services/cce/service.go`: `BuildPromptContext(projectID, opts) (Context, error)` с кэшем версии.
4. Интеграция в генерацию
   - В `GenerateService` перед вызовом AI собрать `ctx := CCE.BuildPromptContext(...)` и включить его в Dev‑prompt.
5. Проверка включения обязательных элементов
   - `CCE.ValidateRequiredInclusion(ctx, page *PageSchema) (FixReport, error)`: наличие payment_url, цен, цитат, палитры.
   - При нарушении — авто‑патч результата до публикации (минимально инвазивно).

Выход: генерации учитывают бренд/продукт, цены/ссылки/цитаты и проходят проверку включения.

### Фаза P2 — Чат‑режим с SGR‑патчами (3–5 дней)
1. Контракт AI‑ответа
   - `{ reasoning_steps: [], patch: { type: "json_patch"|"merge_patch", ops|doc }, summary }`.
2. Промпты (SGR Chat)
   - `SYSTEM_PROMPT_SGR_CHAT`, `DEV_PROMPT_TEMPLATE_CHAT` (см. ранее обсуждённые шаблоны).
3. Применение патча
   - Библиотеки: `github.com/evanphx/json-patch/v5` для RFC6902/RFC7386.
   - Применять к актуальной версии схемы; повторная валидация + авто‑фиксы.
4. Версионность схемы проекта
   - Поле `schema_version` в Project; optimistic locking (если версия изменилась — попытка смарт‑мерджа/конфликт).
5. API
   - `POST /v1/projects/:id/chat { message, patch_mode }` возвращает `{ previewSchema, summary }`.
6. Стриминг (опционально)
   - SSE/WebSocket: события reason:step, patch, preview.

Выход: интеративные правки быстрые и детерминированные.

### Фаза P3 — Провайдеры, надёжность, метрики (2–4 дня)
1. Второй провайдер
   - Добавить `openrouter` или `bedrock`/`gemini` (если есть ключи), выбрать модели с поддержкой structured outputs.
2. Стратегия ретраев
   - JSON‑невалида → уточняющий репромпт; таймаут → деградация на mock; лог причина.
3. Метрики/логирование
   - latency_ms, input_tokens, output_tokens, success, validation_errors_count, autofix_count, patch_apply_success.
   - Экспорт в стандартный логгер +, при включении, Prometheus (observability.metrics.enabled).
4. Тюнинг промптов
   - Few‑shot примеры для sales/leadgen/info, RU/EN варианты.

---

## 3) Изменения по коду (карта работ)

Backend
- Конфиг: `config.yml` + `apps/backend/config` — добавить поля AI.
- AI: `apps/backend/internal/storage/ai/` — `openai_client.go`, `openrouter_client.go`, обновить `mock_client.go` при необходимости; фабрика клиента.
- Промпты: `apps/backend/internal/storage/ai/prompts.go` — константы System/Dev/User (генерация + чат‑патчи).
- Валидация: `apps/backend/internal/validation/pageschema/` — JSON Schema загрузка, валидатор и бизнес‑правила.
- CCE: `apps/backend/internal/services/cce/` — сервис, контракты, проверки включения, кэш.
- GenerateService: включение CCE, валидация/авто‑фиксы, ретраи (файл уже есть `internal/services/generate.go`).
- Chat endpoints: расширить обработчик `GenerateHandler` или ввести `ChatHandler` (роут уже есть: `/v1/projects/:id/chat`).
- Миграции: `apps/backend/migrations` — для BrandProfile/ProductProfile/ContentSnippets и schema_version.
- Observability: добавить метрики/логи в AI‑клиент и сервисы.

Frontend (минимум)
- Форма генерации: добавить поля `goal`, `target_audience`, `tone`, `language`, `paymentURL`.
- Экран «Brand/Product setup»: простые формы под CCE CRUD.
- Чат‑правки: параметр `patch_mode` и вывод summary; предпросмотр после патча.

Docs
- `docs/page_schema.json` — схема для PageSchema.
- `docs/ai_prompts.md` — промпт‑константы и рекомендации по тюнингу.

---

## 4) PageSchema: валидация и правила
- JSON Schema: обязательные разделы hero, features, cta, footer; длины полей (title ≤ 90, subtitle ≤ 140 и т.п.).
- Бизнес‑правила:
  - features ≥ 3; CTA обязателен; ссылки https; action_url|payment_url одно из не null.
  - theme.palette — 2–3 цвета максимум; если бренд‑цвета заданы — должны попасть в palette.
- Авто‑фиксы: обрезка длин; заполнение CTA безопасной фразой; добивка features «заглушками» (с логом autofix).

---

## 5) Промпты (сводка)
- System: режим SGR/Structured Outputs, запрет текста вне JSON, строгая FORMAL_SPEC/Schema.
- Dev: контекст проекта + CCE‑контекст (brand/product/snippets), правила структуры, бизнес‑ограничения.
- User: исходный prompt + доп.параметры (keywords/locale/forbidden/length_hint).
- Chat (SGR Patch): вывод только `{reasoning_steps, patch, summary}`, изменять только запрошенные части, минимальные операции.

---

## 6) Данные CCE (модель и миграции)
Таблицы (ядро):
- BrandProfile(id, project_id, brand_name, brand_tone, brand_colors jsonb, font, style_preset, guidelines jsonb, version, versioned_at)
- ProductProfile(id, project_id, product_name, features jsonb[], pricing jsonb[], links jsonb, audience, differentiators jsonb[], version, versioned_at)
- ContentSnippets(id, project_id, label, content, locale, tags jsonb[], version, versioned_at)
- В Project: schema_version int (для optimistic locking/чат‑патчей)

Валидации CCE:
- HEX‑цвета, https‑ссылки, цены (число + валюта + период), forbidden_words фильтр, лимиты длин.
- FixReport: список автопоправок/предупреждений.

---

## 7) Тестирование
- Unit: 
  - Промпт‑билдеры (подстановка полей), пост‑процессинг/авто‑фиксы, валидатор схемы.
  - CCE.ValidateRequiredInclusion.
  - Патч‑применение (RFC6902/7386) и конфликтные ситуации.
- Интеграция:
  - С мок‑клиентом (стабильный JSON) и с реальным провайдером (build‑tag `ai`).
  - Генерация → валидация → публикация → GET /sites/:slug.
- E2E (минимум):
  - Форма генерации с обязательными полями → предпросмотр → публикация.
  - Чат: 3 сценария патчей (replace title, add feature, update CTA).

---

## 8) Observability и безопасность
- Метрики: latency_ms, input_tokens, output_tokens, success, json_valid, validation_errors_count, autofix_count, patch_apply_success.
- Логи: трассировка request_id, причины ретраев, провайдер/модель, без промптов в прод‑логах (опция dev‑trace=true).
- Безопасность: маскирование PII/секретов, allowlist доменов payment, HTTPS‑проверка ссылок.

---

## 9) Риски и смягчение
- Невалидный JSON от модели → structured outputs + валидатор + ретрай + fallback mock.
- Рост латентности/стоимости из‑за SGR → компактный режим reasoning в проде; лимиты токенов.
- UX‑перегрузка формами → пресеты + обязательный минимум, остальное — «расширенные настройки».
- Несовместимость рендера с расширенной схемой → строгое версионирование схемы и обратная совместимость.

---

## 10) План релиза (инкременты)
- R1 (P0): провайдер + валидация + публикация (без CCE/чата), дока и переменные.
- R2 (P1): CCE минимум + включение обязательных элементов + UI формы бренда/продукта.
- R3 (P2): чат‑патчи + версионность схемы + предпросмотр патчей.
- R4 (P3): второй провайдер, расширенные метрики, стриминг, тюнинг промптов/few‑shot.

---

## 11) Чек‑лист приёмки
- Генерация с реальным провайдером: валидный JSON, публикация успешна.
- CCE: цены/ссылки/цитаты/цвета попадают в результат (или авто‑фиксятся), отчёт о включении проходит.
- Чат‑патчи: 3 базовых сценария проходят, конфликты обрабатываются.
- Метрики отображаются, логи информативны без утечки данных.
- README обновлён: включение AI‑провайдера, переменные, ограничения.

---

## 12) Быстрые задачи (первые PR)
1) Добавить поля AI в конфиг и фабрику клиента, выключить `log.Fatal` для не‑mock.
2) Реализовать `openai_client.go` (или `openrouter_client.go`) с structured outputs.
3) Добавить `docs/page_schema.json` и валидатор с бизнес‑правилами.
4) Включить постпроцессинг и 1 ретрай при невалидном JSON.
5) Обновить README (раздел AI: как включить провайдера, env‑переменные).

---

## Прогресс
- P0a (конфигурация/доки): обновлены `config.yml`, `config.example.yml`, README и `config.Config` (новые поля, таймаут, response_format, OpenRouter-заготовка). Ожидает интеграцию клиента.
- P0b (AI client): добавлен `ai.NewProviderClient` с поддержкой OpenAI/OpenRouter совместимого Chat Completions (HTTP клиент, ретраи, базовый промпт). Backend (`cmd/api`, `internal/server`) теперь инициализируют реальный клиент. Тесты `cd apps/backend && go test ./...` проходят.
