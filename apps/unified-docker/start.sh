#!/bin/sh
set -e

mkdir -p /data/migrations
cp -r /app/migrations/* /data/migrations/

cd /data
/usr/local/bin/migrator -direction up

/usr/local/bin/eev-backend &
BACKEND_PID=$!

echo "Waiting for backend..."
for i in $(seq 1 30); do
    if ! kill -0 $BACKEND_PID 2>/dev/null; then
        echo "Backend crashed on startup"
        exit 1
    fi
    if nc -z localhost 8080 2>/dev/null; then
        echo "Backend ready"
        break
    fi
    sleep 0.5
done

cd /app/frontend
node index.js &
NODE_PID=$!

monitor() {
    while sleep 5; do
        if ! kill -0 $BACKEND_PID 2>/dev/null; then
            echo "Backend died, shutting down"
            exit 1
        fi
        if ! kill -0 $NODE_PID 2>/dev/null; then
            echo "Frontend died, shutting down"
            exit 1
        fi
    done
}
monitor &

caddy run --config /etc/caddy/Caddyfile --adapter caddyfile