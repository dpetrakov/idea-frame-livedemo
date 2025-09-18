# syntax=docker/dockerfile:1

FROM node:20-alpine AS build

WORKDIR /app

# Переменные для сборки Vite (build-time)
ARG VITE_API_BASE=/api
ENV VITE_API_BASE=${VITE_API_BASE}

# Копируем package.json и package-lock.json для кеширования зависимостей
COPY package*.json ./
RUN npm ci --only=production=false

# Копируем исходный код
COPY . .

# Сборка приложения
RUN npm run build

# Статическая выдача файлов через Caddy
FROM caddy:2-alpine

# Копируем собранную статику
COPY --from=build /app/dist /usr/share/caddy

# Создаём Caddyfile с SPA fallback
RUN printf ":3000 {\n  root * /usr/share/caddy\n  file_server\n  try_files {path} /index.html\n}\n" > /etc/caddy/Caddyfile

# Создаём непривилегированного пользователя
RUN addgroup -g 1001 -S caddy && \
    adduser -u 1001 -S caddy -G caddy

# Настраиваем права доступа
RUN chown -R caddy:caddy /usr/share/caddy /etc/caddy

USER caddy

EXPOSE 3000

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:3000 || exit 1