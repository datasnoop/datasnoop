package graphql_test

import (
	"context"
	"github.com/datasnoop/datasnoop/apps/api/internal/investigation"
	"github.com/datasnoop/datasnoop/apps/api/internal/investigation/graphql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type reader struct{}

func (reader) Endpoints(context.Context, string, string, time.Time, time.Time) ([]investigation.EndpointSummary, error) {
	return []investigation.EndpointSummary{{Route: "/orders/:id", Requests: 2, Errors: 1, AverageDuration: time.Millisecond}}, nil
}
func TestEndpointGraphQLContract(t *testing.T) {
	handler, err := graphql.Handler(reader{})
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodPost, "/graphql", strings.NewReader(`{"query":"{ endpoints(service: \"checkout\", environment: \"production\", from: \"2026-08-22T12:00:00Z\", to: \"2026-08-22T13:00:00Z\", first: 1) { items { route requests errors averageDurationNs } nextCursor } platformDiagnostics { restartId } }"}`))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"route":"/orders/:id"`) {
		t.Fatalf("response=%s", response.Body.String())
	}
}
