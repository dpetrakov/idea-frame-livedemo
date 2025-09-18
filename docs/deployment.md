# Стратегия развертывания: Система фрейминга портфеля инициатив

Документ описывает стратегию развертывания для `ideaframe.dimlight.online` с использованием Docker Compose, PostgreSQL, Caddy для TLS и обратного прокси.

## Архитектура развертывания

- **Backend**: Go монолит с HTTP API
- **Frontend**: React SPA (мобильный first)
- **База данных**: PostgreSQL с автоматическими миграциями
- **Proxy/TLS**: Caddy с автоматическими Let's Encrypt сертификатами
- **Контейнеризация**: Docker Compose для оркестрации
- **Миграции**: golang-migrate встроен в backend контейнер

## Сервисы

### 1. postgres
- Образ: `postgres:15-alpine`
- Постоянное хранение в volume `postgres_data`
- Healthcheck для проверки готовности
- Переменные: `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB`

### 2. backend  
- Собирается из `app/backend/Dockerfile`
- Включает golang-migrate для автоматических миграций
- Порт 8080, эндпоинты `/api/*`
- Зависит от postgres (depends_on + condition: service_healthy)
- Entrypoint: миграции → запуск приложения

### 3. frontend
- Собирается из `app/frontend/Dockerfile` 
- React SPA, статическая выдача через Caddy
- Порт 3000

### 4. caddy
- TLS терминация с автоматическими сертификатами
- Проксирование: `/api/*` → backend, остальное → frontend
- Конфигурация из `infra/caddy/Caddyfile`
- Публичный порт 443 (HTTPS)

## Переменные окружения

Файл `.env` в корне репозитория (не коммитится):

```bash
# Целевой домен
APP_URL=ideaframe.dimlight.online

# Email для Let's Encrypt сертификатов
CADDY_EMAIL=admin@dimlight.online

# JWT секрет для аутентификации
JWT_SECRET=super-secret-jwt-key-change-in-production

# PostgreSQL конфигурация  
POSTGRES_USER=ideaframe_user
POSTGRES_PASSWORD=secure-db-password-change-me
POSTGRES_DB=ideaframe_db
POSTGRES_HOST=postgres
POSTGRES_PORT=5432

# Строка подключения к БД (используется backend и миграциями)
DATABASE_URL=postgres://ideaframe_user:secure-db-password-change-me@postgres:5432/ideaframe_db?sslmode=disable

# Порты сервисов
BACKEND_PORT=8080
FRONTEND_PORT=3000
```

## Структура файлов инфраструктуры

```
infra/
├── docker-compose.yml          # Оркестрация сервисов
└── caddy/
    └── Caddyfile              # Конфигурация прокси и TLS

app/
├── backend/
│   └── Dockerfile             # Сборка Go API с migrate
└── frontend/
    └── Dockerfile             # Сборка React SPA

.env.example                   # Шаблон переменных окружения
.env                          # Локальные переменные (не в VCS)
```

## Команды развертывания

### Первоначальный деплой

```bash
# 1. Клонировать репозиторий
git clone <repository-url>
cd idea-frame-livedemo

# 2. Настроить переменные окружения
cp .env.example .env
# Отредактировать .env с реальными значениями

# 3. Сборка и запуск
docker compose --env-file ./.env -f infra/docker-compose.yml build --no-cache
docker compose --env-file ./.env -f infra/docker-compose.yml up -d
```

### Обновление приложения

```bash
# Пересборка и рестарт с новым кодом
docker compose --env-file ./.env -f infra/docker-compose.yml up -d --build
```

### Мониторинг и отладка

```bash
# Логи всех сервисов
docker compose --env-file ./.env -f infra/docker-compose.yml logs -f

# Логи конкретного сервиса
docker compose --env-file ./.env -f infra/docker-compose.yml logs -f backend
docker compose --env-file ./.env -f infra/docker-compose.yml logs -f caddy

# Статус сервисов
docker compose --env-file ./.env -f infra/docker-compose.yml ps
```

### Остановка

```bash
# Остановка без удаления данных
docker compose --env-file ./.env -f infra/docker-compose.yml down

# Полная очистка включая volumes (ВНИМАНИЕ: удалит данные БД)
docker compose --env-file ./.env -f infra/docker-compose.yml down -v
```

## Автоматические миграции БД

Backend контейнер включает golang-migrate и выполняет миграции автоматически:

1. При старте контейнера запускается entrypoint скрипт
2. Выполняется `migrate up` с миграциями из `db/migrations/`  
3. При успехе запускается Go приложение
4. При ошибке миграции контейнер завершается с ошибкой

Миграции располагаются в `db/migrations/` в формате:
- `001_initial.up.sql` / `001_initial.down.sql`
- `002_add_comments.up.sql` / `002_add_comments.down.sql`

