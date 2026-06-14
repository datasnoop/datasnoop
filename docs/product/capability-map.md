# Capability Map

This map describes the intended product territory. A capability's presence here does not imply implementation or an immediate delivery commitment.

## Instrumentation

- Official SDKs with opinionated onboarding
- Direct OTLP compatibility
- Automatic framework and library instrumentation
- Context propagation and correlation
- Application and host metrics

## Ingestion and processing

- Authenticated reception of logs, metrics, and operations
- Normalization into the DataSnoop domain
- Backpressure, batching, and partial acceptance
- Streaming for live monitoring
- Controlled persistence and retention

## Investigation

- Endpoint rate, error, and duration overview
- Log and event search and filtering
- Correlation among errors, requests, operations, and logs
- Time-aligned CPU, memory, and disk context
- Future evolution toward distributed flows

## Lounge

- Service and health overview
- Occurrence investigation
- Live View
- Queries and filters
- Service, token, and retention configuration
- DataSnoop platform operation and health

## Reaction

- Alert rules
- Trigger history
- Webhooks and integrations

## Ecosystem

- Reference Go SDK
- Adoption-oriented SDKs for broadly used languages
- Direct use through OpenTelemetry exporters
- Optional OpenTelemetry Collector compatibility

## States

Use `exploring`, `planned`, `specified`, `in-progress`, `available`, or `deferred` when tracking capabilities. Do not use phase names in specs.
