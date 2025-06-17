# TODO Robin2

> Легенда статусов:
>
> * `[ ]` — задача не выполнена
> * `[~]` — работа ведётся
> * `[x]` — задача выполнена и проверена
>
> Приоритеты:
>
> * **P0** — критично, необходимо закрыть до следующего публичного развёртывания
> * **P1** — важно, желательно закрыть в ближайший минорный релиз (≤ 1 мес.)
> * **P2** — улучшения и косметические доработки

---

## P0 — Критические задачи

### 1. Безопасность

1.1 Параметризация SQL‐запросов вместо конкатенации строк

* `[ ]` **internal/store/base.go**

  * `[ ]` TemplateGet
  * `[ ]` TemplateList
  * `[ ]` TemplateExec
  * `[ ]` ExecQuery
* `[ ]` **internal/store/clickhouse/clickhouse.go** — корректно использовать named/position placeholders
* `[ ]` **internal/store/mysql/mysql.go**, **oracle.go**, **mssql.go** — обновить Prepare/Exec логку
* `[ ]` Создать модульные тесты на попытки SQL-инъекций (пакет `store_test`)

1.2 Авторизация и разграничение прав

* `[ ]` Реализовать middleware `auth.JWT`, читающий секрет из ENV
* `[ ]` Добавить проверку `auth` на эндпойнтах:
  `/templ/*`, `/api/reload`, `/api/log`, `/api/*` (кроме `/api/info`, `/api/health`)
* `[ ]` Добавить защиту Swagger UI: bearer token через `Authorize`

1.3 Ограничить CORS

* `[ ]` Переместить установку заголовка **Access-Control-Allow-Origin** в общее middleware
* `[ ]` Считать разрешённые Origin-ы из `config.CORS.AllowedOrigins` (массив)

1.4 Глобальный recover‑middleware

* `[ ]` Создать `internal/middleware/recover.go` с `defer recover()`
* `[ ]` Зарегистрировать в `app.go` до Log и Timing middleware

1.5 Исправить двойную отправку ответа при ошибке

* `[ ]` `internal/app/handlersApi.go`, функция `handleTemplateAdd`, строка `if err != nil ...` — добавить `return` после `http.Error`

---

## P1 — Важные задачи

### 2. Архитектура / SOLID

2.1 Разделить интерфейс `Store`

* `[ ]` Выделить `TagStore`, `TemplateStore`, `MetaStore`
* `[ ]` Актуализировать импорты во всех местах (`app`, `handlers`)

2.2 Декларативное хранение шаблонов

* `[ ]` Переместить таблицу `runtime.templates` в отдельный модуль `templrepo`
* `[ ]` Реализовать JSON‑бэкенд как альтернативу БД (файл в `data/templates.json`)

2.3 Dependency Injection для `App`

* `[ ]` Конструктор `NewApp(store.Store, cache.Cache, logger.Logger)`
* `[ ]` Поддержать старый вызов для обратной совместимости (deprecated)

### 3. Производительность

3.1 Пакетные запросы

* `[ ]` Переписать `GetTagCount` на один SQL используя агрегирование (ClickHouse: `arrayJoin`, MS SQL: `WITH cte`)
* `[ ]` Переписать `GetTagFromTo` для группового SELECT с `IN` по списку тегов

3.2 Ограничить параллелизм

* `[ ]` Добавить параметр `store.parallel_limit` (default: 8)
* `[ ]` Подключить существующий `WorkerPool` к `GetTagFromTo`

3.3 Кэш TTL

* `[ ]` MemoryCache: janitor goroutine, удаляющая записи старше TTL раз в N минут (N = TTL/4)
* `[ ]` RedisCache: `SetEX`+`TTL` configurable

### 4. Тесты

4.1 Unit‑тесты мок‑Store

* `[ ]` Пакет `internal/store/mock`, реализующий `TagStore`, `TemplateStore`
* `[ ]` Обновить существующие тесты, убрать реальную БД из CI

4.2 Интеграционные тесты

* `[ ]` `docker-compose.test.yml` — ClickHouse + Robin2
* `[ ]` GitHub Actions workflow `ci.yml` запуск `go test ./...`

---

## P2 — Низкий приоритет

### 5. Чистота кода

5.1 Удалить или довести неиспользуемый код

* `[ ]` Пакет `queuedserver` (решить, удалить или интегрировать)
* `[ ]` `internal/cache/memoryByte.go` (доделать MD5‑ключи или убрать)

5.2 Разбивка крупных файлов

* `[ ]` `handlersApi.go` → `handlers_tag.go`, `handlers_template.go`, `handlers_log.go`
* `[ ]` `base.go` → `base_tag.go`, `base_template.go`, `base_cache.go`

5.3 Унификация логирования

* `[ ]` Единый формат: `time | level | message | context`
* `[ ]` Исключить дублирование Error‑логов (выше одного уровня)

### 6. Документация

6.1 README

* `[ ]` Описание схемы БД, миграция `runtime.templates`
* `[ ]` Полное описание переменных .env

6.2 API‑документация

* `[ ]` JSON‑формат ошибок `{ "error": "..." }`, коды 4xx/5xx
* `[ ]` Пример авторизации в Swagger (`Bearer <token>`)

---

> *Последнее обновление: 2025‑06‑11*
