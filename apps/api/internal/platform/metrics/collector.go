// Package metrics collects bounded operational measurements for DataSnoop itself.
package metrics

import (
	"os"
	"runtime"
	"syscall"
	"time"

	"github.com/datasnoop/datasnoop/apps/api/internal/telemetry"
)

// Collector emits DataSnoop's host measurements without mixing them with monitored-service data.
type Collector struct {
	hostID      string
	now         func() time.Time
	processCPU  func() (time.Duration, error)
	statfs      func(string, *syscall.Statfs_t) error
	previousCPU time.Duration
	previousAt  time.Time
}

func NewCollector() *Collector {
	hostID, err := os.Hostname()
	if err != nil || hostID == "" {
		hostID = "datasnoop-host"
	}
	return &Collector{hostID: hostID, now: time.Now, processCPU: readProcessCPU, statfs: syscall.Statfs}
}

func (collector *Collector) Collect() []telemetry.Metric {
	now := collector.now().UTC()
	measurements := make([]telemetry.Metric, 0, 3)
	if cpu, err := collector.processCPU(); err == nil && !collector.previousAt.IsZero() && now.After(collector.previousAt) && cpu >= collector.previousCPU {
		measurements = append(measurements, collector.measurement(now, "process.cpu.utilization", "1", float64(cpu-collector.previousCPU)/float64(now.Sub(collector.previousAt))))
		collector.previousCPU = cpu
		collector.previousAt = now
	} else if err == nil {
		collector.previousCPU = cpu
		collector.previousAt = now
	}
	var memory runtime.MemStats
	runtime.ReadMemStats(&memory)
	if memory.Sys > 0 {
		measurements = append(measurements, collector.measurement(now, "process.memory.usage", "By", float64(memory.Sys)))
	}
	var filesystem syscall.Statfs_t
	if collector.statfs("/", &filesystem) == nil && filesystem.Blocks > 0 {
		measurements = append(measurements, collector.measurement(now, "system.filesystem.utilization", "1", float64(filesystem.Blocks-filesystem.Bavail)/float64(filesystem.Blocks)))
	}
	return measurements
}

func (collector *Collector) measurement(timestamp time.Time, name, unit string, value float64) telemetry.Metric {
	return telemetry.Metric{Resource: telemetry.Resource{ServiceName: "datasnoop-platform"}, Host: telemetry.Host{ID: collector.hostID}, SourceRole: telemetry.SourceRoleDataSnoopPlatform, Timestamp: timestamp, Name: name, Unit: unit, Value: value}
}

func readProcessCPU() (time.Duration, error) {
	var usage syscall.Rusage
	if err := syscall.Getrusage(syscall.RUSAGE_SELF, &usage); err != nil {
		return 0, err
	}
	return time.Duration(usage.Utime.Sec+usage.Stime.Sec)*time.Second + time.Duration(usage.Utime.Usec+usage.Stime.Usec)*time.Microsecond, nil
}
