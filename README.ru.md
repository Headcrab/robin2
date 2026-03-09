# Robin2

[![Go 1.26.1](https://img.shields.io/badge/Go-1.26.1-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![Лицензия MIT](https://img.shields.io/badge/License-MIT-black.svg)](./LICENSE.ru.md)
[![Docker Ready](https://img.shields.io/badge/Docker-ready-2496ED?logo=docker&logoColor=white)](./deploy/docker-compose.dev.yml)
[![Swagger](https://img.shields.io/badge/API-Swagger-85EA2D?logo=swagger&logoColor=222)](http://localhost:8008/api/swagger/)
[![English README](https://img.shields.io/badge/README-EN-1F6FEB)](./Readme.md)

Сервис для доступа к историческим промышленным тегам, HTTP API и встроенного web-интерфейса из одного Go-бинаря.

## Что это такое

Robin2 рассчитан на работу с технологическими данными в среде SCADA / historian. Сервис умеет читать значения тегов на момент времени и за период, агрегировать и дискретизировать ряды, искать теги по маске, декодировать имена тегов, выполнять SQL-шаблоны и отдавать статус, логи, документацию и Swagger из одного процесса.

## Что умеет

- получать значение тега на дату;
- получать диапазоны значений;
- агрегировать данные через `avg`, `sum`, `count`, `min`, `max`;
- дискретизировать диапазон через `count`;
- искать теги по `like`-маске;
- находить моменты `up/down`;
- декодировать имена тегов через `config/tag_classifier.json`;
- хранить и выполнять SQL-шаблоны;
- показывать статус сервиса, логи, docs и Swagger.

## Технологии

- Go `1.26.1`
- ClickHouse, MySQL, MS SQL, Oracle
- Redis или in-memory cache
- HTML templates + JS
- Swagger
- Docker / Docker Compose

## Быстрый старт

1. Скопируй `.env.example` в `.env`.
2. Заполни секреты для активного DB-профиля из `config/Robin.json`.
3. Запусти локально:

```powershell
go run ./cmd
```

Или собери бинарь:

```powershell
go build -o ./bin/Robin.exe ./cmd
```

Или через task:

```powershell
task build
task docker:rebuild
```

## Конфигурация

Основные файлы:

- `config/Robin.json`
- `.env`

Ключевые переменные окружения:

- `PROJECT_NAME`
- `PROJECT_VERSION`
- `PORT`
- `LOG_PATH`
- `LOG_LEVEL`
- `ROBIN_ADMIN_TOKEN`
- `ROBIN_DB_*`

Чувствительные параметры БД подставляются через `${ENV_NAME}` в `config/Robin.json`.
Уровень логирования по умолчанию `warn`, то есть в лог пишутся только предупреждения и ошибки, если ты сам не занизишь порог.

## Основные маршруты API

Публичные маршруты:

- `GET /get/tag/`
- `GET /get/tag/list/`
- `GET /get/tag/up/`
- `GET /get/tag/down/`
- `GET /tag/decode/`
- `GET /api/info/`
- `GET /api/status/`
- `GET /api/log/`
- `GET /api/swagger/`
- `GET /api/swagger/doc.json`

Админские маршруты:

- `POST /api/reload/`
- `GET /templ/list/`
- `POST /templ/add/`
- `GET /templ/get/`
- `POST /templ/edit/`
- `DELETE /templ/delete/`
- `POST /templ/exec/`

Очистка логов:

- `POST /api/log/clear/`
- `DELETE /api/log/clear/`

Admin token можно передавать через:

- `X-Admin-Token`
- `Authorization: Bearer <token>`

## Документация

- English README: [Readme.md](./Readme.md)
- Спецификация: [spec.md](./spec.md)
- Возможности системы: [docs/FUNCTIONAL_CAPABILITIES.md](./docs/FUNCTIONAL_CAPABILITIES.md)
- Индекс доков: [docs/Readme.md](./docs/Readme.md)
- MIT License: [LICENSE](./LICENSE)
- MIT на русском: [LICENSE.ru.md](./LICENSE.ru.md)

## Примеры запросов

Получить значение тега на дату:

```powershell
curl "http://localhost:8008/get/tag/?tag=A20_WT_01&date=2026-02-19T09:00:00"
```

Получить список тегов в JSON:

```powershell
curl "http://localhost:8008/get/tag/list/?like=A20_WT_%25&format=json"
```

Выполнить шаблон:

```powershell
curl -X POST "http://localhost:8008/templ/exec/" `
  -H "X-Admin-Token: change-me" `
  -d "name=example" `
  -d "args=tag=A20_WT_01,limit=10" `
  -d "format=json"
```

Перечитать конфиг:

```powershell
curl -X POST "http://localhost:8008/api/reload/" -H "X-Admin-Token: change-me"
```

## Структура проекта

```text
cmd/        точка входа приложения
internal/   основная логика
config/     runtime-конфиги
docs/       markdown-доки и swagger
deploy/     docker и деплой
web/        шаблоны, скрипты, стили, картинки
```
