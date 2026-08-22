// Package otlp converts the supported OTLP subset into DataSnoop telemetry.
package otlp

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/datasnoop/datasnoop/apps/api/internal/ingestion"
	"github.com/datasnoop/datasnoop/apps/api/internal/telemetry"
	logsv1 "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	metricsv1 "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	tracev1 "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	commonv1 "go.opentelemetry.io/proto/otlp/common/v1"
	logsv1data "go.opentelemetry.io/proto/otlp/logs/v1"
	metricv1 "go.opentelemetry.io/proto/otlp/metrics/v1"
	resourcev1 "go.opentelemetry.io/proto/otlp/resource/v1"
	tracev1data "go.opentelemetry.io/proto/otlp/trace/v1"
)

type Reason string

const (
	ReasonInvalidResource    Reason = "invalid-resource"
	ReasonInvalidAttribute   Reason = "invalid-attribute"
	ReasonInvalidCorrelation Reason = "invalid-correlation"
	ReasonInvalidTimestamp   Reason = "invalid-timestamp"
	ReasonUnsupportedSignal  Reason = "unsupported-signal"
	ReasonPayloadLimit       Reason = "payload-limit"
)

type Result struct {
	Batch    ingestion.Batch
	Rejected int
	Reasons  map[Reason]int
}

func NormalizeTraces(request *tracev1.ExportTraceServiceRequest) Result {
	result := Result{Reasons: make(map[Reason]int)}
	for _, resourceSpans := range request.GetResourceSpans() {
		resource, host, err := normalizeResource(resourceSpans.GetResource())
		for _, scopeSpans := range resourceSpans.GetScopeSpans() {
			for _, span := range scopeSpans.GetSpans() {
				if err != nil {
					result.reject(reasonFor(err))
					continue
				}
				operation, err := normalizeSpan(resource, host, span)
				if err != nil {
					result.reject(reasonFor(err))
					continue
				}
				result.Batch.Operations = append(result.Batch.Operations, operation)
			}
		}
	}
	return result
}

func NormalizeLogs(request *logsv1.ExportLogsServiceRequest) Result {
	result := Result{Reasons: make(map[Reason]int)}
	for _, resourceLogs := range request.GetResourceLogs() {
		resource, host, err := normalizeResource(resourceLogs.GetResource())
		for _, scopeLogs := range resourceLogs.GetScopeLogs() {
			for _, record := range scopeLogs.GetLogRecords() {
				if err != nil {
					result.reject(reasonFor(err))
					continue
				}
				logRecord, err := normalizeLog(resource, host, record)
				if err != nil {
					result.reject(reasonFor(err))
					continue
				}
				result.Batch.Logs = append(result.Batch.Logs, logRecord)
			}
		}
	}
	return result
}

func NormalizeMetrics(request *metricsv1.ExportMetricsServiceRequest) Result {
	result := Result{Reasons: make(map[Reason]int)}
	for _, resourceMetrics := range request.GetResourceMetrics() {
		resource, host, err := normalizeResource(resourceMetrics.GetResource())
		for _, scopeMetrics := range resourceMetrics.GetScopeMetrics() {
			for _, metric := range scopeMetrics.GetMetrics() {
				points := metric.GetGauge().GetDataPoints()
				if metric.GetGauge() == nil {
					result.reject(ReasonUnsupportedSignal)
					continue
				}
				for _, point := range points {
					if err != nil {
						result.reject(reasonFor(err))
						continue
					}
					measurement, err := normalizeMeasurement(resource, host, metric, point, resourceMetrics.GetResource())
					if err != nil {
						result.reject(reasonFor(err))
						continue
					}
					result.Batch.Metrics = append(result.Batch.Metrics, measurement)
				}
			}
		}
	}
	return result
}

func (result *Result) reject(reason Reason) {
	result.Rejected++
	result.Reasons[reason]++
}

