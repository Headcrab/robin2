# Robin2

[![Go 1.26.1](https://img.shields.io/badge/Go-1.26.1-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-black.svg)](./LICENSE)
[![Docker Ready](https://img.shields.io/badge/Docker-ready-2496ED?logo=docker&logoColor=white)](./deploy/docker-compose.dev.yml)
[![Swagger](https://img.shields.io/badge/API-Swagger-85EA2D?logo=swagger&logoColor=222)](http://localhost:8008/api/swagger/)
[![Русская версия](https://img.shields.io/badge/README-RU-1F6FEB)](./README.ru.md)

Industrial tag history service for querying process data, exposing HTTP APIs, and serving a built-in web UI from a single Go binary.

## Overview

Robin2 is built for operational data access in SCADA / historian-style environments. It can query tag values at a specific point in time, return ranges, aggregate and sample series, decode tag names, execute stored SQL templates, and expose service health, logs, and Swagger from the same runtime.

## Highlights

- point-in-time and range reads for process tags;
- aggregation with `avg`, `sum`, `count`, `min`, `max`;
- range sampling with `count`;
- tag lookup by mask;
- up/down event timestamp lookup;
- tag decoding through `config/tag_classifier.json`;
- SQL template CRUD and execution;
- built-in logs, docs, Swagger, and web UI.

## Stack

- Go `1.26.1`
- ClickHouse, MySQL, MS SQL, Oracle backends
- Redis or in-memory cache
- HTML templates + JS frontend
- Swagger-generated API docs
- Docker / Docker Compose deployment

## Quick Start

1. Copy `.env.example` to `.env`.
2. Fill in the secrets for the active database profile from `config/Robin.json`.
3. Run locally:

```powershell
go run ./cmd
```

Or build a binary:

```powershell
go build -o ./bin/Robin.exe ./cmd
```

Or use task:

```powershell
task build
task docker:rebuild
```

## Configuration

Main config files:

- `config/Robin.json`
- `.env`

Important environment variables:

- `PROJECT_NAME`
- `PROJECT_VERSION`
- `PORT`
- `LOG_PATH`
- `LOG_LEVEL`
- `ROBIN_ADMIN_TOKEN`
- `ROBIN_DB_*`

Sensitive database settings are injected via `${ENV_NAME}` placeholders in `config/Robin.json`.
Default log level is `warn`, so only warnings and errors are written unless you explicitly lower it.

## API Surface

Public routes:

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

Admin routes:

- `POST /api/reload/`
- `GET /templ/list/`
- `POST /templ/add/`
- `GET /templ/get/`
- `POST /templ/edit/`
- `DELETE /templ/delete/`
- `POST /templ/exec/`

Log cleanup:

- `POST /api/log/clear/`
- `DELETE /api/log/clear/`

Admin token can be passed through:

- `X-Admin-Token`
- `Authorization: Bearer <token>`

## Docs

- English spec: [spec.md](./spec.md)
- Functional capabilities: [docs/FUNCTIONAL_CAPABILITIES.md](./docs/FUNCTIONAL_CAPABILITIES.md)
- Docs index: [docs/Readme.md](./docs/Readme.md)
- Russian README: [README.ru.md](./README.ru.md)
- MIT License: [LICENSE](./LICENSE)
- MIT License (RU): [LICENSE.ru.md](./LICENSE.ru.md)

## Example Requests

Get a tag value at a date:

```powershell
curl "http://localhost:8008/get/tag/?tag=A20_WT_01&date=2026-02-19T09:00:00"
```

Get tag list as JSON:

```powershell
curl "http://localhost:8008/get/tag/list/?like=A20_WT_%25&format=json"
```

Execute a template:

```powershell
curl -X POST "http://localhost:8008/templ/exec/" `
  -H "X-Admin-Token: change-me" `
  -d "name=example" `
  -d "args=tag=A20_WT_01,limit=10" `
  -d "format=json"
```

Reload config:

```powershell
curl -X POST "http://localhost:8008/api/reload/" -H "X-Admin-Token: change-me"
```

## Project Layout

```text
cmd/        application entrypoint
internal/   core application packages
config/     runtime configuration
docs/       markdown docs + generated swagger
deploy/     docker and deployment assets
web/        templates, scripts, styles, images
```
