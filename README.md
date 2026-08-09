# Авито Тамагочи

MVP виртуального питомца: действия пользователя дают XP, повышают уровень и открывают персональные бонусы.

## Сценарий

1. Пользователь регистрируется и получает сессию.
2. При первом открытии главной страницы создаётся питомец.
3. Зарядка и задания начисляют XP.
4. Новые уровни открывают награды.
5. Сводка, лидерборд и WebSocket показывают изменения.

Правила: [docs/game-design.md](docs/game-design.md). API: [api/openapi.yaml](api/openapi.yaml).

## Стек

- React 19, TypeScript, Vite, ESLint;
- Go 1.26, `net/http`;
- PostgreSQL 18;
- Docker Compose и Caddy.

PostgreSQL — единственный источник истины. Изменения XP, заданий и наград защищены транзакциями и ограничениями БД.

## Запуск

Нужны Docker и Docker Compose.

```bash
cp .env.example .env
docker compose up --build
```

Приложение: <http://localhost:3000>

```bash
docker compose down
```

### HTTPS на домене

```dotenv
SITE_ADDRESS=tamagochi.example.ru
FRONTEND_PORT=80
FRONTEND_HTTPS_PORT=443
APP_SECURE_COOKIES=true
```

DNS должен указывать на сервер. Откройте TCP `80`, TCP `443` и UDP `443`. Caddy получает TLS-сертификат и хранит его в `caddy_data`.

## Проверки

```bash
cd backend && go test ./...
cd backend && golangci-lint run ./...
cd frontend && npm run lint && npm run build
docker compose config --quiet
```

PostgreSQL integration-тесты запускаются отдельно: [docs/testing.md](docs/testing.md). Конфигурация backend-линтера находится в `.golangci.yml`.

## Ограничения MVP

- Caddy — единственная публичная точка входа.
- HTTP API доступен под `/api/v1`, WebSocket — под `/ws`.
- Интеграции с Авито и применение бонусов заменены mock-операциями.
- Нет Redis, очередей, cron, резервного копирования и production-мониторинга.

## Использование ИИ

ИИ использовался для анализа требований, подготовки тестов и проверки согласованности API и документации. Результаты проверялись автоматическими тестами и сборкой.

## Документация

- [Исходный кейс](docs/Кейс%201.pdf)
- [Игровой дизайн](docs/game-design.md)
- [Игровой API](docs/game-api.md)
- [Тестирование](docs/testing.md)
