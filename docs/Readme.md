# Robin2 Docs

This folder contains the project documentation that is still relevant to the current codebase.

Recommended reading order:

1. `../Readme.md` — English project overview.
2. `../README.ru.md` — Russian project overview.
3. `FUNCTIONAL_CAPABILITIES.md` — what the service actually does today.
4. `../spec.md` — implementation-oriented route and config reference.
5. `todo.md` — remaining technical debt and roadmap items.

Generated Swagger artifacts also live here:

- `swagger.json`
- `swagger.yaml`
- `docs.go`

The application page `/docs/` renders Markdown files from this directory, so keep only user-facing or project-facing documents here. Temporary fix logs and one-off migration notes do not belong in this folder.
