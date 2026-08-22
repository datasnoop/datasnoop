package postgres_test

import (
	"context"
	"github.com/datasnoop/datasnoop/apps/api/internal/ingestion"
	ingestpostgres "github.com/datasnoop/datasnoop/apps/api/internal/ingestion/postgres"
	"github.com/datasnoop/datasnoop/apps/api/internal/investigation/postgres"
	"github.com/datasnoop/datasnoop/apps/api/internal/telemetry"
	"github.com/jackc/pgx/v5/pgxpool"
	"os"
	"testing"
	"time"
)

func TestEndpointsAggregatePersistedOperationsWithoutRetransmissionInflation(t *testing.T) {
	url := os.Getenv("DATASNOOP_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("DATASNOOP_TEST_DATABASE_URL is not set")
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
	operation := telemetry.Operation{Resource: telemetry.Resource{ServiceName: "checkout", Environment: "production"}, Correlation: telemetry.Correlation{TraceID: "00112233445566778899aabbccddeeff", SpanID: "0011223344556677"}, Route: "/orders/:orderID", Method: "GET", StatusCode: 500, StartedAt: time.Date(2026, 8, 22, 12, 0, 0, 0, time.UTC), Duration: 25 * time.Millisecond}
	store := ingestpostgres.New(pool)
	store.Persist(ctx, ingestion.Batch{Operations: []telemetry.Operation{operation}})
	store.Persist(ctx, ingestion.Batch{Operations: []telemetry.Operation{operation}})
	summaries, err := postgres.New(pool).Endpoints(ctx, "checkout", "production", operation.StartedAt.Add(-time.Minute), operation.StartedAt.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if len(summaries) != 1 || summaries[0].Requests != 1 || summaries[0].Errors != 1 || summaries[0].AverageDuration != 25*time.Millisecond {
		t.Fatalf("summaries=%#v", summaries)
	}
}

func TestOccurrencesCannotCrossServiceOrEnvironment(t *testing.T) {
	url := os.Getenv("DATASNOOP_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("DATASNOOP_TEST_DATABASE_URL is not set")
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
	store := ingestpostgres.New(pool)
	at := time.Date(2026, 8, 22, 12, 0, 0, 0, time.UTC)
	for _, item := range []struct{ name, environment, trace string }{{"checkout", "production", "00112233445566778899aabbccddeeff"}, {"checkout", "staging", "102132435465768798a9babcbddcedfe"}, {"billing", "production", "202132435465768798a9babcbddcedfe"}} {
		outcome := store.Persist(ctx, ingestion.Batch{Operations: []telemetry.Operation{{Resource: telemetry.Resource{ServiceName: item.name, Environment: item.environment}, Correlation: telemetry.Correlation{TraceID: item.trace, SpanID: "0011223344556677"}, Route: "/orders/:orderID", Method: "GET", StatusCode: 500, StartedAt: at, Duration: time.Millisecond}}})
		if outcome.Failed != nil {
			t.Fatal(outcome.Failed)
		}
	}
	result, err := postgres.New(pool).Occurrences(ctx, "checkout", "production", "/orders/:orderID", 500, at.Add(-time.Minute), at.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 || result[0].TraceID != "00112233445566778899aabbccddeeff" {
		t.Fatalf("isolated occurrences=%#v", result)
	}
}
