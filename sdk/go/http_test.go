package datasnoop_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/datasnoop/datasnoop/sdk/go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestHTTPEmitsNormalizedOperationForSuccessAndError(t *testing.T) {
	recorder := tracetest.NewSpanRecorder()
	provider := trace.NewTracerProvider(trace.WithSpanProcessor(recorder))
	previous := otel.GetTracerProvider()
	otel.SetTracerProvider(provider)
	t.Cleanup(func() { otel.SetTracerProvider(previous) })

	for _, status := range []int{http.StatusOK, http.StatusInternalServerError} {
		handler := datasnoop.HTTP(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) { writer.WriteHeader(status) }), func(*http.Request) string { return "/orders/:orderID" })
		handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/orders/123456", nil))
	}
	spans := recorder.Ended()
	if len(spans) != 2 {
		t.Fatalf("emitted spans = %d, want 2", len(spans))
	}
	for index, span := range spans {
		attributes := span.Attributes()
		wantStatus := int64(http.StatusOK)
		if index == 1 {
			wantStatus = http.StatusInternalServerError
		}
		if stringAttribute(attributes, "http.route") != "/orders/:orderID" || intAttribute(attributes, "http.response.status_code") != wantStatus {
			t.Fatalf("span lacks normalized operation fields: %#v", attributes)
		}
	}
}

func intAttribute(attributes []attribute.KeyValue, key string) int64 {
	for _, attribute := range attributes {
		if string(attribute.Key) == key {
			return attribute.Value.AsInt64()
		}
	}
	return 0
}

func stringAttribute(attributes []attribute.KeyValue, key string) string {
	for _, attribute := range attributes {
		if string(attribute.Key) == key {
			return attribute.Value.AsString()
		}
	}
	return ""
}
