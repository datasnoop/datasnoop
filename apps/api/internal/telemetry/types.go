// Package telemetry defines normalized telemetry records independent from OTLP and
// persistence details.
package telemetry

import "time"

const (
	MaxAttributes        = 64
	MaxAttributeKeyBytes = 128
	MaxAttributeValue    = 4096
	MaxStringBytes       = 16 * 1024
)

type SourceRole string

const (
	SourceRoleMonitoredService  SourceRole = "monitored-service"
	SourceRoleDataSnoopPlatform SourceRole = "datasnoop-platform"
)

type Attribute struct {
	Key   string
	Value string
}

type Resource struct {
	ServiceName string
	Environment string
	Attributes  []Attribute
}

type Service struct {
	Name        string
	Environment string
}

type Host struct {
	ID   string
	Name string
}

type Correlation struct {
	TraceID      string
	SpanID       string
	ParentSpanID string
}

type Operation struct {
	Resource    Resource
	Host        Host
	Correlation Correlation
	Route       string
	Method      string
	StatusCode  int
	StartedAt   time.Time
	Duration    time.Duration
	Attributes  []Attribute
}

type Log struct {
	Resource    Resource
	Host        Host
	Correlation Correlation
	Timestamp   time.Time
	Severity    string
	Message     string
	Attributes  []Attribute
}

type Metric struct {
	Resource   Resource
	Host       Host
	SourceRole SourceRole
	Timestamp  time.Time
	Name       string
	Unit       string
	Value      float64
	Attributes []Attribute
}
