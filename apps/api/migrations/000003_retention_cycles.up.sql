CREATE TABLE IF NOT EXISTS retention_cycles (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    started_at TIMESTAMPTZ NOT NULL,
    finished_at TIMESTAMPTZ NOT NULL,
    duration_ms BIGINT NOT NULL CHECK (duration_ms >= 0),
    retention_interval INTERVAL NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('completed', 'failed')),
    failure TEXT
);
