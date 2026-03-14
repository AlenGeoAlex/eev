FROM golang:1.25-alpine AS backend-builder
RUN apk add --no-cache gcc musl-dev
WORKDIR /src

COPY . .
RUN mkdir -p /app && \
    cp apps/unified-docker/Caddyfile /app/Caddyfile && \
    cp apps/unified-docker/start.sh /app/start.sh
WORKDIR /src/apps/backend-go
RUN go mod download
RUN CGO_ENABLED=1 GOOS=linux go build -o /app/eev-backend ./cmd/app/main.go
RUN CGO_ENABLED=1 GOOS=linux go build -o /app/migrator ./cmd/migrator

RUN cp -r migrations /app/migrations

FROM node:22-alpine AS frontend-builder
RUN corepack enable && corepack prepare pnpm@latest --activate
WORKDIR /src
ENV CI=true
COPY . .
WORKDIR /src/apps/eev-web

RUN pnpm install && \
    pnpm add -D @sveltejs/adapter-node && \
    sed -i "s/'@sveltejs\/adapter-auto'/'@sveltejs\/adapter-node'/g" svelte.config.js && \
    sed -i 's/"@sveltejs\/adapter-auto"/"@sveltejs\/adapter-node"/g' svelte.config.js
RUN pnpm build

FROM caddy:2.9-alpine
RUN apk add --no-cache nodejs npm netcat-openbsd
WORKDIR /app

COPY --from=backend-builder /app/eev-backend /usr/local/bin/eev-backend
COPY --from=backend-builder /app/migrator /usr/local/bin/migrator
COPY --from=backend-builder /app/migrations /app/migrations
COPY --from=backend-builder /app/Caddyfile /etc/caddy/Caddyfile
COPY --from=backend-builder /app/start.sh /usr/local/bin/start.sh
RUN chmod +x /usr/local/bin/start.sh


COPY --from=frontend-builder /src/apps/eev-web/build /app/frontend
COPY --from=frontend-builder /src/apps/eev-web/package.json /app/frontend/package.json

WORKDIR /app/frontend
RUN apk add --no-cache npm && npm install --production --ignore-scripts

ENV APP_PORT=8080
ENV DB_SQLITE_PATH=/data/sqlite.db
ENV PORT=3000
ENV ORIGIN=http://localhost

VOLUME /data
EXPOSE 80

CMD ["/usr/local/bin/start.sh"]
