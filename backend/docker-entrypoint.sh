#!/bin/sh
set -e

# Applique les migrations goose (idempotent) avant de démarrer l'API.
if [ -n "$DATABASE_URL" ]; then
  /app/goose -dir /app/db/migrations postgres "$DATABASE_URL" up
fi

exec /app/api
