CREATE EXTENSION IF NOT EXISTS timescaledb;

CREATE TABLE IF NOT EXISTS services (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name TEXT NOT NULL,
    environment TEXT NOT NULL DEFAULT 'default',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (name, environment)
);

CREATE TABLE IF NOT EXISTS hosts (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    stable_id TEXT NOT NULL,
    name TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (stable_id)
);

CREATE TABLE IF NOT EXISTS resources (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    service_id BIGINT NOT NULL REFERENCES services(id),
    host_id BIGINT REFERENCES hosts(id),
    attributes JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE NULLS NOT DISTINCT (service_id, host_id, attributes)
);

CREATE TABLE IF NOT EXISTS operation_identities (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    service_id BIGINT NOT NULL REFERENCES services(id),
    trace_id TEXT NOT NULL,
    span_id TEXT NOT NULL,
    UNIQUE (service_id, trace_id, span_id)
);

CREATE TABLE IF NOT EXISTS operations (
    id BIGINT GENERATED ALWAYS AS IDENTITY,
    identity_id BIGINT NOT NULL REFERENCES operation_identities(id),
    service_id BIGINT NOT NULL REFERENCES services(id),
    resource_id BIGINT NOT NULL REFERENCES resources(id),
    started_at TIMESTAMPTZ NOT NULL,
    duration_ns BIGINT NOT NULL CHECK (duration_ns > 0),
    route TEXT NOT NULL,
    method TEXT NOT NULL,
    status_code SMALLINT NOT NULL CHECK (status_code BETWEEN 100 AND 599),
    trace_id TEXT NOT NULL,
    span_id TEXT NOT NULL,
    parent_span_id TEXT,
    attributes JSONB NOT NULL DEFAULT '{}'::jsonb
);

CREATE TABLE IF NOT EXISTS logs (
    id BIGINT GENERATED ALWAYS AS IDENTITY,
    service_id BIGINT NOT NULL REFERENCES services(id),
    resource_id BIGINT NOT NULL REFERENCES resources(id),
    timestamp TIMESTAMPTZ NOT NULL,
    severity TEXT,
    message TEXT NOT NULL,
    trace_id TEXT,
    span_id TEXT,
    attributes JSONB NOT NULL DEFAULT '{}'::jsonb
);

CREATE TABLE IF NOT EXISTS metrics (
    id BIGINT GENERATED ALWAYS AS IDENTITY,
    service_id BIGINT NOT NULL REFERENCES services(id),
    resource_id BIGINT NOT NULL REFERENCES resources(id),
    timestamp TIMESTAMPTZ NOT NULL,
    source_role TEXT NOT NULL CHECK (source_role IN ('monitored-service', 'datasnoop-platform')),
    name TEXT NOT NULL,
    unit TEXT NOT NULL,
    value DOUBLE PRECISION NOT NULL,
    attributes JSONB NOT NULL DEFAULT '{}'::jsonb
);

SELECT create_hypertable('operations', by_range('started_at'), if_not_exists => TRUE);
SELECT create_hypertable('logs', by_range('timestamp'), if_not_exists => TRUE);
SELECT create_hypertable('metrics', by_range('timestamp'), if_not_exists => TRUE);
