package otlp_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/datasnoop/datasnoop/apps/api/internal/ingestion/otlp"
	logsv1 "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	metricsv1 "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	tracev1 "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func TestNormalizeIndependentOTLPFixtures(t *testing.T) {
	t.Parallel()
	traces := &tracev1.ExportTraceServiceRequest{}
	decodeFixture(t, "valid-trace.json", traces)
	traceResult := otlp.NormalizeTraces(traces)
	if traceResult.Rejected != 0 || len(traceResult.Batch.Operations) != 1 {
		t.Fatalf("trace result = %#v", traceResult)
	}
	operation := traceResult.Batch.Operations[0]
	if operation.Resource.ServiceName != "checkout" || operation.Route != "/orders/:orderID" || operation.Correlation.TraceID != "00112233445566778899aabbccddeeff" {
		t.Fatalf("normalized operation lost investigation fields: %#v", operation)
	}

	logs := &logsv1.ExportLogsServiceRequest{}
	decodeFixture(t, "correlated-logs.json", logs)
	logResult := otlp.NormalizeLogs(logs)
	if logResult.Rejected != 0 || len(logResult.Batch.Logs) != 1 {
		t.Fatalf("log result = %#v", logResult)
	}
	if logResult.Batch.Logs[0].Correlation.TraceID != "102132435465768798a9babcbddcedfe" {
		t.Fatalf("normalized log lost correlation: %#v", logResult.Batch.Logs[0])
	}

	metrics := &metricsv1.ExportMetricsServiceRequest{}
	decodeFixture(t, "valid-metrics.json", metrics)
	metricResult := otlp.NormalizeMetrics(metrics)
	if metricResult.Rejected != 0 || len(metricResult.Batch.Metrics) != 1 {
		t.Fatalf("metric result = %#v", metricResult)
	}
	if metricResult.Batch.Metrics[0].Host.ID != "host-checkout-01" || metricResult.Batch.Metrics[0].Value != 0.42 {
		t.Fatalf("normalized metric lost host context: %#v", metricResult.Batch.Metrics[0])
	}
}

func TestNormalizeMixedAndOversizedFixturesRejectOnlyAffectedRecords(t *testing.T) {
	t.Parallel()
	mixed := &logsv1.ExportLogsServiceRequest{}
	decodeFixture(t, "mixed-logs.json", mixed)
	result := otlp.NormalizeLogs(mixed)
	if len(result.Batch.Logs) != 1 || result.Rejected != 1 || result.Reasons[otlp.ReasonInvalidTimestamp] != 1 {
		t.Fatalf("mixed log result = %#v", result)
	}

	overSized := &logsv1.ExportLogsServiceRequest{}
	decodeFixture(t, "oversized-attributes.json", overSized)
	result = otlp.NormalizeLogs(overSized)
	if len(result.Batch.Logs) != 0 || result.Rejected != 1 || result.Reasons[otlp.ReasonPayloadLimit] != 1 {
		t.Fatalf("oversized log result = %#v", result)
	}
}

func FuzzNormalizeLogs(f *testing.F) {
	seed, err := os.ReadFile(filepath.Join("..", "..", "..", "testdata", "otlp", "mixed-logs.json"))
	if err != nil {
		f.Fatal(err)
	}
	f.Add(seed)
	f.Fuzz(func(t *testing.T, payload []byte) {
		request := &logsv1.ExportLogsServiceRequest{}
		if protojson.Unmarshal(payload, request) == nil {
			_ = otlp.NormalizeLogs(request)
		}
	})
}

func decodeFixture(t *testing.T, name string, message proto.Message) {
	t.Helper()
	payload, err := os.ReadFile(filepath.Join("..", "..", "..", "testdata", "otlp", name))
	if err != nil {
		t.Fatal(err)
	}
	if err := protojson.Unmarshal(payload, message); err != nil {
		t.Fatalf("decode %s: %v", name, err)
	}
}
