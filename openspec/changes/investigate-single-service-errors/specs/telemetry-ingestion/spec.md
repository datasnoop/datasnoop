## Purpose

Define a secure and predictable OTLP ingestion boundary that accepts the telemetry needed for DataSnoop investigations while remaining interoperable with standard exporters.

## ADDED Requirements

### Requirement: Authenticated OTLP/gRPC ingestion
The system SHALL expose an authenticated OTLP/gRPC endpoint for logs, traces and metrics, and SHALL reject requests with missing or invalid credentials without persisting their records.

#### Scenario: Valid exporter sends telemetry
- **WHEN** a standard OTLP exporter sends a supported signal with valid credentials
- **THEN** the system accepts the request using the OTLP response contract and makes accepted records available to processing

#### Scenario: Invalid credential is rejected
- **WHEN** an exporter sends telemetry with missing or invalid credentials
- **THEN** the system rejects the request and persists none of its records

### Requirement: Explicit partial acceptance
The system MUST validate telemetry at record granularity and SHALL report rejected record counts and a useful reason through the OTLP partial-success contract when valid and invalid records share a request.

#### Scenario: Mixed valid and invalid records
- **WHEN** one export request contains both processable records and records that violate documented limits or required semantics
- **THEN** the system persists the processable records and reports the number and reason for rejected records

#### Scenario: Entire request is invalid
- **WHEN** no record in an export request can be processed
- **THEN** the system returns an explicit non-success result and persists no records

### Requirement: Stable normalized correlation context
The system SHALL normalize accepted telemetry without losing service identity, environment, host identity, source timestamp, trace identifier, span identifier, severity or supported attributes needed for investigation.

#### Scenario: Correlated log and operation arrive
- **WHEN** an operation and a log carry the same valid trace and span correlation identifiers
- **THEN** their normalized representations retain those identifiers so they can be retrieved together

#### Scenario: Attributes exceed documented limits
- **WHEN** a record exceeds a documented attribute count, value size or payload limit
- **THEN** the system applies its documented rejection or truncation policy and surfaces that outcome instead of silently losing data

### Requirement: Bounded overload behavior
The ingestion path MUST use bounded concurrency and buffering, SHALL return a retryable outcome when it cannot safely accept more work, and MUST NOT grow memory without a configured bound.

#### Scenario: Capacity is temporarily exhausted
- **WHEN** the accepted-work capacity is exhausted
- **THEN** the system rejects or throttles additional work with a retryable OTLP-compatible outcome while continuing to process already accepted work

#### Scenario: Backend persistence is unavailable
- **WHEN** persistence remains unavailable beyond the configured ingestion timeout
- **THEN** the system stops accepting work that cannot be held within its bounded capacity and exposes the degraded state

### Requirement: Exporter-independent compatibility
The ingestion contract SHALL be verifiable using an official OpenTelemetry exporter that does not depend on a DataSnoop SDK.

#### Scenario: Independent exporter submits supported signals
- **WHEN** a conforming independent exporter sends the documented supported subset
- **THEN** the resulting service, operations, logs and metrics are available with the same semantics as equivalent data sent by an official DataSnoop SDK

### Requirement: Explicit duplicate delivery behavior
The system SHALL tolerate OTLP retransmission without corrupting accepted telemetry, SHALL apply documented signal-specific identity rules, and MUST NOT inflate endpoint summaries when the same identifiable HTTP operation is delivered more than once.

#### Scenario: Identifiable operation is retransmitted
- **WHEN** an exporter resends an HTTP operation with the same service, environment, trace identifier and span identifier after an uncertain acknowledgement
- **THEN** the system retains one logical operation and its endpoint Rate, Errors and Duration summaries count that operation once

#### Scenario: Records have no reliable identity
- **WHEN** two accepted records have equivalent content but the supported signal provides no reliable identity that proves they are the same record
- **THEN** the system treats them as independent records according to its documented signal policy instead of silently collapsing them by content or payload hash
