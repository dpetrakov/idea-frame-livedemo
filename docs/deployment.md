# MeetAx Next: стратегия развертывания (Docker Compose + Caddy)

Целевой домен: meetax.example.com

Документ описывает единый способ развёртывания для локальной разработки и продакшн‑сервера. Конфиги едины, различаются только значения переменных в файле .env.

## Архитектура окружения

- Caddy (HTTPS, reverse proxy)
- Backend (Go, API `/api/v1/*`, алиас `/api/health`)
- Frontend (React, статическая выдача)
- PostgreSQL (данные в volume)

Роутинг:
- `https://meetax.example.com/api/*` → backend
- Остальные запросы → frontend

## DNS и сеть

- Создайте DNS A/AAAA записи на IP сервера для `meetax.example.com`.
- Откройте на сервере входящие порты 80 и 443 (HTTPS через Caddy / ACME).

## Переменные окружения (.env)

Хранится в корне репозитория: `./.env`. Пример содержимого:

```
APP_URL=meetax.example.com
CADDY_EMAIL=you@example.com

# PostgreSQL
POSTGRES_USER=idea
POSTGRES_PASSWORD=strong_password
POSTGRES_DB=idea
POSTGRES_HOST=postgres
POSTGRES_PORT=5432

# App
BACKEND_PORT=8080
FRONTEND_PORT=3000
JWT_SECRET=change_me
JWT_EXPIRATION=168h  # 7 дней для passwordless входа

# Email verification / SMTP (предпочтительно)
AXENIX_EMAIL_DOMAIN=axenix.pro
EMAIL_CODES_TTL_MINUTES=10
SMTP_HOST=mail.dimlight.online
SMTP_PORT=587
SMTP_USERNAME=noreply@dimlight.online
SMTP_PASSWORD=***_set_in_vault_***
SMTP_FROM=noreply@dimlight.online
SMTP_TLS_SERVER_NAME=hosting.reg.ru   # если TLS сертификат не на SMTP_HOST
SMTP_TLS_INSECURE_SKIP_VERIFY=false   # true только для отладки

# Mailgun удалён. Используем только SMTP

# Полная строка подключения (явно)
DATABASE_URL=postgres://idea:strong_password@postgres:5432/idea?sslmode=disable

# Роли админа (V2)
# Список e-mail через запятую, без пробелов; сравнение в нижнем регистре
# Пример:
ADMIN_EMAILS=admin1@axenix.pro,admin2@axenix.pro
```

Не коммитьте `.env`. Поддерживайте `.env.example` без секретов.

## Сборка и запуск

Команды выполняются из корня репозитория:

```bash
# Первый запуск/обновление
docker compose --env-file ./.env -f infra/docker-compose.yml up -d --build

# Просмотр состояния
docker compose --env-file ./.env -f infra/docker-compose.yml ps

# Логи
docker compose --env-file ./.env -f infra/docker-compose.yml logs -f caddy | cat
docker compose --env-file ./.env -f infra/docker-compose.yml logs -f backend | cat

# Остановка
docker compose --env-file ./.env -f infra/docker-compose.yml down
```

Перед первым продакшн‑запуском убедитесь, что в каталоге `app/frontend/` установлены зависимости для корректной сборки:

```bash
cd app/frontend && npm install && cd ../..
```

## Проверки после запуска (smoketest)

```bash
# Caddy должен быть доступен на 80/443 (внешний доступ)
curl -I http://meetax.example.com | cat
curl -I https://meetax.example.com | cat

# Health чек API
curl https://meetax.example.com/api/health | cat

# OpenAPI сервер из документа
# Основной префикс API в системе: /api/v1
curl https://meetax.example.com/api/v1/health | cat || true
```

Ожидаем 200 и JSON со статусом.

## Caddy конфигурация

Файл: `infra/caddy/Caddyfile` (универсальный для dev/prod). Используется переменная `APP_URL=meetax.example.com` из `.env`, автоматический выпуск сертификатов Let's Encrypt.

Ключевые моменты:
- Заголовки безопасности: HSTS, CSP, X-Frame-Options, X-Content-Type-Options, Referrer-Policy.
- Принудительный редирект HTTP → HTTPS.
- Прокси `/api/*` на `backend:${BACKEND_PORT}`, остальное на `frontend:${FRONTEND_PORT}`.

## Docker Compose

Файл: `infra/docker-compose.yml`.

- `postgres`: БД, volume `postgres_data`, healthcheck `pg_isready`.
- `backend`: собирается из `app/backend/Dockerfile`; healthcheck по `GET /api/health` на 8080.
- `frontend`: собирается из `app/frontend/Dockerfile`; раздаёт статику на 3000.
- `caddy`: image `caddy:2-alpine`, монтирует `infra/caddy/Caddyfile`, открывает 80/443.

## Обновления и релизы

```bash
# Пересборка и быстрый деплой
docker compose --env-file ./.env -f infra/docker-compose.yml up -d --build
```

## Типовые проблемы

- Сертификат не выдаётся: проверьте DNS, открытые порты 80/443, `CADDY_EMAIL`.
- Caddy unhealthy: смотрите логи сервиса caddy.
- Frontend не собирается: установите зависимости `app/frontend/package-lock.json` и выполните `npm install`.

## Чек‑лист

- [x] `APP_URL=meetax.example.com` задан в `.env`.
- [x] `CADDY_EMAIL` заполнен (настоящий e-mail для ACME).
- [x] Порты 80 и 443 доступны снаружи.
- [x] PostgreSQL доступен по `DATABASE_URL`.
- [x] `JWT_SECRET` установлен.
- [x] TTL JWT: 7 дней (env `JWT_EXPIRATION=168h` либо значение по умолчанию в конфиге)
- [x] Запросы `/api/*` идут в backend; остальное в frontend.
- [x] `https://meetax.example.com/api/health` возвращает 200.


