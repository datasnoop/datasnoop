package health_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/datasnoop/datasnoop/apps/api/internal/platform/health"
)

func TestLivenessDoesNotDependOnReadiness(t *testing.T) {
	t.Parallel()
	handler := health.Handler(func(_ context.Context) error { return errors.New("database unavailable") })
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/health/live", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("liveness status = %d, want %d", response.Code, http.StatusOK)
	}
}

func TestReadinessReportsDependencyState(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		name  string
		ready health.Readiness
		want  int
	}{
		{"ready", func(context.Context) error { return nil }, http.StatusOK},
		{"unready", func(context.Context) error { return errors.New("database unavailable") }, http.StatusServiceUnavailable},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			health.Handler(testCase.ready).ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/health/ready", nil))
			if response.Code != testCase.want {
				t.Fatalf("readiness status = %d, want %d", response.Code, testCase.want)
			}
		})
	}
}
