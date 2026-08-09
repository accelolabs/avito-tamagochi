# AGENTS.md

## Context

- Monorepository for the Avito Tamagotchi MVP.
- `frontend/` contains React and TypeScript.
- `backend/` contains the Go modular monolith.
- Product and game rules are documented in `docs/game-design.md`.

## Commands

```bash
docker compose up --build
docker compose down
cd backend && go test ./...
cd frontend && npm run lint && npm run build
```

## Language

- Write technical artifacts in English: source code, comments, identifiers, configuration, commit messages, and pull-request descriptions.
- Write product-facing content in Russian: `README.md`, `LLMUSE.md`, every file under `docs/`, and user-facing UI text. Keep technical names, paths, endpoints, and library names unchanged.
- Keep commands, paths, API routes, and library names unchanged.

## Rules

- Keep the system as a simple modular monolith. Do not add microservices, Redis, or message brokers.
- Caddy is the only public entry point. HTTP API routes use `/api`; WebSocket uses `/ws`.
- PostgreSQL is the single source of truth.
- Only the backend calculates XP, levels, and rewards.
- Use transactions and database constraints to protect state-changing operations from partial updates and duplicate rewards.
- Update `api/openapi.yaml` and relevant documentation when changing the public API.
- Cover new business logic with unit tests and run relevant checks before completing a task.
- Do not commit `.env`, secrets, binaries, `node_modules`, or `dist`.
- Do not run `docker compose down -v` without explicit permission.
- Do not modify unrelated files or add unnecessary dependencies.
