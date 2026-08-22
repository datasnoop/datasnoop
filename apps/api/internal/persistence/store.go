// Package persistence writes normalized telemetry records transactionally.
package persistence

import (
	"context"

	"github.com/datasnoop/datasnoop/apps/api/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Batch struct {
	Operations []domain.Operation
	Logs       []domain.Log
	Metrics    []domain.Metric
}

type Outcome struct {
	Committed int
	Rejected  int
	Failed    error
}

type Store struct{ pool *pgxpool.Pool }

func New(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

func (store *Store) Persist(ctx context.Context, batch Batch) Outcome {
	valid := Batch{}
	outcome := Outcome{}
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
	for _, operation := range valid.Operations {
		if err := writeOperation(ctx, transaction, operation); err != nil {
			outcome.Failed = err
			return outcome
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
	outcome.Committed = len(valid.Operations) + len(valid.Logs) + len(valid.Metrics)
	return outcome
}

func resourceID(ctx context.Context, transaction pgx.Tx, resource domain.Resource, host domain.Host) (int64, int64, error) {
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

func writeOperation(ctx context.Context, transaction pgx.Tx, operation domain.Operation) error {
	serviceID, recordResourceID, err := resourceID(ctx, transaction, operation.Resource, operation.Host)
	if err != nil {
		return err
	}
	_, err = transaction.Exec(ctx, `INSERT INTO operations (service_id, resource_id, started_at, duration_ns, route, method, status_code, trace_id, span_id, parent_span_id) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`, serviceID, recordResourceID, operation.StartedAt, operation.Duration.Nanoseconds(), operation.Route, operation.Method, operation.StatusCode, operation.Correlation.TraceID, operation.Correlation.SpanID, operation.Correlation.ParentSpanID)
	return err
}

func writeLog(ctx context.Context, transaction pgx.Tx, logRecord domain.Log) error {
	serviceID, recordResourceID, err := resourceID(ctx, transaction, logRecord.Resource, logRecord.Host)
	if err != nil {
		return err
	}
	_, err = transaction.Exec(ctx, `INSERT INTO logs (service_id, resource_id, timestamp, severity, message, trace_id, span_id) VALUES ($1,$2,$3,$4,$5,$6,$7)`, serviceID, recordResourceID, logRecord.Timestamp, logRecord.Severity, logRecord.Message, nullable(logRecord.Correlation.TraceID), nullable(logRecord.Correlation.SpanID))
	return err
}

func writeMetric(ctx context.Context, transaction pgx.Tx, metric domain.Metric) error {
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
