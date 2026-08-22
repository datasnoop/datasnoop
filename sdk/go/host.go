package datasnoop

import (
	"bytes"
	"os"
	"runtime"
	"strconv"
	"syscall"
	"time"
)

type HostMeasurement struct {
	HostID     string
	SourceRole string
	Timestamp  time.Time
	Name       string
	Unit       string
	Value      float64
}

type HostCollector struct {
	hostID   string
	now      func() time.Time
	statfs   func(string, *syscall.Statfs_t) error
	readFile func(string) ([]byte, error)
	total    uint64
	idle     uint64
}

func NewHostCollector() *HostCollector {
	hostID, err := os.Hostname()
	if err != nil || hostID == "" {
		hostID = "unknown-host"
	}
	return &HostCollector{hostID: hostID, now: time.Now, statfs: syscall.Statfs, readFile: os.ReadFile}
}

func (collector *HostCollector) Collect() []HostMeasurement {
	now := collector.now().UTC()
	measurements := make([]HostMeasurement, 0, 3)
	if total, idle, ok := collector.cpu(); ok {
		if collector.total > 0 && total > collector.total && idle >= collector.idle {
			measurements = append(measurements, collector.measurement(now, "system.cpu.utilization", "1", 1-float64(idle-collector.idle)/float64(total-collector.total)))
		}
		collector.total, collector.idle = total, idle
	}
	var memory runtime.MemStats
	runtime.ReadMemStats(&memory)
	if memory.Sys > 0 {
		measurements = append(measurements, collector.measurement(now, "system.memory.usage", "By", float64(memory.Sys)))
	}
	var filesystem syscall.Statfs_t
	if collector.statfs("/", &filesystem) == nil && filesystem.Blocks > 0 {
		used := filesystem.Blocks - filesystem.Bavail
		measurements = append(measurements, collector.measurement(now, "system.filesystem.utilization", "1", float64(used)/float64(filesystem.Blocks)))
	}
	return measurements
}

func (collector *HostCollector) cpu() (uint64, uint64, bool) {
	if collector.readFile == nil {
		return 0, 0, false
	}
	contents, err := collector.readFile("/proc/stat")
	if err != nil {
		return 0, 0, false
	}
	fields := bytes.Fields(bytes.SplitN(contents, []byte("\n"), 2)[0])
	if len(fields) < 5 || string(fields[0]) != "cpu" {
		return 0, 0, false
	}
	values := make([]uint64, len(fields)-1)
	for index, field := range fields[1:] {
		value, err := strconv.ParseUint(string(field), 10, 64)
		if err != nil {
			return 0, 0, false
		}
		values[index] = value
	}
	total := uint64(0)
	for _, value := range values {
		total += value
	}
	return total, values[3] + func() uint64 {
		if len(values) > 4 {
			return values[4]
		}
		return 0
	}(), true
}

func (collector *HostCollector) measurement(timestamp time.Time, name, unit string, value float64) HostMeasurement {
	return HostMeasurement{HostID: collector.hostID, SourceRole: "monitored-service", Timestamp: timestamp, Name: name, Unit: unit, Value: value}
}
