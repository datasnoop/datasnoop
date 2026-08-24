package retention

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/datasnoop/datasnoop/apps/api/internal/ingestion"
)

func TestRunnerExpiresChunksAndRecordsCompletedCycle(t *testing.T) {
	database := &recordingExecutor{}
	runner := NewRunner(database, func() Policy { return Policy{Duration: 7 * 24 * time.Hour} })
	runner.now = clock(time.Date(2026, 8, 24, 12, 0, 0, 0, time.UTC))
	cycle := runner.Run(context.Background())
	if cycle.Failure != nil || len(database.calls) != 4 {
		t.Fatalf("cycle=%+v calls=%#v", cycle, database.calls)
	}
	for _, relation := range []string{"operations", "logs", "metrics"} {
		if !contains(database.calls, relation) {
			t.Fatalf("retention did not expire %s: %#v", relation, database.calls)
		}
	}
	if !contains(database.calls, "'completed'") {
		t.Fatalf("completed cycle was not recorded: %#v", database.calls)
	}
}

func TestRunnerRecordsFailureAndScheduleHasBound(t *testing.T) {
	database := &recordingExecutor{failOn: "logs"}
	runner := NewRunner(database, func() Policy { return Policy{Duration: Default} })
	runner.now = clock(time.Date(2026, 8, 24, 12, 0, 0, 0, time.UTC))
	cycle := runner.Run(context.Background())
	if cycle.Failure == nil || !contains(database.calls, "'failed'") {
		t.Fatalf("cycle=%+v calls=%#v", cycle, database.calls)
	}
	if _, err := runner.Schedule(context.Background(), time.Second); err == nil {
		t.Fatal("expected too-short schedule to be rejected")
	}
}

func TestConcurrentRetentionDoesNotConsumeIngestionBudget(t *testing.T) {
	database := &purgeBlockingExecutor{started: make(chan struct{}), release: make(chan struct{})}
	budget, err := NewBudget(1)
	if err != nil {
		t.Fatal(err)
	}
	runner := NewRunner(database, func() Policy { return Policy{Duration: Default} }).WithBudget(budget)
	first := make(chan Cycle, 1)
	go func() { first <- runner.Run(context.Background()) }()
	<-database.started
	if cycle := runner.Run(context.Background()); !errors.Is(cycle.Failure, ErrWorkBudgetExhausted) {
		t.Fatalf("overlapping cycle = %+v", cycle)
	}
	processor, err := ingestion.NewProcessor(ingestion.ProcessorConfig{AdmissionLimit: 1, QueueLimit: 1, Workers: 1, PersistenceTimeout: time.Second}, instantStore{})
	if err != nil {
		t.Fatal(err)
	}
	defer processor.Close()
	started := time.Now()
	if _, err := processor.Persist(context.Background(), ingestion.Batch{}); err != nil {
		t.Fatalf("ingestion during purge: %v", err)
	}
	if elapsed := time.Since(started); elapsed > 100*time.Millisecond {
		t.Fatalf("ingestion degradation budget exceeded: %v", elapsed)
	}
	close(database.release)
	if cycle := <-first; cycle.Failure != nil {
		t.Fatalf("first cycle = %+v", cycle)
	}
}

type recordingExecutor struct {
	calls  []string
	failOn string
}

func (executor *recordingExecutor) Exec(_ context.Context, query string, arguments ...any) error {
	executor.calls = append(executor.calls, query+" "+strings.TrimSpace(strings.Join(stringify(arguments), " ")))
	if executor.failOn != "" && len(arguments) > 0 && arguments[0] == executor.failOn {
		return errors.New("database unavailable")
	}
	return nil
}

func stringify(arguments []any) []string {
	values := make([]string, len(arguments))
	for index, argument := range arguments {
		values[index] = strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(toString(argument), "\n", " "), "\t", " "))
	}
	return values
}

func toString(value any) string {
	if text, ok := value.(string); ok {
		return text
	}
	return "value"
}

func contains(values []string, needle string) bool {
	for _, value := range values {
		if strings.Contains(value, needle) {
			return true
		}
	}
	return false
}

func clock(value time.Time) func() time.Time { return func() time.Time { return value } }

type instantStore struct{}

func (instantStore) Persist(context.Context, ingestion.Batch) ingestion.Outcome {
	return ingestion.Outcome{}
}

type purgeBlockingExecutor struct {
	started chan struct{}
	release chan struct{}
	once    sync.Once
}

func (executor *purgeBlockingExecutor) Exec(ctx context.Context, query string, _ ...any) error {
	if strings.Contains(query, "drop_chunks") {
		executor.once.Do(func() { close(executor.started) })
		select {
		case <-executor.release:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}
