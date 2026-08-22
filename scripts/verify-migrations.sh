#!/usr/bin/env bash
set -euo pipefail

compose_file="infra/postgres/compose.yaml"
database_service="database"
migration_up="apps/api/migrations/000001_telemetry.up.sql"
migration_down="apps/api/migrations/000001_telemetry.down.sql"

cleanup() {
  docker compose -f "$compose_file" down --volumes --remove-orphans
}

trap cleanup EXIT
cleanup
docker compose -f "$compose_file" up --wait

for attempt in $(seq 1 10); do
  if docker compose -f "$compose_file" exec -T "$database_service" pg_isready -U datasnoop -d datasnoop_test; then
    break
  fi
  sleep 1
done

apply_migration() {
  docker compose -f "$compose_file" exec -T "$database_service" \
    psql -v ON_ERROR_STOP=1 -U datasnoop -d datasnoop_test
}

apply_migration < "$migration_up"
apply_migration < "$migration_up"

table_count="$(docker compose -f "$compose_file" exec -T "$database_service" \
  psql -At -U datasnoop -d datasnoop_test -c \
  "SELECT count(*) FROM pg_tables WHERE schemaname = 'public' AND tablename IN ('services', 'hosts', 'resources', 'operations', 'logs', 'metrics');")"
test "$table_count" = 6

apply_migration < "$migration_down"

remaining_count="$(docker compose -f "$compose_file" exec -T "$database_service" \
  psql -At -U datasnoop -d datasnoop_test -c \
  "SELECT count(*) FROM pg_tables WHERE schemaname = 'public' AND tablename IN ('services', 'hosts', 'resources', 'operations', 'logs', 'metrics');")"
test "$remaining_count" = 0
