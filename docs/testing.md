# Тестирование

## Автоматически покрыто

- валидация регистрации, сессии и auth-handlers;
- формулы уровня, стадии, энергии, XP и московской даты;
- отдельные JSON-ответы игровых handlers;
- структура миграций и ограничения БД;
- rollback и отправка событий после commit;
- расчёт дневной сводки.

PostgreSQL integration-тесты требуют отдельную базу. Они применяют миграции и удаляют созданных пользователей после теста.

```bash
cd backend
GAME_TEST_DATABASE_URL="postgres://..." go test ./internal/game/integration -count=1
```

Без `GAME_TEST_DATABASE_URL` integration-тесты пропускаются.

## Проверки проекта

```bash
cd backend && go test ./...
cd backend && golangci-lint run ./...
cd frontend && npm run lint && npm run build
docker compose config --quiet
```

Конкурентные запросы, браузерный E2E и сетевое поведение WebSocket отдельными тестами не покрыты.
