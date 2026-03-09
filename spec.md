# Robin2 Specification

This document describes the current configuration model, HTTP routes, and operational constraints of the project as implemented in the repository.

## 1. Configuration

Main config file: `config/Robin.json`

Key fields:

- `port`: HTTP port.
- `round`: default numeric rounding.
- `date_formats`: accepted input date/time layouts.
- `curr_db`: active database profile name.
- `db`: database connection profiles and SQL query templates.
- `curr_cache`: active cache profile name.
- `cache`: cache profiles.

Supported database types:

- `mssql`
- `mysql`
- `clickhouse`
- `oracle`

Supported cache types:

- `memory`
- `redis`

### Environment Expansion

`config/Robin.json` may contain `${ENV_NAME}` placeholders in database and cache settings. On reload, the service expands them from process environment before selecting the active backend.

If the current database or cache still contains unresolved environment references, startup or reload fails validation.

Relevant environment variables include:

- `ROBIN_ADMIN_TOKEN`
- `ROBIN_DB_HS0_*`
- `ROBIN_DB_APPSRV_*`
- `ROBIN_DB_MYSQL_LOCAL_*`
- `ROBIN_DB_MYSQL_DOCKER_*`
- `ROBIN_DB_CLICKHOUSE_DOCKER_*`

## 2. Authentication and Access Rules

Admin-only routes require `ROBIN_ADMIN_TOKEN` and accept it through:

- `X-Admin-Token`
- `Authorization: Bearer <token>`

Admin-only routes:

- `POST /api/reload/`
- `GET /templ/list/`
- `POST /templ/add/`
- `GET /templ/get/`
- `POST /templ/edit/`
- `DELETE /templ/delete/`
- `POST /templ/exec/`

`/api/log/clear/` is less strict:

- accepts `POST` and `DELETE`;
- allows either a valid admin token or a same-origin request based on `Origin` or `Referer`.

## 3. HTTP API

### 3.1 System

#### `GET /api/info/`

Returns JSON:

```json
{
  "name": "Robin",
  "version": "2.4.99",
  "uptime": "1m12s",
  "op_count": 42
}
```

#### `GET /api/status/`

Returns JSON with database name, type, version, uptime, and application uptime.

#### `GET /api/log/`

Supported `format` values:

- `text`
- `str` -> normalized to `text`
- `raw` -> normalized to `text`
- `json`
- any formatter registered in `internal/format` and supported by the handler path

#### `POST /api/reload/`

Reloads `config/Robin.json`, validates the active backend, and reinitializes cache/store.

Requires admin token.

#### `POST /api/log/clear/`
#### `DELETE /api/log/clear/`

Clears log files.

Requires admin token or same-origin web access.

### 3.2 Tags

#### `GET /get/tag/`

Modes:

- `tag + date`
- `tag + from + to`
- `tag + from + to + group`
- `tag + from + to + count`
- `tag + from + to + count + group`

Parameters:

- `tag`: one tag or comma-separated list.
- `date`: point-in-time lookup.
- `from`, `to`: range.
- `group`: aggregation function such as `avg`, `sum`, `count`, `min`, `max`.
- `count`: number of points for sampled output.
- `round`: decimal precision override.
- `format`: response formatter.

#### `GET /get/tag/list/`

Parameters:

- `like`
- `format`

Default response format is `json`.

#### `GET /get/tag/up/`
#### `GET /get/tag/down/`

Parameters:

- `tag`
- `from`
- `to`
- `count`

Return plain text timestamp or empty body.

#### `GET /tag/decode/`

Parameters:

- `tag`
- `format`

Uses `config/tag_classifier.json`.

Default response format is `json`.

### 3.3 Templates

#### `GET /templ/list/`

Query parameters:

- `like`

Restrictions:

- admin token required;
- `like` must match `^[A-Za-z0-9_.:%-]*$`.

#### `POST /templ/add/`

Form parameters:

- `name`
- `body`

Restrictions:

- admin token required;
- `name` must match `^[A-Za-z0-9_.:-]+$`.

#### `GET /templ/get/`

Query parameters:

- `name`

Restrictions:

- admin token required;
- `name` must match `^[A-Za-z0-9_.:-]+$`.

#### `POST /templ/edit/`

Form parameters:

- `name`
- `body`

Restrictions:

- admin token required;
- `name` must match `^[A-Za-z0-9_.:-]+$`.

#### `DELETE /templ/delete/`

Form/query parameters:

- `name`

Restrictions:

- admin token required;
- `name` must match `^[A-Za-z0-9_.:-]+$`.

#### `POST /templ/exec/`

Form parameters:

- `name`
- `db`
- `format`
- `args` formatted as `k1=v1,k2=v2`

Restrictions:

- admin token required;
- template name and optional database name must match `^[A-Za-z0-9_.:-]+$`;
- argument keys must match `^[A-Za-z0-9_]+$`.

## 4. Web UI

Pages served by the binary:

- `/`
- `/data/`
- `/tags/`
- `/logs/`
- `/charts/`
- `/docs/`
- `/docs/view/`
- `/swagger/`

Static assets:

- `/images/`
- `/scripts/`
- `/css/`
- `/favicon.ico`

The `/docs/` page renders Markdown files from the local `docs` directory.

## 5. API v2

`GET /api/v2/get/...` is present in routing but currently acts as a stub that echoes path segments. It is not a stable data API yet.
