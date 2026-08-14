# Архитектура

## Диаграмма компонентов

```mermaid
flowchart LR
    user[Пользователь]
    cron[Внешний cron]

    subgraph web[Web-приложение]
        caddy[Caddy<br/>единая публичная точка входа]
        frontend[React frontend]
    end

    subgraph monolith[Go backend — модульный монолит]
        http[HTTP handlers<br/>/api/v1]
        auth[Auth<br/>пользователи и сессии]
        game[Game<br/>питомец, задания, XP,<br/>награды, сводка и лидерборды]
        realtime[Realtime hub<br/>/ws]
        repositories[PostgreSQL repositories]
    end

    subgraph commands[Одноразовые backend-команды]
        migrate[migrate]
        seed[seed]
        notifications[notifications]
    end

    postgres[(PostgreSQL<br/>единственный источник истины)]
    mailpit[Mailpit<br/>SMTP и локальный UI]

    user -->|HTTP / HTTPS| caddy
    caddy -->|статические файлы| frontend
    frontend -->|REST /api/v1| caddy
    frontend <-->|WebSocket /ws| caddy
    caddy -->|/api/v1| http
    caddy -->|/ws| realtime

    http --> auth
    http --> game
    auth --> repositories
    game --> repositories
    game -->|события после commit| realtime
    repositories --> postgres

    migrate --> postgres
    seed --> postgres
    cron --> notifications
    notifications --> postgres
    notifications -->|SMTP| mailpit
```

Backend остаётся единым процессом: `auth`, игровые модули и `realtime` не являются
отдельными сервисами. XP, уровни и награды рассчитываются только в backend, а
изменения игрового состояния защищаются транзакциями и ограничениями PostgreSQL.

## Диаграмма развёртывания

```mermaid
flowchart TB
    browser[Браузер пользователя]
    scheduler[cron на Docker-хосте]
    developer[Разработчик<br/>localhost]

    subgraph host[Docker-хост]
        subgraph network[Docker Compose network]
            frontendContainer[frontend container<br/>Caddy + React static files<br/>:80 / :443]
            backendContainer[backend container<br/>Go API + WebSocket<br/>:8080 internal]
            postgresContainer[(postgres container<br/>PostgreSQL 18<br/>:5432 internal)]
            mailpitContainer[mailpit container<br/>SMTP :1025 internal<br/>UI :8025]
            migrateContainer[migrate container<br/>one-shot]
            seedContainer[seed container<br/>profile tools, one-shot]
            notificationsContainer[notifications container<br/>profile tools, one-shot]
        end

        postgresVolume[(postgres_data)]
        caddyData[(caddy_data)]
        caddyConfig[(caddy_config)]
    end

    browser -->|HTTP / HTTPS<br/>публичные 80 / 443| frontendContainer
    frontendContainer -->|reverse proxy<br/>/api/v1 и /ws| backendContainer
    backendContainer -->|DATABASE_URL| postgresContainer

    migrateContainer -->|миграции до запуска backend| postgresContainer
    migrateContainer -.->|успешное завершение| backendContainer
    seedContainer -->|DATABASE_URL| postgresContainer
    scheduler -->|docker compose run --rm| notificationsContainer
    notificationsContainer -->|DATABASE_URL| postgresContainer
    notificationsContainer -->|SMTP_ADDRESS :1025| mailpitContainer
    developer -->|127.0.0.1:8025| mailpitContainer

    postgresContainer --- postgresVolume
    frontendContainer --- caddyData
    frontendContainer --- caddyConfig
```

Caddy — единственная публичная точка входа приложения. `backend` и PostgreSQL
доступны только внутри Compose-сети. Mailpit UI публикуется исключительно на
`127.0.0.1`, а `migrate`, `seed` и `notifications` завершаются после выполнения.
