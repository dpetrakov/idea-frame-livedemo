## Стратегия развертывания (Docker Compose + Caddy)

Домен публикации: ideaframe.dimlight.online

### Основные положения
- Развёртывание через Docker Compose.
- Caddy выступает обратным прокси с автоматическим TLS (Let’s Encrypt).
- Все секреты и служебные параметры — ТОЛЬКО в файле `.env` (в корне репозитория) + `./.env.example` без секретов.
- Контейнеры `backend` и `frontend` собираются на лету из `app/backend/Dockerfile` и `app/frontend/Dockerfile`.
- Миграции PostgreSQL выполняются автоматически при старте бэкенда (golang-migrate), затем запускается приложение.

---

## 1) Предпосылки и требования
- DNS: A/AAAA записи домена `ideaframe.dimlight.online` указывают на сервер.
- Открыты порты 80 и 443 на сервере (файервол, облачные правила).
- Установлены Docker и Docker Compose.
- Git установлен; деплой осуществляется из корня репозитория.

Опционально: e‑mail для Let’s Encrypt уведомлений — переменная `CADDY_EMAIL`.

---

## 2) Переменные окружения (.env)
Создайте файл `.env` в корне репозитория (рядом с `README.md`). Значения приведены как пример; отредактируйте под себя. Обязательно укажите домен.

```dotenv
# Общее
APP_URL=ideaframe.dimlight.online
CADDY_EMAIL=                     # рекомендовано указать (для уведомлений TLS), например, admin@dimlight.online

# Auth/секреты
JWT_SECRET=change_me_strong_secret

# Postgres
POSTGRES_USER=app
POSTGRES_PASSWORD=change_me_db_password
POSTGRES_DB=app
POSTGRES_HOST=postgres
POSTGRES_PORT=5432
DATABASE_URL=postgres://$POSTGRES_USER:$POSTGRES_PASSWORD@$POSTGRES_HOST:$POSTGRES_PORT/$POSTGRES_DB?sslmode=disable

# Порты сервисов (внутри сети Docker)
BACKEND_PORT=8080
FRONTEND_PORT=3000
```

Рядом положите `.env.example` без секретов (подставьте безопасные заглушки).

---

## 3) Обзор Docker Compose (сервисы)
Планируется 4 сервиса:
- `postgres` — PostgreSQL + volume + healthcheck.
- `backend` — Go‑сервис; при старте выполняет миграции и запускает приложение.
- `frontend` — React; собирается в статические файлы.
- `caddy` — HTTPS reverse proxy: `/api/* → backend`, остальное — `frontend`.

Рекомендуемая структура каталогов для инфраструктуры:
```
infra/
  docker-compose.yml
  caddy/
    Caddyfile
```

---

## 4) Пример docker-compose.yml
Разместите файл в `infra/docker-compose.yml`.

```yaml
services:
  postgres:
    image: postgres:16-alpine
    container_name: postgres
    environment:
      POSTGRES_USER: ${POSTGRES_USER}
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
      POSTGRES_DB: ${POSTGRES_DB}
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U $$POSTGRES_USER"]
      interval: 5s
      timeout: 3s
      retries: 20
    volumes:
      - pgdata:/var/lib/postgresql/data
    networks:
      - app

  backend:
    build:
      context: ..
      dockerfile: app/backend/Dockerfile
    container_name: backend
    depends_on:
      postgres:
        condition: service_healthy
    environment:
      DATABASE_URL: ${DATABASE_URL}
      BACKEND_PORT: ${BACKEND_PORT}
      JWT_SECRET: ${JWT_SECRET}
    expose:
      - "${BACKEND_PORT}"
    networks:
      - app

  frontend:
    build:
      context: ../app/frontend
      dockerfile: Dockerfile
    container_name: frontend
    environment:
      - NODE_ENV=production
    command:
      - caddy
      - file-server
      - --root
      - /usr/share/caddy
      - --listen
      - ":${FRONTEND_PORT}"
    expose:
      - "${FRONTEND_PORT}"
    networks:
      - app

  caddy:
    image: caddy:2-alpine
    container_name: caddy
    ports:
      - "80:80"
      - "443:443"
    environment:
      APP_URL: ${APP_URL}
      BACKEND_PORT: ${BACKEND_PORT}
      FRONTEND_PORT: ${FRONTEND_PORT}
      CADDY_EMAIL: ${CADDY_EMAIL}
    volumes:
      - ./caddy/Caddyfile:/etc/caddy/Caddyfile:ro
      - caddy_data:/data
      - caddy_config:/config
    depends_on:
      - backend
      - frontend
    networks:
      - app

networks:
  app:
    driver: bridge

volumes:
  pgdata:
  caddy_data:
  caddy_config:
```

