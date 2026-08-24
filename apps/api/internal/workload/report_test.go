package workload

import "testing"

func TestReportsMakeOverloadAndRecoveryExplicit(t *testing.T) {
	report := Report{Profile: "saturation-v1", Sent: 400, Accepted: 300, Throttled: 80, Dropped: 20, CPU: 80, Memory: 128, Connections: 2, QueueDepth: 1, HistoricalVisibilityMS: 10, LiveVisibilityMS: 20, InvestigationLatencyMS: 15, SSEGaps: 1, SemanticCorrect: true}
	if err := report.Validate(); err != nil {
		t.Fatal(err)
	}
	report.SemanticCorrect = false
	if err := report.Validate(); err == nil {
		t.Fatal("invalid semantics were accepted")
	}
}