## Безопасность

### TLS/HTTPS
- Автоматические Let's Encrypt сертификаты через Caddy
- Принудительное перенаправление HTTP → HTTPS
- HSTS заголовки

### Аутентификация  
- JWT токены со сроком жизни 24 часа
- Секретный ключ в переменной `JWT_SECRET`
- HTTP-Only cookies или Authorization заголовки

### Сеть
- Внутренняя Docker сеть между сервисами
- Только Caddy экспонирует публичные порты (80, 443)
- PostgreSQL доступна только внутри сети

### Переменные окружения
- Все секреты только в `.env` (исключен из VCS)
- `.env.example` содержит шаблон без секретных значений

## Health проверки

### Backend
- Эндпоинт: `GET /api/health`
- Проверяет подключение к БД
- Возвращает статус сервиса

### PostgreSQL  
- Docker healthcheck с `pg_isready`
- Backend ждет готовности БД перед стартом

### Caddy
- Встроенный health endpoint
- Проверка статуса upstream сервисов

## Мониторинг и логи

### Структурированные логи
- JSON формат для всех сервисов
- Корреляция запросов по request_id
- Уровни логирования: ERROR, WARN, INFO, DEBUG

### Метрики
- HTTP запросы: количество, статус коды, время ответа
- Бизнес метрики: создание инициатив, комментарии, оценки
- Системные метрики: использование ресурсов

### Сбор логов
- Логи Docker контейнеров доступны через `docker compose logs`
- Ротация логов настроена в Docker daemon
- Опционально: экспорт в внешние системы (ELK, Grafana)

## Резервное копирование

### База данных
```bash
# Создание backup
docker compose --env-file ./.env -f infra/docker-compose.yml exec postgres \
  pg_dump -U ideaframe_user ideaframe_db > backup_$(date +%Y%m%d_%H%M%S).sql

# Восстановление из backup  
docker compose --env-file ./.env -f infra/docker-compose.yml exec -T postgres \
  psql -U ideaframe_user ideaframe_db < backup_file.sql
```

### Volumes
- Основные данные в `postgres_data` volume
- Backup volumes через `docker volume` команды

## Масштабирование

### Вертикальное масштабирование
- Увеличение CPU/RAM для Docker контейнеров
- Настройка в `docker-compose.yml` через `deploy.resources`

### Горизонтальное масштабирование  
- Для демо не требуется
- В будущем: load balancer, репликация БД, multiple backend instances

## Troubleshooting

### Типичные проблемы

1. **Сертификаты не выпускаются**
   - Проверить домен указывает на сервер
   - Проверить доступность портов 80/443
   - Логи Caddy: `docker compose logs caddy`

2. **Backend не стартует**  
   - Проверить подключение к БД
   - Логи миграций в логах backend
   - Проверить `DATABASE_URL`

3. **Frontend недоступен**
   - Проверить сборку React приложения  
   - Проверить Caddy конфигурацию
   - Логи frontend контейнера

### Smoketest после деплоя

```bash
# Проверка HTTPS доступности
curl -I https://ideaframe.dimlight.online

# Проверка API health
curl https://ideaframe.dimlight.online/api/health

# Проверка перенаправления HTTP -> HTTPS  
curl -I http://ideaframe.dimlight.online
```

## Обновления

### Rolling updates
- Обновление кода через `git pull` + `docker compose up -d --build`
- Zero-downtime через healthchecks и graceful shutdown
- Откат через `git checkout` предыдущего коммита

### Миграции БД
- Всегда backwards compatible миграции
- Тестирование на staging окружении
- Backup БД перед применением

## Требования к серверу

### Минимальные требования
- OS: Ubuntu 20.04+ / CentOS 8+ / Debian 11+
- CPU: 2 vCPU  
- RAM: 4 GB
- Storage: 20 GB SSD
- Network: публичный IP, порты 80/443

### Рекомендуемые требования
- CPU: 4 vCPU
- RAM: 8 GB  
- Storage: 50 GB SSD
- Мониторинг и алерты

---

## Чек-лист готовности к деплою

- [x] Целевой URL определен: `ideaframe.dimlight.online`
- [ ] `.env` файл настроен с реальными секретами
- [ ] DNS записи указывают на сервер
- [ ] Docker и Docker Compose установлены
- [ ] Файлы инфраструктуры созданы:
  - [ ] `infra/docker-compose.yml`
  - [ ] `infra/caddy/Caddyfile`  
  - [ ] `app/backend/Dockerfile`
  - [ ] `app/frontend/Dockerfile`
  - [ ] `.env.example`
- [ ] Миграции БД подготовлены в `db/migrations/`
- [ ] Smoke test после деплоя выполнен успешно
