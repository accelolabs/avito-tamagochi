# Авито Тамагочи

MVP виртуального питомца, который развивается от полезных действий пользователя на классифайде и открывает персональные бонусы Авито.

## Пользовательский сценарий

1. Пользователь регистрируется и получает питомца.
2. Действия изменяют состояние питомца и начисляют опыт.
3. Опыт повышает уровень и открывает персональные награды.
4. WebSocket, лидерборд и ежедневная сводка показывают прогресс и мотивируют вернуться.

> Точные параметры, действия, формула прогрессии и награды будут зафиксированы до реализации бизнес-логики.

## Стек

- Frontend: React 19, TypeScript, Vite, ESLint, Nginx.
- Backend: Go 1.26, стандартный `net/http`.
- Data: PostgreSQL 18.
- Infrastructure: Docker Compose.

## Запуск

Требуется Docker с Docker Compose.

```bash
cp .env.example .env
docker compose up --build
```

- Приложение: <http://localhost:3000>
- Healthcheck: <http://localhost:3000/api/health>

```bash
docker compose ps
docker compose down
```

## Проверки

```bash
cd backend && go test ./...
cd frontend && npm run lint && npm run build
```

## Документация

- [Исходный кейс](docs/Кейс%201.pdf)
