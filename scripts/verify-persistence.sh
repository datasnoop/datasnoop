#!/usr/bin/env bash
set -euo pipefail

compose_file="infra/postgres/compose.yaml"
cleanup() { docker compose -f "$compose_file" down --volumes --remove-orphans; }
trap cleanup EXIT
cleanup
docker compose -f "$compose_file" up --wait
for attempt in $(seq 1 10); do
  if docker compose -f "$compose_file" exec -T database pg_isready -U datasnoop -d datasnoop_test; then break; fi
  sleep 1
done
docker compose -f "$compose_file" exec -T database psql -v ON_ERROR_STOP=1 -U datasnoop -d datasnoop_test < apps/api/migrations/000001_telemetry.up.sql
docker compose -f "$compose_file" exec -T database psql -v ON_ERROR_STOP=1 -U datasnoop -d datasnoop_test < apps/api/migrations/000003_retention_cycles.up.sql
(
  cd apps/api
  DATASNOOP_TEST_DATABASE_URL="postgres://datasnoop:datasnoop@127.0.0.1:54329/datasnoop_test?sslmode=disable" go test ./internal/ingestion/postgres -run '^TestStorePersistOutcomes$'
  DATASNOOP_TEST_DATABASE_URL="postgres://datasnoop:datasnoop@127.0.0.1:54329/datasnoop_test?sslmode=disable" go test ./internal/platform/retention -run '^TestTimescaleRetentionRemovesExpiredChunks$'
  DATASNOOP_TEST_DATABASE_URL="postgres://datasnoop:datasnoop@127.0.0.1:54329/datasnoop_test?sslmode=disable" go test ./internal/workload -run '^TestSteadyProfileCompletesHistoricalIncidentJourney$'
)
