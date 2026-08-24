package metrics

import (
	"errors"
	"syscall"
	"testing"
	"time"

	"github.com/datasnoop/datasnoop/apps/api/internal/telemetry"
)

func TestCollectorLabelsPlatformMetricsOnSharedHost(t *testing.T) {
	stamps := []time.Time{time.Date(2026, 8, 24, 12, 0, 0, 0, time.UTC), time.Date(2026, 8, 24, 12, 0, 1, 0, time.UTC)}
	index := 0
	collector := &Collector{hostID: "shared-host", now: func() time.Time { value := stamps[index]; index++; return value }, processCPU: sequenceCPU(100*time.Millisecond, 300*time.Millisecond), statfs: func(_ string, info *syscall.Statfs_t) error { info.Blocks, info.Bavail = 100, 40; return nil }}
	collector.Collect()
	measurements := collector.Collect()
	if len(measurements) != 3 {
		t.Fatalf("measurements=%#v", measurements)
	}
	for _, measurement := range measurements {
		if measurement.Resource.ServiceName != "datasnoop-platform" || measurement.Host.ID != "shared-host" || measurement.SourceRole != telemetry.SourceRoleDataSnoopPlatform {
			t.Fatalf("platform identity lost: %#v", measurement)
		}
	}
	if measurements[0].Name != "process.cpu.utilization" || measurements[0].Value != .2 {
		t.Fatalf("cpu measurement=%#v", measurements[0])
	}
}

func TestCollectorOmitsUnavailablePlatformMeasurements(t *testing.T) {
	collector := &Collector{hostID: "platform", now: time.Now, processCPU: func() (time.Duration, error) { return 0, errors.New("unsupported") }, statfs: func(string, *syscall.Statfs_t) error { return errors.New("unsupported") }}
	measurements := collector.Collect()
	if len(measurements) != 1 || measurements[0].Name != "process.memory.usage" {
		t.Fatalf("measurements=%#v", measurements)
	}
}

func sequenceCPU(values ...time.Duration) func() (time.Duration, error) {
	index := 0
	return func() (time.Duration, error) { value := values[index]; index++; return value, nil }
}
