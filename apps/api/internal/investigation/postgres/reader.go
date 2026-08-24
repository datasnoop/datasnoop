// Package postgres adapts DataSnoop investigation queries to PostgreSQL.
package postgres

import (
	"context"
	"github.com/datasnoop/datasnoop/apps/api/internal/investigation"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type Reader struct{ pool *pgxpool.Pool }

func New(pool *pgxpool.Pool) *Reader { return &Reader{pool: pool} }
func (reader *Reader) Endpoints(ctx context.Context, service, environment string, from, to time.Time) ([]investigation.EndpointSummary, error) {
	rows, err := reader.pool.Query(ctx, `SELECT operations.route, count(*), count(*) FILTER (WHERE operations.status_code >= 500), avg(operations.duration_ns)::bigint FROM operations JOIN services ON services.id = operations.service_id WHERE services.name=$1 AND services.environment=$2 AND operations.started_at >= $3 AND operations.started_at < $4 GROUP BY operations.route ORDER BY count(*) FILTER (WHERE operations.status_code >= 500) DESC, operations.route`, service, environment, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []investigation.EndpointSummary{}
	for rows.Next() {
		var summary investigation.EndpointSummary
		var duration int64
		if err := rows.Scan(&summary.Route, &summary.Requests, &summary.Errors, &duration); err != nil {
			return nil, err
		}
		summary.AverageDuration = time.Duration(duration)
		result = append(result, summary)
	}
	return result, rows.Err()
}

func (reader *Reader) Occurrences(ctx context.Context, service, environment, route string, status int, from, to time.Time) ([]investigation.Occurrence, error) {
	rows, err := reader.pool.Query(ctx, `SELECT operations.trace_id, operations.span_id, operations.route, operations.status_code, operations.started_at, operations.duration_ns FROM operations JOIN services ON services.id=operations.service_id WHERE services.name=$1 AND services.environment=$2 AND operations.route=$3 AND operations.status_code=$4 AND operations.started_at >= $5 AND operations.started_at < $6 ORDER BY operations.started_at DESC`, service, environment, route, status, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []investigation.Occurrence{}
	for rows.Next() {
		var item investigation.Occurrence
		var duration int64
		if err := rows.Scan(&item.TraceID, &item.SpanID, &item.Route, &item.StatusCode, &item.StartedAt, &duration); err != nil {
			return nil, err
		}
		item.Duration = time.Duration(duration)
		result = append(result, item)
	}
	return result, rows.Err()
}

func (reader *Reader) CorrelatedLogs(ctx context.Context, service, environment, traceID, spanID, severity string) ([]investigation.LogEntry, error) {
	rows, err := reader.pool.Query(ctx, `SELECT logs.timestamp, COALESCE(logs.severity,''), logs.message, logs.trace_id, logs.span_id FROM logs JOIN services ON services.id=logs.service_id WHERE services.name=$1 AND services.environment=$2 AND logs.trace_id=$3 AND logs.span_id=$4 AND ($5='' OR logs.severity=$5) ORDER BY logs.timestamp ASC`, service, environment, traceID, spanID, severity)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []investigation.LogEntry{}
	for rows.Next() {
		var item investigation.LogEntry
		if err := rows.Scan(&item.Timestamp, &item.Severity, &item.Message, &item.TraceID, &item.SpanID); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (reader *Reader) HostMeasurements(ctx context.Context, service, environment string, at time.Time, window time.Duration) ([]investigation.HostContext, error) {
	rows, err := reader.pool.Query(ctx, `SELECT metrics.name, metrics.value, metrics.timestamp FROM metrics JOIN services ON services.id=metrics.service_id WHERE services.name=$1 AND services.environment=$2 AND metrics.source_role='monitored-service' AND metrics.timestamp >= $3 AND metrics.timestamp <= $4 ORDER BY metrics.timestamp ASC`, service, environment, at.Add(-window), at.Add(window))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []investigation.HostContext{}
	for rows.Next() {
		var item investigation.HostContext
		if err := rows.Scan(&item.Name, &item.Value, &item.Timestamp); err != nil {
			return nil, err
		}
		item.State = "available"
		result = append(result, item)
	}
	if len(result) == 0 {
		var exists bool
		if err := reader.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM metrics JOIN services ON services.id=metrics.service_id WHERE services.name=$1 AND services.environment=$2 AND metrics.source_role='monitored-service' AND metrics.timestamp < $3)`, service, environment, at.Add(-window)).Scan(&exists); err != nil {
			return nil, err
		}
		state := "missing"
		if exists {
			state = "stale"
		}
		return []investigation.HostContext{{State: state}}, nil
	}
	return result, rows.Err()
}
