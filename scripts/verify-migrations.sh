#!/usr/bin/env bash
set -euo pipefail

compose_file="infra/postgres/compose.yaml"
database_service="database"
migration_ups=(
  "apps/api/migrations/000001_telemetry.up.sql"
  "apps/api/migrations/000002_investigation_indexes.up.sql"
  "apps/api/migrations/000003_retention_cycles.up.sql"
)
migration_downs=(
  "apps/api/migrations/000003_retention_cycles.down.sql"
  "apps/api/migrations/000002_investigation_indexes.down.sql"
  "apps/api/migrations/000001_telemetry.down.sql"
)

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

for migration in "${migration_ups[@]}"; do apply_migration < "$migration"; done
for migration in "${migration_ups[@]}"; do apply_migration < "$migration"; done

table_count="$(docker compose -f "$compose_file" exec -T "$database_service" \
  psql -At -U datasnoop -d datasnoop_test -c \
  "SELECT count(*) FROM pg_tables WHERE schemaname = 'public' AND tablename IN ('services', 'hosts', 'resources', 'operation_identities', 'operations', 'logs', 'metrics', 'retention_cycles');")"
test "$table_count" = 8

for migration in "${migration_downs[@]}"; do apply_migration < "$migration"; done

remaining_count="$(docker compose -f "$compose_file" exec -T "$database_service" \
  psql -At -U datasnoop -d datasnoop_test -c \
  "SELECT count(*) FROM pg_tables WHERE schemaname = 'public' AND tablename IN ('services', 'hosts', 'resources', 'operation_identities', 'operations', 'logs', 'metrics', 'retention_cycles');")"
test "$remaining_count" = 0
