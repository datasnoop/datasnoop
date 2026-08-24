package receiver_test

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/datasnoop/datasnoop/apps/api/internal/ingestion"
	"github.com/datasnoop/datasnoop/apps/api/internal/ingestion/otlp"
	"github.com/datasnoop/datasnoop/apps/api/internal/ingestion/receiver"
	logsv1 "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	metricsv1 "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	tracev1 "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func TestAuthenticatedOTLPExportServices(t *testing.T) {
	t.Parallel()
	sink := &recordingSink{}
	connection := startReceiver(t, sink)
	contextWithToken := metadata.AppendToOutgoingContext(context.Background(), receiver.TokenMetadataKey, "test-token")

	if _, err := logsv1.NewLogsServiceClient(connection).Export(contextWithToken, &logsv1.ExportLogsServiceRequest{}); err != nil {
		t.Fatalf("export logs: %v", err)
	}
	if _, err := tracev1.NewTraceServiceClient(connection).Export(contextWithToken, &tracev1.ExportTraceServiceRequest{}); err != nil {
		t.Fatalf("export traces: %v", err)
	}
	if _, err := metricsv1.NewMetricsServiceClient(connection).Export(contextWithToken, &metricsv1.ExportMetricsServiceRequest{}); err != nil {
		t.Fatalf("export metrics: %v", err)
	}
	if got := sink.total(); got != 3 {
		t.Fatalf("accepted requests = %d, want 3", got)
	}
}

func TestOTLPExportRejectsMissingOrInvalidCredentialsWithoutPassingRequestsToSink(t *testing.T) {
	t.Parallel()
	sink := &recordingSink{}
	connection := startReceiver(t, sink)
	client := tracev1.NewTraceServiceClient(connection)

	for _, requestContext := range []context.Context{
		context.Background(),
		metadata.AppendToOutgoingContext(context.Background(), receiver.TokenMetadataKey, "wrong-token"),
	} {
		_, err := client.Export(requestContext, &tracev1.ExportTraceServiceRequest{})
		if got := status.Code(err); got != codes.Unauthenticated {
			t.Fatalf("credential error code = %s, want %s", got, codes.Unauthenticated)
		}
	}
	if got := sink.total(); got != 0 {
		t.Fatalf("unauthenticated requests reached sink %d times", got)
	}
}

func TestOTLPPartialSuccessAndWhollyInvalidMappings(t *testing.T) {
	t.Parallel()
	store := &committingStore{}
	processor, err := ingestion.NewProcessor(ingestion.ProcessorConfig{AdmissionLimit: 1, QueueLimit: 1, Workers: 1, PersistenceTimeout: time.Second}, store)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(processor.Close)
	sink := receiver.NewProcessingSink(processor)
	connection := startReceiver(t, sink)
	contextWithToken := metadata.AppendToOutgoingContext(context.Background(), receiver.TokenMetadataKey, "test-token")

	mixed := &logsv1.ExportLogsServiceRequest{}
	decodeReceiverFixture(t, "mixed-logs.json", mixed)
	response, err := logsv1.NewLogsServiceClient(connection).Export(contextWithToken, mixed)
	if err != nil || response.GetPartialSuccess().GetRejectedLogRecords() != 1 {
		t.Fatalf("mixed export response = %#v, error = %v", response, err)
	}
	if store.records != 1 {
		t.Fatalf("persisted records = %d, want 1", store.records)
	}
	if snapshot := sink.Snapshot(); snapshot.Accepted != 1 || snapshot.Rejected != 1 || snapshot.RestartID == "" {
		t.Fatalf("diagnostic snapshot = %#v", snapshot)
	}

	overSized := &logsv1.ExportLogsServiceRequest{}
	decodeReceiverFixture(t, "oversized-attributes.json", overSized)
	_, err = logsv1.NewLogsServiceClient(connection).Export(contextWithToken, overSized)
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("wholly invalid status = %s, want %s", status.Code(err), codes.InvalidArgument)
	}
}

