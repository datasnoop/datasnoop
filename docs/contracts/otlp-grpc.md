# OTLP/gRPC Ingestion Contract

## Status and scope

This document defines the supported OTLP/gRPC subset for the single-service
DataSnoop investigation journey. It is versioned as contract version `v1`.
Unsupported fields and signal shapes are rejected explicitly; they are never
silently discarded. OTLP is the transport contract, while DataSnoop uses the
product terms service, operation, log, and measurement after normalization.

The receiver implements the standard unary `Export` methods for:

- `opentelemetry.proto.collector.logs.v1.LogsService`;
- `opentelemetry.proto.collector.trace.v1.TraceService`; and
- `opentelemetry.proto.collector.metrics.v1.MetricsService`.

OTLP/HTTP, client-streaming RPCs, and Collector-specific configuration are not
supported in this contract version.

## Authentication and transport

Clients connect to the configured gRPC endpoint with TLS in production. The
receiver supports the OTLP-required `none` and `gzip` transport compression
options. Every `Export` request MUST include exactly one `x-datasnoop-token` metadata value.
The value is an opaque service credential. Credentials MUST NOT be supplied in
URLs, resource attributes, log attributes, span attributes, or metric labels.

Missing, malformed, expired, or unknown credentials return gRPC
`Unauthenticated`. The receiver persists no record from an unauthenticated
request. A request that exceeds the configured 4 MiB decoded-message limit is
rejected before record validation with non-retryable gRPC
`ResourceExhausted`.

The receiver enforces a 10-second export deadline and accepts at most 64
concurrent export calls. Capacity exhaustion returns retryable gRPC
`Unavailable` with a `RetryInfo` delay; callers SHOULD honor that delay and use
bounded exponential backoff.

## Shared resource and attribute rules

Each accepted record MUST resolve to a service identity through the enclosing
resource attribute `service.name`. `service.name` is a non-empty UTF-8 string
of at most 128 bytes. `deployment.environment.name`, when present, is a UTF-8
string of at most 128 bytes. A missing environment normalizes to `default`.

The receiver recognizes `host.id` or `host.name` for host identity. At least
one is required for supported host measurements. `host.id` takes precedence
when both are supplied.

For every supported resource, span, log, or data-point attribute set:

- at most 64 attributes are accepted;
- attribute keys are non-empty UTF-8 strings of at most 128 bytes;
- string attribute values are at most 4 KiB UTF-8 bytes;
- byte-array attribute values are at most 4 KiB;
- array values contain at most 32 scalar values; and
- nested maps and arrays of arrays are unsupported.

An attribute rule violation rejects that record. The receiver does not truncate
accepted record attributes. `service.name`, environment, host identity,
correlation identifiers, timestamps, severity, and the semantic attributes
listed below are normalized into typed fields. Supported remaining attributes
are retained as bounded additional attributes.

## Supported traces: HTTP operations

DataSnoop accepts server spans that describe HTTP operations. A supported span
MUST have a valid non-zero 16-byte trace ID, valid non-zero 8-byte span ID, a
positive end time not earlier than its start time, and `service.name`.

The following semantic attributes are recognized:

| OTLP attribute | Normalized operation field | Rule |
| --- | --- | --- |
| `http.request.method` | method | Required non-empty string, at most 16 bytes. |
| `http.route` | normalized route | Required non-empty string, at most 256 bytes. |
| `http.response.status_code` | HTTP status | Required integer from 100 through 599. |
| `server.address` | server address | Optional string. |
| `url.scheme` | scheme | Optional string. |

The receiver preserves the parent span ID when valid. Span status `ERROR` or an
HTTP status from 500 through 599 marks the normalized operation as an error.
The operation timestamp is the span start time and duration is end minus start.
Client, producer, consumer, internal, and spans without the required HTTP
fields are unsupported records in this version.

## Supported logs

DataSnoop accepts OTLP log records with `service.name`, a non-zero observed or
event timestamp, and a body that is a UTF-8 string of at most 16 KiB. If both
timestamps are supplied, the event timestamp is used; otherwise the observed
timestamp is used. Severity number and severity text are retained when present.

