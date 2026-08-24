package workload

import "fmt"

// Report records bounded-load evidence for a versioned profile.
type Report struct {
	Profile                                                          string
	Sent, Accepted, Rejected, Throttled, Dropped                     int
	CPU, Memory, Connections, QueueDepth                             int
	HistoricalVisibilityMS, LiveVisibilityMS, InvestigationLatencyMS int
	SSEGaps                                                          int
	SemanticCorrect                                                  bool
}

func (report Report) Validate() error {
	if report.Profile == "" || report.Sent < report.Accepted+report.Rejected+report.Throttled+report.Dropped || report.QueueDepth < 0 || !report.SemanticCorrect {
		return fmt.Errorf("invalid workload report for %q", report.Profile)
	}
	return nil
}
