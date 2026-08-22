CREATE INDEX IF NOT EXISTS operations_service_started_idx ON operations (service_id, started_at DESC);
CREATE INDEX IF NOT EXISTS operations_route_status_started_idx ON operations (service_id, route, status_code, started_at DESC);
CREATE INDEX IF NOT EXISTS operations_trace_span_idx ON operations (trace_id, span_id);
CREATE INDEX IF NOT EXISTS logs_trace_span_timestamp_idx ON logs (trace_id, span_id, timestamp ASC);
CREATE INDEX IF NOT EXISTS metrics_resource_timestamp_idx ON metrics (resource_id, timestamp DESC);
