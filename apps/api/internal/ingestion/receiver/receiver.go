// Package receiver implements DataSnoop's authenticated OTLP/gRPC boundary.
package receiver

import (
	"context"
	"crypto/subtle"
	"errors"
	"strings"

	logsv1 "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	metricsv1 "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	tracev1 "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const TokenMetadataKey = "x-datasnoop-token"

// Sink is a temporary boundary for authenticated raw OTLP requests. Subsequent
// ingestion tasks replace it with record normalization and bounded admission.
type Sink interface {
	AcceptLogs(context.Context, *logsv1.ExportLogsServiceRequest) error
	AcceptTraces(context.Context, *tracev1.ExportTraceServiceRequest) error
	AcceptMetrics(context.Context, *metricsv1.ExportMetricsServiceRequest) error
}

type responseSink interface {
	ExportLogs(context.Context, *logsv1.ExportLogsServiceRequest) (*logsv1.ExportLogsServiceResponse, error)
	ExportTraces(context.Context, *tracev1.ExportTraceServiceRequest) (*tracev1.ExportTraceServiceResponse, error)
	ExportMetrics(context.Context, *metricsv1.ExportMetricsServiceRequest) (*metricsv1.ExportMetricsServiceResponse, error)
}

type Receiver struct {
	logsv1.UnimplementedLogsServiceServer
	metricsv1.UnimplementedMetricsServiceServer
	tracev1.UnimplementedTraceServiceServer

	token string
	sink  Sink
}

func New(token string, sink Sink) (*Receiver, error) {
	if strings.TrimSpace(token) == "" {
		return nil, errors.New("receiver credential is required")
	}
	return &Receiver{token: token, sink: sink}, nil
}

func (receiver *Receiver) Export(ctx context.Context, request *logsv1.ExportLogsServiceRequest) (*logsv1.ExportLogsServiceResponse, error) {
	if err := receiver.authenticate(ctx); err != nil {
		return nil, err
	}
	if sink, ok := receiver.sink.(responseSink); ok {
		return sink.ExportLogs(ctx, request)
	}
	if receiver.sink != nil {
		if err := receiver.sink.AcceptLogs(ctx, request); err != nil {
			return nil, status.Error(codes.Unavailable, "telemetry receiver is temporarily unavailable")
		}
	}
	return &logsv1.ExportLogsServiceResponse{}, nil
}

func (receiver *Receiver) ExportTraces(ctx context.Context, request *tracev1.ExportTraceServiceRequest) (*tracev1.ExportTraceServiceResponse, error) {
	if err := receiver.authenticate(ctx); err != nil {
		return nil, err
	}
	if sink, ok := receiver.sink.(responseSink); ok {
		return sink.ExportTraces(ctx, request)
	}
	if receiver.sink != nil {
		if err := receiver.sink.AcceptTraces(ctx, request); err != nil {
			return nil, status.Error(codes.Unavailable, "telemetry receiver is temporarily unavailable")
		}
	}
	return &tracev1.ExportTraceServiceResponse{}, nil
}

func (receiver *Receiver) ExportMetrics(ctx context.Context, request *metricsv1.ExportMetricsServiceRequest) (*metricsv1.ExportMetricsServiceResponse, error) {
	if err := receiver.authenticate(ctx); err != nil {
		return nil, err
	}
	if sink, ok := receiver.sink.(responseSink); ok {
		return sink.ExportMetrics(ctx, request)
	}
	if receiver.sink != nil {
		if err := receiver.sink.AcceptMetrics(ctx, request); err != nil {
			return nil, status.Error(codes.Unavailable, "telemetry receiver is temporarily unavailable")
		}
	}
	return &metricsv1.ExportMetricsServiceResponse{}, nil
}

func (receiver *Receiver) authenticate(ctx context.Context) error {
	values := metadata.ValueFromIncomingContext(ctx, TokenMetadataKey)
	if len(values) != 1 || subtle.ConstantTimeCompare([]byte(values[0]), []byte(receiver.token)) != 1 {
		return status.Error(codes.Unauthenticated, "a valid DataSnoop credential is required")
	}
	return nil
}

func Register(server grpc.ServiceRegistrar, receiver *Receiver) {
	logsv1.RegisterLogsServiceServer(server, receiver)
	tracev1.RegisterTraceServiceServer(server, traceService{receiver})
	metricsv1.RegisterMetricsServiceServer(server, metricsService{receiver})
}

type traceService struct{ *Receiver }

func (service traceService) Export(ctx context.Context, request *tracev1.ExportTraceServiceRequest) (*tracev1.ExportTraceServiceResponse, error) {
	return service.Receiver.ExportTraces(ctx, request)
}

type metricsService struct{ *Receiver }

func (service metricsService) Export(ctx context.Context, request *metricsv1.ExportMetricsServiceRequest) (*metricsv1.ExportMetricsServiceResponse, error) {
	return service.Receiver.ExportMetrics(ctx, request)
}
