# !/bin/sh

set -e #exit immediately when the command return none 0 code

echo "run db migration"
/app/migrate -path /app/migration -database "$DB_SOURCE" -verbose up

echo "start the app"
exec "$@"