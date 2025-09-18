# syntax=docker/dockerfile:1

FROM golang:1.23-alpine AS build

WORKDIR /src

# Устанавливаем git для загрузки зависимостей
RUN apk add --no-cache git ca-certificates

# Копируем go.mod и go.sum для кеширования зависимостей
COPY app/backend/go.mod app/backend/go.sum ./
RUN go mod download

# Копируем исходный код
COPY app/backend/ .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o /out/app ./cmd/server

# Устанавливаем golang-migrate CLI с поддержкой PostgreSQL
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Финальный образ
FROM alpine:3.20

# Устанавливаем необходимые пакеты
RUN apk add --no-cache ca-certificates tzdata

# Создаём непривилегированного пользователя
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

WORKDIR /app

# Копируем приложение и migrate из builder образа
COPY --from=build /out/app /app/app
COPY --from=build /go/bin/migrate /usr/local/bin/migrate

# Копируем миграции
COPY db/migrations /app/db/migrations

# Устанавливаем владельца файлов
RUN chown -R appuser:appgroup /app

# Переключаемся на непривилегированного пользователя
USER appuser

# Настройки для production
ENV GIN_MODE=release
ENV PORT=8080

EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/v1/health || exit 1

# Запуск приложения
ENTRYPOINT ["/app/app"]