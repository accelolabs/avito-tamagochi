# Тестирование

## Автоматически покрыто

- валидация регистрации, сессии и auth-handlers;
- формулы уровня, стадии, энергии, XP и московской даты;
- отдельные JSON-ответы игровых handlers;
- отклонение trailing JSON и неизвестных полей;
- ограничение пула PostgreSQL backend;
- same-origin политика и штатные close codes WebSocket;
- структура миграций и ограничения БД;
- rollback и отправка событий после commit;
- расчёт дневной сводки;
- расчёт текущей и рекордной серии и растущей дневной награды;
- смерть питомца и согласованность изменяющих XP операций;
- общий, недельный, месячный и streak-лидерборды;
- идемпотентное заполнение всех лидербордов служебными данными.

PostgreSQL integration-тесты требуют отдельную базу. Они применяют миграции и удаляют созданных пользователей после теста.

```bash
cd backend
GAME_TEST_DATABASE_URL="postgres://..." go test ./internal/game/integration -count=1
```

Без `GAME_TEST_DATABASE_URL` integration-тесты пропускаются.

## Проверки проекта

```bash
cd backend && go test ./...
cd backend && golangci-lint run --config ../.golangci.yml ./...
cd frontend && npm run lint && npm run build
docker compose config --quiet
```

Политика WebSocket покрыта unit-тестами. Конкурентные запросы, браузерный E2E и
полное сетевое поведение WebSocket постоянным автоматическим harness не покрыты.

## Проверка сидирования

После запуска PostgreSQL и миграций:

```bash
docker compose run --rm seed
docker compose run --rm seed
```

Оба запуска должны завершиться успешно. Во всех четырёх лидербордах должны находиться десять `seed-*` пользователей; второй запуск не должен создавать дополнительные события.
