package retention

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
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
