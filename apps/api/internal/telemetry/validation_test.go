package telemetry

import (
	"math"
	"strings"
	"testing"
	"time"
)

func validResource() Resource {
	return Resource{ServiceName: "checkout", Environment: "production"}
}

func validCorrelation() Correlation {
	return Correlation{TraceID: "00112233445566778899aabbccddeeff", SpanID: "0011223344556677"}
}

func TestOperationValidationAcceptsCorrelationAndTimestamps(t *testing.T) {
	operation := Operation{
		Resource:    validResource(),
		Correlation: validCorrelation(),
		Route:       "/orders/:orderID",
		Method:      "GET",
		StatusCode:  500,
		StartedAt:   time.Now().UTC(),
		Duration:    10 * time.Millisecond,
	}
	if err := operation.Validate(); err != nil {
		t.Fatalf("validate operation: %v", err)
	}
}

func TestValidationRejectsInvalidCorrelationAndTimestamps(t *testing.T) {
	logRecord := Log{Resource: validResource(), Correlation: Correlation{TraceID: "bad", SpanID: "0011223344556677"}, Message: "failed"}
	if err := logRecord.Validate(); err == nil {
		t.Fatal("expected invalid correlation and timestamp to fail")
	}
}

func TestValidationRejectsInvalidAttributes(t *testing.T) {
	attributes := make([]Attribute, MaxAttributes+1)
	for index := range attributes {
		attributes[index] = Attribute{Key: "key", Value: "value"}
	}
	if err := (Resource{ServiceName: "checkout", Attributes: attributes}).Validate(); err == nil {
		t.Fatal("expected too many attributes to fail")
	}
	if err := (Resource{ServiceName: "checkout", Attributes: []Attribute{{Key: "key", Value: strings.Repeat("x", MaxAttributeValue+1)}}}).Validate(); err == nil {
		t.Fatal("expected oversized attribute value to fail")
	}
}

func TestMetricValidationRequiresKnownSourceRoleAndHost(t *testing.T) {
	metric := Metric{Resource: validResource(), SourceRole: "unknown", Timestamp: time.Now().UTC(), Name: "system.cpu.utilization", Unit: "1", Value: 0.5}
	if err := metric.Validate(); err == nil {
		t.Fatal("expected unknown source role to fail")
	}
	metric.SourceRole = SourceRoleMonitoredService
	if err := metric.Validate(); err == nil {
		t.Fatal("expected missing host to fail")
	}
	metric.Host = Host{ID: "host-01"}
	metric.Value = math.Inf(1)
	if err := metric.Validate(); err == nil {
		t.Fatal("expected non-finite value to fail")
	}
}
