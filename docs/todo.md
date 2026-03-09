# TODO Robin2

Status legend:

- `[ ]` not done
- `[~]` in progress
- `[x]` done

## P0

### Security

- `[ ]` Parameterize template metadata queries and template execution paths in `internal/store/base.go`.
- `[x]` Protect `/templ/*` with admin token and non-GET methods for mutating operations.
- `[x]` Protect `/api/reload/` with admin token and `POST`.
- `[x]` Limit `/api/log/clear/` to `POST|DELETE` and gate it by same-origin or admin token.
- `[ ]` Add a global recover middleware so panics become controlled HTTP failures.
- `[ ]` Replace mixed `200 OK + #Error:` responses with proper HTTP status codes on template endpoints.

### Secrets and config

- `[x]` Remove hardcoded DB credentials from `config/Robin.json`.
- `[x]` Expand `${ENV_NAME}` placeholders from environment during config reload.
- `[x]` Fail startup or reload when the active DB/cache still depends on unresolved env secrets.

## P1

### Store and SQL portability

- `[ ]` Make template CRUD portable across supported drivers.
- `[ ]` Respect selected DB name in template execution and close temporary connections correctly.
- `[ ]` Apply configured DB timeouts to ping and connection flows.

### Web and runtime safety

- `[ ]` Remove or scope global `pageCache` so users do not share cached UI state.
- `[ ]` Escape log lines before rendering them into HTML.
- `[ ]` Make `op_count` atomic.

### Testing

- `[~]` Add focused tests for admin-token enforcement and config env expansion.
- `[ ]` Separate unit tests from integration tests that require a live database.
- `[ ]` Decide whether Excel serial dates should resolve to UTC or local time and align tests with code.

## P2

### Cleanup

- `[ ]` Split large handler/store files into smaller units.
- `[ ]` Decide whether `/api/v2/get/...` should be implemented or removed from public routing.
- `[ ]` Revisit web page validation for `/data/` and add clearer user-facing errors.

### Documentation

- `[x]` Sync README, spec, and docs folder with the current route set and security model.
- `[x]` Remove one-off fix notes from `docs/`.
- `[ ]` Document error response conventions once handlers stop mixing plain text and HTTP errors.