---

## 5) Конфигурация Caddy
Создайте файл `infra/caddy/Caddyfile` со следующим содержимым. Здесь используются переменные окружения из `.env`:

```caddyfile
{
  email {$CADDY_EMAIL}
}

{$APP_URL} {
  encode gzip

  @api path /api/*
  handle @api {
    reverse_proxy backend:{$BACKEND_PORT}
  }

  handle {
    reverse_proxy frontend:{$FRONTEND_PORT}
  }
}
```

Пояснения:
- `email` — рекомендуем указать `CADDY_EMAIL` для уведомлений Let’s Encrypt.
- Caddy автоматически выпустит и обновит TLS‑сертификат для `APP_URL`.

---

## 6) Dockerfiles (сборка на лету)

### Backend — `app/backend/Dockerfile`
Образ включает бинарь приложения и `migrate`. Рекомендуется запуск через entrypoint‑скрипт, выполняющий миграции перед стартом.

```dockerfile
# syntax=docker/dockerfile:1

FROM golang:1.22-alpine AS build
RUN apk add --no-cache git
WORKDIR /src
COPY app/backend/go.mod app/backend/go.sum ./
RUN go mod download
COPY app/backend/ .
# Собираем приложение
RUN CGO_ENABLED=0 go build -o /out/app ./cmd/server
# Устанавливаем migrate CLI
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

FROM alpine:3.20
RUN apk add --no-cache ca-certificates bash
WORKDIR /app
COPY --from=build /out/app /app/app
COPY --from=build /go/bin/migrate /usr/local/bin/migrate
# Скопируйте миграции в образ (при сборке backend):
COPY db/migrations /app/migrations
ENV GIN_MODE=release
EXPOSE 8080
COPY app/backend/entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh
ENTRYPOINT ["/app/entrypoint.sh"]
```

`app/backend/entrypoint.sh`:
```bash
#!/usr/bin/env bash
set -euo pipefail

echo "[backend] Waiting for database..."
for i in {1..60}; do
  if migrate -path /app/migrations -database "$DATABASE_URL" version >/dev/null 2>&1; then
    break
  fi
  sleep 1
done

echo "[backend] Running migrations..."
migrate -path /app/migrations -database "$DATABASE_URL" up

echo "[backend] Starting app..."
exec /app/app
```

### Frontend — `app/frontend/Dockerfile`

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
EXPOSE 3000
```

---

## 7) Команды
Первый запуск:
```bash
cd infra
docker compose build --no-cache
docker compose up -d
```

Пересборка и рестарт:
```bash
cd infra
docker compose up -d --build
```

Логи:
```bash
cd infra
docker compose logs -f backend
docker compose logs -f caddy
```

Остановка:
```bash
cd infra
docker compose down
```

---

## 8) Healthcheck и smoketest
- Postgres: настраивается в compose через `pg_isready`.
- Backend: проверьте `/api/health` (должен возвращать 200 OK).
- Caddy/Frontend: корень сайта.

Smoketest после старта:
```bash
curl -I https://ideaframe.dimlight.online/
curl -I https://ideaframe.dimlight.online/api/health
```

Оба запроса должны вернуть 200.

---

## 9) Обновления и откаты
- Обновление приложения: `docker compose up -d --build`.
- Откат: переключитесь на предыдущий commit в Git и выполните ту же команду.

---

## 11) Чек‑лист перед продом
- [ ] DNS настроен: `ideaframe.dimlight.online → сервер`.
- [ ] Открыты порты 80/443.
- [ ] Заполнен `.env`, указан `APP_URL=ideaframe.dimlight.online`.
- [ ] `CADDY_EMAIL` указан (рекомендуется).
- [ ] Бэкенд запускает миграции при старте; при ошибке — контейнер падает.
- [ ] Caddy проксирует `/api/*` → `backend`, остальное → `frontend`.
- [ ] Smoketest проходит: `/` и `/api/health` возвращают 200.

---

## Примечания
- Значения секретов в этом документе — примеры; используйте уникальные сильные пароли.
- В проде не публикуйте порты `backend` и `frontend` наружу — только через `caddy` (80/443).
- Если у сервера нет публичного IPv6, проверьте DNS AAAA-записи; при необходимости удалите.