func TestOfficialOTLPClientPreservesIndependentExporterSemantics(t *testing.T) {
	sink := &recordingSink{}
	connection := startReceiver(t, sink)
	request := &tracev1.ExportTraceServiceRequest{}
	decodeReceiverFixture(t, "correlated-trace.json", request)
	contextWithToken := metadata.AppendToOutgoingContext(context.Background(), receiver.TokenMetadataKey, "test-token")
	// This generated OTLP gRPC client is used directly; no DataSnoop SDK is imported.
	if _, err := tracev1.NewTraceServiceClient(connection).Export(contextWithToken, request); err != nil {
		t.Fatal(err)
	}
	sink.mu.Lock()
	received := sink.lastTrace
	sink.mu.Unlock()
	result := otlp.NormalizeTraces(received)
	if result.Rejected != 0 || len(result.Batch.Operations) != 1 {
		t.Fatalf("normalized result=%#v", result)
	}
	operation := result.Batch.Operations[0]
	if operation.Resource.ServiceName != "checkout" || operation.Route != "/orders" || operation.StatusCode != 500 || operation.Correlation.TraceID == "" || operation.Correlation.SpanID == "" {
		t.Fatalf("independent exporter semantics=%#v", operation)
	}
}

func startReceiver(t *testing.T, sink receiver.Sink) *grpc.ClientConn {
	t.Helper()
	listener := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()
	receiverInstance, err := receiver.New("test-token", sink)
	if err != nil {
		t.Fatalf("new receiver: %v", err)
	}
	receiver.Register(server, receiverInstance)
	go func() { _ = server.Serve(listener) }()
	t.Cleanup(server.Stop)

	connection, err := grpc.NewClient("passthrough:///bufnet", grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
		return listener.Dial()
	}), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("dial receiver: %v", err)
	}
	t.Cleanup(func() { _ = connection.Close() })
	return connection
}

type recordingSink struct {
	mu        sync.Mutex
	logs      int
	traces    int
	metrics   int
	lastTrace *tracev1.ExportTraceServiceRequest
}

type committingStore struct {
	mu      sync.Mutex
	records int
}

func (store *committingStore) Persist(_ context.Context, batch ingestion.Batch) ingestion.Outcome {
	store.mu.Lock()
	defer store.mu.Unlock()
	store.records += len(batch.Operations) + len(batch.Logs) + len(batch.Metrics)
	return ingestion.Outcome{Committed: store.records}
}

func decodeReceiverFixture(t *testing.T, name string, message proto.Message) {
	t.Helper()
	payload, err := os.ReadFile(filepath.Join("..", "..", "..", "testdata", "otlp", name))
	if err != nil {
		t.Fatal(err)
	}
	if err := protojson.Unmarshal(payload, message); err != nil {
		t.Fatalf("decode %s: %v", name, err)
	}
}

func (sink *recordingSink) AcceptLogs(context.Context, *logsv1.ExportLogsServiceRequest) error {
	sink.mu.Lock()
	defer sink.mu.Unlock()
	sink.logs++
	return nil
}

func (sink *recordingSink) AcceptTraces(_ context.Context, request *tracev1.ExportTraceServiceRequest) error {
	sink.mu.Lock()
	defer sink.mu.Unlock()
	sink.traces++
	sink.lastTrace = proto.Clone(request).(*tracev1.ExportTraceServiceRequest)
	return nil
}

func (sink *recordingSink) AcceptMetrics(context.Context, *metricsv1.ExportMetricsServiceRequest) error {
	sink.mu.Lock()
	defer sink.mu.Unlock()
	sink.metrics++
	return nil
}

func (sink *recordingSink) total() int {
	sink.mu.Lock()
	defer sink.mu.Unlock()
	return sink.logs + sink.traces + sink.metrics
}
