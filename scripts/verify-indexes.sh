#!/usr/bin/env bash
set -euo pipefail

compose_file="infra/postgres/compose.yaml"
database_service="database"

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

sql() {
  docker compose -f "$compose_file" exec -T "$database_service" psql -v ON_ERROR_STOP=1 -U datasnoop -d datasnoop_test
}

sql < apps/api/migrations/000001_telemetry.up.sql
sql < apps/api/migrations/000002_investigation_indexes.up.sql
sql <<'SQL'
INSERT INTO services (name, environment) VALUES ('checkout', 'production');
INSERT INTO hosts (stable_id, name) VALUES ('host-01', 'checkout-host');
INSERT INTO resources (service_id, host_id) VALUES (1, 1);
INSERT INTO operations (service_id, resource_id, started_at, duration_ns, route, method, status_code, trace_id, span_id)
SELECT 1, 1, now() - (value || ' seconds')::interval, 1000000, CASE WHEN value % 10 = 0 THEN '/orders/:orderID' ELSE '/health' END, 'GET', CASE WHEN value % 10 = 0 THEN 500 ELSE 200 END, lpad(to_hex(value), 32, '0'), lpad(to_hex(value), 16, '0') FROM generate_series(1, 1000) AS value;
INSERT INTO logs (service_id, resource_id, timestamp, message, trace_id, span_id) VALUES (1, 1, now(), 'payment failed', lpad(to_hex(500), 32, '0'), lpad(to_hex(500), 16, '0'));
INSERT INTO metrics (service_id, resource_id, timestamp, source_role, name, unit, value) VALUES (1, 1, now(), 'monitored-service', 'system.cpu.utilization', '1', 0.5);
ANALYZE;
SQL

plan="$(sql -At <<'SQL'
SET enable_seqscan TO off;
EXPLAIN (COSTS OFF) SELECT * FROM operations WHERE service_id = 1 AND route = '/orders/:orderID' AND status_code = 500 ORDER BY started_at DESC LIMIT 10;
EXPLAIN (COSTS OFF) SELECT * FROM logs WHERE trace_id = lpad(to_hex(500), 32, '0') AND span_id = lpad(to_hex(500), 16, '0') ORDER BY timestamp;
EXPLAIN (COSTS OFF) SELECT * FROM metrics WHERE resource_id = 1 ORDER BY timestamp DESC;
SQL
)"

echo "$plan"
echo "$plan" | grep -q operations_route_status_started_idx
echo "$plan" | grep -q logs_trace_span_timestamp_idx
echo "$plan" | grep -q metrics_resource_timestamp_idx
