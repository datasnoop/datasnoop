package datasnoop_test

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/datasnoop/datasnoop/sdk/go"
	"go.opentelemetry.io/otel/trace"
)

func TestSlogExportsRequestAndBackgroundRecords(t *testing.T) {
	var records []datasnoop.LogRecord
	logger := slog.New(datasnoop.Slog(nil, func(_ context.Context, record datasnoop.LogRecord) { records = append(records, record) }))
	traceID, _ := trace.TraceIDFromHex("00112233445566778899aabbccddeeff")
	spanID, _ := trace.SpanIDFromHex("0011223344556677")
	requestContext := trace.ContextWithSpanContext(context.Background(), trace.NewSpanContext(trace.SpanContextConfig{TraceID: traceID, SpanID: spanID, TraceFlags: trace.FlagsSampled}))
	logger.InfoContext(requestContext, "payment failed", "attempt", 2)
	logger.Warn("background maintenance")
	if len(records) != 2 || records[0].TraceID != traceID.String() || records[0].SpanID != spanID.String() || records[1].TraceID != "" || records[1].Message != "background maintenance" || records[0].Timestamp.IsZero() {
		t.Fatalf("unexpected exported records: %#v", records)
	}
	if records[0].Severity != slog.LevelInfo || len(records[0].Attributes) != 1 || records[0].Attributes[0].Key != "attempt" {
		t.Fatalf("record content not preserved: %#v", records[0])
	}
}

func TestSlogPreservesDownstreamHandler(t *testing.T) {
	called := false
	next := slog.Handler(slog.NewTextHandler(discardWriter{}, nil))
	logger := slog.New(datasnoop.Slog(next, func(context.Context, datasnoop.LogRecord) { called = true }))
	logger.LogAttrs(context.Background(), slog.LevelError, "preserved", slog.Time("at", time.Now()))
	if !called {
		t.Fatal("exporter was not called")
	}
}

type discardWriter struct{}

func (discardWriter) Write(payload []byte) (int, error) { return len(payload), nil }
