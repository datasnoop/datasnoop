// Package postgres implements ingestion persistence with PostgreSQL and TimescaleDB.
package postgres

import (
	"context"

	"github.com/datasnoop/datasnoop/apps/api/internal/ingestion"
	"github.com/datasnoop/datasnoop/apps/api/internal/telemetry"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct{ pool *pgxpool.Pool }

func New(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

func (store *Store) Persist(ctx context.Context, batch ingestion.Batch) ingestion.Outcome {
	valid := ingestion.Batch{}
	outcome := ingestion.Outcome{}
	for _, operation := range batch.Operations {
		if operation.Validate() != nil {
			outcome.Rejected++
			continue
		}
		valid.Operations = append(valid.Operations, operation)
	}
	for _, logRecord := range batch.Logs {
		if logRecord.Validate() != nil {
			outcome.Rejected++
			continue
		}
		valid.Logs = append(valid.Logs, logRecord)
	}
	for _, metric := range batch.Metrics {
		if metric.Validate() != nil {
			outcome.Rejected++
			continue
		}
		valid.Metrics = append(valid.Metrics, metric)
	}
	transaction, err := store.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		outcome.Failed = err
		return outcome
	}
	defer transaction.Rollback(ctx)
	committed := len(valid.Logs) + len(valid.Metrics)
	repeated := 0
	for _, operation := range valid.Operations {
		inserted, err := writeOperation(ctx, transaction, operation)
		if err != nil {
			outcome.Failed = err
			return outcome
		}
		if inserted {
			committed++
		} else {
			repeated++
		}
	}
	for _, logRecord := range valid.Logs {
		if err := writeLog(ctx, transaction, logRecord); err != nil {
			outcome.Failed = err
			return outcome
		}
	}
	for _, metric := range valid.Metrics {
		if err := writeMetric(ctx, transaction, metric); err != nil {
			outcome.Failed = err
			return outcome
		}
	}
	if err := transaction.Commit(ctx); err != nil {
		outcome.Failed = err
		return outcome
	}
	outcome.Committed = committed
	outcome.Repeated = repeated
	return outcome
}

func resourceID(ctx context.Context, transaction pgx.Tx, resource telemetry.Resource, host telemetry.Host) (int64, int64, error) {
	environment := resource.Environment
	if environment == "" {
		environment = "default"
	}
	var serviceID int64
	if err := transaction.QueryRow(ctx, `INSERT INTO services (name, environment) VALUES ($1, $2) ON CONFLICT (name, environment) DO UPDATE SET name = EXCLUDED.name RETURNING id`, resource.ServiceName, environment).Scan(&serviceID); err != nil {
		return 0, 0, err
	}
	var hostID *int64
	if host.ID != "" || host.Name != "" {
		stableID := host.ID
		if stableID == "" {
			stableID = host.Name
		}
		var value int64
		if err := transaction.QueryRow(ctx, `INSERT INTO hosts (stable_id, name) VALUES ($1, $2) ON CONFLICT (stable_id) DO UPDATE SET name = COALESCE(EXCLUDED.name, hosts.name) RETURNING id`, stableID, host.Name).Scan(&value); err != nil {
			return 0, 0, err
		}
		hostID = &value
	}
	var resourceID int64
	if err := transaction.QueryRow(ctx, `INSERT INTO resources (service_id, host_id) VALUES ($1, $2) ON CONFLICT (service_id, host_id, attributes) DO UPDATE SET service_id = EXCLUDED.service_id RETURNING id`, serviceID, hostID).Scan(&resourceID); err != nil {
		return 0, 0, err
	}
	return serviceID, resourceID, nil
}

func writeOperation(ctx context.Context, transaction pgx.Tx, operation telemetry.Operation) (bool, error) {
	serviceID, recordResourceID, err := resourceID(ctx, transaction, operation.Resource, operation.Host)
	if err != nil {
		return false, err
	}
	var identityID int64
	err = transaction.QueryRow(ctx, `INSERT INTO operation_identities (service_id, trace_id, span_id) VALUES ($1,$2,$3) ON CONFLICT (service_id, trace_id, span_id) DO NOTHING RETURNING id`, serviceID, operation.Correlation.TraceID, operation.Correlation.SpanID).Scan(&identityID)
	if err == pgx.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	_, err = transaction.Exec(ctx, `INSERT INTO operations (identity_id, service_id, resource_id, started_at, duration_ns, route, method, status_code, trace_id, span_id, parent_span_id) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`, identityID, serviceID, recordResourceID, operation.StartedAt, operation.Duration.Nanoseconds(), operation.Route, operation.Method, operation.StatusCode, operation.Correlation.TraceID, operation.Correlation.SpanID, operation.Correlation.ParentSpanID)
	return err == nil, err
}

func writeLog(ctx context.Context, transaction pgx.Tx, logRecord telemetry.Log) error {
	serviceID, recordResourceID, err := resourceID(ctx, transaction, logRecord.Resource, logRecord.Host)
	if err != nil {
		return err
	}
	_, err = transaction.Exec(ctx, `INSERT INTO logs (service_id, resource_id, timestamp, severity, message, trace_id, span_id) VALUES ($1,$2,$3,$4,$5,$6,$7)`, serviceID, recordResourceID, logRecord.Timestamp, logRecord.Severity, logRecord.Message, nullable(logRecord.Correlation.TraceID), nullable(logRecord.Correlation.SpanID))
	return err
}

func writeMetric(ctx context.Context, transaction pgx.Tx, metric telemetry.Metric) error {
	serviceID, recordResourceID, err := resourceID(ctx, transaction, metric.Resource, metric.Host)
	if err != nil {
		return err
	}
	_, err = transaction.Exec(ctx, `INSERT INTO metrics (service_id, resource_id, timestamp, source_role, name, unit, value) VALUES ($1,$2,$3,$4,$5,$6,$7)`, serviceID, recordResourceID, metric.Timestamp, metric.SourceRole, metric.Name, metric.Unit, metric.Value)
	return err
}

func nullable(value string) interface{} {
	if value == "" {
		return nil
	}
	return value
}
