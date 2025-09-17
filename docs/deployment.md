---
title: Стратегия развертывания — IdeaFrame
status: draft
lastUpdated: 2025-09-17
dependsOn: docs/architecture.md, docs/openapi.yaml
appUrl: ideaframe.dimlight.online
---

## Обзор

Цель: развернуть приложение IdeaFrame на домене ideaframe.dimlight.online с автоматическими TLS‑сертификатами и маршрутизацией через Caddy. Развёртывание выполняется с помощью Docker Compose, все секреты и переменные — только в файле `.env` в корне репозитория.

Ключевые решения:
- **Docker Compose** управляет сервисами: `postgres`, `backend`, `frontend`, `caddy`.
- **Caddy** завершает TLS (Let's Encrypt) и проксирует:
  - `https://ideaframe.dimlight.online/api/* → backend`;
  - весь остальной трафик → `frontend`.
- **Миграции БД** запускаются автоматически из контейнера `backend` (через `migrate`) при старте.

## Предварительные требования

- Сервер с публичным IP (Linux, Docker Engine 25+, Docker Compose V2).
- DNS‑запись для `ideaframe.dimlight.online`:
  - `A` (и при необходимости `AAAA`) указывает на публичный IP сервера.
- Открытые порты: `80/tcp` и `443/tcp` (входящий трафик к Caddy).

## Переменные окружения (.env)

Файл `.env` хранится в корне репозитория. Пример шаблона (`.env.example`):

```dotenv
APP_URL=ideaframe.dimlight.online
CADDY_EMAIL=admin@example.com

JWT_SECRET=change_me

POSTGRES_USER=idea
POSTGRES_PASSWORD=change_me
POSTGRES_DB=idea
POSTGRES_HOST=postgres
POSTGRES_PORT=5432

# Явная строка подключения; не используйте вложенные переменные
DATABASE_URL=postgres://idea:change_me@postgres:5432/idea?sslmode=disable

BACKEND_PORT=8080
FRONTEND_PORT=80
```

Запуск команд `docker compose` — всегда из корня репозитория с явной передачей `.env` и файла Compose:

```bash
docker compose --env-file ./.env -f infra/docker-compose.yml up -d
```

## Docker Compose (infra/docker-compose.yml)

Compose файл управляет сборкой образов и запуском контейнеров. Не используйте устаревший ключ `version`.

```yaml
services:
  postgres:
    image: postgres:16-alpine
    container_name: idea-postgres
    environment:
      POSTGRES_USER: ${POSTGRES_USER}
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
      POSTGRES_DB: ${POSTGRES_DB}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U $$POSTGRES_USER -d $$POSTGRES_DB"]
      interval: 5s
      timeout: 3s
      retries: 10
    restart: unless-stopped

  backend:
    build:
      context: ../app/backend
      dockerfile: Dockerfile
    container_name: idea-backend
    depends_on:
      postgres:
        condition: service_healthy
    environment:
      DATABASE_URL: ${DATABASE_URL}
      JWT_SECRET: ${JWT_SECRET}
    volumes:
      - ../db/migrations:/migrations:ro
    command: sh -c "migrate -path /migrations -database ${DATABASE_URL} up && /app/app"
    expose:
      - "${BACKEND_PORT}"
    restart: unless-stopped

  frontend:
    build:
      context: ../app/frontend
      dockerfile: Dockerfile
    container_name: idea-frontend
    expose:
      - "${FRONTEND_PORT}"
    restart: unless-stopped

  caddy:
    image: caddy:2-alpine
    container_name: idea-caddy
    depends_on:
      - backend
      - frontend
    ports:
      - "80:80"
      - "443:443"
    environment:
      APP_URL: ${APP_URL}
      CADDY_EMAIL: ${CADDY_EMAIL}
      BACKEND_PORT: ${BACKEND_PORT}
      FRONTEND_PORT: ${FRONTEND_PORT}
    volumes:
      - ./caddy/Caddyfile:/etc/caddy/Caddyfile:ro
      - caddy_data:/data
      - caddy_config:/config
    restart: unless-stopped

volumes:
  postgres_data:
  caddy_data:
  caddy_config:
```

Пути `context` и монтирование `../db/migrations` указаны относительно каталога `infra/`.

## Caddyfile (infra/caddy/Caddyfile)

Обратный прокси с автоматическим TLS и маршрутизацией.

```caddyfile
{
  email {$CADDY_EMAIL}
}

{$APP_URL} {
  encode gzip

  @api path /api/*
  handle @api {
    reverse_proxy backend:${BACKEND_PORT}
  }

  handle {
    reverse_proxy frontend:${FRONTEND_PORT}
  }
}
```

## Dockerfiles

Backend (`app/backend/Dockerfile`) — многостадийная сборка, включает `migrate`:

```dockerfile
# syntax=docker/dockerfile:1

FROM golang:1.22-alpine AS build
WORKDIR /src
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/app ./cmd/server
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=build /out/app /app/app
COPY --from=build /go/bin/migrate /usr/local/bin/migrate
ENV GIN_MODE=release
EXPOSE 8080
ENTRYPOINT ["/app/app"]
```

Frontend (`app/frontend/Dockerfile`) — сборка и статика, отдаётся самим контейнером (порт 80):

```dockerfile
# syntax=docker/dockerfile:1

FROM node:20-alpine AS build
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM caddy:2-alpine
COPY --from=build /app/build /usr/share/caddy
EXPOSE 80
```

## Процедуры запуска и обновления

Первый запуск:

```bash
docker compose --env-file ./.env -f infra/docker-compose.yml build --no-cache
docker compose --env-file ./.env -f infra/docker-compose.yml up -d
```

Обновление с пересборкой:

```bash
git pull
docker compose --env-file ./.env -f infra/docker-compose.yml up -d --build
```

Остановка:

```bash
docker compose --env-file ./.env -f infra/docker-compose.yml down
```

## Healthchecks и smoketest

Проверка состояния контейнеров:

```bash
docker compose --env-file ./.env -f infra/docker-compose.yml ps
docker compose --env-file ./.env -f infra/docker-compose.yml logs -f caddy
docker compose --env-file ./.env -f infra/docker-compose.yml logs -f backend
```

Smoketest внешнего контура:

```bash
curl -fsSL https://ideaframe.dimlight.online/api/v1/health
```

Ожидаемый ответ:

```json
{ "status": "ok" }
```

## Роллбек

- Если обновление некорректно: `git checkout <предыдущий_тэг_или_commit>` и повторить `up -d --build`.
- При необходимости вернуть данные БД — восстановить из резервной копии (см. ниже).

## Резервное копирование БД

Экспорт:

```bash
docker compose --env-file ./.env -f infra/docker-compose.yml exec -T postgres \
  pg_dump -U "$POSTGRES_USER" -d "$POSTGRES_DB" -F t > backup_$(date +%Y%m%d_%H%M%S).tar
```

Импорт:

```bash
cat backup_YYYYMMDD_HHMMSS.tar | \
docker compose --env-file ./.env -f infra/docker-compose.yml exec -T postgres \
  pg_restore -U "$POSTGRES_USER" -d "$POSTGRES_DB" --clean
```

## Безопасность и эксплуатация

- Все секреты только в `.env`; не коммитить `.env` в VCS. Коммитить `.env.example` без значений.
- Минимальные привилегии БД, надёжные пароли, ротация секретов.
- Ограничить SSH‑доступ, настроить автоматические обновления ОС/патчей.

## Чек‑лист

- [ ] DNS‑запись `ideaframe.dimlight.online` указывает на сервер.
- [ ] `APP_URL=ideaframe.dimlight.online` в `.env`.
- [ ] Аккаунт для LE (`CADDY_EMAIL`) указан.
- [ ] Все секреты заполнены в `.env`; `.env.example` добавлен в репозиторий.
- [ ] Миграции доступны в `db/migrations`; backend применяет их при старте.
- [ ] `Caddyfile` проксирует `/api/*` на backend, остальное — на frontend.
- [ ] Smoketest `/api/v1/health` возвращает `{"status":"ok"}` по HTTPS.

## Примечания по архитектуре

Изначально в архитектуре миграции упоминались как отдельная job, однако для упрощения эксплуатации миграции выполняются внутри контейнера `backend` при старте (через `migrate`), что соответствует принятой стратегиe развертывания.


