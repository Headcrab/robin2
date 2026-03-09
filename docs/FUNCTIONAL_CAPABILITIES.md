# Robin2: Functional Capabilities

## Purpose

`Robin2` is a Go service that reads process history data from industrial databases and exposes it through HTTP endpoints and a small embedded web UI.

## Core Capabilities

### Tag data

- read a single tag value at a specific date;
- read one or more tags for a time range;
- aggregate values by `avg`, `sum`, `count`, `min`, `max`;
- sample a period into a fixed number of points with `count`;
- round output values;
- return data as `text`, `json`, `xml`, `html`, or `grafana` where supported by the handler.

### Tag search and decoding

- search tag names by `like` mask;
- decode tag names using `config/tag_classifier.json`;
- return decoded structures for one or more tags.

### State transitions

- find down events with `GET /get/tag/down/`;
- find up events with `GET /get/tag/up/`;
- select a specific event by index with `count`.

### Templates

- list stored SQL templates;
- read template body;
- create or edit templates;
- delete templates;
- execute templates with arguments and optional database override.

These operations are intentionally restricted because they touch SQL execution.

### Service operations

- return app info and uptime;
- return database status;
- stream logs as text or JSON;
- clear logs;
- reload runtime configuration;
- serve Swagger UI and JSON spec.

## Current Access Model

Admin token source:

- `ROBIN_ADMIN_TOKEN`

Accepted headers:

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

Special rule:

- `/api/log/clear/` allows either a valid admin token or same-origin web requests.

## Input Validation

Template subsystem validation:

- template names: `^[A-Za-z0-9_.:-]+$`
- template list mask: `^[A-Za-z0-9_.:%-]*$`
- template argument keys: `^[A-Za-z0-9_]+$`

Date parsing:

- uses `date_formats` from `config/Robin.json`;
- also accepts Excel serial dates;
- also accepts large numeric Unix timestamps interpreted as milliseconds.

## Supported Infrastructure

Databases:

- `clickhouse`
- `mysql`
- `mssql`
- `oracle`

Cache:

- `memory`
- `redis`

## Web UI

Built-in pages:

- `/`
- `/data/`
- `/tags/`
- `/logs/`
- `/charts/`
- `/docs/`
- `/docs/view/`
- `/swagger/`

The UI also serves static files from `/images/`, `/scripts/`, and `/css/`.

## Operational Notes

- config is loaded from `config/Robin.json`;
- secret values can be injected via `${ENV_NAME}` placeholders;
- startup and config reload validate the active backend before connecting;
- request timing is exposed through middleware with `X-Execution-Time`;
- `/api/v2/get/...` exists but is still only a stub, not a stable feature.
