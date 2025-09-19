## Стратегия развертывания IdeaFrame (MVP)

Документ описывает, как развернуть и сопровождать систему из `docs/prd.md` и `docs/architecture.md`. Цель — простой, надёжный и повторяемый деплой через Docker, с Caddy (TLS), PostgreSQL и миграциями.

---

## Окружения

### Local (разработка)
- Цель: быстрый запуск всей системы локально одной командой.
- Инструменты: Docker Compose, локальный `.env`.
- Доступ: `http://localhost` (без TLS), API на `http://localhost:8080`.

### Staging (по желанию)
- Цель: проверка релизов перед продом.
- Один сервер (VM) с Docker Compose, реальный домен вида `stage.example.com`, TLS от Caddy.

### Production
- Цель: минимально сложная эксплуатация с HTTPS и бэкапами.
- Один сервер (VM) с Docker Compose, домен `app.example.com`, TLS от Caddy, ежедневные бэкапы БД.

---

## Компоненты развертывания
- Caddy — обратный прокси и TLS (Let’s Encrypt), статика SPA.
- API (Go) — HTTP сервис, авторизация по JWT.
- PostgreSQL — транзакционная БД.
- Migrate job — одноразовый запуск миграций при каждом деплое.

Сетевые связи (внутри сети Docker): Caddy → API (HTTP), API → Postgres (TCP), Migrate → Postgres (DDL). Внешний трафик: Клиент → Caddy (HTTP/HTTPS).

---

## Переменные окружения (.env)
Используются во всех окружениях; секреты не хранятся в Git. Для локали допустим `.env` в репозитории как пример, для staging/prod — только секреты CI/CD или секрет‑хранилище.

Рекомендуемый набор:

```dotenv
# Общие
APP_ENV=local                # local | staging | production
JWT_SECRET=replace-me-32bytes-min

# API / БД
DB_HOST=db
DB_PORT=5432
DB_USER=idea
DB_PASSWORD=idea
DB_NAME=idea
DB_SSLMODE=disable           # production: prefer require/verify-full при наличии TLS к БД
DB_DSN=postgres://idea:idea@db:5432/idea?sslmode=disable

# Миграции
MIGRATIONS_DIR=/migrations

# Caddy / TLS
CADDY_DOMAIN=localhost       # production: app.example.com
CADDY_EMAIL=admin@example.com
```

---

## Docker Compose (рекомендованный план)
Файл предполагается в `infra/docker-compose.yml` и может применяться как локально, так и на сервере. Ниже — ориентир.

```yaml
version: "3.9"
services:
  db:
    image: postgres:16
    environment:
      POSTGRES_DB: ${DB_NAME}
      POSTGRES_USER: ${DB_USER}
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - db-data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U $$POSTGRES_USER -d $$POSTGRES_DB"]
      interval: 5s
      timeout: 5s
      retries: 20
    ports:
      - "5432:5432" # только для локали; в staging/prod убрать публикацию

  migrate:
    image: migrate/migrate:4
    command: [ "-path", "/migrations", "-database", "${DB_DSN}", "up" ]
    volumes:
      - ../db/migrations:/migrations:ro
    depends_on:
      db:
        condition: service_healthy
    restart: "on-failure"

  api:
    # В локали можно билдить из исходников; в prod — указывать готовый образ
    build:
      context: ../app/backend
    environment:
      APP_ENV: ${APP_ENV}
      JWT_SECRET: ${JWT_SECRET}
      DB_DSN: ${DB_DSN}
    depends_on:
      migrate:
        condition: service_completed_successfully
    healthcheck:
      test: ["CMD", "curl", "-fsS", "http://localhost:8080/healthz"]
      interval: 10s
      timeout: 3s
      retries: 10
    ports:
      - "8080:8080" # только для локали; в staging/prod убрать публикацию

  caddy:
    image: caddy:2
    depends_on:
      api:
        condition: service_healthy
    volumes:
      - ./Caddyfile:/etc/caddy/Caddyfile:ro
      - caddy_data:/data
      - caddy_config:/config
      # Если SPA билдится, положить артефакты в app/frontend/dist и смонтировать:
      - ../app/frontend/dist:/srv/www:ro
    ports:
      - "80:80"
      - "443:443" # для локали можно опустить 443

volumes:
  db-data:
  caddy_data:
  caddy_config:
```

---

## Конфигурация Caddy

### Local (без TLS)
```caddyfile
{
  auto_https off
  admin off
}

http://localhost:80 {
  encode gzip

  @api path /api/* /auth/*
  handle @api {
    reverse_proxy api:8080
  }

  handle {
    root * /srv/www
    try_files {path} /index.html
    file_server
  }
}
```

### Production/Staging (с TLS от Let’s Encrypt)
```caddyfile
{
  email {env.CADDY_EMAIL}
}

{env.CADDY_DOMAIN} {
  encode gzip

  @api path /api/* /auth/*
  handle @api {
    reverse_proxy api:8080
  }

  handle {
    root * /srv/www
    try_files {path} /index.html
    file_server
  }
}
```

Примечания:
- По необходимости добавьте CORS/заголовки безопасности на стороне API.
- При желании можно включить rate limiting/бот‑защиту (через плагины Caddy или на уровне API).

