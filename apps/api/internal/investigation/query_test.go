package investigation_test

import (
	"github.com/datasnoop/datasnoop/apps/api/internal/investigation"
	"testing"
	"time"
)

func TestEndpointSummaryCarriesRateErrorAndDurationFields(t *testing.T) {
	summary := investigation.EndpointSummary{Route: "/orders/:orderID", Requests: 2, Errors: 1, AverageDuration: 25 * time.Millisecond}
	if summary.Route == "" || summary.Requests != 2 || summary.Errors != 1 || summary.AverageDuration <= 0 {
		t.Fatalf("invalid endpoint summary: %#v", summary)
	}
}
