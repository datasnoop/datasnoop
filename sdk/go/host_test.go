package datasnoop

import (
	"errors"
	"syscall"
	"testing"
	"time"
)

func TestHostCollectorUsesStableIdentityAndMonitoredServiceRole(t *testing.T) {
	collector := &HostCollector{hostID: "checkout-host", now: func() time.Time { return time.Unix(1724356800, 0) }, statfs: func(_ string, info *syscall.Statfs_t) error { info.Blocks, info.Bavail = 100, 25; return nil }, readFile: func(string) ([]byte, error) { return []byte("cpu  10 0 0 10 0\n"), nil }}
	measurements := collector.Collect()
	if len(measurements) == 0 {
		t.Fatal("expected supported host measurements")
	}
	for _, measurement := range measurements {
		if measurement.HostID != "checkout-host" || measurement.SourceRole != "monitored-service" || measurement.Timestamp.IsZero() {
			t.Fatalf("invalid host measurement: %#v", measurement)
		}
	}
}

func TestHostCollectorOmitsUnavailableFilesystemMeasurement(t *testing.T) {
	collector := &HostCollector{hostID: "checkout-host", now: time.Now, statfs: func(string, *syscall.Statfs_t) error { return errors.New("unsupported") }, readFile: func(string) ([]byte, error) { return nil, errors.New("unsupported") }}
	for _, measurement := range collector.Collect() {
		if measurement.Name == "system.filesystem.utilization" {
			t.Fatalf("fabricated unavailable measurement: %#v", measurement)
		}
	}
}