---

## CI/CD (ориентир)
Инструменты: GitHub Actions (пример), Docker Registry (GHCR/Docker Hub), SSH‑deploy на VM со `docker compose`.

Основной pipeline:
1) Lint/Build: собрать Docker образы `api`, (опционально) `frontend`.
2) Push в реестр с тегами `:commit-sha` и `:latest` (или семантическая версия).
3) Deploy: по SSH выполнить `docker compose pull && docker compose up -d --remove-orphans` в каталоге `infra/`.
4) Smoke‑тесты: `GET /api/healthz`, открытие главной страницы SPA.

Пример шага деплоя (фрагмент, без секретов):

```yaml
jobs:
  deploy:
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'
    steps:
      - uses: actions/checkout@v4
      - name: Copy files
        uses: appleboy/scp-action@v0.1.7
        with:
          host: ${{ secrets.SSH_HOST }}
          username: ${{ secrets.SSH_USER }}
          key: ${{ secrets.SSH_KEY }}
          source: "infra/*"
          target: "~/idea-frame/infra"
      - name: Deploy
        uses: appleboy/ssh-action@v0.1.10
        with:
          host: ${{ secrets.SSH_HOST }}
          username: ${{ secrets.SSH_USER }}
          key: ${{ secrets.SSH_KEY }}
          script: |
            cd ~/idea-frame/infra
            docker compose pull
            docker compose up -d --remove-orphans
```

Секреты (минимум): `JWT_SECRET`, креды БД (если внешняя), SSH ключи, пароли реестра. На сервере — `.env` рядом с `docker-compose.yml` или экспорт переменных окружения в сервис‑менеджере.

---

## Миграции БД
- Инструмент: `golang-migrate` (контейнер `migrate/migrate`).
- Запускается как одноразовая job перед стартом API (см. `depends_on`).
- Идемпотентность: повторный запуск безопасен, если нет новых миграций.
- Практика изменений: "expand → migrate → contract" (безопасные изменения вперёд, удаление — отдельным релизом).
- Для откатов допускается наличие `down`‑скриптов; применять только если откат схемы совместим с кодом.

---

## Бэкапы и восстановление

### Бэкапы
- Ежедневный `pg_dump` БД на сервере с ротацией (например, 7/30/180 дней).
- Хранение: локальный зашифрованный диск + объектное хранилище (S3/Wasabi) при необходимости.

Пример cron‑задачи на сервере:
```bash
0 2 * * * docker exec -t idea-frame-db pg_dump -U idea -d idea | gzip > /backups/idea-$(date +\%F).sql.gz
```

### Восстановление
```bash
gunzip -c /backups/idea-2025-09-19.sql.gz | docker exec -i idea-frame-db psql -U idea -d idea
```

---

## Наблюдаемость и здоровье
- Health‑эндпоинты API: `/healthz` (liveness), `/readyz` (readiness) — используются в compose.
- Логи: структурированные логи API; логи Caddy. Сбор — сначала через `docker logs`, при росте — стэк логирования.
- Внешний аптайм‑монитор (UptimeRobot/BetterStack) на главную страницу и `/api/healthz`.

---

## Процедуры

### Первый запуск (staging/prod)
1. Подготовить VM: Docker, Docker Compose plugin, открыты порты 80/443.
2. Настроить DNS `A`/`AAAA` на сервер (`CADDY_DOMAIN`).
3. Скопировать каталог `infra/` (compose, Caddyfile) и `.env` (секреты).
4. Выполнить `docker compose pull` (если используются готовые образы).
5. Выполнить `docker compose up -d`.
6. Проверить `https://<domain>/api/healthz` и загрузку SPA.

### Релиз
1. CI собирает и пушит образы с тегами.
2. На сервере: `docker compose pull && docker compose up -d --remove-orphans`.
3. Проверить health и базовые пользовательские сценарии.

### Откат
1. Выбрать предыдущие теги образов (например, `api:<prev-sha>`).
2. Обновить compose/переменные, выполнить `docker compose up -d`.
3. Откат миграций выполнять только при наличии совместимых `down`‑скриптов и понимании рисков.

---

## Безопасность
- Секреты — только в секрет‑хранилище CI/CD и переменных окружения на сервере.
- TLS — автоматически через Caddy (Let’s Encrypt). Регулярно проверять почту на уведомления LE.
- CORS — ограничить домен фронтенда в API.
- Лимиты запросов на аутентификацию (на уровне API или прокси) и троттлинг.
- Обновления базовых образов (Postgres/Caddy) — регулярно.

---

## Файлы инфраструктуры (план)
- `infra/docker-compose.yml` — основной плейбук контейнеров.
- `infra/Caddyfile` — конфигурация прокси и TLS.
- `infra/.env.example` — пример переменных окружения (без секретов).
- `docs/deployment.md` — данный документ.

---

## Критерии готовности к прод‑деплою
- Домен и TLS работают (Caddy выдаёт валидный сертификат).
- Миграции применяются автоматически и идемпотентно.
- API проходит health‑чеки, базовые сценарии выполняются.
- Настроены ежедневные бэкапы и тест восстановление из последнего дампа.
- CI/CD доставляет изменения без ручных шагов, откат проверен.


