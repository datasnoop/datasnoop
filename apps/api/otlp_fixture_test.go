package api

import (
	"os"
	"path/filepath"
	"testing"

	logsv1 "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	metricsv1 "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	tracev1 "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func readFixture(t *testing.T, name string) []byte {
	t.Helper()

	contents, err := os.ReadFile(filepath.Join("testdata", "otlp", name))
	if err != nil {
		t.Fatalf("read fixture %q: %v", name, err)
	}

	return contents
}

func TestOTLPFixturesDecodeWithOfficialProtobufTypes(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		fixture string
		message proto.Message
	}{
		{"valid trace", "valid-trace.json", &tracev1.ExportTraceServiceRequest{}},
		{"valid metrics", "valid-metrics.json", &metricsv1.ExportMetricsServiceRequest{}},
		{"mixed logs", "mixed-logs.json", &logsv1.ExportLogsServiceRequest{}},
		{"correlated trace", "correlated-trace.json", &tracev1.ExportTraceServiceRequest{}},
		{"correlated logs", "correlated-logs.json", &logsv1.ExportLogsServiceRequest{}},
		{"oversized attributes", "oversized-attributes.json", &logsv1.ExportLogsServiceRequest{}},
		{"retransmitted trace", "retransmitted-trace.json", &tracev1.ExportTraceServiceRequest{}},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if err := protojson.Unmarshal(readFixture(t, testCase.fixture), testCase.message); err != nil {
				t.Fatalf("decode %s: %v", testCase.fixture, err)
			}
		})
	}
}

func TestRetransmissionFixtureRepeatsOperationIdentity(t *testing.T) {
	t.Parallel()

	var request tracev1.ExportTraceServiceRequest
	if err := protojson.Unmarshal(readFixture(t, "retransmitted-trace.json"), &request); err != nil {
		t.Fatalf("decode retransmitted trace: %v", err)
	}
	spans := request.ResourceSpans[0].ScopeSpans[0].Spans
	if len(spans) != 2 {
		t.Fatalf("retransmission fixture contains %d spans, want 2", len(spans))
	}
	if string(spans[0].TraceId) != string(spans[1].TraceId) || string(spans[0].SpanId) != string(spans[1].SpanId) {
		t.Fatal("retransmitted spans must have the same trace and span identifiers")
	}
}

func TestFixturesPreserveCorrelationAndValidationEdges(t *testing.T) {
	t.Parallel()

	var traceRequest tracev1.ExportTraceServiceRequest
	if err := protojson.Unmarshal(readFixture(t, "correlated-trace.json"), &traceRequest); err != nil {
		t.Fatalf("decode correlated trace: %v", err)
	}
	var logsRequest logsv1.ExportLogsServiceRequest
	if err := protojson.Unmarshal(readFixture(t, "correlated-logs.json"), &logsRequest); err != nil {
		t.Fatalf("decode correlated logs: %v", err)
	}

	span := traceRequest.ResourceSpans[0].ScopeSpans[0].Spans[0]
	logRecord := logsRequest.ResourceLogs[0].ScopeLogs[0].LogRecords[0]
	if string(span.TraceId) != string(logRecord.TraceId) || string(span.SpanId) != string(logRecord.SpanId) {
		t.Fatal("correlated fixtures must use identical trace and span identifiers")
	}

	var mixedRequest logsv1.ExportLogsServiceRequest
	if err := protojson.Unmarshal(readFixture(t, "mixed-logs.json"), &mixedRequest); err != nil {
		t.Fatalf("decode mixed logs: %v", err)
	}
	if got := len(mixedRequest.ResourceLogs[0].ScopeLogs[0].LogRecords); got != 2 {
		t.Fatalf("mixed fixture contains %d log records, want 2", got)
	}

	var oversizedRequest logsv1.ExportLogsServiceRequest
	if err := protojson.Unmarshal(readFixture(t, "oversized-attributes.json"), &oversizedRequest); err != nil {
		t.Fatalf("decode oversized attributes: %v", err)
	}
	if got := len(oversizedRequest.ResourceLogs[0].Resource.Attributes); got != 65 {
		t.Fatalf("oversized fixture contains %d attributes, want 65", got)
	}
}
