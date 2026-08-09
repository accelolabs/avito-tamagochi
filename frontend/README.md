# Frontend

React-приложение Авито Тамагочи. Использует REST под `/api/v1` и WebSocket `/ws`.

```bash
npm ci
npm run dev
```

Dev-server проксирует `/api` и `/ws` на backend по `localhost:8080`.

Проверки:

```bash
npm run lint
npm run build
```

Production-сборку раздаёт Caddy; неизвестные маршруты перенаправляются на `index.html`.
