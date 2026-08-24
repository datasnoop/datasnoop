package datasnoop

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"testing"
)

func TestDemonstrationProducesCanonicalIncidentDataset(t *testing.T) {
	var output bytes.Buffer
	if err := Demonstration(&output); err != nil {
		t.Fatal(err)
	}
	decoder := json.NewDecoder(&output)
	events := []DemonstrationEvent{}
	for {
		var event DemonstrationEvent
		if err := decoder.Decode(&event); errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			t.Fatal(err)
		}
		events = append(events, event)
	}
	if len(events) != 7 {
		t.Fatalf("events=%#v", events)
	}
	if events[0].Status != 200 || events[1].Route != "/orders/:orderID" || events[1].Status != 500 || events[3].TraceID != events[1].TraceID || events[2].SpanID != events[1].SpanID {
		t.Fatalf("operation and log correlation lost: %#v", events)
	}
	if events[4].Name != "system.cpu.utilization" || events[5].Name != "system.memory.usage" || events[6].Name != "system.filesystem.utilization" {
		t.Fatalf("host measurements missing: %#v", events)
	}
}
