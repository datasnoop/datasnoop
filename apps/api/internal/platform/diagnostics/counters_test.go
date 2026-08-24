package diagnostics_test

import (
	"testing"

	"github.com/datasnoop/datasnoop/apps/api/internal/platform/diagnostics"
)

func TestCountersRemainMonotonicWithinOneRestart(t *testing.T) {
	t.Parallel()
	counters := diagnostics.NewCounters()
	counters.Accept(2)
	counters.Reject(1)
	counters.Throttle(3)
	counters.Drop(4)
	first := counters.Snapshot()
	counters.Accept(1)
	second := counters.Snapshot()
	if first.RestartID == "" || second.RestartID != first.RestartID || second.Accepted <= first.Accepted {
		t.Fatalf("unexpected monotonic snapshots: %#v then %#v", first, second)
	}
}

func TestNewCountersHaveDistinctRestartIdentity(t *testing.T) {
	t.Parallel()
	previous := diagnostics.NewCounters()
	previous.Accept(10)
	restarted := diagnostics.NewCounters()
	current := restarted.Snapshot()
	if current.RestartID == previous.Snapshot().RestartID || current.Accepted != 0 {
		t.Fatalf("counter reset lacks restart boundary: previous=%#v current=%#v", previous.Snapshot(), current)
	}
}
