# syntax=docker/dockerfile:1

FROM node:20-alpine AS build
WORKDIR /app
COPY app/frontend/package*.json ./
RUN npm ci --legacy-peer-deps
COPY app/frontend/ .
RUN npm run build

FROM caddy:2-alpine
COPY --from=build /app/build /usr/share/caddy
COPY app/frontend/Caddyfile.frontend /etc/caddy/Caddyfile
EXPOSE 3000
