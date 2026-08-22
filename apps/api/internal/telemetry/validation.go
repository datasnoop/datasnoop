package telemetry

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
	"unicode/utf8"
)

func (resource Resource) Validate() error {
	if err := validateBoundedString("service name", resource.ServiceName, MaxAttributeKeyBytes); err != nil {
		return err
	}
	if resource.Environment != "" {
		if err := validateBoundedString("environment", resource.Environment, MaxAttributeKeyBytes); err != nil {
			return err
		}
	}
	return validateAttributes(resource.Attributes)
}

func (host Host) Validate(required bool) error {
	if !required && host.ID == "" && host.Name == "" {
		return nil
	}
	if host.ID == "" && host.Name == "" {
		return errors.New("host ID or name is required")
	}
	if host.ID != "" {
		return validateBoundedString("host ID", host.ID, MaxAttributeKeyBytes)
	}
	return validateBoundedString("host name", host.Name, MaxAttributeKeyBytes)
}

func (correlation Correlation) Validate(required bool) error {
	if correlation.TraceID == "" && correlation.SpanID == "" && correlation.ParentSpanID == "" {
		if required {
			return errors.New("trace and span IDs are required")
		}
		return nil
	}
	if !isHexIdentifier(correlation.TraceID, 32) {
		return errors.New("trace ID must be 32 hexadecimal characters")
	}
	if !isHexIdentifier(correlation.SpanID, 16) {
		return errors.New("span ID must be 16 hexadecimal characters")
	}
	if correlation.ParentSpanID != "" && !isHexIdentifier(correlation.ParentSpanID, 16) {
		return errors.New("parent span ID must be 16 hexadecimal characters")
	}
	return nil
}

func (operation Operation) Validate() error {
	return errors.Join(
		operation.Resource.Validate(),
		operation.Host.Validate(false),
		operation.Correlation.Validate(true),
		validateBoundedString("route", operation.Route, 256),
		validateBoundedString("method", operation.Method, 16),
		validateStatusCode(operation.StatusCode),
		validateTimestamp(operation.StartedAt),
		validatePositiveDuration(operation.Duration),
		validateAttributes(operation.Attributes),
	)
}

func (logRecord Log) Validate() error {
	return errors.Join(
		logRecord.Resource.Validate(),
		logRecord.Host.Validate(false),
		logRecord.Correlation.Validate(false),
		validateTimestamp(logRecord.Timestamp),
		validateBoundedString("message", logRecord.Message, MaxStringBytes),
		validateAttributes(logRecord.Attributes),
	)
}

func (metric Metric) Validate() error {
	if metric.SourceRole != SourceRoleMonitoredService && metric.SourceRole != SourceRoleDataSnoopPlatform {
		return fmt.Errorf("unsupported source role %q", metric.SourceRole)
	}
	if math.IsNaN(metric.Value) || math.IsInf(metric.Value, 0) {
		return errors.New("metric value must be finite")
	}
	return errors.Join(
		metric.Resource.Validate(),
		metric.Host.Validate(true),
		validateTimestamp(metric.Timestamp),
		validateBoundedString("metric name", metric.Name, MaxAttributeKeyBytes),
		validateBoundedString("metric unit", metric.Unit, 32),
		validateAttributes(metric.Attributes),
	)
}

func validateAttributes(attributes []Attribute) error {
	if len(attributes) > MaxAttributes {
		return fmt.Errorf("attribute count exceeds %d", MaxAttributes)
	}
	for _, attribute := range attributes {
		if err := validateBoundedString("attribute key", attribute.Key, MaxAttributeKeyBytes); err != nil {
			return err
		}
		if err := validateBoundedString("attribute value", attribute.Value, MaxAttributeValue); err != nil {
			return err
		}
	}
	return nil
}

func validateBoundedString(name, value string, maximum int) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s is required", name)
	}
	if !utf8.ValidString(value) {
		return fmt.Errorf("%s must be valid UTF-8", name)
	}
	if len(value) > maximum {
		return fmt.Errorf("%s exceeds %d bytes", name, maximum)
	}
	return nil
}

func validateStatusCode(statusCode int) error {
	if statusCode < 100 || statusCode > 599 {
		return errors.New("HTTP status code must be between 100 and 599")
	}
	return nil
}

func validateTimestamp(timestamp time.Time) error {
	if timestamp.IsZero() {
		return errors.New("timestamp is required")
	}
	return nil
}

func validatePositiveDuration(duration time.Duration) error {
	if duration <= 0 {
		return errors.New("duration must be positive")
	}
	return nil
}

func isHexIdentifier(value string, length int) bool {
	if len(value) != length || strings.Trim(value, "0") == "" {
		return false
	}
	for _, character := range value {
		if !(character >= '0' && character <= '9') && !(character >= 'a' && character <= 'f') && !(character >= 'A' && character <= 'F') {
			return false
		}
	}
	return true
}
