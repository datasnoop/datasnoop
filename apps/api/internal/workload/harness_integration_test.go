package workload_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/datasnoop/datasnoop/apps/api/internal/ingestion"
	ingestpostgres "github.com/datasnoop/datasnoop/apps/api/internal/ingestion/postgres"
	investigationpostgres "github.com/datasnoop/datasnoop/apps/api/internal/investigation/postgres"
	"github.com/datasnoop/datasnoop/apps/api/internal/telemetry"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestSteadyProfileCompletesHistoricalIncidentJourney(t *testing.T) {
	url := os.Getenv("DATASNOOP_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("DATASNOOP_TEST_DATABASE_URL is required")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	if _, err := pool.Exec(ctx, "TRUNCATE metrics, logs, operations, operation_identities, resources, hosts, services RESTART IDENTITY CASCADE"); err != nil {
		t.Fatal(err)
	}
	at := time.Date(2026, 8, 24, 12, 0, 0, 0, time.UTC)
	resource := telemetry.Resource{ServiceName: "checkout", Environment: "system"}
	correlation := telemetry.Correlation{TraceID: "00112233445566778899aabbccddeeff", SpanID: "0011223344556677"}
	host := telemetry.Host{ID: "system-host"}
	batch := ingestion.Batch{Operations: []telemetry.Operation{{Resource: resource, Host: host, Correlation: correlation, Route: "/orders/:orderID", Method: "GET", StatusCode: 500, StartedAt: at, Duration: time.Millisecond}, {Resource: resource, Host: host, Correlation: telemetry.Correlation{TraceID: "102132435465768798a9babcbddcedfe", SpanID: "1021324354657687"}, Route: "/health", Method: "GET", StatusCode: 200, StartedAt: at, Duration: time.Millisecond}}, Logs: []telemetry.Log{{Resource: resource, Host: host, Correlation: correlation, Timestamp: at, Severity: "ERROR", Message: "payment provider failed"}}, Metrics: []telemetry.Metric{{Resource: resource, Host: host, SourceRole: telemetry.SourceRoleMonitoredService, Timestamp: at, Name: "system.cpu.utilization", Unit: "1", Value: .82}}}
	if outcome := ingestpostgres.New(pool).Persist(ctx, batch); outcome.Failed != nil {
		t.Fatal(outcome.Failed)
	}
	reader := investigationpostgres.New(pool)
	summaries, err := reader.Endpoints(ctx, "checkout", "system", at.Add(-time.Minute), at.Add(time.Minute))
	if err != nil || len(summaries) != 2 || summaries[0].Errors != 1 {
		t.Fatalf("summaries=%#v err=%v", summaries, err)
	}
	occurrences, err := reader.Occurrences(ctx, "checkout", "system", "/orders/:orderID", 500, at.Add(-time.Minute), at.Add(time.Minute))
	if err != nil || len(occurrences) != 1 {
		t.Fatalf("occurrences=%#v err=%v", occurrences, err)
	}
	logs, err := reader.CorrelatedLogs(ctx, "checkout", "system", correlation.TraceID, correlation.SpanID, "ERROR")
	if err != nil || len(logs) != 1 {
		t.Fatalf("logs=%#v err=%v", logs, err)
	}
	hostContext, err := reader.HostMeasurements(ctx, "checkout", "system", at, time.Minute)
	if err != nil || len(hostContext) != 1 || hostContext[0].State != "available" {
		t.Fatalf("host=%#v err=%v", hostContext, err)
	}
}
