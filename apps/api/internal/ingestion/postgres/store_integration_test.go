package postgres

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/datasnoop/datasnoop/apps/api/internal/ingestion"
	"github.com/datasnoop/datasnoop/apps/api/internal/telemetry"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestStorePersistOutcomes(t *testing.T) {
	databaseURL := os.Getenv("DATASNOOP_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATASNOOP_TEST_DATABASE_URL is not set")
	}
	context := context.Background()
	pool, err := pgxpool.New(context, databaseURL)
	if err != nil {
		t.Fatalf("open database pool: %v", err)
	}
	defer pool.Close()
	if _, err := pool.Exec(context, "TRUNCATE metrics, logs, operations, operation_identities, resources, hosts, services RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("truncate telemetry tables: %v", err)
	}

	store := New(pool)
	valid := telemetry.Log{Resource: telemetry.Resource{ServiceName: "checkout"}, Timestamp: time.Now().UTC(), Message: "payment failed"}
	outcome := store.Persist(context, ingestion.Batch{Logs: []telemetry.Log{valid, {}}})
	if outcome.Failed != nil || outcome.Committed != 1 || outcome.Rejected != 1 {
		t.Fatalf("unexpected mixed batch outcome: %+v", outcome)
	}
	var count int
	if err := pool.QueryRow(context, "SELECT count(*) FROM logs").Scan(&count); err != nil || count != 1 {
		t.Fatalf("committed logs = %d, err = %v; want 1", count, err)
	}

	operation := telemetry.Operation{
		Resource: telemetry.Resource{ServiceName: "checkout", Environment: "production"},
		Correlation: telemetry.Correlation{
			TraceID: "00112233445566778899aabbccddeeff",
			SpanID:  "0011223344556677",
		},
		Route:      "/orders/:orderID",
		Method:     "GET",
		StatusCode: 500,
		StartedAt:  time.Date(2026, 8, 22, 12, 0, 0, 0, time.UTC),
		Duration:   25 * time.Millisecond,
	}
	first := store.Persist(context, ingestion.Batch{Operations: []telemetry.Operation{operation}})
	second := store.Persist(context, ingestion.Batch{Operations: []telemetry.Operation{operation}})
	if first.Failed != nil || first.Committed != 1 || first.Repeated != 0 {
		t.Fatalf("unexpected first-delivery outcome: %+v", first)
	}
	if second.Failed != nil || second.Committed != 0 || second.Repeated != 1 {
		t.Fatalf("unexpected retransmission outcome: %+v", second)
	}
	if err := pool.QueryRow(context, "SELECT count(*) FROM operations").Scan(&count); err != nil || count != 1 {
		t.Fatalf("persisted operations = %d, err = %v; want 1", count, err)
	}

	if _, err := pool.Exec(context, "DROP TABLE logs"); err != nil {
		t.Fatalf("drop logs table: %v", err)
	}
	outcome = store.Persist(context, ingestion.Batch{Logs: []telemetry.Log{valid}})
	if outcome.Failed == nil || outcome.Committed != 0 {
		t.Fatalf("unexpected failed batch outcome: %+v", outcome)
	}
}