Trace and span IDs are optional for logs. When supplied, each MUST be valid and
non-zero; a supplied span ID requires a supplied trace ID. Valid identifiers
are stored unchanged so logs and operations can be correlated. Logs without
correlation are retained as service logs and are not presented as correlated
evidence for an operation.

Log attributes follow the shared attribute rules. Non-string log bodies,
invalid timestamps, invalid correlation identifiers, or oversized bodies reject
the affected log record.

## Supported metrics: host measurements

DataSnoop accepts only numeric gauge data points that describe the monitored
application host. Each accepted point MUST have `service.name`, `host.id` or
`host.name`, a non-zero timestamp, a finite numeric value, and a unit matching
the supported name.

| OTLP metric name | Unit | Normalized measurement |
| --- | --- | --- |
| `system.cpu.utilization` | `1` | CPU utilization from 0 through 1. |
| `system.memory.utilization` | `1` | Memory utilization from 0 through 1. |
| `system.filesystem.utilization` | `1` | Filesystem utilization from 0 through 1. |
| `system.memory.usage` | `By` | Memory usage in bytes. |
| `system.filesystem.usage` | `By` | Filesystem usage in bytes. |

Metric resources MUST carry `datasnoop.source.role=monitored-service`. Points
with an unsupported metric type, name, unit, aggregation temporality, missing
host identity, non-finite value, or out-of-range utilization are rejected.
DataSnoop-platform metrics use the same subset with
`datasnoop.source.role=datasnoop-platform` after platform collection is added;
they are not accepted through a monitored-service credential in this version.

## Validation, outcomes, and partial success

Validation is record-granular: one span, log record, or numeric data point is
one record. A rejected record contributes one count to its signal's
partial-success response. Its reason category is one of:

- `invalid-resource`;
- `invalid-attribute`;
- `invalid-correlation`;
- `invalid-timestamp`;
- `unsupported-signal`; or
- `payload-limit`.

When an authenticated request contains at least one valid record, the receiver
persists only valid records and returns the OTLP partial-success response for
that signal. `rejected_*` equals the rejected record count, and the response
error message summarizes the reason categories without including sensitive
attribute values.

When every record in an authenticated request is invalid, the receiver returns
gRPC `InvalidArgument`, persists no record, and reports the applicable reason
categories. Authentication, encoded-message limits, overload, and persistence
timeouts fail the whole request before partial-success handling. A persistence
failure after validation returns gRPC `Unavailable` and persists no partial
batch.

## Scenario traceability

Every telemetry-ingestion specification scenario maps to a contract rule:

| Specification scenario | Contract section | Expected contract behavior |
| --- | --- | --- |
| Valid exporter sends telemetry | Status and scope; Authentication and transport | Standard unary `Export` accepts authenticated supported records. |
| Invalid credential is rejected | Authentication and transport | `Unauthenticated`; no persistence. |
| Mixed valid and invalid records | Validation, outcomes, and partial success | Valid records persist; rejected count and categories return through partial success. |
| Entire request is invalid | Validation, outcomes, and partial success | `InvalidArgument`; no persistence. |
| Correlated log and operation arrive | Supported traces: HTTP operations; Supported logs | Valid trace and span IDs are retained unchanged. |
| Attributes exceed documented limits | Shared resource and attribute rules | The affected record is rejected with `invalid-attribute` or `payload-limit`. |
| Capacity is temporarily exhausted | Authentication and transport | `Unavailable` with `RetryInfo` is retryable and bounded. |
| Backend persistence is unavailable | Validation, outcomes, and partial success | `Unavailable`; no partial batch persists. |
| Independent exporter submits supported signals | Status and scope; Supported traces; Supported logs; Supported metrics | Standard OTLP contracts define the accepted subset without SDK dependencies. |

## Compatibility references

This contract follows the [OTLP/gRPC protocol specification](https://opentelemetry.io/docs/specs/otlp/), including unary export services, partial-success responses, retry guidance, compression support, and message-limit handling. HTTP operation fields follow the [OpenTelemetry HTTP semantic conventions](https://opentelemetry.io/docs/specs/semconv/http/http-spans/).
