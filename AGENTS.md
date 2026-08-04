# AGENTS.md

## Context

- Monorepository for the "Avito Tamagotchi" MVP.
- Product requirements: `docs/Кейс 1.pdf`.
- `frontend/` contains React and TypeScript; `backend/` contains the Go monolith.

## Commands

```bash
docker compose up --build
docker compose down
cd backend && go test ./...
cd frontend && npm run lint && npm run build
```

## Language

- Write technical artifacts in English: `AGENTS.md`, source-code comments, identifiers, configuration comments, commit messages, and pull-request descriptions.
- Write product-facing content in Russian: `README.md`, `LLMUSE.md`, all files under `docs/`, and user-facing UI text.
- Keep commands, paths, API routes, and library names unchanged.

## Rules

- Keep a simple modular monolith. Do not add microservices, Redis, or message brokers without a team decision.
- Nginx is the only public entry point. HTTP API routes use `/api`; WebSocket uses `/ws`.
- PostgreSQL is the single source of truth.
- Only the backend calculates XP, levels, and rewards.
- Use transactions and idempotency for state-changing operations.
- Do not change the public API without updating its documentation.
- Cover new business logic with unit tests.
- Run all relevant checks before completing a task.
- Do not commit `.env`, secrets, binaries, `node_modules`, or `dist`.
- Do not run `docker compose down -v` without explicit permission.
- Do not modify unrelated files or add unnecessary dependencies.
