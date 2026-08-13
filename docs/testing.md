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
- пороги email-уведомлений, дедупликация, retry после SMTP-ошибки, UTF-8 шаблоны и HTML/plain MIME;
- инкрементальная зарядка, ежедневное поглаживание и сброс доступных почтовых порогов;

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

## Ручная проверка Mailpit

1. Запустить `docker compose up --build -d` и открыть <http://localhost:8025>.
2. Зарегистрировать пользователя через `/api/v1/auth/register`, сохранив cookie `session_id`.
3. Вызвать `GET /api/v1/pet`, затем `docker compose --profile tools run --rm --no-deps notifications`. Начальные 50% должны создать письмо на email регистрации.
4. В PostgreSQL последовательно установить `energy_percent` в 25, 5 и 0, обновляя `energy_updated_at = CURRENT_TIMESTAMP`. После каждого изменения запустить команду и проверить соответствующий порог.
5. Повторить команду без изменения энергии и убедиться, что письмо не дублируется.
6. Зарядить питомца выше ранее отправленного порога, снова снизить энергию и убедиться, что письмо этого порога снова доступно.
7. В каждом письме проверить получателя, тему `Питомец ждёт вас`, emoji, plain-text, HTML и жирную первую строку.
8. Для 0% проверить death reset: текущие XP, активная серия и неиспользованные награды сбрасываются, а история XP, рекорд серии, использованные награды и прогресс заданий сохраняются.

Cron на Docker-хосте может ежечасно выполнять:

```cron
0 * * * * cd /path/to/avito-tamagochi && /usr/bin/docker compose --profile tools run --rm --no-deps notifications >> /var/log/avito-tamagochi-notifications.log 2>&1
```

После проверки остановить сервисы командой `docker compose down`; `docker compose down -v` не использовать.

## Проверка сидирования

После запуска PostgreSQL и миграций:

```bash
docker compose run --rm seed
docker compose run --rm seed
```

Оба запуска должны завершиться успешно. Во всех четырёх лидербордах должны находиться десять `seed-*` пользователей; второй запуск не должен создавать дополнительные события.
