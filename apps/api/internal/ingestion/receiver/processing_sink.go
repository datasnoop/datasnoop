package receiver

import (
	"context"
	"errors"
	"sort"
	"strings"

	"github.com/datasnoop/datasnoop/apps/api/internal/ingestion"
	"github.com/datasnoop/datasnoop/apps/api/internal/ingestion/otlp"
	"github.com/datasnoop/datasnoop/apps/api/internal/live"
	"github.com/datasnoop/datasnoop/apps/api/internal/platform/diagnostics"
	logsv1 "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	metricsv1 "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	tracev1 "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ProcessingSink joins normalization and bounded persistence behind the OTLP
// transport while keeping both responsibilities independently testable.
type ProcessingSink struct {
	processor *ingestion.Processor
	counters  *diagnostics.Counters
	publisher interface{ Publish(live.Event) }
}

func NewProcessingSink(processor *ingestion.Processor) *ProcessingSink {
	return &ProcessingSink{processor: processor, counters: diagnostics.NewCounters()}
}

func NewProcessingSinkWithPublisher(processor *ingestion.Processor, publisher interface{ Publish(live.Event) }) *ProcessingSink {
	sink := NewProcessingSink(processor)
	sink.publisher = publisher
	return sink
}

func (sink *ProcessingSink) Snapshot() diagnostics.Snapshot { return sink.counters.Snapshot() }

func (sink *ProcessingSink) AcceptLogs(context.Context, *logsv1.ExportLogsServiceRequest) error {
	return nil
}
func (sink *ProcessingSink) AcceptTraces(context.Context, *tracev1.ExportTraceServiceRequest) error {
	return nil
}
func (sink *ProcessingSink) AcceptMetrics(context.Context, *metricsv1.ExportMetricsServiceRequest) error {
	return nil
}

func (sink *ProcessingSink) ExportLogs(ctx context.Context, request *logsv1.ExportLogsServiceRequest) (*logsv1.ExportLogsServiceResponse, error) {
	result, err := sink.persist(ctx, otlp.NormalizeLogs(request))
	if err != nil {
		return nil, err
	}
	return &logsv1.ExportLogsServiceResponse{PartialSuccess: &logsv1.ExportLogsPartialSuccess{RejectedLogRecords: int64(result.Rejected), ErrorMessage: reasonMessage(result)}}, nil
}

func (sink *ProcessingSink) ExportTraces(ctx context.Context, request *tracev1.ExportTraceServiceRequest) (*tracev1.ExportTraceServiceResponse, error) {
	result, err := sink.persist(ctx, otlp.NormalizeTraces(request))
	if err != nil {
		return nil, err
	}
	return &tracev1.ExportTraceServiceResponse{PartialSuccess: &tracev1.ExportTracePartialSuccess{RejectedSpans: int64(result.Rejected), ErrorMessage: reasonMessage(result)}}, nil
}

func (sink *ProcessingSink) ExportMetrics(ctx context.Context, request *metricsv1.ExportMetricsServiceRequest) (*metricsv1.ExportMetricsServiceResponse, error) {
	result, err := sink.persist(ctx, otlp.NormalizeMetrics(request))
	if err != nil {
		return nil, err
	}
	return &metricsv1.ExportMetricsServiceResponse{PartialSuccess: &metricsv1.ExportMetricsPartialSuccess{RejectedDataPoints: int64(result.Rejected), ErrorMessage: reasonMessage(result)}}, nil
}

func (sink *ProcessingSink) persist(ctx context.Context, result otlp.Result) (otlp.Result, error) {
	sink.counters.Reject(uint64(result.Rejected))
	if result.Rejected > 0 && len(result.Batch.Operations)+len(result.Batch.Logs)+len(result.Batch.Metrics) == 0 {
		return result, status.Error(codes.InvalidArgument, reasonMessage(result))
	}
	if len(result.Batch.Operations)+len(result.Batch.Logs)+len(result.Batch.Metrics) == 0 {
		return result, nil
	}
	outcome, err := sink.processor.Persist(ctx, result.Batch)
	if err != nil {
		if errors.Is(err, ingestion.ErrAdmissionFull) || errors.Is(err, ingestion.ErrQueueFull) || errors.Is(err, ingestion.ErrPersistenceTimeout) {
			sink.counters.Throttle(1)
			return result, status.Error(codes.Unavailable, "telemetry capacity is temporarily unavailable")
		}
		return result, status.Error(codes.Unavailable, "telemetry persistence is temporarily unavailable")
	}
	sink.counters.Accept(uint64(outcome.Committed + outcome.Repeated))
	if sink.publisher != nil && outcome.Committed > 0 {
		sink.publisher.Publish(live.Event{Type: "telemetry", Data: map[string]int{"accepted": outcome.Committed}})
	}
	return result, nil
}

func reasonMessage(result otlp.Result) string {
	if len(result.Reasons) == 0 {
		return ""
	}
	reasons := make([]string, 0, len(result.Reasons))
	for reason := range result.Reasons {
		reasons = append(reasons, string(reason))
	}
	sort.Strings(reasons)
	return "rejected records: " + strings.Join(reasons, ", ")
}
