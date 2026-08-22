// Package ingestion owns the application contract for accepting normalized telemetry.
package ingestion

import (
	"context"

	"github.com/datasnoop/datasnoop/apps/api/internal/telemetry"
)

type Batch struct {
	Operations []telemetry.Operation
	Logs       []telemetry.Log
	Metrics    []telemetry.Metric
}

type Outcome struct {
	Committed int
	Repeated  int
	Rejected  int
	Failed    error
}

type BatchStore interface {
	Persist(context.Context, Batch) Outcome
}
