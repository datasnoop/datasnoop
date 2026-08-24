package retention_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/datasnoop/datasnoop/apps/api/internal/platform/retention"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestTimescaleRetentionRemovesExpiredChunks(t *testing.T) {
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
	if _, err := pool.Exec(ctx, "TRUNCATE retention_cycles, metrics, logs, operations, operation_identities, resources, hosts, services RESTART IDENTITY CASCADE"); err != nil {
		t.Fatal(err)
	}
	var serviceID, hostID, resourceID int64
	if err := pool.QueryRow(ctx, "INSERT INTO services (name, environment) VALUES ('retention-test', 'test') RETURNING id").Scan(&serviceID); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, "INSERT INTO hosts (stable_id) VALUES ('retention-host') RETURNING id").Scan(&hostID); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, "INSERT INTO resources (service_id, host_id) VALUES ($1,$2) RETURNING id", serviceID, hostID).Scan(&resourceID); err != nil {
		t.Fatal(err)
	}
	old := time.Now().UTC().AddDate(-2, 0, 0)
	fresh := time.Now().UTC().Add(-time.Hour)
	if _, err := pool.Exec(ctx, "INSERT INTO logs (service_id, resource_id, timestamp, message) VALUES ($1,$2,$3,'expired'),($1,$2,$4,'recent')", serviceID, resourceID, old, fresh); err != nil {
		t.Fatal(err)
	}
	runner := retention.NewPoolRunner(pool, func() retention.Policy { return retention.Policy{Duration: 30 * 24 * time.Hour} })
	cycle := runner.Run(ctx)
	if cycle.Failure != nil {
		t.Fatal(cycle.Failure)
	}
	var expired, recent, completed int
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM logs WHERE message='expired'").Scan(&expired); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM logs WHERE message='recent'").Scan(&recent); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM retention_cycles WHERE status='completed'").Scan(&completed); err != nil {
		t.Fatal(err)
	}
	if expired != 0 || recent != 1 || completed != 1 {
		t.Fatalf("expired=%d recent=%d completed=%d", expired, recent, completed)
	}
}
