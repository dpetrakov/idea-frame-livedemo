---
title: Архитектура — IdeaFrame
status: draft
lastUpdated: 2025-09-17
dependsOn: docs/prd.md
---

## Обзор и цели

Цель системы — поддержать сбор инициатив, их обсуждение и приоритизацию через единый веб‑клиент. Основные сценарии: создание инициатив с markdown‑описанием, чат‑комментарии, оценка атрибутов ценность скорость стоимость, назначение ответственного, фильтрация и сортировка по вычисленному весу. Архитектура ориентирована на простоту, предсказуемость и быструю реализацию MVP.

## Технологический стек
- Backend: Go сервис
- База данных: PostgreSQL
- Миграции: golang migrate как job в docker compose
- Frontend: React приложение
- Прокси и TLS: Caddy обратный прокси
- Аутентификация: JWT срок 24 часа

## Диаграмма компонентов

```mermaid
flowchart TD
    Browser[Browser]
    Caddy[Caddy TLS proxy]
    Frontend[React app]
    Backend[Go API service]
    DB[PostgreSQL]
    Migrate[Migration job]
    Obs[Logs and health]

    Browser <--> Caddy
    Caddy --> Frontend
    Caddy --> Backend
    Frontend <--> Backend
    Backend --> DB
    Migrate --> DB
    Backend --> Obs
    Caddy --> Obs
```

Комментарии:
- React app отдает статический бандл через Caddy.
- Общение клиента и сервера по HTTPS через Caddy.
- Аутентификация между Frontend и Backend через заголовок Authorization с JWT.
- Миграции запускаются как отдельная задача, применяются к PostgreSQL до старта сервиса.

## Основные потоки

### Регистрация и вход
```mermaid
sequenceDiagram
    actor User
    participant Browser
    participant Frontend as Frontend app
    participant Backend as Backend API
    participant DB as Database

    User->>Browser: Open login page
    Browser->>Frontend: Submit registration or login
    Frontend->>Backend: Auth request
    Backend->>DB: Select user by login
    DB-->>Backend: User data with hash
    Backend-->>Frontend: JWT token ttl 24h
    Frontend-->>Browser: Store token and redirect
```

### Создание инициативы
```mermaid
sequenceDiagram
    actor User
    participant Browser
    participant Frontend as Frontend app
    participant Backend as Backend API
    participant DB as Database

    User->>Browser: Fill initiative form
    Browser->>Frontend: Submit form
    Frontend->>Backend: Create initiative
    Backend->>DB: Insert initiative
    DB-->>Backend: Insert ok
    Backend-->>Frontend: Initiative created
    Frontend-->>Browser: Navigate to details
```

### Обновление атрибутов и пересчет веса
```mermaid
sequenceDiagram
    actor User
    participant Browser
    participant Frontend as Frontend app
    participant Backend as Backend API
    participant DB as Database

    User->>Browser: Set value velocity cost
    Browser->>Frontend: Submit attributes
    Frontend->>Backend: Update attributes
    Backend->>Backend: Compute weight per PRD formula
    Backend->>DB: Update attributes and weight
    DB-->>Backend: Update ok
    Backend-->>Frontend: Initiative updated
    Frontend-->>Browser: Update list order
```

### Комментарии с лонг поллинг
```mermaid
sequenceDiagram
    actor User
    participant Browser
    participant Frontend as Frontend app
    participant Backend as Backend API
    participant DB as Database

    User->>Browser: Type comment and send
    Browser->>Frontend: Submit comment
    Frontend->>Backend: Create comment
    Backend->>DB: Insert comment
    DB-->>Backend: Insert ok
    Backend-->>Frontend: Comment accepted
    Frontend-->>Browser: Show new message

    loop Long poll
        Browser->>Frontend: Fetch new comments since ts
        Frontend->>Backend: Get comments by ts
        Backend->>DB: Select comments by ts
        DB-->>Backend: Rows
        Backend-->>Frontend: List of comments
        Frontend-->>Browser: Append messages
    end
```

### Список инициатив, фильтры и сортировка
```mermaid
sequenceDiagram
    participant Browser
    participant Frontend as Frontend app
    participant Backend as Backend API
    participant DB as Database

    Browser->>Frontend: Load list state
    Frontend->>Backend: Get initiatives with filter
    Backend->>DB: Query order by weight desc nulls last
    DB-->>Backend: Rows
    Backend-->>Frontend: Page of initiatives
    Frontend-->>Browser: Render cards
```

## Данные и модель

Сущности и поля определены в PRD. Логическая модель данных будет отражена в файле db/schema.dbml и создана позже. Ключевые моменты реализации:
- Уникальный индекс по полю login.
- Ссылочная целостность для связей authorId assigneeId initiativeId.
- Диапазоны для атрибутов value velocity cost строго 1 до 5.
- Поле weight хранится как кеш и пересчитывается на сервере при изменении атрибутов.
- Список сортируется по weight по убыванию, элементы без weight в конце, затем по createdAt по убыванию.

## API контур

Подробная спецификация будет оформлена в docs/openapi.yaml. На уровне архитектуры предусмотрены ресурсы: auth users initiatives comments. Операции включают регистрацию и вход, создание инициатив, обновление атрибутов и ответственного, создание комментариев, выборки со списками и фильтрами.

## Безопасность
- Хранение паролей только в виде хеша с солью, рекомендуемый алгоритм bcrypt с достаточной стоимостью.
- JWT срок 24 часа, формат Bearer в заголовке Authorization, проверка на всех защищенных маршрутах.
- CORS настроен на домены фронтенда, методы и заголовки ограничены необходимым минимумом.
- Ввод валидируется на клиенте и на сервере. Ошибки возвращаются в явном виде без утечки лишней информации.
- Роли в текущем объеме единые, авторизация простая на уровне принадлежности сущностей текущему пользователю там где применимо.

## Наблюдаемость
- Логирование ошибок и ключевых событий на стороне Backend с уровнем info warn error.
- Health чек endpoint для статуса приложения и доступности базы данных.
- Базовые метрики запросов и ошибок на уровне прокси и Backend, при необходимости вывод в stdout и интеграция с внешними средствами.

## Архитектурные решения и компромиссы
- Реал тайм чат реализуется лонг поллинг для упрощения, WebSocket вне текущего объема.
- Вес инициативы хранится в поле weight для быстрой сортировки, источник истины вычисляется на сервере при записи атрибутов.
- Сортировка по весу выполняется на стороне базы данных, что упрощает реализацию и повышает предсказуемость.
- Одна роль пользователя снижает сложность авторизации.
- Прокси Caddy используется для TLS и маршрутизации, что упрощает выпуск сертификатов и скрывает внутренний контур.

## Границы и допущения
- Функциональность строго по PRD. Расширенные уведомления и полноценные пуш механизмы вне текущего объема.
- Хранение файлов и вложений не предусмотрено.
- Масштабирование горизонтальное не закладывается на этапе MVP, при необходимости Backend может быть масштабирован и вынесен state из памяти.

## Следующие шаги
1) Подготовить логическую модель данных в db/schema.dbml.
2) Описать API в docs/openapi.yaml.
3) Реализовать сервисы и фронтенд согласно документации.


