package persistence

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/datasnoop/datasnoop/apps/api/internal/domain"
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
	if _, err := pool.Exec(context, "TRUNCATE metrics, logs, operations, resources, hosts, services RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("truncate telemetry tables: %v", err)
	}

	store := New(pool)
	valid := domain.Log{Resource: domain.Resource{ServiceName: "checkout"}, Timestamp: time.Now().UTC(), Message: "payment failed"}
	outcome := store.Persist(context, Batch{Logs: []domain.Log{valid, {}}})
	if outcome.Failed != nil || outcome.Committed != 1 || outcome.Rejected != 1 {
		t.Fatalf("unexpected mixed batch outcome: %+v", outcome)
	}
	var count int
	if err := pool.QueryRow(context, "SELECT count(*) FROM logs").Scan(&count); err != nil || count != 1 {
		t.Fatalf("committed logs = %d, err = %v; want 1", count, err)
	}

	if _, err := pool.Exec(context, "DROP TABLE logs"); err != nil {
		t.Fatalf("drop logs table: %v", err)
	}
	outcome = store.Persist(context, Batch{Logs: []domain.Log{valid}})
	if outcome.Failed == nil || outcome.Committed != 0 {
		t.Fatalf("unexpected failed batch outcome: %+v", outcome)
	}
}