func normalizeResource(resource *resourcev1.Resource) (telemetry.Resource, telemetry.Host, error) {
	attributes, values, err := normalizeAttributes(resource.GetAttributes())
	if err != nil {
		return telemetry.Resource{}, telemetry.Host{}, err
	}
	serviceName := values["service.name"]
	if serviceName == "" {
		return telemetry.Resource{}, telemetry.Host{}, classified{ReasonInvalidResource, errors.New("service.name is required")}
	}
	return telemetry.Resource{ServiceName: serviceName, Environment: values["deployment.environment.name"], Attributes: attributes}, telemetry.Host{ID: values["host.id"], Name: values["host.name"]}, nil
}

func normalizeSpan(resource telemetry.Resource, host telemetry.Host, span *tracev1data.Span) (telemetry.Operation, error) {
	if span.GetKind() != tracev1data.Span_SPAN_KIND_SERVER {
		return telemetry.Operation{}, classified{ReasonUnsupportedSignal, errors.New("only server spans are supported")}
	}
	attributes, values, err := normalizeAttributes(span.GetAttributes())
	if err != nil {
		return telemetry.Operation{}, err
	}
	status, err := strconv.Atoi(values["http.response.status_code"])
	if err != nil {
		return telemetry.Operation{}, classified{ReasonUnsupportedSignal, errors.New("HTTP status is required")}
	}
	start, err := timestamp(span.GetStartTimeUnixNano())
	if err != nil {
		return telemetry.Operation{}, err
	}
	end, err := timestamp(span.GetEndTimeUnixNano())
	if err != nil || !end.After(start) {
		return telemetry.Operation{}, classified{ReasonInvalidTimestamp, errors.New("span end time must follow start time")}
	}
	operation := telemetry.Operation{Resource: resource, Host: host, Correlation: telemetry.Correlation{TraceID: hex.EncodeToString(span.GetTraceId()), SpanID: hex.EncodeToString(span.GetSpanId()), ParentSpanID: hex.EncodeToString(span.GetParentSpanId())}, Route: values["http.route"], Method: values["http.request.method"], StatusCode: status, StartedAt: start, Duration: end.Sub(start), Attributes: attributes}
	if err := operation.Validate(); err != nil {
		return telemetry.Operation{}, classified{ReasonInvalidCorrelation, err}
	}
	return operation, nil
}

func normalizeLog(resource telemetry.Resource, host telemetry.Host, record *logsv1data.LogRecord) (telemetry.Log, error) {
	attributes, _, err := normalizeAttributes(record.GetAttributes())
	if err != nil {
		return telemetry.Log{}, err
	}
	body, ok := record.GetBody().GetValue().(*commonv1.AnyValue_StringValue)
	if !ok {
		return telemetry.Log{}, classified{ReasonUnsupportedSignal, errors.New("log body must be a string")}
	}
	nanoseconds := record.GetTimeUnixNano()
	if nanoseconds == 0 {
		nanoseconds = record.GetObservedTimeUnixNano()
	}
	stamp, err := timestamp(nanoseconds)
	if err != nil {
		return telemetry.Log{}, err
	}
	logRecord := telemetry.Log{Resource: resource, Host: host, Correlation: telemetry.Correlation{TraceID: hex.EncodeToString(record.GetTraceId()), SpanID: hex.EncodeToString(record.GetSpanId())}, Timestamp: stamp, Severity: record.GetSeverityText(), Message: body.StringValue, Attributes: attributes}
	if err := logRecord.Validate(); err != nil {
		return telemetry.Log{}, classified{ReasonInvalidCorrelation, err}
	}
	return logRecord, nil
}

