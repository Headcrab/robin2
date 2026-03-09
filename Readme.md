# Robin2

`Robin2` is a Go service for reading historical process tags from industrial databases, exposing them over HTTP API, and serving a small built-in web UI.

## What It Does

- reads tag values at a point in time or over a period;
- aggregates and samples time series;
- searches tags by mask;
- returns up/down event timestamps;
- decodes tag names using `config/tag_classifier.json`;
- executes stored SQL templates;
- serves logs, status, Swagger, and web pages from the same binary.

## Supported Backends

Databases configured in the project:

- `mssql`
- `mysql`
- `clickhouse`
- `oracle`

Cache backends:

- `memory`
- `redis`

The active database and cache are selected through `config/Robin.json` with `curr_db` and `curr_cache`.

## Configuration

Project settings live in:

- `config/Robin.json`
- `.env`

`config/Robin.json` now uses environment placeholders for sensitive fields, for example `${ROBIN_DB_CLICKHOUSE_DOCKER_USER}`.

Important environment variables:

- `PROJECT_NAME`
- `PROJECT_VERSION`
- `PORT`
- `LOG_PATH`
- `LOG_LEVEL`
- `ROBIN_ADMIN_TOKEN`
- `ROBIN_DB_*`

`ROBIN_ADMIN_TOKEN` protects admin endpoints. If it is empty, admin-only routes return `503 Service Unavailable`.

## Run Locally

1. Copy `.env.example` to `.env` and fill in the required secrets.
2. Make sure the active database from `config/Robin.json` is reachable.
3. Start the service:

```powershell
go run ./cmd
```

Or build it:

```powershell
go build -o ./bin/robin.exe ./cmd
```

## Main Routes

Public API:

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

Admin-only API:

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

Access to `/api/log/clear/` is allowed either with an admin token or from the same origin as the web UI.

Admin token can be passed as:

- `X-Admin-Token: <token>`
- `Authorization: Bearer <token>`

## Examples

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

Swagger UI is available at [http://localhost:8008/api/swagger/](http://localhost:8008/api/swagger/).
