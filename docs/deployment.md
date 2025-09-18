# Стратегия развертывания: Idea Frame Livedemo

Целевой домен: ideaframe.dimlight.online

Цель: развернуть демо‑приложение (Go API + PostgreSQL + React SPA) на одном хосте с использованием Docker Compose и Caddy (TLS и обратный прокси). Конфигурация должна быть одинаковой для prod и dev, отличаться только `.env`.

## Архитектура развертывания

- Обратный прокси и TLS: Caddy (автоматические сертификаты Let's Encrypt)
- Backend: монолит на Go, HTTP API
- БД: PostgreSQL (персистентный volume)
- Frontend: React SPA, статическая выдача
- Миграции: golang‑migrate как отдельная job в Docker Compose (см. Примечание о вариантах)

Трафик: `https://ideaframe.dimlight.online` → Caddy →
- `/api/*` → backend:8080
- остальные пути → frontend:3000

## Предпосылки и требования к среде

1) DNS
- A/AAAA запись для `ideaframe.dimlight.online` указывает на публичный IP сервера.

2) Сеть и firewall
- Открыты входящие TCP порты 80 и 443.

3) ПО на сервере
- Docker Engine + Docker Compose v2

Установка (Ubuntu 22.04+):
```bash
sudo apt-get update
sudo apt-get install -y ca-certificates curl gnupg lsb-release
sudo install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
  $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
sudo apt-get update
sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
sudo usermod -aG docker $USER
newgrp docker
docker compose version | cat
```

## Конфигурация через `.env`

Файл `.env` хранится в корне репозитория и не коммитится. Шаблон — `.env.example`.

Ключевые переменные:
```
APP_URL=ideaframe.dimlight.online
CADDY_EMAIL=
JWT_SECRET=
POSTGRES_USER=
POSTGRES_PASSWORD=
POSTGRES_DB=
POSTGRES_HOST=postgres
POSTGRES_PORT=5432
DATABASE_URL=postgres://USER:PASSWORD@HOST:5432/DB?sslmode=disable
BACKEND_PORT=8080
FRONTEND_PORT=3000
```

Рекомендации:
- Используйте достаточно длинный `JWT_SECRET` (не менее 32 байт).
- Для `DATABASE_URL` указывайте полную строку, не ссылаясь на другие переменные.

## Docker Compose (план)

Сервисы:
- `postgres`: официальное изображение PostgreSQL, volume `pgdata`, healthcheck.
- `backend`: образ Go приложения; зависит от `postgres`.
- `frontend`: образ React SPA (статическая выдача); внутренний порт 3000.
- `caddy`: внешний вход; проксирует `/api/*` на backend и остальное на frontend; выпускает TLS.
- `migrate`: одноразовая job с `golang-migrate`, применяет миграции к БД.

Особенности:
- Хранение данных БД в именованном volume.
- Конфигурация Caddy берёт домен и почту из `.env`.

Файлы инфраструктуры (будут добавлены отдельно):
- `infra/docker-compose.yml`
- `infra/caddy/Caddyfile`

Запуск команд всегда из корня проекта с явным указанием `.env` и compose‑файла:
```bash
docker compose --env-file ./.env -f infra/docker-compose.yml build --no-cache
docker compose --env-file ./.env -f infra/docker-compose.yml up -d
```

Обновление:
```bash
docker compose --env-file ./.env -f infra/docker-compose.yml up -d --build
```

Логи (пример):
```bash
docker compose --env-file ./.env -f infra/docker-compose.yml logs -f caddy | cat
docker compose --env-file ./.env -f infra/docker-compose.yml logs -f backend | cat
```

Остановка:
```bash
docker compose --env-file ./.env -f infra/docker-compose.yml down
```

## Caddy (маршрутизация и TLS)

Правила:
- Всегда HTTPS с жёстким редиректом HTTP → HTTPS.
- Прод: автоматический выпуск сертификатов Let's Encrypt (ACME) с `CADDY_EMAIL`.
- Базовые заголовки безопасности: HSTS, X-Content-Type-Options, X-Frame-Options, Referrer-Policy, CSP.

## Миграции БД

Выбран целевой вариант по архитектуре: отдельная job `migrate` (одиночный запуск на деплой). Альтернативно допустим запуск `migrate` в entrypoint контейнера `backend` (вариант из правила deploy‑rules). Выбор варианта не меняет остальную конфигурацию.

Рекомендации:
- При первом деплое запустите job вручную либо автоматически через `depends_on` и условие `service_completed_successfully`.
- При ошибке миграции деплой должен считаться неуспешным.

## Чек‑лист продакшн‑готовности

- APP_URL явно установлен: `ideaframe.dimlight.online`.
- DNS указывает на сервер, порты 80/443 открыты.
- Все секреты и URL — только в `.env`.
- Миграции выполняются до старта приложения; при ошибке — остановка.
- Caddy проксирует `/api/*` на backend, остальное — на frontend.
- Включён редирект HTTP → HTTPS.
- Заголовки безопасности выставлены (HSTS, CSP, X‑Frame‑Options, X‑Content‑Type‑Options, Referrer‑Policy).
- Указан `CADDY_EMAIL`, сертификаты выпускаются автоматически.

## Смоук‑тест после деплоя

1) Главная страница SPA
```bash
curl -I https://ideaframe.dimlight.online | cat
# Ожидаем: 200/304, заголовки hsts, csp; cert валидный
```

2) Редирект HTTP → HTTPS
```bash
curl -I http://ideaframe.dimlight.online | cat
# Ожидаем: 301 на https
```

3) API health (примерный эндпоинт)
```bash
curl -i https://ideaframe.dimlight.online/api/health | cat
# Ожидаем: 200 и JSON с состоянием
```

4) Доступ к БД (внутри контейнера backend)
```bash
docker compose --env-file ./.env -f infra/docker-compose.yml exec backend /bin/sh -lc 'echo ok'
```

## Процедуры

Бэкап БД (пример с pg_dump):
```bash
docker compose --env-file ./.env -f infra/docker-compose.yml exec postgres \
  pg_dump -U "$POSTGRES_USER" -d "$POSTGRES_DB" -F c -f /var/lib/postgresql/data/backup.dump
```

Ролбэк контейнеров:
```bash
docker compose --env-file ./.env -f infra/docker-compose.yml down
docker compose --env-file ./.env -f infra/docker-compose.yml up -d --build
```

## Примечания

- Выбор способа миграций: по архитектуре — отдельная job; альтернативный вариант — миграции в entrypoint `backend`. Подтвердите желаемый режим — инфраструктурные файлы будут подготовлены в соответствии с выбором.
- Конфиги едины для prod и dev; отличие только в `.env` (например, `APP_URL=app.localhost` в dev).


