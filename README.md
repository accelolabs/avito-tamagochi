# Авито Тамагочи

<p align="center">
  <img src="frontend/src/assets/teen.png" alt="Питомец Авито Тамагочи" width="220">
</p>

MVP виртуального питомца: действия пользователя дают XP, повышают уровень и открывают персональные бонусы.

## Сценарий

1. Пользователь регистрируется и получает сессию.
2. При первом открытии главной страницы создаётся питомец.
3. Зарядка добавляет 20% энергии и 2 XP; первая зарядка дня также увеличивает серию и даёт дневную награду.
4. Ежедневное поглаживание и задания тоже дают XP, а новые уровни открывают награды.
5. Сводка и четыре лидерборда показывают изменения.
6. При 0% энергии питомец умирает и теряет текущий прогресс.
7. Пользователю отправляются уведомления о состоянии зарядки питомца.

Правила: [docs/game-design.md](docs/game-design.md). API: [api/openapi.yaml](api/openapi.yaml).

## Стек

- React 19, TypeScript, Vite, ESLint;
- Go 1.26, `net/http`;
- PostgreSQL 18;
- Docker Compose, Caddy и development-only Mailpit.

PostgreSQL — единственный источник истины. Изменения XP, заданий и наград защищены транзакциями и ограничениями БД.

## Структура проекта

- `backend/cmd` — API, миграции, seed и пакетная рассылка уведомлений;
- `backend/internal` — auth, игровые домены, realtime и инфраструктура;
- `backend/migrations` — схема и ограничения PostgreSQL;
- `frontend/src` — React-приложение и страницы;
- `api` — OpenAPI-контракт;
- `docs` — игровой дизайн, API, описание тестирования.

История коммитов доступна в [GitHub-репозитории](https://github.com/accelolabs/avito-tamagochi).

## Особенности реализации

- Backend — единственный источник расчёта XP, уровней и наград.
- Изменения состояния выполняются транзакционно в PostgreSQL.
- Ограничения БД предотвращают повторное начисление XP и использование награды.
- Пул backend ограничен 25 соединениями, чтобы burst-нагрузка не исчерпывала
  подключения PostgreSQL.
- После изменений frontend получает WebSocket-событие и обновляет данные через API.

## Запуск

Нужны Docker и Docker Compose.

```bash
cp .env.example .env
docker compose up --build
```

- Локально: <http://localhost:3000>
- Mailpit: <http://localhost:8025>
- Production: <https://42plusplus-team.accelolabs.com>

```bash
docker compose down
```

### Email-уведомления через Mailpit

Mailpit перехватывает письма локально. При `--no-deps` PostgreSQL и Mailpit должны быть уже запущены:

```bash
docker compose --profile tools run --rm --no-deps notifications
```

Команда проверяет пользователей с питомцем. Письмо текущего порога не повторяется, пока зарядка не поднимет энергию выше него. UI Mailpit доступен на <http://localhost:8025>; порт и отправитель настраиваются через `MAILPIT_UI_PORT` и `MAIL_FROM`.

Пример ежечасного запуска через cron на Docker-хосте:

```cron
0 * * * * cd /path/to/avito-tamagochi && /usr/bin/docker compose --profile tools run --rm --no-deps notifications >> /var/log/avito-tamagochi-notifications.log 2>&1
```

### Тестовые лидерборды

После запуска приложения отдельная команда создаёт или обновляет десять служебных пользователей для общего, недельного, месячного и streak-лидербордов:

```bash
docker compose run --rm seed
```

Команда не запускается автоматически и не изменяет обычных пользователей.

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
cd backend && golangci-lint run --config ../.golangci.yml ./...
cd frontend && npm run lint && npm run build
docker compose config --quiet
```

CI использует `golangci-lint` версии `v2.10.1`. PostgreSQL integration-тесты запускаются отдельно: [docs/testing.md](docs/testing.md).

## Ограничения MVP

- Caddy — единственная публичная точка входа.
- HTTP API доступен под `/api/v1`, WebSocket — под `/ws`.
- Интеграции с Авито и применение бонусов заменены mock-операциями.
- Email-уведомления работают только через локальный Mailpit; расписание обеспечивает внешний cron, а дедупликацию — PostgreSQL.
- При полной разрядке сбрасываются текущие XP, активная серия и неиспользованные награды; история XP, рекорд серии, использованные награды и прогресс заданий сохраняются.
- Нет Redis, очередей, встроенного scheduler, резервного копирования и production-мониторинга.

## Использование ИИ

ИИ использовался для анализа требований, планирования реализации и написания кода в соответствии с заданными спецификациями. Также он применялся для подготовки тестов и проверки согласованности API и документации. Все результаты дополнительно проверялись автоматическими тестами и сборкой проекта.

## Документация

- [Исходный кейс](docs/Кейс%201.pdf)
- [Игровой дизайн](docs/game-design.md)
- [Игровой API](docs/game-api.md)
- [Архитектура](docs/architecture.md)
- [Тестирование](docs/testing.md)

## Индивидуальный вклад участников

**iconfire7 / @sabirov_amin (Амин) [Репозиторий](https://github.com/iconfire7/avito-iconfire7)** — разработал аутентификацию и сессии, включая регистрацию, вход, безопасность и PostgreSQL-репозитории; реализовал WebSocket-взаимодействие, недельный и месячный лидерборды по приросту XP, механику серии и смерти питомца со сбросом прогресса, а также сидирование тестовых лидербордов.

**accelolabs (Даниил) [Репозиторий](https://github.com/accelolabs/avito-accelolabs)** — разработал основные игровые backend-модули и PostgreSQL-миграции; реализовал ежедневную награду, которая увеличивается вместе с серией зарядок, и лидерборд активных серий; настроил Docker/Caddy, деплой, OpenAPI, тесты, линтер и проектную документацию.

**Kartoshkasmyasom / @nastyalesk (Анастасия)** — реализовала зарядку питомца несколькими нажатиями, дополнительное ежедневное действие «Погладить» и почтовый сервис уведомлений; настроила CI и добавила тесты отката изменений в базе данных при ошибках.

**Alexandra-Vikhareva (Александра)** — разработала дизайн интерфейса и маскота проекта, а также реализовала новый frontend на React и TypeScript: страницы авторизации, питомца, заданий, наград и лидербордов.

## Скриншоты

### Авторизация

<p align="center">
  <img src="docs/screenshots/auth.png" alt="Авторизация" width="700">
</p>

### Питомец

<p align="center">
  <img src="docs/screenshots/pet.png" alt="Питомец" width="700">
</p>

### Задания

<p align="center">
  <img src="docs/screenshots/task.png" alt="Задания" width="700">
</p>

### Награды

<p align="center">
  <img src="docs/screenshots/reward.png" alt="Награды" width="700">
</p>

### Лидерборд

<p align="center">
  <img src="docs/screenshots/leaderboard.png" alt="Лидерборд" width="700">
</p>
