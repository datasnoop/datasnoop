## Purpose

Provide an idiomatic Go onboarding path that instruments a service with safe defaults while keeping OpenTelemetry configuration out of the standard user journey.

## ADDED Requirements

### Requirement: Opinionated bootstrap
The Go SDK SHALL let an application configure service identity, endpoint and credential through a single documented bootstrap flow without requiring the user to construct OpenTelemetry providers or exporters.

#### Scenario: Application starts with minimum configuration
- **WHEN** a developer supplies a service name, DataSnoop endpoint and valid credential
- **THEN** the SDK configures the supported telemetry pipeline and reports that it is ready to export

#### Scenario: Required configuration is invalid
- **WHEN** bootstrap receives a missing service name or malformed endpoint
- **THEN** it returns an actionable configuration error before instrumented traffic is processed

### Requirement: HTTP server instrumentation
The Go SDK SHALL instrument supported `net/http` server requests with normalized route, method, status, duration and correlation identifiers while preserving application behavior.

#### Scenario: Successful request is handled
- **WHEN** an instrumented handler completes an HTTP request
- **THEN** the SDK emits an operation containing the normalized route, method, status, duration and correlation context

#### Scenario: Handler returns an error status
- **WHEN** an instrumented handler completes with a server error status
- **THEN** the emitted operation is marked as an error and can contribute to endpoint error analysis

#### Scenario: High-cardinality URL is handled
- **WHEN** a request path contains a resource identifier and a normalized route is available
- **THEN** the SDK records the normalized route for aggregation rather than using the raw path as the endpoint identity

### Requirement: Structured log correlation
The Go SDK SHALL provide an idiomatic structured logging integration that preserves the original log content and attaches active operation correlation when present.

#### Scenario: Log is emitted during a request
- **WHEN** an integrated logger emits a record inside an instrumented request context
- **THEN** the exported log retains its timestamp, severity, message and attributes and includes the active trace and span identifiers

#### Scenario: Log has no active request
- **WHEN** an integrated logger emits a background record without active correlation context
- **THEN** the log is still exported with service and resource identity and no fabricated trace or span identifier

### Requirement: Failure isolation and bounded buffering
The Go SDK MUST NOT propagate backend availability failures into request handling, MUST bound queued telemetry, and SHALL expose counts of telemetry dropped by local limits.

#### Scenario: DataSnoop is unavailable
- **WHEN** export attempts fail or time out
- **THEN** the application continues serving requests while the SDK follows its bounded retry and discard policy

#### Scenario: Local queue is full
- **WHEN** new telemetry arrives after the configured local queue reaches capacity
- **THEN** the SDK applies its documented discard policy and increments an observable dropped-record counter without blocking indefinitely

### Requirement: Graceful shutdown
The Go SDK SHALL provide a bounded shutdown operation that attempts to flush accepted local telemetry and reports whether the flush completed.

#### Scenario: Application shuts down normally
- **WHEN** the application invokes SDK shutdown with sufficient time and the backend is available
- **THEN** accepted queued telemetry is exported before shutdown returns

#### Scenario: Shutdown deadline expires
- **WHEN** the flush cannot complete before the caller's deadline
- **THEN** shutdown returns a timeout result and does not wait beyond that deadline
