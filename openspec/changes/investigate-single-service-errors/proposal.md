## Why

Developers on cost-conscious infrastructure need to discover quickly why an API is failing without assembling an observability stack or learning OpenTelemetry. DataSnoop's first vertical slice must prove this value end to end: locate an error-prone endpoint and open its correlated logs with enough context to distinguish an application failure from host pressure.

## What Changes

- Add authenticated OTLP/gRPC ingestion for the signals required by single-service investigation, with explicit acceptance, limits, and backpressure behavior.
- Normalize resources, HTTP operations, logs, and host metrics into a queryable DataSnoop domain without exposing OTLP terminology in the default experience.
- Create a reference Go SDK that configures instrumentation, integrates `net/http` and `slog`, exports batches, and fails safely when the backend is unavailable.
- Let users identify endpoints with a high incidence of errors, open a failing operation, and inspect its correlated logs.
- Present the minimum service overview, investigation journey, and basic time-aligned CPU, memory, and disk context in the Lounge.
- Provide operational health and bounded retention so the self-hosted installation remains predictable.
- Validate receiver compatibility through an OTLP exporter independent from the DataSnoop SDK.
- Keep distributed tracing, adoption-oriented SDKs for other languages, advanced Boolean filters, multiple alert channels, OTLP/HTTP, and alternative analytical storage outside the current scope; their resumption triggers remain in `docs/roadmap/opportunities.md`.

## Capabilities

### New Capabilities

- `telemetry-ingestion`: Authenticated OTLP/gRPC reception, supported signals, normalization, partial acceptance, limits, and backpressure.
- `go-instrumentation`: Opinionated bootstrap, HTTP and log instrumentation, batched export, and safe failure for the reference Go SDK.
- `request-investigation`: Identification of problematic endpoints and navigation from a failing operation to its correlated logs.
- `host-monitoring`: Collection and presentation of CPU, memory, and disk associated with the service and investigation window.
- `platform-operations`: DataSnoop health, connectivity diagnostics, and predictable telemetry retention.

### Modified Capabilities

None. No main specs have been published yet.

## Impact

- Introduces public OTLP/gRPC and authentication contracts that require versioned compatibility.
- Creates a Go SDK module distributed separately or with an explicit dependency boundary from the core.
- Affects ingestion, normalization, temporal persistence, aggregations, live streaming, and the Lounge interface.
- Adds OpenTelemetry/OTLP ecosystem dependencies and requires conformance, resilience, and modest-hardware load testing.
- Establishes the first durable OpenSpec behavior contracts for the product.
