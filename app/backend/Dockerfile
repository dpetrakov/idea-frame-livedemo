# syntax=docker/dockerfile:1

# --- build stage ---
FROM golang:1.21-alpine AS build
WORKDIR /src
RUN apk add --no-cache git ca-certificates

# Кешируем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем код
COPY . .

# Сборка статического бинарника
ENV CGO_ENABLED=0 GOOS=linux
RUN go build -ldflags="-s -w" -o /out/app ./cmd/server

# --- runtime stage ---
FROM alpine:3.20
RUN adduser -D -H -u 10001 appuser && apk add --no-cache ca-certificates
USER appuser
WORKDIR /app
COPY --from=build /out/app /app/app
EXPOSE 8080
ENTRYPOINT ["/app/app"]