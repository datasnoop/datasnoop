// Package investigation owns historical incident query use cases.
package investigation

import (
	"context"
	"time"
)

type EndpointSummary struct {
	Route            string
	Requests, Errors int64
	AverageDuration  time.Duration
}
type EndpointReader interface {
	Endpoints(context.Context, string, string, time.Time, time.Time) ([]EndpointSummary, error)
}

type Occurrence struct {
	TraceID, SpanID, Route string
	StatusCode             int
	StartedAt              time.Time
	Duration               time.Duration
}

type LogEntry struct {
	Timestamp         time.Time
	Severity, Message string
	TraceID, SpanID   string
}

type HostContext struct {
	Name, State string
	Value       float64
	Timestamp   time.Time
}
