package datasnoop

import (
	"encoding/json"
	"io"
	"time"
)

// DemonstrationEvent is one deterministic item in the single-service incident dataset.
type DemonstrationEvent struct {
	Kind      string    `json:"kind"`
	Timestamp time.Time `json:"timestamp"`
	Route     string    `json:"route,omitempty"`
	Status    int       `json:"status,omitempty"`
	Message   string    `json:"message,omitempty"`
	TraceID   string    `json:"traceId,omitempty"`
	SpanID    string    `json:"spanId,omitempty"`
	Name      string    `json:"name,omitempty"`
	Value     float64   `json:"value,omitempty"`
}

// Demonstration emits the canonical successful-request and repeated-error incident dataset.
func Demonstration(writer io.Writer) error {
	stamp := time.Date(2026, 8, 24, 12, 0, 0, 0, time.UTC)
	traceID, spanID := "00112233445566778899aabbccddeeff", "0011223344556677"
	events := []DemonstrationEvent{
		{Kind: "operation", Timestamp: stamp, Route: "/health", Status: 200},
		{Kind: "operation", Timestamp: stamp.Add(time.Second), Route: "/orders/:orderID", Status: 500, TraceID: traceID, SpanID: spanID},
		{Kind: "log", Timestamp: stamp.Add(1100 * time.Millisecond), Message: "payment provider failed", TraceID: traceID, SpanID: spanID},
		{Kind: "operation", Timestamp: stamp.Add(2 * time.Second), Route: "/orders/:orderID", Status: 500, TraceID: traceID, SpanID: spanID},
		{Kind: "metric", Timestamp: stamp.Add(2 * time.Second), Name: "system.cpu.utilization", Value: .82},
		{Kind: "metric", Timestamp: stamp.Add(2 * time.Second), Name: "system.memory.usage", Value: 1048576},
		{Kind: "metric", Timestamp: stamp.Add(2 * time.Second), Name: "system.filesystem.utilization", Value: .41},
	}
	encoder := json.NewEncoder(writer)
	for _, event := range events {
		if err := encoder.Encode(event); err != nil {
			return err
		}
	}
	return nil
}