func normalizeMeasurement(resource telemetry.Resource, host telemetry.Host, metric *metricv1.Metric, point *metricv1.NumberDataPoint, rawResource *resourcev1.Resource) (telemetry.Metric, error) {
	if !supportedMetric(metric.GetName(), metric.GetUnit()) {
		return telemetry.Metric{}, classified{ReasonUnsupportedSignal, errors.New("unsupported host measurement")}
	}
	_, values, err := normalizeAttributes(rawResource.GetAttributes())
	if err != nil {
		return telemetry.Metric{}, err
	}
	if values["datasnoop.source.role"] != string(telemetry.SourceRoleMonitoredService) {
		return telemetry.Metric{}, classified{ReasonInvalidResource, errors.New("monitored-service source role is required")}
	}
	attributes, _, err := normalizeAttributes(point.GetAttributes())
	if err != nil {
		return telemetry.Metric{}, err
	}
	stamp, err := timestamp(point.GetTimeUnixNano())
	if err != nil {
		return telemetry.Metric{}, err
	}
	value, ok := number(point)
	if !ok || math.IsNaN(value) || math.IsInf(value, 0) || (metric.GetUnit() == "1" && (value < 0 || value > 1)) {
		return telemetry.Metric{}, classified{ReasonUnsupportedSignal, errors.New("invalid measurement value")}
	}
	measurement := telemetry.Metric{Resource: resource, Host: host, SourceRole: telemetry.SourceRoleMonitoredService, Timestamp: stamp, Name: metric.GetName(), Unit: metric.GetUnit(), Value: value, Attributes: attributes}
	if err := measurement.Validate(); err != nil {
		return telemetry.Metric{}, classified{ReasonInvalidResource, err}
	}
	return measurement, nil
}

func normalizeAttributes(attributes []*commonv1.KeyValue) ([]telemetry.Attribute, map[string]string, error) {
	if len(attributes) > telemetry.MaxAttributes {
		return nil, nil, classified{ReasonPayloadLimit, errors.New("too many attributes")}
	}
	result := make([]telemetry.Attribute, 0, len(attributes))
	values := make(map[string]string, len(attributes))
	for _, attribute := range attributes {
		if attribute.GetKey() == "" || attribute.GetValue() == nil {
			return nil, nil, classified{ReasonInvalidAttribute, errors.New("attribute key and value are required")}
		}
		value, ok := scalar(attribute.GetValue())
		if !ok {
			return nil, nil, classified{ReasonInvalidAttribute, fmt.Errorf("unsupported value for %q", attribute.GetKey())}
		}
		result = append(result, telemetry.Attribute{Key: attribute.GetKey(), Value: value})
		values[attribute.GetKey()] = value
	}
	if err := (telemetry.Resource{ServiceName: "attributes", Attributes: result}).Validate(); err != nil {
		return nil, nil, classified{ReasonInvalidAttribute, err}
	}
	return result, values, nil
}

func scalar(value *commonv1.AnyValue) (string, bool) {
	switch typed := value.GetValue().(type) {
	case *commonv1.AnyValue_StringValue:
		return typed.StringValue, true
	case *commonv1.AnyValue_BoolValue:
		return strconv.FormatBool(typed.BoolValue), true
	case *commonv1.AnyValue_IntValue:
		return strconv.FormatInt(typed.IntValue, 10), true
	case *commonv1.AnyValue_DoubleValue:
		return strconv.FormatFloat(typed.DoubleValue, 'g', -1, 64), true
	case *commonv1.AnyValue_BytesValue:
		return string(typed.BytesValue), true
	default:
		return "", false
	}
}

func timestamp(nanoseconds uint64) (time.Time, error) {
	if nanoseconds == 0 || nanoseconds > math.MaxInt64 {
		return time.Time{}, classified{ReasonInvalidTimestamp, errors.New("timestamp is required")}
	}
	return time.Unix(0, int64(nanoseconds)).UTC(), nil
}

func number(point *metricv1.NumberDataPoint) (float64, bool) {
	switch value := point.GetValue().(type) {
	case *metricv1.NumberDataPoint_AsDouble:
		return value.AsDouble, true
	case *metricv1.NumberDataPoint_AsInt:
		return float64(value.AsInt), true
	default:
		return 0, false
	}
}

func supportedMetric(name, unit string) bool {
	return (name == "system.cpu.utilization" || name == "system.memory.utilization" || name == "system.filesystem.utilization") && unit == "1" || (name == "system.memory.usage" || name == "system.filesystem.usage") && unit == "By"
}

type classified struct {
	reason Reason
	err    error
}

func (error classified) Error() string { return error.err.Error() }

func reasonFor(err error) Reason {
	var classifiedError classified
	if errors.As(err, &classifiedError) {
		return classifiedError.reason
	}
	return ReasonUnsupportedSignal
}
